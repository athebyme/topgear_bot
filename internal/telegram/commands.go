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

// registerCommandHandlers registers all command handlers
func (b *Bot) registerCommandHandlers() {
	b.CommandHandlers = map[string]CommandHandler{
		"start":     b.handleStart,
		"register":  b.handleRegister,
		"driver":    b.handleDriver,
		"seasons":   b.handleSeasons,
		"races":     b.handleRaces,
		"newrace":   b.handleNewRace,
		"results":   b.handleResults,
		"help":      b.handleHelp,
		"addresult": b.handleAddResult,
		"cancel":    b.handleCancel,
		"joinrace":  b.handleJoinRace,
		"leaverage": b.handleLeaveRace,
		"mycar":     b.handleMyCar,
	}
}

// handleStart provides main menu and starting point
func (b *Bot) handleStart(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is already registered
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
	}

	// Create keyboard with main commands
	keyboard := MainKeyboard()

	var messageText string
	if driver != nil {
		messageText = fmt.Sprintf("🏎️ *Top Gear Racing Club* 🏎️\n\nПривет, %s! Выбери действие:", driver.Name)
	} else {
		messageText = "🏎️ *Top Gear Racing Club* 🏎️\n\nДобро пожаловать! Для начала зарегистрируйтесь как гонщик с помощью команды /register."
	}

	b.sendMessageWithKeyboard(chatID, messageText, keyboard)
}

// handleDriver with corrected message
func (b *Bot) handleDriver(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Get driver data
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении вашей карточки гонщика.")
		return
	}

	if driver == nil {
		// FIXED: Changed from "/start" to "/register"
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Получаем статистику гонщика
	stats, err := b.DriverRepo.GetStats(driver.ID)
	if err != nil {
		log.Printf("Ошибка получения статистики гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении вашей статистики.")
		return
	}

	// Формируем карточку гонщика
	text := fmt.Sprintf("👨‍🏎️ *Карточка гонщика*\n\n*%s*\n", driver.Name)

	if driver.Description != "" {
		text += fmt.Sprintf("📋 *Описание:* %s\n\n", driver.Description)
	}

	text += fmt.Sprintf("🏆 *Всего очков:* %d\n", stats.TotalScore)
	text += fmt.Sprintf("🏁 *Гонок:* %d\n\n", stats.TotalRaces)

	if len(stats.RecentRaces) > 0 {
		text += "*Последние гонки:*\n"
		for _, race := range stats.RecentRaces {
			text += fmt.Sprintf("• %s: %d очков\n", race.RaceName, race.Score)
		}
	} else {
		text += "*Пока нет завершенных гонок*"
	}

	// Клавиатура для редактирования профиля
	keyboard := DriverProfileKeyboard()

	if driver.PhotoURL != "" {
		// Отправляем фото с подписью
		b.sendPhotoWithKeyboard(chatID, driver.PhotoURL, text, keyboard)
	} else {
		// Отправляем только текст
		b.sendMessageWithKeyboard(chatID, text, keyboard)
	}
}

// handleSeasons обрабатывает команду /seasons
func (b *Bot) handleSeasons(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Получаем все сезоны
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	if len(seasons) == 0 {
		b.sendMessage(chatID, "🏁 Сезоны еще не созданы.")
		return
	}

	// Отображаем список сезонов с кнопками для просмотра деталей
	text := "🏆 *Сезоны Top Gear Racing Club*\n\n"

	for _, season := range seasons {
		var status string
		if season.Active {
			status = "🟢 Активный"
		} else {
			status = "🔴 Завершен"
		}

		dateRange := fmt.Sprintf("%s", b.formatDate(season.StartDate))
		if !season.EndDate.IsZero() {
			dateRange += fmt.Sprintf(" - %s", b.formatDate(season.EndDate))
		} else {
			dateRange += " - н.в."
		}

		text += fmt.Sprintf("*%s* (%s)\n%s\n\n", season.Name, status, dateRange)
	}

	// Создаем клавиатуру с сезонами
	keyboard := SeasonsKeyboard(seasons, b.IsAdmin(userID))

	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleNewRace обрабатывает команду /newrace
func (b *Bot) handleNewRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для создания новых гонок")
		return
	}

	// Получаем активный сезон
	activeSeason, err := b.SeasonRepo.GetActive()
	if err != nil {
		log.Printf("Ошибка получения активного сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении активного сезона.")
		return
	}

	if activeSeason == nil {
		b.sendMessage(chatID, "⚠️ Не найден активный сезон. Создайте сезон перед добавлением гонки.")
		return
	}

	// Создаем состояние для пользователя
	b.StateManager.SetState(userID, "new_race_name", map[string]interface{}{
		"season_id": activeSeason.ID,
	})

	b.sendMessage(chatID, fmt.Sprintf("🏁 Создание новой гонки для *%s*\n\nВведите название гонки:", activeSeason.Name))
}

