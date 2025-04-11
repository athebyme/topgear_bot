package telegram

import (
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/repository"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleDriversList обрабатывает вывод списка всех гонщиков
func (b *Bot) handleDriversList(chatID int64) {
	// Получаем всех гонщиков с их статистикой
	drivers, statsMap, err := b.DriverRepo.GetAllWithStats()
	if err != nil {
		log.Printf("Ошибка получения гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонщиков.")
		return
	}

	if len(drivers) == 0 {
		b.sendMessage(chatID, "👨‍🏎️ Пока нет зарегистрированных гонщиков.")
		return
	}

	// Формируем сообщение со списком гонщиков и статистикой
	text := "👨‍🏎️ *Гонщики Top Gear Racing Club*\n\n"

	for _, driver := range drivers {
		stats := statsMap[driver.ID]
		text += fmt.Sprintf("*%s* - %d очков (%d гонок)\n", driver.Name, stats.TotalScore, stats.TotalRaces)
	}

	// Создаем клавиатуру для выбора гонщика
	keyboard := DriversKeyboard(drivers)

	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleSeasonRaces обрабатывает просмотр гонок определенного сезона
// handleSeasonRaces обрабатывает просмотр гонок определенного сезона
func (b *Bot) handleSeasonRaces(chatID int64, seasonID int, userID int64) {
	log.Printf("handleSeasonRaces: запрос гонок сезона ID=%d", seasonID)

	// Получаем информацию о сезоне
	season, err := b.SeasonRepo.GetByID(seasonID)
	if err != nil {
		log.Printf("Ошибка получения сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о сезоне.")
		return
	}

	if season == nil {
		b.sendMessage(chatID, "⚠️ Сезон не найден.")
		return
	}

	// Получаем гонки сезона
	races, err := b.RaceRepo.GetBySeason(seasonID)
	if err != nil {
		log.Printf("Ошибка получения гонок: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонок.")
		return
	}

	log.Printf("Найдено %d гонок для сезона ID=%d", len(races), seasonID)

	// Подсчет гонок по статусу
	var activeCount, upcomingCount, completedCount int
	for _, race := range races {
		// Проверка на nil или пустую строку для безопасности
		if race.State == "" {
			// Если state пустой, предполагаем статус по флагу Completed
			if race.Completed {
				completedCount++
				// Устанавливаем state для дальнейшего использования
				race.State = models.RaceStateCompleted
			} else {
				upcomingCount++
				// Устанавливаем state для дальнейшего использования
				race.State = models.RaceStateNotStarted
			}
			log.Printf("Гонка ID=%d не имеет состояния, установлено по флагу Completed: %v",
				race.ID, race.State)
		} else {
			switch race.State {
			case models.RaceStateInProgress:
				activeCount++
			case models.RaceStateNotStarted:
				upcomingCount++
			case models.RaceStateCompleted:
				completedCount++
			default:
				log.Printf("Неизвестное состояние гонки: %s для ID=%d", race.State, race.ID)
				// Предполагаем, что это предстоящая гонка
				upcomingCount++
				race.State = models.RaceStateNotStarted
			}
		}
	}

	// Формируем сообщение по новому формату
	text := fmt.Sprintf("🏁 *Гонки %s*\n\n", season.Name)

	// Добавляем статистику по гонкам
	text += fmt.Sprintf("*Сводка:* %d активных, %d предстоящих, %d завершенных\n\n",
		activeCount, upcomingCount, completedCount)

	if len(races) == 0 {
		text += "В этом сезоне пока нет запланированных гонок."
	} else {
		text += "Используйте кнопки ниже для выбора гонки. Символы указывают на статус:\n"
		text += "🏎️ - активная гонка\n"
		text += "⏳ - предстоящая гонка\n"
		text += "✅ - завершенная гонка\n"
		text += "Отметка ✅ рядом с названием означает, что вы зарегистрированы на гонку."
	}

	// Просто используем прямое создание клавиатуры без дополнительного слоя абстракции
	var keyboard [][]tgbotapi.InlineKeyboardButton

	if len(races) > 0 {
		// Группируем гонки по статусу
		var activeRaces, upcomingRaces, completedRaces []*models.Race
		for _, race := range races {
			switch race.State {
			case models.RaceStateInProgress:
				activeRaces = append(activeRaces, race)
			case models.RaceStateNotStarted:
				upcomingRaces = append(upcomingRaces, race)
			case models.RaceStateCompleted:
				completedRaces = append(completedRaces, race)
			}
		}

		// Добавляем предстоящие гонки с приоритетом
		if len(upcomingRaces) > 0 {
			for _, race := range upcomingRaces {
				// Проверяем регистрацию пользователя
				var isRegistered bool
				if driver, err := b.DriverRepo.GetByTelegramID(userID); err == nil && driver != nil {
					registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID)
					if err == nil {
						isRegistered = registered
					}
				}

				// Имя кнопки с индикатором регистрации
				var buttonText string
				if isRegistered {
					buttonText = fmt.Sprintf("⏳ %s ✅", race.Name)
				} else {
					buttonText = fmt.Sprintf("⏳ %s", race.Name)
				}

				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						buttonText,
						fmt.Sprintf("race_details:%d", race.ID),
					),
				))
			}
		}

		// Добавляем активные гонки
		if len(activeRaces) > 0 {
			for _, race := range activeRaces {
				var isRegistered bool
				if driver, err := b.DriverRepo.GetByTelegramID(userID); err == nil && driver != nil {
					registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID)
					if err == nil {
						isRegistered = registered
					}
				}

				var buttonText string
				if isRegistered {
					buttonText = fmt.Sprintf("🏎️ %s ✅", race.Name)
				} else {
					buttonText = fmt.Sprintf("🏎️ %s", race.Name)
				}

				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						buttonText,
						fmt.Sprintf("race_details:%d", race.ID),
					),
				))
			}
		}

		// Добавляем завершенные гонки
		if len(completedRaces) > 0 {
			for _, race := range completedRaces {
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						fmt.Sprintf("✅ %s", race.Name),
						fmt.Sprintf("race_results:%d", race.ID),
					),
				))
			}
		}
	}

	// Кнопка создания новой гонки для админов
	if b.IsAdmin(userID) {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новую гонку",
				"new_race",
			),
		))
	}

	// Кнопка возврата
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к сезонам",
			"seasons",
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// startNewSeasonCreation начинает процесс создания нового сезона
func (b *Bot) startNewSeasonCreation(chatID, userID int64) {
	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для создания нового сезона")
		return
	}

	// Устанавливаем состояние для создания нового сезона
	b.StateManager.SetState(userID, "new_season_name", make(map[string]interface{}))

	b.sendMessage(chatID, "🏆 Создание нового сезона\n\nВведите название сезона:")
}

