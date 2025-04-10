package telegram

import (
	"fmt"
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
func (b *Bot) handleSeasonRaces(chatID int64, seasonID int) {
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

	// Формируем сообщение со списком гонок
	text := fmt.Sprintf("🏁 *Гонки %s*\n\n", season.Name)

	if len(races) == 0 {
		text += "В этом сезоне пока нет запланированных гонок."
	} else {
		for _, race := range races {
			var status string
			if race.Completed {
				status = "✅ Завершена"
			} else {
				status = "🕑 Предстоит"
			}

			text += fmt.Sprintf("*%s* (%s)\n", race.Name, status)
			text += fmt.Sprintf("📅 %s\n", b.formatDate(race.Date))
			text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
			text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))
		}
	}

	// Создаем клавиатуру для управления гонками
	isAdmin := false // Этот параметр нужно будет заменить на проверку админов
	keyboard := RacesKeyboard(races, isAdmin)

	b.sendMessageWithKeyboard(chatID, text, keyboard)
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

// handleResultDiscipline обрабатывает ввод результата дисциплины при добавлении результата
func (b *Bot) handleResultDiscipline(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, что введено корректное число
	place, err := strconv.Atoi(message.Text)
	if err != nil || place < 0 || place > 3 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, введите число от 0 до 3 (0 - не участвовал, 1-3 - место).")
		return
	}

	// Получаем текущие данные
	disciplines := state.ContextData["disciplines"].([]string)
	currentIdx := state.ContextData["current_idx"].(int)
	results := state.ContextData["results"].(map[string]int)

	// Сохраняем результат текущей дисциплины
	currentDiscipline := disciplines[currentIdx]
	results[currentDiscipline] = place

	// Переходим к следующей дисциплине или завершаем
	currentIdx++

	if currentIdx < len(disciplines) {
		// Еще есть дисциплины
		b.StateManager.SetState(userID, "add_result_discipline", map[string]interface{}{
			"race_id":     state.ContextData["race_id"],
			"car_number":  state.ContextData["car_number"],
			"car_name":    state.ContextData["car_name"],
			"car_photo":   state.ContextData["car_photo"],
			"disciplines": disciplines,
			"current_idx": currentIdx,
			"results":     results,
		})

		// Запрашиваем результат следующей дисциплины
		disciplineName := disciplines[currentIdx]
		keyboard := PlacesKeyboard(disciplineName)

		b.sendMessageWithKeyboard(
			chatID,
			fmt.Sprintf("Выберите ваше место в дисциплине '%s':", disciplineName),
			keyboard,
		)
	} else {
		// Все дисциплины заполнены, сохраняем результат
		// Получаем ID гонщика
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

		// Вычисляем общий счет
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

		// Создаем результат гонки
		result := &models.RaceResult{
			RaceID:      state.ContextData["race_id"].(int),
			DriverID:    driver.ID,
			CarNumber:   state.ContextData["car_number"].(int),
			CarName:     state.ContextData["car_name"].(string),
			CarPhotoURL: state.ContextData["car_photo"].(string),
			Results:     results,
			TotalScore:  totalScore,
		}

		// Сохраняем результат в БД
		_, err = b.ResultRepo.Create(result)
		if err != nil {
			log.Printf("Ошибка сохранения результата: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при сохранении результатов.")
			return
		}

		// Очищаем состояние
		b.StateManager.ClearState(userID)

		b.sendMessage(
			chatID,
			fmt.Sprintf("✅ Результаты успешно сохранены! Вы набрали %d очков в этой гонке.", totalScore),
		)

		// Показываем результаты гонки
		b.showRaceResults(chatID, result.RaceID)
	}
}