// handleResults обрабатывает команду /results
func (b *Bot) handleResults(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Получаем все сезоны
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	text := "📊 *Результаты гонок*\n\nВыберите сезон для просмотра результатов:"

	if len(seasons) == 0 {
		text += "\n\nПока нет созданных сезонов."
		b.sendMessage(chatID, text)
		return
	}

	// Создаем клавиатуру с сезонами
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, season := range seasons {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				season.Name,
				fmt.Sprintf("season_results:%d", season.ID),
			),
		))
	}

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// handleHelp provides documentation for all commands
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	text := `🏎️ *Top Gear Racing Club Бот* 🏎️

*Доступные команды:*

/start - Главное меню
/register - Регистрация нового гонщика
/driver - Просмотр карточки гонщика
/seasons - Просмотр сезонов
/races - Просмотр гонок текущего сезона
/results - Просмотр результатов гонок
/help - Эта справка
/cancel - Отмена текущего действия

*Команды для участия в гонках:*
/joinrace - Регистрация на предстоящую гонку
/leaverage - Отмена регистрации на гонку
/mycar - Просмотр назначенной машины для текущей гонки
/addresult - Добавить свой результат в текущей гонке

*Система подсчета очков:*
🥇 1 место - 3 очка
🥈 2 место - 2 очка
🥉 3 место - 1 очко
⚠️ Реролл машины - штраф -1 очко

*Дисциплины:*
• Визуал
• Драг
• Круговая гонка (3 круга)
• Офроад
• Гонка от А к Б
• Ралли (на время)

*Процесс гонки:*
1. Регистрация на гонку через /joinrace
2. После начала гонки всем участникам будут назначены машины
3. Вы можете принять машину или использовать реролл (со штрафом -1 очко)
4. После подтверждения машины вводите результаты по дисциплинам`

	b.sendMessage(chatID, text)
}

// handleCancel обрабатывает команду /cancel
func (b *Bot) handleCancel(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	if b.StateManager.HasState(userID) {
		b.StateManager.ClearState(userID)
		b.sendMessage(chatID, "🚫 Текущее действие отменено.")
	} else {
		b.sendMessage(chatID, "🤔 Нет активных действий для отмены.")
	}
}

// handleStateInput routes input to appropriate handler based on state
func (b *Bot) handleStateInput(message *tgbotapi.Message, state models.UserState) {
	switch state.State {
	case "register_name":
		b.handleRegisterName(message, state)
	case "register_description":
		b.handleRegisterDescription(message, state)
	case "register_photo":
		b.handleRegisterPhoto(message, state)
	case "new_race_name":
		b.handleNewRaceName(message, state)
	case "new_race_date":
		b.handleNewRaceDate(message, state)
	case "new_race_car_class":
		b.handleNewRaceCarClass(message, state)
	case "edit_driver_name":
		b.handleEditDriverName(message, state)
	case "edit_driver_description":
		b.handleEditDriverDescription(message, state)
	case "edit_driver_photo":
		b.handleEditDriverPhoto(message, state)
	case "add_result_car_number":
		b.handleResultCarNumber(message, state)
	case "add_result_car_name":
		b.handleResultCarName(message, state)
	case "add_result_car_photo":
		b.handleResultCarPhoto(message, state)
	case "add_result_discipline":
		b.handleResultDiscipline(message, state)
	case "new_season_name":
		b.handleNewSeasonName(message, state)
	case "new_season_start_date":
		b.handleNewSeasonStartDate(message, state)
	default:
		b.sendMessage(message.Chat.ID, "⚠️ Неизвестное состояние. Используйте /cancel для отмены текущего действия.")
	}
}

// Обработчики состояний пользователя

// handleRegisterDescription обрабатывает ввод описания при регистрации
func (b *Bot) handleRegisterDescription(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	description := message.Text
	if description == "-" {
		description = ""
	}

	// Сохраняем описание в контексте и запрашиваем фото
	b.StateManager.SetState(userID, "register_photo", map[string]interface{}{
		"name":        state.ContextData["name"],
		"description": description,
	})

	b.sendMessage(chatID, "Отлично! Теперь отправьте фото для вашей карточки гонщика (или отправьте '-' чтобы пропустить):")
}