// startAddRaceResult начинает процесс добавления результата для конкретной гонки
func (b *Bot) startAddRaceResult(chatID, userID int64, raceID int) {
	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Проверяем, не добавлял ли уже гонщик результат для этой гонки
	exists, err := b.ResultRepo.CheckDriverResultExists(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки результата: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке результатов.")
		return
	}

	if exists {
		b.sendMessage(chatID, "⚠️ Вы уже добавили результат для этой гонки.")
		return
	}

	// Устанавливаем состояние для добавления результата
	b.StateManager.SetState(userID, "add_result_car_number", map[string]interface{}{
		"race_id": raceID,
	})

	b.sendMessage(chatID, "Введите номер вашей машины:")
}

// getCarPlaceEmoji возвращает эмодзи для места в гонке
func getCarPlaceEmoji(place int) string {
	switch place {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	default:
		return "➖"
	}
}

// parseDate парсит строку даты из формата ДД.ММ.ГГГГ
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02.01.2006", dateStr)
}

// Переименуем обработчики для избежания конфликтов с handlers_car.go
func (b *Bot) handleResultCarNumber(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, что введено число
	carNumber, err := strconv.Atoi(message.Text)
	if err != nil || carNumber < 1 || carNumber > 999 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, введите корректный номер машины (число от 1 до 999).")
		return
	}

	// Сохраняем номер машины и запрашиваем название машины
	b.StateManager.SetState(userID, "add_result_car_name", map[string]interface{}{
		"race_id":    state.ContextData["race_id"],
		"car_number": carNumber,
	})

	b.sendMessage(chatID, "Введите название вашей машины:")
}

// handleResultCarName обрабатывает ввод названия машины при добавлении результата
func (b *Bot) handleResultCarName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность названия
	carName := strings.TrimSpace(message.Text)
	if len(carName) < 2 || len(carName) > 50 {
		b.sendMessage(chatID, "⚠️ Название машины должно содержать от 2 до 50 символов. Пожалуйста, введите корректное название:")
		return
	}

	// Сохраняем название машины и запрашиваем фото
	b.StateManager.SetState(userID, "add_result_car_photo", map[string]interface{}{
		"race_id":    state.ContextData["race_id"],
		"car_number": state.ContextData["car_number"],
		"car_name":   carName,
	})

	b.sendMessage(chatID, "Отправьте фото вашей машины (или '-' чтобы пропустить):")
}

// handleResultCarPhoto обрабатывает отправку фото машины при добавлении результата
func (b *Bot) handleResultCarPhoto(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	var photoURL string

	if message.Text == "-" {
		photoURL = ""
	} else if message.Photo != nil && len(message.Photo) > 0 {
		// Получаем ID фото для сохранения
		photo := message.Photo[len(message.Photo)-1]
		photoURL = photo.FileID
	} else {
		b.sendMessage(chatID, "⚠️ Пожалуйста, отправьте фото или '-' для пропуска.")
		return
	}

	// Получаем гонку для определения дисциплин
	raceID := state.ContextData["race_id"].(int)
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонки.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		b.StateManager.ClearState(userID)
		return
	}

	// Сохраняем данные и переходим к вводу результатов первой дисциплины
	b.StateManager.SetState(userID, "add_result_discipline", map[string]interface{}{
		"race_id":     raceID,
		"car_number":  state.ContextData["car_number"],
		"car_name":    state.ContextData["car_name"],
		"car_photo":   photoURL,
		"disciplines": race.Disciplines,
		"current_idx": 0,
		"results":     make(map[string]int),
	})

	// Запрашиваем результат первой дисциплины
	disciplineName := race.Disciplines[0]
	keyboard := PlacesKeyboard(disciplineName)

	b.sendMessageWithKeyboard(
		chatID,
		fmt.Sprintf("Выберите ваше место в дисциплине '%s':", disciplineName),
		keyboard,
	)
}

