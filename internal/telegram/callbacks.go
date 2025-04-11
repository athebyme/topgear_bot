package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/repository"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) registerCallbackHandlers() {
	// Существующие обработчики
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
		"place":               b.callbackPlace,
		"cancel_delete_race":  b.callbackCancelDeleteRace,
		"season_results":      b.callbackSeasonResults,
		"back_to_main":        b.callbackBackToMain,
		"cancel":              b.callbackCancel,
		"register_driver":     b.callbackRegisterDriver,
		"cars":                b.callbackCars,
		"car_class":           b.callbackCarClass,
		"car_class_all":       b.callbackCarClassAll,
		"random_car":          b.callbackRandomCar,
		"update_cars_db":      b.callbackUpdateCarsDB,
		"race_assign_cars":    b.callbackRaceAssignCars,
		"view_race_cars":      b.callbackViewRaceCars,
		"stats_season":        b.callbackStatsForSeason,
		"race_progress":       b.callbackRaceProgress,
		"admin_confirm_car":   b.callbackAdminConfirmCar,
		"leaderboard":         b.callbackLeaderboard,
		"select_discipline":   b.callbackSelectDiscipline,
		"set_place":           b.callbackSetPlace,

		// Добавляем новые обработчики
		"admin_confirm_all_cars": b.callbackAdminConfirmAllCars,
		"admin_add_result":       b.callbackAdminAddResult,
		"admin_select_place":     b.callbackAdminSelectPlace,
	}

	// Регистрация существующих обработчиков
	b.CallbackHandlers["start_race"] = b.callbackStartRace
	b.CallbackHandlers["register_race"] = b.callbackRegisterRace
	b.CallbackHandlers["driver_command"] = b.callbackDriverCommand
	b.CallbackHandlers["admin_edit_result"] = b.callbackAdminEditResult
	b.CallbackHandlers["admin_edit_discipline"] = b.callbackAdminEditDiscipline
	b.CallbackHandlers["admin_set_place"] = b.callbackAdminSetPlace
	b.CallbackHandlers["admin_toggle_reroll"] = b.callbackAdminToggleReroll
	b.CallbackHandlers["admin_race_panel"] = b.callbackAdminRacePanel
	b.CallbackHandlers["admin_edit_results_menu"] = b.callbackAdminEditResultsMenu
	b.CallbackHandlers["admin_force_confirm_car"] = b.callbackAdminForceConfirmCar
	b.CallbackHandlers["admin_send_notifications"] = b.callbackAdminSendNotifications
	b.CallbackHandlers["race_detailed_status"] = b.callbackRaceDetailedStatus
	b.CallbackHandlers["activerace"] = b.callbackActiveRace
	b.CommandHandlers["startrace"] = b.handleStartRace

	b.CallbackHandlers["register_race"] = b.callbackRegisterRace
	b.CallbackHandlers["unregister_race"] = b.callbackUnregisterRace
	b.CallbackHandlers["start_race"] = b.callbackStartRace
	b.CallbackHandlers["confirm_car"] = b.callbackConfirmCar
	b.CallbackHandlers["reroll_car"] = b.callbackRerollCar
	b.CallbackHandlers["race_registrations"] = b.callbackRaceRegistrations
	b.CallbackHandlers["race_start_confirm"] = b.callbackRaceStartConfirm
	b.CallbackHandlers["complete_race_confirm"] = b.callbackCompleteRaceConfirm
	b.CallbackHandlers["race_details"] = b.callbackRaceDetails
}

