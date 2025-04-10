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

// registerCallbackHandlers регистрирует обработчики callback-запросов
func (b *Bot) registerCallbackHandlers() {
	// Создаем обработчики для каждого типа callback-запроса
	b.CallbackHandlers = map[string]CallbackHandler{
		"races":               b.callbackRaces,
		"seasons":             b.callbackSeasons,
		"drivers":             b.callbackDrivers,
		"results":             b.callbackResults,
		"season_races":        b.callbackSeasonRaces,
		"race_results":        b.callbackRaceResults,
		"driver_card":         b.callbackDriverCard,
		"edit_driver_name":    b.callbackEditDriverName,
		"edit_driver_desc":    b.callbackEditDriverDescription,
		"edit_driver_photo":   b.callbackEditDriverPhoto,
		"new_race":            b.callbackNewRace,
		"new_season":          b.callbackNewSeason,
		"add_result":          b.callbackAddResult,
		"discipline":          b.callbackDiscipline,
		"disciplines_done":    b.callbackDisciplinesDone,
		"complete_race":       b.callbackCompleteRace,
		"edit_race":           b.callbackEditRace,
		"delete_race":         b.callbackDeleteRace,
		"confirm_delete_race": b.callbackConfirmDeleteRace,
		"cancel_delete_race":  b.callbackCancelDeleteRace,
		"season_results":      b.callbackSeasonResults,
		"back_to_main":        b.callbackBackToMain,
		"cancel":              b.callbackCancel,
		// Новые обработчики для работы с машинами
		"cars":             b.callbackCars,
		"car_class":        b.callbackCarClass,
		"car_class_all":    b.callbackCarClassAll,
		"random_car":       b.callbackRandomCar,
		"update_cars_db":   b.callbackUpdateCarsDB,
		"race_assign_cars": b.callbackRaceAssignCars,
		"view_race_cars":   b.callbackViewRaceCars,
	}
}

// handleCallbackQuery обрабатывает callback-запросы от кнопок
func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// Отправляем уведомление о получении запроса
	b.answerCallbackQuery(query.ID, "", false)

	// Разбираем данные запроса
	data := query.Data
	parts := strings.Split(data, ":")
	action := parts[0]

	// Вызываем соответствующий обработчик
	if handler, exists := b.CallbackHandlers[action]; exists {
		handler(query)
	} else {
		// Если обработчик не найден, отправляем сообщение об ошибке
		b.sendMessage(query.Message.Chat.ID, "⚠️ Неизвестное действие.")
	}
}