// handleRaces обрабатывает запрос списка гонок
func (b *Bot) handleRaces(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	log.Printf("Запрошен список гонок пользователем %d", userID)

	// Получаем активный сезон
	activeSeason, err := b.SeasonRepo.GetActive()
	if err != nil {
		log.Printf("Ошибка получения активного сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении активного сезона.")
		return
	}

	// Выводим информацию о активном сезоне
	if activeSeason != nil {
		log.Printf("Найден активный сезон ID=%d, Name='%s'", activeSeason.ID, activeSeason.Name)
	} else {
		log.Printf("Активный сезон не найден")
	}

	var races []*models.Race
	var seasonName string

	// Получаем все гонки независимо от наличия активного сезона
	log.Printf("Пробуем получить все гонки...")

	// Используем GetAll() вместо условной логики
	races, err = b.RaceRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения всех гонок: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонок.")
		return
	}

	// Выбираем название для заголовка
	if activeSeason != nil {
		seasonName = activeSeason.Name
	} else {
		seasonName = "Все сезоны"
	}

	// Подробно логируем найденные гонки
	log.Printf("Найдено %d гонок для отображения", len(races))
	for i, race := range races {
		log.Printf("Гонка %d: ID=%d, Название='%s', State='%s', SeasonID=%d, Дата=%v",
			i+1, race.ID, race.Name, race.State, race.SeasonID, race.Date)
	}

	// Проверка наличия гонок
	if len(races) == 0 {
		log.Printf("Нет доступных гонок для отображения")

		// Для администраторов показываем кнопку создания гонки
		if b.IsAdmin(userID) {
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"➕ Создать новую гонку",
						"new_race",
					),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"🔙 Главное меню",
						"back_to_main",
					),
				),
			)

			b.sendMessageWithKeyboard(
				chatID,
				"🏁 *Список гонок*\n\nВ настоящее время нет доступных гонок.\n\nВы можете создать новую гонку, нажав кнопку ниже.",
				keyboard,
			)
		} else {
			b.sendMessageWithKeyboard(
				chatID,
				"🏁 *Список гонок*\n\nВ настоящее время нет доступных гонок.",
				tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(
							"🔙 Главное меню",
							"back_to_main",
						),
					),
				),
			)
		}
		return
	}

	// Считаем количество гонок каждого типа
	var activeCount, upcomingCount, completedCount int
	for _, race := range races {
		// Проверка на nil или пустую строку для безопасности
		if race.State == "" {
			// Если state пустой, предполагаем статус по флагу Completed
			if race.Completed {
				completedCount++
				// Устанавливаем state для дальнейшего использования
				race.State = models.RaceStateCompleted
			} else {
				upcomingCount++
				// Устанавливаем state для дальнейшего использования
				race.State = models.RaceStateNotStarted
			}
			log.Printf("Гонка ID=%d не имеет состояния, установлено по флагу Completed: %v",
				race.ID, race.State)
		} else {
			switch race.State {
			case models.RaceStateInProgress:
				activeCount++
			case models.RaceStateNotStarted:
				upcomingCount++
			case models.RaceStateCompleted:
				completedCount++
			default:
				log.Printf("Неизвестное состояние гонки: %s для ID=%d", race.State, race.ID)
				// Предполагаем, что это предстоящая гонка
				upcomingCount++
				race.State = models.RaceStateNotStarted
			}
		}
	}

	// Формируем сообщение
	text := fmt.Sprintf("🏁 *Гонки %s*\n\n", seasonName)

	// Добавляем статистику по гонкам
	text += fmt.Sprintf("*Сводка:* %d активных, %d предстоящих, %d завершенных\n\n",
		activeCount, upcomingCount, completedCount)

	text += "Используйте кнопки ниже для выбора гонки. Символы указывают на статус:\n"
	text += "🏎️ - активная гонка\n"
	text += "⏳ - предстоящая гонка\n"
	text += "✅ - завершенная гонка\n"
	text += "Отметка ✅ рядом с названием означает, что вы зарегистрированы на гонку."

	// Просто используем прямое создание клавиатуры без дополнительного слоя абстракции
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Группируем гонки по статусу
	var activeRaces, upcomingRaces, completedRaces []*models.Race
	for _, race := range races {
		switch race.State {
		case models.RaceStateInProgress:
			activeRaces = append(activeRaces, race)
		case models.RaceStateNotStarted:
			upcomingRaces = append(upcomingRaces, race)
		case models.RaceStateCompleted:
			completedRaces = append(completedRaces, race)
		}
	}

	// Добавляем предстоящие гонки с приоритетом
	if len(upcomingRaces) > 0 {
		for _, race := range upcomingRaces {
			var isRegistered bool
			if driver, err := b.DriverRepo.GetByTelegramID(userID); err == nil && driver != nil {
				registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID)
				if err == nil {
					isRegistered = registered
				}
			}

			// Имя кнопки с индикатором регистрации
			var buttonText string
			if isRegistered {
				buttonText = fmt.Sprintf("⏳ %s ✅", race.Name)
			} else {
				buttonText = fmt.Sprintf("⏳ %s", race.Name)
			}

			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					buttonText,
					fmt.Sprintf("race_details:%d", race.ID),
				),
			))
		}
	}

	// Добавляем активные гонки
	if len(activeRaces) > 0 {

		for _, race := range activeRaces {
			// Проверяем регистрацию пользователя
			var isRegistered bool
			if driver, err := b.DriverRepo.GetByTelegramID(userID); err == nil && driver != nil {
				registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID)
				if err == nil {
					isRegistered = registered
				}
			}

			// Имя кнопки с индикатором регистрации
			var buttonText string
			if isRegistered {
				buttonText = fmt.Sprintf("🏎️ %s ✅", race.Name)
			} else {
				buttonText = fmt.Sprintf("🏎️ %s", race.Name)
			}

			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					buttonText,
					fmt.Sprintf("race_details:%d", race.ID),
				),
			))
		}
	}

	// Добавляем завершенные гонки
	if len(completedRaces) > 0 {
		for _, race := range completedRaces {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✅ %s", race.Name),
					fmt.Sprintf("race_results:%d", race.ID),
				),
			))
		}
	}

	// Добавляем кнопку создания новой гонки для админов
	if b.IsAdmin(userID) {

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новую гонку",
				"new_race",
			),
		))
	}

	// Кнопка возврата в главное меню
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Главное меню",
			"back_to_main",
		),
	))

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	sentMsg, err := b.API.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения со списком гонок: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при отображении списка гонок.")
		return
	}

	log.Printf("Сообщение со списком гонок успешно отправлено, ID: %d", sentMsg.MessageID)
}