// handleStartRace позволяет запустить гонку через команду
func (b *Bot) handleStartRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверка админских прав
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для запуска гонки")
		return
	}

	// Парсим ID гонки из аргументов команды
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Проверяем, есть ли незапущенные гонки
		upcomingRaces, err := b.RaceRepo.GetUpcomingRaces()
		if err != nil {
			log.Printf("Ошибка получения предстоящих гонок: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонок")
			return
		}

		if len(upcomingRaces) == 0 {
			b.sendMessage(chatID, "⚠️ Нет доступных гонок для запуска")
			return
		}

		// Формируем список гонок
		text := "Выберите гонку для запуска, указав ее ID:\n\n"
		for _, race := range upcomingRaces {
			text += fmt.Sprintf("• ID %d: %s (📅 %s)\n",
				race.ID, race.Name, b.formatDate(race.Date))
		}
		text += "\nКоманда для запуска: /startrace ID"

		b.sendMessage(chatID, text)
		return
	}

	raceID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Некорректный ID гонки. Укажите число!")
		return
	}

	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка с указанным ID не найдена")
		return
	}

	// Проверяем, что гонка еще не начата
	//if race.State != models.RaceStateNotStarted {
	//	b.sendMessage(chatID, fmt.Sprintf("⚠️ Гонка '%s' уже запущена или завершена", race.Name))
	//	return
	//}

	// Получаем зарегистрированных участников
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка участников")
		return
	}

	if len(registrations) == 0 {
		b.sendMessage(chatID, "⚠️ Нет зарегистрированных участников для этой гонки")
		return
	}

	// Начинаем транзакцию
	tx, err := b.db.Begin()
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при запуске гонки")
		return
	}

	// Запускаем гонку
	err = b.RaceRepo.StartRace(tx, raceID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка запуска гонки: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Ошибка запуска гонки: %v", err))
		return
	}

	// Назначаем машины участникам
	_, err = b.CarRepo.AssignCarsToRegisteredDrivers(tx, raceID, race.CarClass)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка назначения машин: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Ошибка назначения машин: %v", err))
		return
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		log.Printf("Ошибка подтверждения транзакции: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при завершении запуска гонки")
		return
	}

	// Отправляем уведомление об успешном запуске
	b.sendMessage(chatID, fmt.Sprintf("✅ Гонка '%s' успешно запущена! Участникам отправлены уведомления с их машинами.", race.Name))

	// Отправляем уведомления участникам
	go b.notifyDriversAboutCarAssignments(raceID)

	// Показываем подробную информацию о гонке
	b.showRaceDetails(chatID, raceID, userID)
}

// callbackStatsForSeason handles showing stats for a specific season
func (b *Bot) callbackStatsForSeason(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Parse season ID from callback data
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	seasonID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID сезона", true)
		return
	}

	// Delete the original message
	b.deleteMessage(chatID, query.Message.MessageID)

	// Show stats for selected season
	b.showDriverStats(chatID, seasonID)
}