// handleRegisterPhoto обрабатывает отправку фото при регистрации
func (b *Bot) handleRegisterPhoto(message *tgbotapi.Message, state models.UserState) {
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

	// Создаем нового гонщика
	driver := &models.Driver{
		TelegramID:  userID,
		Name:        state.ContextData["name"].(string),
		Description: state.ContextData["description"].(string),
		PhotoURL:    photoURL,
	}

	// Сохраняем в БД
	_, err := b.DriverRepo.Create(driver)
	if err != nil {
		log.Printf("Ошибка сохранения гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при регистрации. Пожалуйста, попробуйте еще раз.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Регистрация успешно завершена! Используйте /driver чтобы увидеть свою карточку гонщика.")

	// Показываем главное меню
	b.handleStart(message)
}

// handleNewRaceName обрабатывает ввод названия гонки
func (b *Bot) handleNewRaceName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность названия
	name := strings.TrimSpace(message.Text)
	if len(name) < 3 || len(name) > 50 {
		b.sendMessage(chatID, "⚠️ Название должно содержать от 3 до 50 символов. Пожалуйста, введите корректное название:")
		return
	}

	// Сохраняем название в контексте и запрашиваем дату
	b.StateManager.SetState(userID, "new_race_date", map[string]interface{}{
		"season_id": state.ContextData["season_id"],
		"name":      name,
	})

	b.sendMessage(chatID, "Введите дату гонки в формате ДД.ММ.ГГГГ:")
}

// handleNewRaceDate обрабатывает ввод даты гонки
func (b *Bot) handleNewRaceDate(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем формат даты
	dateStr := message.Text
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный формат даты. Используйте формат ДД.ММ.ГГГГ (например, 15.04.2025):")
		return
	}

	// Сохраняем дату в контексте и запрашиваем класс автомобилей
	b.StateManager.SetState(userID, "new_race_car_class", map[string]interface{}{
		"season_id": state.ContextData["season_id"],
		"name":      state.ContextData["name"],
		"date":      date.Format("2006-01-02"),
	})

	b.sendMessage(chatID, "Введите класс автомобилей для гонки:")
}

// handleNewRaceCarClass обрабатывает ввод класса автомобилей
func (b *Bot) handleNewRaceCarClass(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность класса
	carClass := strings.TrimSpace(message.Text)
	if len(carClass) < 1 || len(carClass) > 30 {
		b.sendMessage(chatID, "⚠️ Класс автомобилей должен содержать от 1 до 30 символов. Пожалуйста, введите корректный класс:")
		return
	}

	// Сохраняем класс в контексте и предлагаем выбрать дисциплины
	b.StateManager.SetState(userID, "new_race_disciplines", map[string]interface{}{
		"season_id":   state.ContextData["season_id"],
		"name":        state.ContextData["name"],
		"date":        state.ContextData["date"],
		"car_class":   carClass,
		"disciplines": []string{}, // Пустой массив для выбранных дисциплин
	})

	// Создаем клавиатуру для выбора дисциплин
	keyboard := DisciplinesKeyboard([]string{})

	b.sendMessageWithKeyboard(chatID, "Выберите дисциплины для гонки (можно выбрать несколько):", keyboard)
}

// handleEditDriverName обрабатывает изменение имени гонщика
func (b *Bot) handleEditDriverName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность имени
	name := strings.TrimSpace(message.Text)
	if len(name) < 2 || len(name) > 30 {
		b.sendMessage(chatID, "⚠️ Имя должно содержать от 2 до 30 символов. Пожалуйста, введите корректное имя:")
		return
	}

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
		b.StateManager.ClearState(userID)
		return
	}

	// Обновляем имя в БД
	err = b.DriverRepo.UpdateName(driver.ID, name)
	if err != nil {
		log.Printf("Ошибка обновления имени гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении имени.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Имя успешно обновлено!")

	// Перезагружаем карточку гонщика
	b.handleDriver(message)
}