// handleAddResult with corrected message
func (b *Bot) handleAddResult(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Get driver data
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Get active race instead of incomplete races
	activeRace, err := b.RaceRepo.GetActiveRace()
	if err != nil {
		log.Printf("Ошибка получения активной гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении активной гонки.")
		return
	}

	if activeRace == nil {
		b.sendMessage(chatID, "⚠️ Нет активной гонки для добавления результатов.")
		return
	}

	// Check if driver is registered for this race
	registered, err := b.RaceRepo.CheckDriverRegistered(activeRace.ID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке регистрации на гонку.")
		return
	}

	if !registered {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы на текущую активную гонку.")
		return
	}

	// Check if driver has confirmed their car
	var carConfirmed bool
	err = b.db.QueryRow(`
		SELECT car_confirmed FROM race_registrations
		WHERE race_id = $1 AND driver_id = $2
	`, activeRace.ID, driver.ID).Scan(&carConfirmed)

	if err != nil {
		log.Printf("Ошибка проверки подтверждения машины: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке подтверждения машины.")
		return
	}

	if !carConfirmed {
		b.sendMessage(chatID, "⚠️ Вы должны сначала подтвердить свою машину для этой гонки. Используйте /mycar чтобы увидеть и подтвердить вашу машину.")
		return
	}

	// Check if result already exists
	exists, err := b.ResultRepo.CheckDriverResultExists(activeRace.ID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки результата: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке результатов.")
		return
	}

	if exists {
		b.sendMessage(chatID, "⚠️ Вы уже добавили результат для этой гонки.")
		return
	}

	// Get car assignment for this driver
	assignment, err := b.CarRepo.GetDriverCarAssignment(activeRace.ID, driver.ID)
	if err != nil {
		log.Printf("Ошибка получения назначения машины: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о вашей машине.")
		return
	}

	if assignment == nil {
		b.sendMessage(chatID, "⚠️ У вас нет назначенной машины для этой гонки.")
		return
	}

	// Set state for adding result, pre-filling car info
	b.StateManager.SetState(userID, "add_result_discipline", map[string]interface{}{
		"race_id":     activeRace.ID,
		"car_number":  assignment.AssignmentNumber,
		"car_name":    assignment.Car.Name + " (" + assignment.Car.Year + ")",
		"car_photo":   assignment.Car.ImageURL,
		"disciplines": activeRace.Disciplines,
		"current_idx": 0,
		"results":     make(map[string]int),
	})

	// Ask for first discipline result
	disciplineName := activeRace.Disciplines[0]
	keyboard := PlacesKeyboard(disciplineName)

	b.sendMessageWithKeyboard(
		chatID,
		fmt.Sprintf("Ввод результатов для гонки '%s'.\n\nВыберите ваше место в дисциплине '%s':",
			activeRace.Name, disciplineName),
		keyboard,
	)
}

// handleResultDiscipline with improved place selection
func (b *Bot) handleResultDiscipline(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Check valid place
	place, err := strconv.Atoi(message.Text)
	if err != nil || place < 0 || place > 3 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, введите число от 0 до 3 (0 - не участвовал, 1-3 - место).")
		return
	}

	// Get state data
	disciplines := state.ContextData["disciplines"].([]string)
	currentIdx := state.ContextData["current_idx"].(int)
	results := state.ContextData["results"].(map[string]int)

	// Save current discipline result
	currentDiscipline := disciplines[currentIdx]
	results[currentDiscipline] = place

	// Move to next discipline or finish
	currentIdx++

	if currentIdx < len(disciplines) {
		// More disciplines to go
		b.StateManager.SetState(userID, "add_result_discipline", map[string]interface{}{
			"race_id":     state.ContextData["race_id"],
			"car_number":  state.ContextData["car_number"],
			"car_name":    state.ContextData["car_name"],
			"car_photo":   state.ContextData["car_photo"],
			"disciplines": disciplines,
			"current_idx": currentIdx,
			"results":     results,
		})

		// Ask for next discipline
		disciplineName := disciplines[currentIdx]
		keyboard := PlacesKeyboard(disciplineName)

		b.sendMessageWithKeyboard(
			chatID,
			fmt.Sprintf("Выберите ваше место в дисциплине '%s':", disciplineName),
			keyboard,
		)
	} else {
		// All disciplines done, save result
		driver, err := b.DriverRepo.GetByTelegramID(userID)
		if err != nil {
			log.Printf("Ошибка получения гонщика: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
			return
		}

		if driver == nil {
			b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
			b.StateManager.ClearState(userID)
			return
		}

		// Calculate total score
		totalScore := 0
		for _, place := range results {
			switch place {
			case 1:
				totalScore += 3
			case 2:
				totalScore += 2
			case 3:
				totalScore += 1
			}
		}

		// Check if driver used reroll for this race
		rerollUsed, err := b.ResultRepo.GetDriverRerollStatus(state.ContextData["race_id"].(int), driver.ID)
		if err != nil {
			log.Printf("Ошибка проверки статуса реролла: %v", err)
			rerollUsed = false // Assume not used if error
		}

		// Apply reroll penalty if used
		rerollPenalty := 0
		if rerollUsed {
			rerollPenalty = 1
			totalScore -= rerollPenalty
		}

		// Create race result
		result := &models.RaceResult{
			RaceID:        state.ContextData["race_id"].(int),
			DriverID:      driver.ID,
			CarNumber:     state.ContextData["car_number"].(int),
			CarName:       state.ContextData["car_name"].(string),
			CarPhotoURL:   state.ContextData["car_photo"].(string),
			Results:       results,
			TotalScore:    totalScore,
			RerollPenalty: rerollPenalty,
		}

		// Save result to DB
		var _ int
		if rerollPenalty > 0 {
			_, err = b.ResultRepo.CreateWithRerollPenalty(result)
		} else {
			_, err = b.ResultRepo.Create(result)
		}

		if err != nil {
			log.Printf("Ошибка сохранения результата: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при сохранении результатов.")
			return
		}

		// Clear state
		b.StateManager.ClearState(userID)

		// Format success message with penalties
		successMsg := fmt.Sprintf("✅ Результаты успешно сохранены!")
		if rerollPenalty > 0 {
			successMsg += fmt.Sprintf("\n\n⚠️ Учтен штраф -%d балл за реролл машины.", rerollPenalty)
		}
		successMsg += fmt.Sprintf("\n\nВы набрали %d очков в этой гонке.", totalScore)
		b.sendMessage(chatID, successMsg)

		// Show race results
		b.showRaceResults(chatID, result.RaceID)
	}
}