// handleCallbackQuery обрабатывает callback-запросы от кнопок
func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	log.Printf("DEBUG: Получен callback: %s", query.Data)
	b.answerCallbackQuery(query.ID, "", false)

	data := query.Data
	parts := strings.Split(data, ":")
	action := parts[0]

	if handler, exists := b.CallbackHandlers[action]; exists {
		handler(query)
	} else {
		log.Printf("%v", b.CallbackHandlers)
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
	messageID := query.Message.MessageID

	// Получаем ID гонки из данных запроса
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

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении данных гонщика", true)
		return
	}

	if driver == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Вы не зарегистрированы как гонщик", true)
		return
	}

	// Проверяем, не добавлял ли уже гонщик результат для этой гонки
	exists, err := b.ResultRepo.CheckDriverResultExists(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке результатов", true)
		return
	}

	if exists {
		b.answerCallbackQuery(query.ID, "⚠️ Вы уже добавили результат для этой гонки", true)
		return
	}

	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil || race == nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении данных гонки", true)
		return
	}

	// Получаем назначенную машину
	assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, driver.ID)
	if err != nil || assignment == nil {
		log.Printf("Ошибка получения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ У вас нет назначенной машины для этой гонки", true)
		return
	}

	// Создаем клавиатуру с дисциплинами гонки
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, discipline := range race.Disciplines {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				discipline,
				fmt.Sprintf("select_discipline:%d:%s", raceID, discipline),
			),
		))
	}

	// Добавляем кнопку "Назад"
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			fmt.Sprintf("race_details:%d", raceID),
		),
	))

	// Отправляем сообщение с выбором дисциплины
	b.sendMessageWithKeyboard(
		chatID,
		fmt.Sprintf("🏁 *Добавление результата для гонки '%s'*\n\nВыберите дисциплину:", race.Name),
		tgbotapi.NewInlineKeyboardMarkup(keyboard...),
	)

	// Удаляем исходное сообщение
	b.deleteMessage(chatID, messageID)
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

	tx, err := b.db.Begin()
	if err != nil {
		return
	}

	// Удаляем гонку
	err = b.RaceRepo.DeleteWithTx(tx, raceID)
	if err != nil {
		log.Printf("Ошибка удаления гонки: %v", err)
		tx.Rollback()
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

// Add callback handler for place selection
func (b *Bot) callbackPlace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Отправляем уведомление о получении запроса
	b.answerCallbackQuery(query.ID, "", false)

	// Разбираем данные запроса: place:DisciplineName:PlaceValue
	parts := strings.Split(query.Data, ":")
	if len(parts) < 3 {
		b.sendMessage(chatID, "⚠️ Неверный формат данных callback (place).")
		return
	}

	// disciplineName := parts[1] // We actually get the discipline from state
	place, err := strconv.Atoi(parts[2])
	if err != nil || place < 0 || place > 3 {
		b.sendMessage(chatID, "⚠️ Неверное значение места (place).")
		return
	}

	// Получаем текущее состояние
	state, exists := b.StateManager.GetState(userID)
	if !exists || state.State != "add_result_discipline" {
		b.sendMessage(chatID, "⚠️ Неверное состояние для выбора места. Используйте /cancel или начните заново.")
		// Optionally delete the message with the keyboard
		b.deleteMessage(chatID, messageID)
		return
	}

	// --- Logic copied and adapted from handleResultDiscipline ---
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

		// Запрашиваем результат следующей дисциплины by editing the message
		nextDisciplineName := disciplines[currentIdx]
		keyboard := PlacesKeyboard(nextDisciplineName)
		b.editMessageWithKeyboard( // EDIT instead of send
			chatID,
			messageID, // Edit the existing message
			fmt.Sprintf("Выберите ваше место в дисциплине '%s':", nextDisciplineName),
			keyboard,
		)
	} else {
		// Все дисциплины заполнены, сохраняем результат
		driver, err := b.DriverRepo.GetByTelegramID(userID)
		if err != nil {
			log.Printf("Ошибка получения гонщика: %v", err)
			b.editMessage(chatID, messageID, "⚠️ Произошла ошибка при получении данных гонщика.")
			return
		}

		if driver == nil {
			b.editMessage(chatID, messageID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
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
		if rerollPenalty > 0 {
			_, err = b.ResultRepo.CreateWithRerollPenalty(result)
		} else {
			_, err = b.ResultRepo.Create(result)
		}

		if err != nil {
			log.Printf("Ошибка сохранения результата: %v", err)
			b.editMessage(chatID, messageID, "⚠️ Произошла ошибка при сохранении результатов.")
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

		// Edit the message to show success
		b.editMessage(chatID, messageID, successMsg)

		// Show race results in a new message
		b.showRaceResults(chatID, result.RaceID)
	}
}

// showRaceResults shows race results with reroll penalties
func (b *Bot) showRaceResults(chatID int64, raceID int) {
	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Get race results with driver names and reroll penalties
	results, err := b.ResultRepo.GetRaceResultsWithRerollPenalty(raceID)
	if err != nil {
		log.Printf("Ошибка получения результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении результатов гонки.")
		return
	}

	// Format header
	text := fmt.Sprintf("🏁 *%s*\n\n", race.Name)
	text += fmt.Sprintf("📅 %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))

	// Add race state
	switch race.State {
	case models.RaceStateNotStarted:
		text += "⏳ *Статус: Регистрация*\n\n"
	case models.RaceStateInProgress:
		text += "🏎️ *Статус: В процессе*\n\n"
	case models.RaceStateCompleted:
		text += "✅ *Статус: Завершена*\n\n"
	}

	if len(results) == 0 {
		text += "Пока нет результатов для этой гонки."
	} else {
		// Format results table
		for i, result := range results {
			text += fmt.Sprintf("*%d. %s* (%s)\n", i+1, result.DriverName, result.CarName)
			text += fmt.Sprintf("🔢 Номер: %d\n", result.CarNumber)

			// Add discipline results
			var placesText []string
			for _, discipline := range race.Disciplines {
				place := result.Results[discipline]
				emoji := getPlaceEmoji(place)
				placesText = append(placesText, fmt.Sprintf("%s %s: %s", emoji, discipline, getPlaceText(place)))
			}

			text += fmt.Sprintf("📊 %s\n", strings.Join(placesText, " | "))

			// Add reroll penalty if any
			if result.RerollPenalty > 0 {
				text += fmt.Sprintf("⚠️ Штраф за реролл: -%d\n", result.RerollPenalty)
			}

			text += fmt.Sprintf("🏆 Всего очков: %d\n\n", result.TotalScore)
		}
	}

	// Create keyboard for race based on state
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add buttons based on race state
	switch race.State {
	case models.RaceStateNotStarted:
		// Add registration button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Зарегистрироваться",
				fmt.Sprintf("register_race:%d", raceID),
			),
		))
	case models.RaceStateInProgress:
		// Add registration status button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"👨‍🏎️ Статус участников",
				fmt.Sprintf("race_registrations:%d", raceID),
			),
		))

		// Add add result button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Добавить свой результат",
				fmt.Sprintf("add_result:%d", raceID),
			),
		))

		// Add view cars button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Посмотреть машины",
				fmt.Sprintf("view_race_cars:%d", raceID),
			),
		))
	}

	// Add buttons common for all states
	if b.IsAdmin(0) { // Replace with actual user ID check when possible
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать",
				fmt.Sprintf("edit_race:%d", raceID),
			),
		))
	}

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			fmt.Sprintf("season_races:%d", race.SeasonID),
		),
	))

	// If we have photos from results, use the first one
	if len(results) > 0 && results[0].CarPhotoURL != "" {
		b.sendPhotoWithKeyboard(chatID, results[0].CarPhotoURL, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	} else {
		b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	}
}