// callbackRaces обрабатывает запрос на просмотр гонок
func (b *Bot) callbackRaces(query *tgbotapi.CallbackQuery) {
	// Имитируем команду /races
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleRaces(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackSeasons обрабатывает запрос на просмотр сезонов
func (b *Bot) callbackSeasons(query *tgbotapi.CallbackQuery) {
	// Имитируем команду /seasons
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleSeasons(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackDrivers обрабатывает запрос на просмотр гонщиков
func (b *Bot) callbackDrivers(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

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

	// Сортируем гонщиков по общему счету (можно реализовать более сложную сортировку)

	// Формируем сообщение со списком гонщиков
	text := "👨‍🏎️ *Гонщики Top Gear Racing Club*\n\n"

	for _, driver := range drivers {
		stats := statsMap[driver.ID]
		text += fmt.Sprintf("*%s* - %d очков\n", driver.Name, stats.TotalScore)
	}

	// Создаем клавиатуру для выбора гонщика
	keyboard := DriversKeyboard(drivers)

	b.sendMessageWithKeyboard(chatID, text, keyboard)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackResults обрабатывает запрос на просмотр результатов
func (b *Bot) callbackResults(query *tgbotapi.CallbackQuery) {
	// Имитируем команду /results
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleResults(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackSeasonRaces обрабатывает запрос на просмотр гонок сезона
func (b *Bot) callbackSeasonRaces(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Получаем ID сезона из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	seasonID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID сезона.")
		return
	}

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

	// Создаем клавиатуру с гонками и кнопкой создания новой гонки (для админов)
	keyboard := RacesKeyboard(races, b.IsAdmin(userID))

	b.sendMessageWithKeyboard(chatID, text, keyboard)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackRaceResults обрабатывает запрос на просмотр результатов гонки
func (b *Bot) callbackRaceResults(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Получаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки.")
		return
	}

	// Показываем результаты гонки
	b.showRaceResults(chatID, raceID)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// showRaceResults показывает результаты гонки
func (b *Bot) showRaceResults(chatID int64, raceID int) {
	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Получаем результаты гонки с именами гонщиков
	results, err := b.ResultRepo.GetRaceResultsWithDriverNames(raceID)
	if err != nil {
		log.Printf("Ошибка получения результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении результатов гонки.")
		return
	}

	// Формируем заголовок
	text := fmt.Sprintf("🏁 *%s*\n\n", race.Name)
	text += fmt.Sprintf("📅 %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))

	if len(results) == 0 {
		text += "Пока нет результатов для этой гонки."
	} else {
		// Формируем таблицу результатов
		for i, result := range results {
			text += fmt.Sprintf("*%d. %s* (%s)\n", i+1, result.DriverName, result.CarName)
			text += fmt.Sprintf("🔢 Номер: %d\n", result.CarNumber)

			// Добавляем результаты по дисциплинам
			var placesText []string
			for _, discipline := range race.Disciplines {
				place := result.Results[discipline]
				if place > 0 {
					placesText = append(placesText, fmt.Sprintf("%s: %d место", discipline, place))
				} else {
					placesText = append(placesText, fmt.Sprintf("%s: -", discipline))
				}
			}

			text += fmt.Sprintf("📊 %s\n", strings.Join(placesText, " | "))
			text += fmt.Sprintf("🏆 Всего очков: %d\n\n", result.TotalScore)
		}
	}

	// Создаем клавиатуру для результатов гонки
	isAdmin := false // Заглушка, заменить на проверку админа
	keyboard := RaceResultsKeyboard(raceID, race.Completed, isAdmin)

	// Если есть фотографии машин, отправляем их
	if len(results) > 0 && results[0].CarPhotoURL != "" {
		b.sendPhotoWithKeyboard(chatID, results[0].CarPhotoURL, text, keyboard)
	} else {
		b.sendMessageWithKeyboard(chatID, text, keyboard)
	}
}

// callbackDriverCard обрабатывает запрос на просмотр карточки гонщика
func (b *Bot) callbackDriverCard(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Получаем ID гонщика из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	driverID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонщика.")
		return
	}

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByID(driverID)
	if err != nil {
		log.Printf("Ошибка получения гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден.")
		return
	}

	// Получаем статистику гонщика
	stats, err := b.DriverRepo.GetStats(driverID)
	if err != nil {
		log.Printf("Ошибка получения статистики: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении статистики гонщика.")
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

	// Отправляем карточку гонщика
	// Если пользователь смотрит свою карточку, добавляем кнопки редактирования
	if driver.TelegramID == query.From.ID {
		keyboard := DriverProfileKeyboard()

		if driver.PhotoURL != "" {
			b.sendPhotoWithKeyboard(chatID, driver.PhotoURL, text, keyboard)
		} else {
			b.sendMessageWithKeyboard(chatID, text, keyboard)
		}
	} else {
		if driver.PhotoURL != "" {
			b.sendPhoto(chatID, driver.PhotoURL, text)
		} else {
			b.sendMessage(chatID, text)
		}
	}

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackEditDriverName обрабатывает запрос на изменение имени гонщика
func (b *Bot) callbackEditDriverName(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Устанавливаем состояние для изменения имени
	b.StateManager.SetState(userID, "edit_driver_name", make(map[string]interface{}))

	// Отправляем запрос на ввод нового имени
	msg := b.sendMessage(chatID, "Введите новое имя гонщика:")

	// Сохраняем ID сообщения для удаления после ввода
	b.StateManager.SetContextValue(userID, "message_id", msg.MessageID)
}

// callbackEditDriverDescription обрабатывает запрос на изменение описания гонщика
func (b *Bot) callbackEditDriverDescription(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Устанавливаем состояние для изменения описания
	b.StateManager.SetState(userID, "edit_driver_description", make(map[string]interface{}))

	// Отправляем запрос на ввод нового описания
	msg := b.sendMessage(chatID, "Введите новое описание гонщика:")

	// Сохраняем ID сообщения для удаления после ввода
	b.StateManager.SetContextValue(userID, "message_id", msg.MessageID)
}

// callbackEditDriverPhoto обрабатывает запрос на изменение фото гонщика
func (b *Bot) callbackEditDriverPhoto(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Устанавливаем состояние для изменения фото
	b.StateManager.SetState(userID, "edit_driver_photo", make(map[string]interface{}))

	// Отправляем запрос на отправку нового фото
	msg := b.sendMessage(chatID, "Отправьте новое фото для вашей карточки гонщика:")

	// Сохраняем ID сообщения для удаления после ввода
	b.StateManager.SetContextValue(userID, "message_id", msg.MessageID)
}

// callbackNewRace обрабатывает запрос на создание новой гонки
func (b *Bot) callbackNewRace(query *tgbotapi.CallbackQuery) {
	// Имитируем команду /newrace
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleNewRace(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackNewSeason обрабатывает запрос на создание нового сезона
func (b *Bot) callbackNewSeason(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для создания нового сезона")
		return
	}

	// Устанавливаем состояние для создания нового сезона
	b.StateManager.SetState(userID, "new_season_name", make(map[string]interface{}))

	b.sendMessage(chatID, "🏆 Создание нового сезона\n\nВведите название сезона:")

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackAddResult обрабатывает запрос на добавление результата
func (b *Bot) callbackAddResult(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Получаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки.")
		return
	}

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения гонщика: %v", err)
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

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackDiscipline обрабатывает выбор дисциплины для гонки
func (b *Bot) callbackDiscipline(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Получаем индекс дисциплины из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	disciplineIdx, err := strconv.Atoi(parts[1])
	if err != nil || disciplineIdx < 0 || disciplineIdx >= len(models.DefaultDisciplines) {
		b.sendMessage(chatID, "⚠️ Неверный индекс дисциплины.")
		return
	}

	// Получаем текущее состояние
	state, exists := b.StateManager.GetState(userID)
	if !exists || state.State != "new_race_disciplines" {
		b.sendMessage(chatID, "⚠️ Неверное состояние. Начните создание гонки заново.")
		return
	}

	// Получаем текущий список выбранных дисциплин
	disciplines, ok := state.ContextData["disciplines"].([]string)
	if !ok {
		disciplines = []string{}
	}

	// Добавляем или удаляем дисциплину из списка
	discipline := models.DefaultDisciplines[disciplineIdx]
	found := false

	for i, d := range disciplines {
		if d == discipline {
			// Удаляем дисциплину из списка
			disciplines = append(disciplines[:i], disciplines[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		// Добавляем дисциплину в список
		disciplines = append(disciplines, discipline)
	}

	// Обновляем состояние
	newContextData := map[string]interface{}{
		"season_id":   state.ContextData["season_id"],
		"name":        state.ContextData["name"],
		"date":        state.ContextData["date"],
		"car_class":   state.ContextData["car_class"],
		"disciplines": disciplines,
	}

	b.StateManager.SetState(userID, "new_race_disciplines", newContextData)

	// Обновляем клавиатуру с отметками выбранных дисциплин
	keyboard := DisciplinesKeyboard(disciplines)

	// Обновляем сообщение с новой клавиатурой
	b.editMessageWithKeyboard(chatID, messageID, "Выберите дисциплины для гонки (можно выбрать несколько):", keyboard)
}

// callbackDisciplinesDone обрабатывает завершение выбора дисциплин
func (b *Bot) callbackDisciplinesDone(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Получаем текущее состояние
	state, exists := b.StateManager.GetState(userID)
	if !exists || state.State != "new_race_disciplines" {
		b.sendMessage(chatID, "⚠️ Неверное состояние. Начните создание гонки заново.")
		return
	}

	// Получаем выбранные дисциплины
	disciplines, ok := state.ContextData["disciplines"].([]string)
	if !ok || len(disciplines) == 0 {
		b.sendMessage(chatID, "⚠️ Необходимо выбрать хотя бы одну дисциплину.")
		return
	}

	// Создаем новую гонку
	date, err := time.Parse("2006-01-02", state.ContextData["date"].(string))
	if err != nil {
		log.Printf("Ошибка разбора даты: %v", err)
		b.sendMessage(chatID, "⚠️ Ошибка в формате даты. Начните создание гонки заново.")
		return
	}

	race := &models.Race{
		SeasonID:    state.ContextData["season_id"].(int),
		Name:        state.ContextData["name"].(string),
		Date:        date,
		CarClass:    state.ContextData["car_class"].(string),
		Disciplines: disciplines,
		Completed:   false,
	}

	// Сохраняем гонку в БД
	_, err = b.RaceRepo.Create(race)
	if err != nil {
		log.Printf("Ошибка создания гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при создании гонки.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Новая гонка успешно создана!")

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)

	// Показываем гонки сезона
	b.callbackSeasonRaces(&tgbotapi.CallbackQuery{
		Data: fmt.Sprintf("season_races:%d", race.SeasonID),
		From: query.From,
		Message: &tgbotapi.Message{
			Chat: query.Message.Chat,
		},
	})
}

// callbackCompleteRace обрабатывает завершение гонки
func (b *Bot) callbackCompleteRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для завершения гонки")
		return
	}

	// Получаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки.")
		return
	}

	// Проверяем, есть ли результаты для этой гонки
	count, err := b.ResultRepo.GetResultCountByRaceID(raceID)
	if err != nil {
		log.Printf("Ошибка проверки результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке результатов гонки.")
		return
	}

	if count == 0 {
		b.sendMessage(chatID, "⚠️ Нельзя завершить гонку без результатов. Сначала добавьте результаты участников.")
		return
	}

	// Обновляем статус гонки
	err = b.RaceRepo.UpdateCompleted(raceID, true)
	if err != nil {
		log.Printf("Ошибка завершения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при завершении гонки.")
		return
	}

	b.sendMessage(chatID, "✅ Гонка успешно завершена!")

	// Показываем обновленные результаты гонки
	b.showRaceResults(chatID, raceID)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackEditRace обрабатывает редактирование гонки
func (b *Bot) callbackEditRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для редактирования гонки")
		return
	}

	// Получаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки.")
		return
	}

	// Получаем данные гонки
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонки.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Устанавливаем состояние для редактирования гонки
	b.StateManager.SetState(userID, "edit_race_name", map[string]interface{}{
		"race_id": raceID,
	})

	b.sendMessage(chatID, fmt.Sprintf("🏁 Редактирование гонки\n\nВведите новое название гонки (текущее: %s):", race.Name))

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackDeleteRace обрабатывает удаление гонки
func (b *Bot) callbackDeleteRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для удаления гонки")
		return
	}

	// Получаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки.")
		return
	}

	// Получаем данные гонки
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонки.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Проверяем, есть ли результаты для этой гонки
	count, err := b.ResultRepo.GetResultCountByRaceID(raceID)
	if err != nil {
		log.Printf("Ошибка проверки результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке результатов гонки.")
		return
	}

	// Запрашиваем подтверждение удаления
	text := fmt.Sprintf("Вы действительно хотите удалить гонку *%s*?", race.Name)
	if count > 0 {
		text += fmt.Sprintf("\n\n⚠️ У этой гонки есть %d результатов, которые тоже будут удалены!", count)
	}

	keyboard := ConfirmationKeyboard("delete_race", raceID)

	b.sendMessageWithKeyboard(chatID, text, keyboard)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackConfirmDeleteRace обрабатывает подтверждение удаления гонки
func (b *Bot) callbackConfirmDeleteRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для удаления гонки")
		return
	}

	// Получаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки.")
		return
	}

	// Получаем данные гонки для запоминания сезона
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонки.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Запоминаем ID сезона для возврата к списку гонок сезона
	seasonID := race.SeasonID

	// Удаляем гонку
	err = b.RaceRepo.Delete(raceID)
	if err != nil {
		log.Printf("Ошибка удаления гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при удалении гонки.")
		return
	}

	b.sendMessage(chatID, "✅ Гонка успешно удалена!")

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)

	// Показываем гонки сезона
	b.callbackSeasonRaces(&tgbotapi.CallbackQuery{
		Data: fmt.Sprintf("season_races:%d", seasonID),
		From: query.From,
		Message: &tgbotapi.Message{
			Chat: query.Message.Chat,
		},
	})
}

// callbackCancelDeleteRace обрабатывает отмену удаления гонки
func (b *Bot) callbackCancelDeleteRace(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	b.sendMessage(chatID, "❌ Удаление гонки отменено.")

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackSeasonResults обрабатывает запрос на просмотр результатов сезона
func (b *Bot) callbackSeasonResults(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Получаем ID сезона из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса.")
		return
	}

	seasonID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID сезона.")
		return
	}

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

	if len(races) == 0 {
		b.sendMessage(chatID, fmt.Sprintf("⚠️ В сезоне '%s' пока нет гонок.", season.Name))
		return
	}

	// Формируем сообщение с результатами сезона
	text := fmt.Sprintf("📊 *Результаты сезона '%s'*\n\n", season.Name)

	// Создаем клавиатуру с гонками сезона
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, race := range races {
		var status string
		if race.Completed {
			status = "✅"
		} else {
			status = "🕑"
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", status, race.Name),
				fmt.Sprintf("race_results:%d", race.ID),
			),
		))
	}

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackBackToMain обрабатывает возврат в главное меню
func (b *Bot) callbackBackToMain(query *tgbotapi.CallbackQuery) {
	// Имитируем команду /start
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleStart(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackCancel обрабатывает отмену действия
func (b *Bot) callbackCancel(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Очищаем состояние пользователя
	if b.StateManager.HasState(userID) {
		b.StateManager.ClearState(userID)
		b.sendMessage(chatID, "🚫 Действие отменено.")
	}

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}