func (b *Bot) handleRegister(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	log.Printf("Starting driver registration for user ID: %d", userID)

	exists, _ := b.DriverRepo.CheckExists(userID)

	log.Printf("Driver exists check result: %v", exists)

	if exists {
		b.sendMessage(chatID, "✅ Вы уже зарегистрированы как гонщик. Используйте /driver для просмотра своей карточки.")
		return
	}

	registrationContext := make(map[string]interface{})
	registrationContext["messageIDs"] = []int{}

	log.Printf("Setting user state to register_name")
	b.StateManager.SetState(userID, "register_name", registrationContext)

	msg := b.sendMessage(chatID, "📝 *Регистрация нового гонщика*\n\nВведите ваше гоночное имя (от 2 до 30 символов):")

	b.addMessageIDToState(userID, msg.MessageID)

	b.deleteMessage(chatID, message.MessageID)
}

func (b *Bot) addMessageIDToState(userID int64, messageID int) {
	state, exists := b.StateManager.GetState(userID)
	if !exists {
		return
	}

	messageIDs, ok := state.ContextData["messageIDs"].([]int)
	if !ok {
		messageIDs = []int{}
	}

	messageIDs = append(messageIDs, messageID)
	b.StateManager.SetContextValue(userID, "messageIDs", messageIDs)
}

func (b *Bot) handleJoinRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	upcomingRaces, err := b.RaceRepo.GetUpcomingRaces()
	if err != nil {
		log.Printf("Ошибка получения предстоящих гонок: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка предстоящих гонок.")
		return
	}

	if len(upcomingRaces) == 0 {
		b.sendMessage(chatID, "⚠️ Сейчас нет предстоящих гонок для регистрации.")
		return
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, race := range upcomingRaces {
		registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID)
		if err != nil {
			log.Printf("Ошибка проверки регистрации: %v", err)
			continue
		}

		var buttonText string
		var callbackData string

		if registered {
			buttonText = fmt.Sprintf("✅ %s", race.Name)
			callbackData = fmt.Sprintf("unregister_race:%d", race.ID)
		} else {
			buttonText = race.Name
			callbackData = fmt.Sprintf("register_race:%d", race.ID)
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		))
	}

	b.sendMessageWithKeyboard(
		chatID,
		"🏁 *Регистрация на гонку*\n\nВыберите гонку для регистрации:",
		tgbotapi.NewInlineKeyboardMarkup(keyboard...),
	)
}

func (b *Bot) handleMyCar(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Get driver information
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Get active race
	activeRace, err := b.RaceRepo.GetActiveRace()
	if err != nil {
		log.Printf("Ошибка получения активной гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации об активной гонке.")
		return
	}

	if activeRace == nil {
		b.sendMessage(chatID, "⚠️ Сейчас нет активной гонки.")
		return
	}

	// Check if driver is registered for this race
	registered, err := b.RaceRepo.CheckDriverRegistered(activeRace.ID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке регистрации.")
		return
	}

	if !registered {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы на текущую гонку.")
		return
	}

	// Show car information
	showCarForRace(b, chatID, activeRace.ID, driver.ID)
}