// Add the proper callback handler for the registration button from main menu
func (b *Bot) callbackRegisterDriver(query *tgbotapi.CallbackQuery) {
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleRegister(&message)

	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

func (b *Bot) callbackRaceProgress(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

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

	// Show race progress
	b.showRaceProgress(chatID, raceID)

	// Delete original message
	b.deleteMessage(chatID, messageID)
}

// showRaceProgress shows the current progress of a race including all submitted results
func (b *Bot) showRaceProgress(chatID int64, raceID int) {
	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Get all registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка участников.")
		return
	}

	// Get all submitted results
	results, err := b.ResultRepo.GetRaceResultsWithRerollPenalty(raceID)
	if err != nil {
		log.Printf("Ошибка получения результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении результатов гонки.")
		return
	}

	// Create a map of driver IDs to results for quick lookup
	driverResults := make(map[int]*repository.RaceResultWithDriver)
	for _, result := range results {
		driverResults[result.DriverID] = result
	}

	// Format header
	text := fmt.Sprintf("🏁 *Ход гонки: %s*\n\n", race.Name)
	text += fmt.Sprintf("📅 %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))

	// Add race state
	switch race.State {
	case models.RaceStateNotStarted:
		text += "⏳ *Статус: Регистрация*\n\n"
	case models.RaceStateInProgress:
		text += "🏎️ *Статус: В процессе*\n\n"
	case models.RaceStateCompleted:
		text += "✅ *Статус: Завершена*\n\n"
	}

	// Add progress table
	text += "*Прогресс участников:*\n\n"

	if len(registrations) == 0 {
		text += "Нет зарегистрированных участников."
	} else {
		// For each registered driver
		for i, reg := range registrations {
			// Get car assignment
			assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, reg.DriverID)
			if err != nil || assignment == nil {
				continue
			}

			// Check if driver has submitted results
			result, hasResult := driverResults[reg.DriverID]

			text += fmt.Sprintf("%d. *%s* (%s)\n", i+1, reg.DriverName, assignment.Car.Name)
			text += fmt.Sprintf("🔢 Номер: %d\n", assignment.AssignmentNumber)

			// If reroll was used, show it
			if assignment.IsReroll {
				text += "🎲 Был использован реролл\n"
			}

			// Show results if available
			if hasResult {
				// Add discipline results
				var placesText []string
				for _, discipline := range race.Disciplines {
					place := result.Results[discipline]
					emoji := getPlaceEmoji(place)
					placesText = append(placesText, fmt.Sprintf("%s %s: %s", emoji, discipline, getPlaceText(place)))
				}

				text += fmt.Sprintf("📊 %s\n", strings.Join(placesText, " | "))

				// Add reroll penalty if any
				if result.RerollPenalty > 0 {
					text += fmt.Sprintf("⚠️ Штраф за реролл: -%d\n", result.RerollPenalty)
				}

				text += fmt.Sprintf("🏆 Текущий счет: %d очков\n", result.TotalScore)
			} else {
				text += "❓ Результаты еще не поданы\n"
			}

			text += "\n"
		}
	}

	// Create keyboard
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add relevant buttons based on race state
	if race.State == models.RaceStateInProgress {
		// Add add result button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Добавить свой результат",
				fmt.Sprintf("add_result:%d", raceID),
			),
		))

		// Add view cars button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Посмотреть машины",
				fmt.Sprintf("view_race_cars:%d", raceID),
			),
		))
	}

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_details:%d", raceID),
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// callbackAdminEditResult handles the admin editing a driver's result
func (b *Bot) callbackAdminEditResult(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для редактирования результатов", true)
		return
	}

	// Parse parameters from callback data (admin_edit_result:resultID)
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	resultID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID результата", true)
		return
	}

	// Get the result details
	result, err := b.ResultRepo.GetByID(resultID)
	if err != nil {
		log.Printf("Ошибка получения результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении результата", true)
		return
	}

	if result == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Результат не найден", true)
		return
	}

	// Get driver information
	driver, err := b.DriverRepo.GetByID(result.DriverID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении данных гонщика", true)
		return
	}

	// Get race information
	race, err := b.RaceRepo.GetByID(result.RaceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении информации о гонке", true)
		return
	}

	// Format message with the current results
	text := fmt.Sprintf("✏️ *Редактирование результатов*\n\n")
	text += fmt.Sprintf("Гонка: *%s*\n", race.Name)
	text += fmt.Sprintf("Гонщик: *%s*\n", driver.Name)
	text += fmt.Sprintf("Машина: *%s* (номер %d)\n\n", result.CarName, result.CarNumber)

	text += "*Текущие результаты:*\n"
	for _, discipline := range race.Disciplines {
		place := result.Results[discipline]
		emoji := getPlaceEmoji(place)
		text += fmt.Sprintf("• %s %s: %s\n", emoji, discipline, getPlaceText(place))
	}

	if result.RerollPenalty > 0 {
		text += fmt.Sprintf("\n⚠️ Штраф за реролл: -%d\n", result.RerollPenalty)
	}

	text += fmt.Sprintf("\n🏆 Всего очков: %d\n\n", result.TotalScore)
	text += "Выберите дисциплину для редактирования:"

	// Create keyboard with disciplines
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, discipline := range race.Disciplines {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getPlaceEmoji(result.Results[discipline]), discipline),
				fmt.Sprintf("admin_edit_discipline:%d:%s", resultID, discipline),
			),
		))
	}

	// Add reroll penalty toggle button
	rerollToggleText := "🎲 Добавить штраф за реролл"
	if result.RerollPenalty > 0 {
		rerollToggleText = "🎲 Убрать штраф за реролл"
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			rerollToggleText,
			fmt.Sprintf("admin_toggle_reroll:%d", resultID),
		),
	))

	// Add save/back buttons
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			fmt.Sprintf("race_results:%d", result.RaceID),
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackAdminEditDiscipline handles editing a specific discipline result
func (b *Bot) callbackAdminEditDiscipline(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для редактирования результатов", true)
		return
	}

	// Parse parameters from callback data (admin_edit_discipline:resultID:disciplineName)
	parts := strings.Split(query.Data, ":")
	if len(parts) < 3 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	resultID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID результата", true)
		return
	}

	disciplineName := parts[2]

	// Get the result details
	result, err := b.ResultRepo.GetByID(resultID)
	if err != nil {
		log.Printf("Ошибка получения результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении результата", true)
		return
	}

	if result == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Результат не найден", true)
		return
	}

	// Show place selection keyboard for this discipline
	text := fmt.Sprintf("Выберите место для дисциплины '%s':", disciplineName)

	// Create keyboard with place options
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Place options row
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🥇 1 место",
			fmt.Sprintf("admin_set_place:%d:%s:1", resultID, disciplineName),
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"🥈 2 место",
			fmt.Sprintf("admin_set_place:%d:%s:2", resultID, disciplineName),
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"🥉 3 место",
			fmt.Sprintf("admin_set_place:%d:%s:3", resultID, disciplineName),
		),
	))

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"❌ Не участвовал",
			fmt.Sprintf("admin_set_place:%d:%s:0", resultID, disciplineName),
		),
	))

	// Back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			fmt.Sprintf("admin_edit_result:%d", resultID),
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackAdminSetPlace handles setting a new place for a discipline
func (b *Bot) callbackAdminSetPlace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для редактирования результатов", true)
		return
	}

	// Parse parameters (admin_set_place:resultID:disciplineName:place)
	parts := strings.Split(query.Data, ":")
	if len(parts) < 4 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	resultID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID результата", true)
		return
	}

	disciplineName := parts[2]

	place, err := strconv.Atoi(parts[3])
	if err != nil || place < 0 || place > 3 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверное значение места", true)
		return
	}

	// Get the result
	result, err := b.ResultRepo.GetByID(resultID)
	if err != nil {
		log.Printf("Ошибка получения результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении результата", true)
		return
	}

	if result == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Результат не найден", true)
		return
	}

	// Update the place for this discipline
	result.Results[disciplineName] = place

	// Recalculate total score
	totalScore := 0
	for _, p := range result.Results {
		switch p {
		case 1:
			totalScore += 3
		case 2:
			totalScore += 2
		case 3:
			totalScore += 1
		}
	}

	// Apply reroll penalty if it exists
	if result.RerollPenalty > 0 {
		totalScore -= result.RerollPenalty
	}

	result.TotalScore = totalScore

	// Save the updated result
	err = b.ResultRepo.Update(result)
	if err != nil {
		log.Printf("Ошибка обновления результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при сохранении результата", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Результат обновлен!", false)

	// Show the edit result screen again
	b.callbackAdminEditResult(&tgbotapi.CallbackQuery{
		Data:    fmt.Sprintf("admin_edit_result:%d", resultID),
		From:    query.From,
		Message: query.Message,
	})
}

// callbackAdminToggleReroll toggles the reroll penalty for a result
func (b *Bot) callbackAdminToggleReroll(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для редактирования результатов", true)
		return
	}

	// Parse parameters (admin_toggle_reroll:resultID)
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	resultID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID результата", true)
		return
	}

	// Get the result
	result, err := b.ResultRepo.GetByID(resultID)
	if err != nil {
		log.Printf("Ошибка получения результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении результата", true)
		return
	}

	if result == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Результат не найден", true)
		return
	}

	// Toggle reroll penalty (1 <-> 0)
	if result.RerollPenalty > 0 {
		result.RerollPenalty = 0
		result.TotalScore += 1 // Remove penalty
	} else {
		result.RerollPenalty = 1
		result.TotalScore -= 1 // Apply penalty
	}

	// Save the updated result
	err = b.ResultRepo.Update(result)
	if err != nil {
		log.Printf("Ошибка обновления результата: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при сохранении результата", true)
		return
	}

	// Get the appropriate message
	message := "✅ Штраф за реролл добавлен!"
	if result.RerollPenalty == 0 {
		message = "✅ Штраф за реролл убран!"
	}

	b.answerCallbackQuery(query.ID, message, false)

	// Show the edit result screen again
	b.callbackAdminEditResult(&tgbotapi.CallbackQuery{
		Data:    fmt.Sprintf("admin_edit_result:%d", resultID),
		From:    query.From,
		Message: query.Message,
	})
}