// handleEditDriverDescription обрабатывает изменение описания гонщика
func (b *Bot) handleEditDriverDescription(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	description := message.Text

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
		b.StateManager.ClearState(userID)
		return
	}

	// Обновляем описание в БД
	err = b.DriverRepo.UpdateDescription(driver.ID, description)
	if err != nil {
		log.Printf("Ошибка обновления описания гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении описания.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Описание успешно обновлено!")

	// Перезагружаем карточку гонщика
	b.handleDriver(message)
}

// handleEditDriverPhoto обрабатывает изменение фото гонщика
func (b *Bot) handleEditDriverPhoto(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, что отправлено фото
	if message.Photo == nil || len(message.Photo) == 0 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, отправьте фото.")
		return
	}

	// Получаем ID фото
	photo := message.Photo[len(message.Photo)-1]
	photoURL := photo.FileID

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
		b.StateManager.ClearState(userID)
		return
	}

	// Обновляем фото в БД
	err = b.DriverRepo.UpdatePhoto(driver.ID, photoURL)
	if err != nil {
		log.Printf("Ошибка обновления фото гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении фото.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Фото успешно обновлено!")

	// Перезагружаем карточку гонщика
	b.handleDriver(message)
}

// handleAddResultCarName обрабатывает ввод названия машины
func (b *Bot) handleAddResultCarName(message *tgbotapi.Message, state models.UserState) {
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

// handleAddResultCarPhoto обрабатывает отправку фото машины
func (b *Bot) handleAddResultCarPhoto(message *tgbotapi.Message, state models.UserState) {
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
	b.sendMessage(
		chatID,
		fmt.Sprintf("Введите ваше место в дисциплине '%s' (1-3 или 0 если не участвовали):",
			race.Disciplines[0]),
	)
}

// handleAddResultDiscipline обрабатывает ввод результата дисциплины
func (b *Bot) handleAddResultDiscipline(message *tgbotapi.Message, state models.UserState) {
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
		b.sendMessage(
			chatID,
			fmt.Sprintf("Введите ваше место в дисциплине '%s' (1-3 или 0 если не участвовали):",
				disciplines[currentIdx]),
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

// handleNewSeasonName обрабатывает ввод названия сезона
func (b *Bot) handleNewSeasonName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность названия
	name := strings.TrimSpace(message.Text)
	if len(name) < 3 || len(name) > 30 {
		b.sendMessage(chatID, "⚠️ Название должно содержать от 3 до 30 символов. Пожалуйста, введите корректное название:")
		return
	}

	// Сохраняем название в контексте и запрашиваем дату начала
	b.StateManager.SetState(userID, "new_season_start_date", map[string]interface{}{
		"name": name,
	})

	b.sendMessage(chatID, "Введите дату начала сезона в формате ДД.ММ.ГГГГ:")
}

// handleNewSeasonStartDate обрабатывает ввод даты начала сезона
func (b *Bot) handleNewSeasonStartDate(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем формат даты
	dateStr := message.Text
	startDate, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный формат даты. Используйте формат ДД.ММ.ГГГГ (например, 15.04.2025):")
		return
	}

	// Создаем новый сезон
	season := &models.Season{
		Name:      state.ContextData["name"].(string),
		StartDate: startDate,
		Active:    true, // Новый сезон сразу активен
	}

	// Сохраняем сезон в БД
	_, err = b.SeasonRepo.Create(season)
	if err != nil {
		log.Printf("Ошибка создания сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при создании нового сезона.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Новый сезон успешно создан и активирован!")

	// Показываем список сезонов
	b.handleSeasons(message)
}

// handleLeaveRace with corrected message
func (b *Bot) handleLeaveRace(message *tgbotapi.Message) {
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
		// FIXED: Changed from "/start" to "/register"
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

// Updated handleRegisterName to properly handle driver registration
func (b *Bot) handleRegisterName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	log.Printf("Processing driver name for user ID: %d", userID)

	// Check name validity
	name := strings.TrimSpace(message.Text)
	log.Printf("Provided name: '%s', length: %d", name, len(name))

	if len(name) < 2 || len(name) > 30 {
		b.sendMessage(chatID, "⚠️ Имя должно содержать от 2 до 30 символов. Пожалуйста, введите корректное имя:")
		return
	}

	// Save name in context and request description
	log.Printf("Setting state to register_description with name: %s", name)
	b.StateManager.SetState(userID, "register_description", map[string]interface{}{
		"name": name,
	})

	b.sendMessage(chatID, fmt.Sprintf("Отлично, %s! Теперь введите краткое описание о себе как о гонщике (или отправьте '-' чтобы пропустить):", name))
}