func showCarForRace(b *Bot, chatID int64, raceID int, driverID int) {
	// Get car assignment
	assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, driverID)
	if err != nil {
		log.Printf("Ошибка получения назначения машины: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о вашей машине.")
		return
	}

	if assignment == nil {
		b.sendMessage(chatID, "⚠️ Машина еще не назначена для этой гонки.")
		return
	}

	// Get race info
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil || race == nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке.")
		return
	}

	// Check if driver has confirmed their car
	var confirmed bool
	err = b.db.QueryRow(`
        SELECT car_confirmed FROM race_registrations
        WHERE race_id = $1 AND driver_id = $2
    `, raceID, driverID).Scan(&confirmed)

	if err != nil {
		log.Printf("Ошибка получения статуса подтверждения: %v", err)
		confirmed = false // Default to false if error
	}

	// Check if reroll was already used
	rerollUsed, err := b.ResultRepo.GetDriverRerollStatus(raceID, driverID)
	if err != nil {
		log.Printf("Ошибка проверки статуса реролла: %v", err)
		rerollUsed = false // Default to false if error
	}

	// Format car information
	car := assignment.Car
	text := fmt.Sprintf("🚗 *Ваша машина для гонки '%s'*\n\n", race.Name)
	text += fmt.Sprintf("*%s (%s)*\n", car.Name, car.Year)
	text += fmt.Sprintf("🔢 Номер: %d\n", assignment.AssignmentNumber)
	text += fmt.Sprintf("💰 Цена: %d CR\n", car.Price)
	text += fmt.Sprintf("⭐ Редкость: %s\n\n", car.Rarity)
	text += "*Характеристики:*\n"
	text += fmt.Sprintf("🏁 Скорость: %.1f/10\n", car.Speed)
	text += fmt.Sprintf("🔄 Управление: %.1f/10\n", car.Handling)
	text += fmt.Sprintf("⚡ Ускорение: %.1f/10\n", car.Acceleration)
	text += fmt.Sprintf("🚦 Старт: %.1f/10\n", car.Launch)
	text += fmt.Sprintf("🛑 Торможение: %.1f/10\n\n", car.Braking)
	text += fmt.Sprintf("🏆 Класс: %s %d\n", car.ClassLetter, car.ClassNumber)

	if assignment.IsReroll {
		text += "\n*Машина получена после реролла!*"
	}

	// Create keyboard for confirmation or reroll
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Only show confirmation/reroll buttons if not yet confirmed
	if !confirmed {
		// Add confirm button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Подтвердить выбор машины",
				fmt.Sprintf("confirm_car:%d", raceID),
			),
		))

		// Add reroll button if not used yet
		if !rerollUsed {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🎲 Реролл (-1 балл)",
					fmt.Sprintf("reroll_car:%d", raceID),
				),
			))
		}
	} else {
		// If car is confirmed, show button to view race status
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📊 Статус гонки",
				fmt.Sprintf("race_progress:%d", raceID),
			),
		))

		// Add button to add results if the race is in progress
		if race.State == models.RaceStateInProgress {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"➕ Добавить результат",
					fmt.Sprintf("add_result:%d", raceID),
				),
			))
		}
	}

	// Add back button - важно! Всегда возвращаться к гонке, а не общему списку
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_details:%d", raceID),
		),
	))

	// Send message with keyboard and car image if available
	if car.ImageURL != "" {
		b.sendPhotoWithKeyboard(
			chatID,
			car.ImageURL,
			text,
			tgbotapi.NewInlineKeyboardMarkup(keyboard...),
		)
	} else {
		b.sendMessageWithKeyboard(
			chatID,
			text,
			tgbotapi.NewInlineKeyboardMarkup(keyboard...),
		)
	}
}

// handleLeaveRace with corrected message
func (b *Bot) handleLeaveRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Get upcoming races
	upcomingRaces, err := b.RaceRepo.GetUpcomingRaces()
	if err != nil {
		log.Printf("Ошибка получения предстоящих гонок: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка предстоящих гонок.")
		return
	}

	// Filter races where driver is registered
	var registeredRaces []*models.Race

	for _, race := range upcomingRaces {
		registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID)
		if err != nil {
			log.Printf("Ошибка проверки регистрации: %v", err)
			continue
		}

		if registered {
			registeredRaces = append(registeredRaces, race)
		}
	}

	if len(registeredRaces) == 0 {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы ни на одну предстоящую гонку.")
		return
	}

	// Create keyboard with registered races
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, race := range registeredRaces {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				race.Name,
				fmt.Sprintf("unregister_race:%d", race.ID),
			),
		))
	}

	b.sendMessageWithKeyboard(
		chatID,
		"🏁 *Отмена регистрации на гонку*\n\nВыберите гонку для отмены регистрации:",
		tgbotapi.NewInlineKeyboardMarkup(keyboard...),
	)
}

func (b *Bot) callbackAdminConfirmAllCars(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Parse race ID from callback data
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		return
	}

	// Get all registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении списка участников", true)
		return
	}

	// Confirm all unconfirmed cars
	var confirmedCount int
	for _, reg := range registrations {
		if !reg.CarConfirmed {
			err = b.RaceRepo.UpdateCarConfirmation(raceID, reg.DriverID, true)
			if err != nil {
				log.Printf("Ошибка подтверждения машины для гонщика %d: %v", reg.DriverID, err)
				continue
			}
			confirmedCount++
		}
	}

	// Send confirmation message
	b.answerCallbackQuery(query.ID, fmt.Sprintf("✅ Подтверждено %d машин", confirmedCount), false)
	b.sendMessage(chatID, fmt.Sprintf("✅ Вы подтвердили машины для %d гонщиков", confirmedCount))

	// Refresh admin panel
	b.showAdminRacePanel(chatID, raceID)

	// Delete the original message
	b.deleteMessage(chatID, query.Message.MessageID)
}