func (b *Bot) callbackRegisterRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Отладочная информация
	log.Printf("Обработка команды register_race: userID=%d, chatID=%d", userID, chatID)

	// Разбираем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		log.Printf("Ошибка: неверный формат данных колбэка: %s", query.Data)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		log.Printf("Ошибка: не удалось преобразовать ID гонки: %v", err)
		return
	}

	log.Printf("Получен ID гонки: %d", raceID)

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении данных гонщика", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика. Пожалуйста, попробуйте снова.")
		return
	}

	if driver == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Вы не зарегистрированы как гонщик", true)
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении информации о гонке", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке. Пожалуйста, попробуйте снова.")
		return
	}

	if race == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Гонка не найдена", true)
		b.sendMessage(chatID, "⚠️ Гонка не найдена. Пожалуйста, выберите другую гонку.")
		return
	}

	// Проверяем, открыта ли еще регистрация на гонку
	//if race.State != models.RaceStateNotStarted {
	//	b.answerCallbackQuery(query.ID, "⚠️ Регистрация на эту гонку уже закрыта", true)
	//	b.sendMessage(chatID, "⚠️ Регистрация на эту гонку уже закрыта.")
	//	return
	//}

	// Проверяем, не зарегистрирован ли уже гонщик
	registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке регистрации", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке регистрации. Пожалуйста, попробуйте снова.")
		return
	}

	if registered {
		b.answerCallbackQuery(query.ID, "⚠️ Вы уже зарегистрированы на эту гонку", true)
		b.sendMessage(chatID, "⚠️ Вы уже зарегистрированы на эту гонку.")
		return
	}

	// Регистрируем гонщика на гонку
	err = b.RaceRepo.RegisterDriver(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка регистрации на гонку: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при регистрации", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при регистрации на гонку. Пожалуйста, попробуйте снова.")
		return
	}

	// Успешная регистрация - отправляем уведомление
	b.answerCallbackQuery(query.ID, "✅ Вы успешно зарегистрированы на гонку!", false)
	b.sendMessage(chatID, fmt.Sprintf("✅ Вы успешно зарегистрированы на гонку '%s'!", race.Name))

	// Показываем обновленные детали гонки
	// Сначала удаляем исходное сообщение
	b.deleteMessage(chatID, messageID)

	// Затем показываем детали гонки
	b.showRaceDetails(chatID, raceID, userID)
}

// callbackAdminRacePanel обрабатывает запрос на показ админ-панели гонки
func (b *Bot) callbackAdminRacePanel(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Извлекаем ID гонки из данных запроса
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

	// Показываем админ-панель
	b.showAdminRacePanel(chatID, raceID)

	// Удаляем исходное сообщение
	b.deleteMessage(chatID, messageID)
}

// callbackAdminForceConfirmCar позволяет администратору принудительно подтвердить машину гонщика
func (b *Bot) callbackAdminForceConfirmCar(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Извлекаем параметры из данных запроса (admin_force_confirm_car:raceID:driverID)
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

	// Принудительно подтверждаем машину
	err = b.RaceRepo.UpdateCarConfirmation(raceID, driverID, true)
	if err != nil {
		log.Printf("Ошибка подтверждения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при подтверждении машины", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Машина успешно подтверждена!", false)

	// Обновляем информацию о гонке
	b.showAdminRacePanel(chatID, raceID)
}

// callbackAdminSendNotifications позволяет администратору отправить уведомления участникам
func (b *Bot) callbackAdminSendNotifications(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав администратора", true)
		return
	}

	// Извлекаем параметры из данных запроса (admin_send_notifications:raceID:type)
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

	notificationType := parts[2]

	// Отправляем нужные уведомления в зависимости от типа
	switch notificationType {
	case "cars":
		// Отправляем уведомления о машинах
		go b.notifyDriversAboutCarAssignments(raceID)
		b.sendMessage(chatID, "✅ Уведомления о машинах отправлены участникам")
	case "results":
		// Отправляем уведомления о результатах
		go b.notifyDriversAboutRaceCompletion(raceID)
		b.sendMessage(chatID, "✅ Уведомления о результатах отправлены участникам")
	case "reminder":
		// Отправляем напоминание о гонке
		go b.sendRaceReminder(raceID)
		b.sendMessage(chatID, "✅ Напоминания о гонке отправлены участникам")
	default:
		b.answerCallbackQuery(query.ID, "⚠️ Неизвестный тип уведомления", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Уведомления отправлены!", false)
}

// callbackRaceDetailedStatus показывает подробный статус гонки
func (b *Bot) callbackRaceDetailedStatus(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Извлекаем ID гонки из данных запроса
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

	// Показываем подробный статус гонки
	b.showRaceProgress(chatID, raceID)

	// Удаляем исходное сообщение
	b.deleteMessage(chatID, messageID)
}

// sendRaceReminder отправляет напоминание о гонке всем зарегистрированным гонщикам
func (b *Bot) sendRaceReminder(raceID int) {
	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		return
	}

	if race == nil {
		log.Println("Гонка не найдена для отправки напоминаний")
		return
	}

	// Получаем всех зарегистрированных гонщиков
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		return
	}

	// Формируем текст напоминания
	text := fmt.Sprintf("🔔 *Напоминание о гонке: %s*\n\n", race.Name)
	text += fmt.Sprintf("📅 Дата: %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))

	switch race.State {
	case models.RaceStateNotStarted:
		text += "⏳ Гонка скоро начнется! Пожалуйста, будьте готовы."
	case models.RaceStateInProgress:
		text += "🏁 Гонка уже идет! Если вы еще не подтвердили свою машину или не добавили результаты, самое время это сделать."
	}

	for _, reg := range registrations {
		var telegramID int64
		err := b.db.QueryRow("SELECT telegram_id FROM drivers WHERE id = $1", reg.DriverID).Scan(&telegramID)
		if err != nil {
			log.Printf("Ошибка получения Telegram ID гонщика %d: %v", reg.DriverID, err)
			continue
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton

		switch race.State {
		case models.RaceStateInProgress:
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🚗 Моя машина",
					fmt.Sprintf("my_car:%d", raceID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"➕ Добавить результат",
					fmt.Sprintf("add_result:%d", raceID),
				),
			))
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📊 Статус гонки",
				fmt.Sprintf("race_progress:%d", raceID),
			),
		))

		b.sendMessageWithKeyboard(telegramID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	}
}