func (b *Bot) callbackAdminEditResultsMenu(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Parse race ID from callback data
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		return
	}

	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении информации о гонке", true)
		return
	}

	if race == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Гонка не найдена", true)
		return
	}

	// Get all registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении списка участников", true)
		return
	}

	// Get race results
	results, err := b.ResultRepo.GetRaceResultsWithDriverNames(raceID)
	if err != nil {
		log.Printf("Ошибка получения результатов: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении результатов", true)
		return
	}

	// Create a map of driver IDs to results
	resultsByDriverID := make(map[int]*repository.RaceResultWithDriver)
	for _, result := range results {
		resultsByDriverID[result.DriverID] = result
	}

	// Format message
	text := fmt.Sprintf("✏️ *Редактирование результатов гонки: %s*\n\n", race.Name)
	text += fmt.Sprintf("📅 Дата: %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))

	text += "*Участники и результаты:*\n\n"

	// Create keyboard
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add buttons for each driver - either to edit existing result or add new one
	for _, reg := range registrations {
		result, hasResult := resultsByDriverID[reg.DriverID]

		// Add driver info to text
		if hasResult {
			text += fmt.Sprintf("• *%s* - %d очков ✅\n", reg.DriverName, result.TotalScore)

			// Add button to edit result
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✏️ %s", reg.DriverName),
					fmt.Sprintf("admin_edit_result:%d", result.ID),
				),
			))
		} else {
			text += fmt.Sprintf("• *%s* - нет результата ❌\n", reg.DriverName)

			// Add button to add result
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("➕ %s", reg.DriverName),
					fmt.Sprintf("admin_add_result:%d:%d", raceID, reg.DriverID),
				),
			))
		}
	}

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к админ-панели",
			fmt.Sprintf("admin_race_panel:%d", raceID),
		),
	))

	// Send message with keyboard
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))

	// Delete the original message
	b.deleteMessage(chatID, messageID)
}

func (b *Bot) callbackAdminAddResult(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Parse parameters from callback data (admin_add_result:raceID:driverID)
	parts := strings.Split(query.Data, ":")
	if len(parts) < 3 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		return
	}

	driverID, err := strconv.Atoi(parts[2])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонщика", true)
		return
	}

	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении информации о гонке", true)
		return
	}

	// Get driver information
	var driverName string
	err = b.db.QueryRow("SELECT name FROM drivers WHERE id = $1", driverID).Scan(&driverName)
	if err != nil {
		log.Printf("Ошибка получения имени гонщика: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении данных гонщика", true)
		return
	}

	// Get car assignment
	assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, driverID)
	if err != nil || assignment == nil {
		log.Printf("Ошибка получения назначения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Гонщику не назначена машина", true)
		return
	}

	// Set up state for adding result with pre-filled car info
	b.StateManager.SetState(userID, "admin_add_result_discipline", map[string]interface{}{
		"race_id":     raceID,
		"driver_id":   driverID,
		"driver_name": driverName,
		"car_number":  assignment.AssignmentNumber,
		"car_name":    assignment.Car.Name + " (" + assignment.Car.Year + ")",
		"car_photo":   assignment.Car.ImageURL,
		"disciplines": race.Disciplines,
		"current_idx": 0,
		"results":     make(map[string]int),
	})

	// Format message
	text := fmt.Sprintf("✏️ *Добавление результата для гонщика %s*\n\n", driverName)
	text += fmt.Sprintf("🚗 Машина: %s (№%d)\n\n", assignment.Car.Name, assignment.AssignmentNumber)

	// Check if driver used reroll
	var rerollUsed bool
	err = b.db.QueryRow(`
        SELECT reroll_used FROM race_registrations
        WHERE race_id = $1 AND driver_id = $2
    `, raceID, driverID).Scan(&rerollUsed)

	if err == nil && rerollUsed {
		text += "⚠️ *Был использован реролл* (-1 балл к результату)\n\n"
	}

	// First discipline
	disciplineName := race.Disciplines[0]
	text += fmt.Sprintf("Выберите место в дисциплине '*%s*':", disciplineName)

	// Create place selection keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🥇 1 место",
				fmt.Sprintf("admin_select_place:%d:%d:%s:1", raceID, driverID, disciplineName),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🥈 2 место",
				fmt.Sprintf("admin_select_place:%d:%d:%s:2", raceID, driverID, disciplineName),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🥉 3 место",
				fmt.Sprintf("admin_select_place:%d:%d:%s:3", raceID, driverID, disciplineName),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Не участвовал",
				fmt.Sprintf("admin_select_place:%d:%d:%s:0", raceID, driverID, disciplineName),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Отмена",
				fmt.Sprintf("admin_edit_results_menu:%d", raceID),
			),
		),
	)

	// Send message with keyboard
	b.sendMessageWithKeyboard(chatID, text, keyboard)

	// Delete the original message
	b.deleteMessage(chatID, messageID)
}

func (b *Bot) callbackAdminSelectPlace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Parse parameters from callback data (admin_select_place:raceID:driverID:discipline:place)
	parts := strings.Split(query.Data, ":")
	if len(parts) < 5 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		return
	}

	driverID, err := strconv.Atoi(parts[2])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонщика", true)
		return
	}

	disciplineName := parts[3]

	place, err := strconv.Atoi(parts[4])
	if err != nil || place < 0 || place > 3 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверное значение места", true)
		return
	}

	// Get state
	state, exists := b.StateManager.GetState(userID)
	if !exists || state.State != "admin_add_result_discipline" {
		b.answerCallbackQuery(query.ID, "⚠️ Неверное состояние", true)
		return
	}

	// Update results in state
	results := state.ContextData["results"].(map[string]int)
	results[disciplineName] = place

	// Get race disciplines
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Ошибка получения информации о гонке", true)
		return
	}

	// Get current discipline index
	var currentIdx int
	for i, d := range race.Disciplines {
		if d == disciplineName {
			currentIdx = i
			break
		}
	}

	// Move to next discipline or complete
	currentIdx++

	if currentIdx < len(race.Disciplines) {
		// Update state for next discipline
		b.StateManager.SetState(userID, "admin_add_result_discipline", map[string]interface{}{
			"race_id":     state.ContextData["race_id"],
			"driver_id":   state.ContextData["driver_id"],
			"driver_name": state.ContextData["driver_name"],
			"car_number":  state.ContextData["car_number"],
			"car_name":    state.ContextData["car_name"],
			"car_photo":   state.ContextData["car_photo"],
			"disciplines": race.Disciplines,
			"current_idx": currentIdx,
			"results":     results,
		})

		// Show next discipline selection
		nextDiscipline := race.Disciplines[currentIdx]

		// Update message with progress and next discipline
		text := fmt.Sprintf("✏️ *Добавление результата для гонщика %s*\n\n", state.ContextData["driver_name"])
		text += fmt.Sprintf("🚗 Машина: %s\n\n", state.ContextData["car_name"])

		// Show previous selections
		text += "*Выбранные места:*\n"
		for i := 0; i < currentIdx; i++ {
			disc := race.Disciplines[i]
			placeEmoji := getPlaceEmoji(results[disc])
			placeText := getPlaceText(results[disc])
			text += fmt.Sprintf("• %s: %s %s\n", disc, placeEmoji, placeText)
		}

		text += fmt.Sprintf("\nВыберите место в дисциплине '*%s*':", nextDiscipline)

		// Create keyboard for next discipline
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🥇 1 место",
					fmt.Sprintf("admin_select_place:%d:%d:%s:1", raceID, driverID, nextDiscipline),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"🥈 2 место",
					fmt.Sprintf("admin_select_place:%d:%d:%s:2", raceID, driverID, nextDiscipline),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"🥉 3 место",
					fmt.Sprintf("admin_select_place:%d:%d:%s:3", raceID, driverID, nextDiscipline),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"❌ Не участвовал",
					fmt.Sprintf("admin_select_place:%d:%d:%s:0", raceID, driverID, nextDiscipline),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🔙 Отмена",
					fmt.Sprintf("admin_edit_results_menu:%d", raceID),
				),
			),
		)

		// Update message
		b.editMessageWithKeyboard(chatID, messageID, text, keyboard)
	} else {
		// All disciplines completed, save result
		// Calculate total score
		totalScore := 0
		for _, place := range results {
			switch place {
			case 1:
				totalScore += 3
			case 2:
				totalScore += 2
			case 3:
				totalScore += 1
			}
		}

		// Check if driver used reroll
		rerollUsed, err := b.ResultRepo.GetDriverRerollStatus(raceID, driverID)
		if err != nil {
			log.Printf("Ошибка проверки статуса реролла: %v", err)
			rerollUsed = false // Default to false if error
		}

		// Apply reroll penalty
		rerollPenalty := 0
		if rerollUsed {
			rerollPenalty = 1
			totalScore -= rerollPenalty
		}

		// Get car assignment for photo
		assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, driverID)
		if err != nil || assignment == nil {
			log.Printf("Ошибка получения назначения машины: %v", err)

			// Clear state and show error
			b.StateManager.ClearState(userID)
			b.editMessage(chatID, messageID, "⚠️ Произошла ошибка при получении информации о машине.")
			return
		}

		// Create race result
		result := &models.RaceResult{
			RaceID:        raceID,
			DriverID:      driverID,
			CarNumber:     state.ContextData["car_number"].(int),
			CarName:       state.ContextData["car_name"].(string),
			CarPhotoURL:   state.ContextData["car_photo"].(string),
			Results:       results,
			TotalScore:    totalScore,
			RerollPenalty: rerollPenalty,
		}

		// Save to database
		var resultID int
		if rerollPenalty > 0 {
			resultID, err = b.ResultRepo.CreateWithRerollPenalty(result)
		} else {
			resultID, err = b.ResultRepo.Create(result)
		}

		if err != nil {
			log.Printf("Ошибка сохранения результата: %v", err)

			// Clear state and show error
			b.StateManager.ClearState(userID)
			b.editMessage(chatID, messageID, "⚠️ Произошла ошибка при сохранении результата.")
			return
		}

		// Clear state
		b.StateManager.ClearState(userID)

		// Show success message
		text := fmt.Sprintf("✅ *Результат для гонщика %s успешно добавлен!*\n\n", state.ContextData["driver_name"])
		text += "*Итоговые места:*\n"

		for _, discipline := range race.Disciplines {
			placeEmoji := getPlaceEmoji(results[discipline])
			placeText := getPlaceText(results[discipline])
			text += fmt.Sprintf("• %s: %s %s\n", discipline, placeEmoji, placeText)
		}

		if rerollPenalty > 0 {
			text += fmt.Sprintf("\n⚠️ Штраф за реролл: -%d\n", rerollPenalty)
		}

		text += fmt.Sprintf("\n🏆 Всего очков: %d\n", totalScore)

		// Add buttons to edit result or go back to menu
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✏️ Редактировать этот результат",
					fmt.Sprintf("admin_edit_result:%d", resultID),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🔙 Назад к списку результатов",
					fmt.Sprintf("admin_edit_results_menu:%d", raceID),
				),
			),
		)

		// Update message
		b.editMessageWithKeyboard(chatID, messageID, text, keyboard)
	}
}