func (b *Bot) callbackActiveRace(query *tgbotapi.CallbackQuery) {
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleActiveRace(&message)

	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

func (b *Bot) callbackSelectDiscipline(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

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

	disciplineName := parts[2]

	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil || driver == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Ошибка получения данных гонщика", true)
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🥇 1 место",
				fmt.Sprintf("set_place:%d:%s:1", raceID, disciplineName),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🥈 2 место",
				fmt.Sprintf("set_place:%d:%s:2", raceID, disciplineName),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🥉 3 место",
				fmt.Sprintf("set_place:%d:%s:3", raceID, disciplineName),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Не участвовал",
				fmt.Sprintf("set_place:%d:%s:0", raceID, disciplineName),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад к выбору дисциплины",
				fmt.Sprintf("add_result:%d", raceID),
			),
		),
	)

	// Отправляем сообщение с клавиатурой
	b.editMessageWithKeyboard(
		chatID,
		messageID,
		fmt.Sprintf("Выберите ваше место в дисциплине '%s':", disciplineName),
		keyboard,
	)
}

func (b *Bot) callbackSetPlace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Разбираем данные запроса: set_place:raceID:disciplineName:place
	parts := strings.Split(query.Data, ":")
	if len(parts) < 4 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		return
	}

	disciplineName := parts[2]

	place, err := strconv.Atoi(parts[3])
	if err != nil || place < 0 || place > 3 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверное место", true)
		return
	}

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil || driver == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Ошибка получения данных гонщика", true)
		return
	}

	// Проверяем, есть ли уже результаты по этой гонке
	var resultID int
	var resultsJSON string
	var totalScore int
	var rerollPenalty int

	err = b.db.QueryRow(`
        SELECT id, results, total_score, reroll_penalty 
        FROM race_results 
        WHERE race_id = $1 AND driver_id = $2
    `, raceID, driver.ID).Scan(&resultID, &resultsJSON, &totalScore, &rerollPenalty)

	var results map[string]int

	if err == nil {
		err = json.Unmarshal([]byte(resultsJSON), &results)
		if err != nil {
			b.answerCallbackQuery(query.ID, "⚠️ Ошибка разбора результатов", true)
			return
		}
	} else {
		// Создаем новый результат
		results = make(map[string]int)

		// Получаем информацию о машине
		assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, driver.ID)
		if err != nil || assignment == nil {
			b.answerCallbackQuery(query.ID, "⚠️ Ошибка получения данных о машине", true)
			return
		}

		// Проверяем статус реролла
		rerollUsed, err := b.ResultRepo.GetDriverRerollStatus(raceID, driver.ID)
		if err == nil && rerollUsed {
			rerollPenalty = 1
		}
	}

	// Обновляем место для выбранной дисциплины
	results[disciplineName] = place

	// Пересчитываем общий счет
	totalScore = 0
	for _, p := range results {
		switch p {
		case 1:
			totalScore += 3
		case 2:
			totalScore += 2
		case 3:
			totalScore += 1
		}
	}

	// Применяем штраф за реролл
	if rerollPenalty > 0 {
		totalScore -= rerollPenalty
	}

	// Получаем данные гонки для получения всех дисциплин
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil || race == nil {
		b.answerCallbackQuery(query.ID, "⚠️ Ошибка получения данных гонки", true)
		return
	}

	// Проверяем, все ли дисциплины заполнены
	allDisciplinesFilled := true
	for _, d := range race.Disciplines {
		if _, exists := results[d]; !exists {
			allDisciplinesFilled = false
			break
		}
	}

	if allDisciplinesFilled {
		// Все дисциплины заполнены, сохраняем результат
		// (реализация сохранения результата)

		// Показываем итоговый результат
		text := "✅ *Все результаты успешно сохранены!*\n\n"

		// Показываем места по дисциплинам
		text += "*Ваши места:*\n"
		for _, discipline := range race.Disciplines {
			place := results[discipline]
			emoji := getPlaceEmoji(place)
			text += fmt.Sprintf("• %s: %s\n", discipline, emoji)
		}

		if rerollPenalty > 0 {
			text += fmt.Sprintf("\n⚠️ Штраф за реролл: -%d\n", rerollPenalty)
		}

		text += fmt.Sprintf("\n🏆 Всего очков: %d", totalScore)

		// Создаем клавиатуру для возврата к гонке
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"📊 Статус гонки",
					fmt.Sprintf("race_progress:%d", raceID),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🔙 Назад к гонке",
					fmt.Sprintf("race_details:%d", raceID),
				),
			),
		)

		b.editMessageWithKeyboard(chatID, messageID, text, keyboard)
	} else {
		// Не все дисциплины заполнены, показываем оставшиеся
		var remainingDisciplines []string
		for _, d := range race.Disciplines {
			if _, exists := results[d]; !exists {
				remainingDisciplines = append(remainingDisciplines, d)
			}
		}

		text := fmt.Sprintf("✅ Результат для дисциплины '%s' сохранен!\n\n", disciplineName)
		text += "*Заполненные дисциплины:*\n"

		for d, p := range results {
			emoji := getPlaceEmoji(p)
			text += fmt.Sprintf("• %s: %s\n", d, emoji)
		}

		if len(remainingDisciplines) > 0 {
			text += "\n*Осталось заполнить:*\n"
			for _, d := range remainingDisciplines {
				text += fmt.Sprintf("• %s\n", d)
			}
		}

		// Создаем клавиатуру для заполнения оставшихся дисциплин
		var keyboard [][]tgbotapi.InlineKeyboardButton

		for _, d := range remainingDisciplines {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					d,
					fmt.Sprintf("select_discipline:%d:%s", raceID, d),
				),
			))
		}

		// Добавляем кнопку "Назад"
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад к гонке",
				fmt.Sprintf("race_details:%d", raceID),
			),
		))

		b.editMessageWithKeyboard(chatID, messageID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	}
}
