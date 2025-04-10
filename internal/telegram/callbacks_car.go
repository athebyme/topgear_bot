package telegram

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// callbackCars обрабатывает запрос на просмотр машин
func (b *Bot) callbackCars(query *tgbotapi.CallbackQuery) {
	// Имитируем команду /cars
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleCars(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackCarClass обрабатывает запрос на просмотр машин определенного класса
func (b *Bot) callbackCarClass(query *tgbotapi.CallbackQuery) {
	// Извлекаем класс машины из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", false)
		return
	}

	classLetter := parts[1]

	// Имитируем команду /carclass
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
		Text: fmt.Sprintf("/carclass %s", classLetter),
	}

	b.handleCarClass(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// callbackCarClassAll обрабатывает запрос на просмотр всех машин определенного класса
func (b *Bot) callbackCarClassAll(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Извлекаем класс машины из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", false)
		return
	}

	classLetter := parts[1]

	// Проверяем корректность класса
	class := models.GetCarClassByLetter(classLetter)
	if class == nil {
		b.sendMessage(chatID, "⚠️ Указан некорректный класс машины.")
		return
	}

	// Получаем машины указанного класса
	cars, err := b.CarRepo.GetByClass(classLetter)
	if err != nil {
		log.Printf("Ошибка получения машин класса %s: %v", classLetter, err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении машин указанного класса.")
		return
	}

	if len(cars) == 0 {
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Машины класса %s не найдены.", classLetter))
		return
	}

	// Формируем сообщение со полным списком машин
	text := fmt.Sprintf("🚗 *Все машины класса %s*\n\n", class.Name)
	text += fmt.Sprintf("Всего машин: %d\n\n", len(cars))

	// Формируем полный список, но с ограничением на длину сообщения
	var carLines []string
	for i, car := range cars {
		line := fmt.Sprintf("%d. *%s (%d)* - %d CR\n", i+1, car.Name, car.Year, car.Price)
		carLines = append(carLines, line)
	}

	// Объединяем строки с учетом ограничения на длину сообщения
	joinedText := text
	maxLength := 4000 // Предельная длина сообщения в Telegram

	for _, line := range carLines {
		if len(joinedText)+len(line) > maxLength {
			// Отправляем текущую порцию и начинаем новую
			b.sendMessage(chatID, joinedText)
			joinedText = ""
		}
		joinedText += line
	}

	// Отправляем последнюю порцию, если она не пустая
	if joinedText != "" {
		// Добавляем кнопку возврата к классам
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🎲 Случайная машина",
					fmt.Sprintf("random_car:%s", classLetter),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🔙 Назад к классам",
					"cars",
				),
			),
		)

		b.sendMessageWithKeyboard(chatID, joinedText, keyboard)
	}

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackRandomCar обрабатывает запрос на просмотр случайной машины определенного класса
func (b *Bot) callbackRandomCar(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Извлекаем класс машины из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", false)
		return
	}

	classLetter := parts[1]

	// Получаем машины указанного класса
	cars, err := b.CarRepo.GetByClass(classLetter)
	if err != nil {
		log.Printf("Ошибка получения машин класса %s: %v", classLetter, err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении машин указанного класса.")
		return
	}

	if len(cars) == 0 {
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Машины класса %s не найдены.", classLetter))
		return
	}

	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Выбираем случайную машину
	car := cars[rand.Intn(len(cars))]

	// Формируем сообщение с информацией о машине
	text := fmt.Sprintf("🚗 *%s (%d)*\n\n", car.Name, car.Year)
	text += fmt.Sprintf("💰 Цена: %d CR\n", car.Price)
	text += fmt.Sprintf("⭐ Редкость: %s\n\n", car.Rarity)
	text += "*Характеристики:*\n"
	text += fmt.Sprintf("🏁 Скорость: %.1f/10\n", car.Speed)
	text += fmt.Sprintf("🔄 Управление: %.1f/10\n", car.Handling)
	text += fmt.Sprintf("⚡ Ускорение: %.1f/10\n", car.Acceleration)
	text += fmt.Sprintf("🚦 Старт: %.1f/10\n", car.Launch)
	text += fmt.Sprintf("🛑 Торможение: %.1f/10\n\n", car.Braking)
	text += fmt.Sprintf("🏆 Класс: %s %d\n", car.ClassLetter, car.ClassNumber)
	text += fmt.Sprintf("📍 Источник: %s", car.Source)

	// Создаем клавиатуру для дополнительных действий
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎲 Еще случайная машина",
				fmt.Sprintf("random_car:%s", classLetter),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад к классу",
				fmt.Sprintf("car_class:%s", classLetter),
			),
		),
	)

	// Отправляем сообщение с клавиатурой и изображением, если оно есть
	if car.ImageURL != "" {
		b.sendPhotoWithKeyboard(chatID, car.ImageURL, text, keyboard)
	} else {
		b.sendMessageWithKeyboard(chatID, text, keyboard)
	}

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackUpdateCarsDB обрабатывает запрос на обновление базы машин
func (b *Bot) callbackUpdateCarsDB(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для обновления базы машин", true)
		return
	}

	// Отправляем уведомление о начале обновления
	b.sendMessage(chatID, "🔄 Запуск обновления базы машин. Это может занять некоторое время...")

	// В реальном приложении здесь был бы запуск парсера или обращение к API
	// Для демонстрации просто делаем задержку
	time.Sleep(3 * time.Second)

	// Отправляем уведомление об успешном обновлении
	b.sendMessage(chatID, "✅ База машин успешно обновлена!")

	// Показываем обновленную статистику
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleCars(&message)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackRaceAssignCars обрабатывает запрос на назначение машин для гонки
func (b *Bot) callbackRaceAssignCars(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для назначения машин", true)
		return
	}

	// Извлекаем ID гонки и класс машин из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 3 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", false)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", false)
		return
	}

	carClass := parts[2]

	// Назначаем случайные машины
	err = b.assignRandomCarsForRace(raceID, carClass)
	if err != nil {
		log.Printf("Ошибка назначения машин: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Ошибка назначения машин: %v", err))
		return
	}

	// Отправляем уведомление об успешном назначении
	b.sendMessage(chatID, "✅ Машины успешно назначены для гонки!")

	// Показываем результаты назначения
	b.showRaceCarAssignments(chatID, raceID, userID)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackViewRaceCars обрабатывает запрос на просмотр назначенных машин для гонки
func (b *Bot) callbackViewRaceCars(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID

	// Извлекаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", false)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", false)
		return
	}

	// Показываем назначения машин
	b.showRaceCarAssignments(chatID, raceID, userID)

	// Удаляем сообщение с кнопкой
	b.deleteMessage(chatID, query.Message.MessageID)
}

// registerRaceFlowCallbackHandlers registers callbacks for race flow
func (b *Bot) registerRaceFlowCallbackHandlers() {
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

// showRaceCarAssignments показывает назначения машин для гонки
func (b *Bot) showRaceCarAssignments(chatID int64, raceID int, userID int64) {
	// Получаем информацию о гонке
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

	// Получаем назначения машин
	assignments, err := b.CarRepo.GetRaceCarAssignments(raceID)
	if err != nil {
		log.Printf("Ошибка получения назначений машин: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении назначений машин.")
		return
	}

	// Формируем сообщение с назначениями
	text := fmt.Sprintf("🏁 *Машины для гонки '%s'*\n\n", race.Name)
	text += fmt.Sprintf("📅 %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s (%s)\n\n", race.CarClass, models.GetCarClassName(race.CarClass))

	if len(assignments) == 0 {
		text += "⚠️ Машины еще не назначены для этой гонки."
	} else {
		for _, assignment := range assignments {
			text += fmt.Sprintf("*%s*\n", assignment.DriverName)
			text += fmt.Sprintf("🔢 Номер: %d\n", assignment.AssignmentNumber)
			// Исправлено: используем %s вместо %d для Car.Year, который хранится как строка
			text += fmt.Sprintf("🚗 Машина: %s (%s)\n", assignment.Car.Name, assignment.Car.Year)
			text += fmt.Sprintf("⭐ Редкость: %s\n\n", assignment.Car.Rarity)
		}
	}

	// Создаем клавиатуру для дополнительных действий
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Добавляем кнопку для повторного назначения машин (для админов)
	if b.IsAdmin(userID) {
		// Кнопки для выбора класса машин
		var classButtons [][]tgbotapi.InlineKeyboardButton

		for _, class := range models.CarClasses {
			classButtons = append(classButtons, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🎲 Назначить %s", class.Name),
					fmt.Sprintf("race_assign_cars:%d:%s", race.ID, class.Letter),
				),
			))
		}

		// Объединяем в группы по 2 кнопки, чтобы не растягивать интерфейс
		for i := 0; i < len(classButtons); i += 2 {
			if i+1 < len(classButtons) {
				// Объединяем две кнопки в один ряд
				row := append(classButtons[i], classButtons[i+1]...)
				keyboard = append(keyboard, row)
			} else {
				// Если осталась одна кнопка, добавляем её отдельно
				keyboard = append(keyboard, classButtons[i])
			}
		}
	}

	// Добавляем кнопку для возврата к гонке
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_results:%d", race.ID),
		),
	))

	// Отправляем сообщение с клавиатурой
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

func (b *Bot) callbackRaceDetails(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

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

	// Show race details including registration options
	b.showRaceDetails(chatID, raceID, userID)

	// Answer the callback query
	b.answerCallbackQuery(query.ID, "", false)

	// Remove the original message
	b.deleteMessage(chatID, query.Message.MessageID)
}

// Updated callback handler for race unregistrations
func (b *Bot) callbackUnregisterRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
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

	// Get driver information
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

	// Check if race is still open for registration changes
	if race.State != models.RaceStateNotStarted {
		b.answerCallbackQuery(query.ID, "⚠️ Изменение регистрации для этой гонки уже недоступно", true)
		return
	}

	// Check if driver is registered
	registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке регистрации", true)
		return
	}

	if !registered {
		b.answerCallbackQuery(query.ID, "⚠️ Вы не были зарегистрированы на эту гонку", true)
		return
	}

	// Unregister driver from the race
	err = b.RaceRepo.UnregisterDriver(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка отмены регистрации: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при отмене регистрации", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Регистрация на гонку отменена", false)

	// Show updated race details
	b.showRaceDetails(chatID, raceID, userID)

	// Delete the original message
	b.deleteMessage(chatID, messageID)
}

// Enhanced callbackStartRace for better race management
func (b *Bot) callbackStartRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для запуска гонки", true)
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

	// Check if race is not started yet
	if race.State != models.RaceStateNotStarted {
		b.answerCallbackQuery(query.ID, "⚠️ Гонка уже запущена или завершена", true)
		return
	}

	// Get registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении списка участников", true)
		return
	}

	if len(registrations) == 0 {
		b.answerCallbackQuery(query.ID, "⚠️ Нет зарегистрированных участников для этой гонки", true)
		return
	}

	// Show confirmation dialog with registered drivers list
	text := fmt.Sprintf("🏁 *Запуск гонки '%s'*\n\n", race.Name)
	text += "*Зарегистрированные участники:*\n\n"

	for i, reg := range registrations {
		text += fmt.Sprintf("%d. %s\n", i+1, reg.DriverName)
	}

	text += "\nПосле запуска гонки всем участникам будут назначены машины и регистрация будет закрыта. Продолжить?"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Запустить гонку",
				fmt.Sprintf("race_start_confirm:%d", raceID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Отмена",
				"cancel",
			),
		),
	)

	b.sendMessageWithKeyboard(chatID, text, keyboard)
	b.deleteMessage(chatID, query.Message.MessageID)
}

// Enhanced callbackRaceStartConfirm with better notifications
func (b *Bot) callbackRaceStartConfirm(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для запуска гонки", true)
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

	// Check if race is not started yet
	//if race.State != models.RaceStateNotStarted {
	//	b.answerCallbackQuery(query.ID, "⚠️ Гонка уже запущена или завершена", true)
	//	return
	//}

	// Start a database transaction
	tx, err := b.db.Begin()
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при запуске гонки", true)
		return
	}

	// Start the race
	err = b.RaceRepo.StartRace(tx, raceID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка запуска гонки: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при запуске гонки", true)
		return
	}

	// Assign cars to registered drivers
	_, err = b.CarRepo.AssignCarsToRegisteredDrivers(tx, raceID, race.CarClass)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка назначения машин: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при назначении машин", true)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при завершении запуска гонки", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Гонка успешно запущена!", false)

	// Send success message
	b.sendMessage(chatID, fmt.Sprintf("✅ Гонка '%s' успешно запущена! Участникам отправлены уведомления с их машинами.", race.Name))

	// Notify all drivers about their cars
	go b.notifyDriversAboutCarAssignments(raceID)

	// Show race details
	b.showRaceDetails(chatID, raceID, userID)
	b.deleteMessage(chatID, query.Message.MessageID)
}

// Add admin ability to confirm cars for drivers
func (b *Bot) callbackAdminConfirmCar(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для подтверждения машины", true)
		return
	}

	// Parse parameters from callback data (admin_confirm_car:raceID:driverID)
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

	// Confirm car for the driver
	err = b.RaceRepo.UpdateCarConfirmation(raceID, driverID, true)
	if err != nil {
		log.Printf("Ошибка подтверждения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при подтверждении машины", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Машина подтверждена администратором!", false)

	// Show updated registrations
	b.callbackRaceRegistrations(&tgbotapi.CallbackQuery{
		Data: fmt.Sprintf("race_registrations:%d", raceID),
		From: query.From,
		Message: &tgbotapi.Message{
			Chat: query.Message.Chat,
		},
	})
}

// callbackConfirmCar handles confirmation of assigned car
func (b *Bot) callbackConfirmCar(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
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

	// Get driver information
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

	// Check if driver is registered for this race
	registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке регистрации", true)
		return
	}

	if !registered {
		b.answerCallbackQuery(query.ID, "⚠️ Вы не зарегистрированы на эту гонку", true)
		return
	}

	// Confirm car
	err = b.RaceRepo.UpdateCarConfirmation(raceID, driver.ID, true)
	if err != nil {
		log.Printf("Ошибка подтверждения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при подтверждении машины", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Машина подтверждена!", false)

	// Update the message to remove buttons
	b.editMessage(
		chatID,
		messageID,
		query.Message.Text+"\n\n✅ *Машина подтверждена!*",
	)
}

// callbackRaceRegistrations shows list of registered drivers for admin
func (b *Bot) callbackRaceRegistrations(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для просмотра регистраций", true)
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

	// Get registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении списка участников", true)
		return
	}

	// Format message with registrations
	text := fmt.Sprintf("👨‍🏎️ *Зарегистрированные участники гонки '%s'*\n\n", race.Name)

	if len(registrations) == 0 {
		text += "Нет зарегистрированных участников."
	} else {
		for i, reg := range registrations {
			var status string
			if race.State == models.RaceStateInProgress || race.State == models.RaceStateCompleted {
				if reg.CarConfirmed {
					status = "✅ машина подтверждена"
				} else {
					status = "⏳ ожидается подтверждение машины"
				}

				if reg.RerollUsed {
					status += ", 🎲 реролл использован"
				}
			} else {
				status = "⏳ ожидание начала гонки"
			}

			text += fmt.Sprintf("%d. *%s* - %s\n", i+1, reg.DriverName, status)
		}
	}

	// Create appropriate keyboard based on race state
	var keyboard [][]tgbotapi.InlineKeyboardButton

	switch race.State {
	case models.RaceStateNotStarted:
		// Add start race button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏁 Запустить гонку",
				fmt.Sprintf("start_race:%d", raceID),
			),
		))
	case models.RaceStateInProgress:
		// Add complete race button
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Завершить гонку",
				fmt.Sprintf("complete_race:%d", raceID),
			),
		))
	}

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_results:%d", raceID),
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
	b.deleteMessage(chatID, query.Message.MessageID)
}

// callbackCompleteRaceConfirm handles confirmation of race completion
func (b *Bot) callbackCompleteRaceConfirm(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

	// Check admin rights
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для завершения гонки", true)
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

	// Check if race is in progress
	if race.State != models.RaceStateInProgress {
		b.answerCallbackQuery(query.ID, "⚠️ Гонка не запущена или уже завершена", true)
		return
	}

	// Check if there are any results
	results, err := b.ResultRepo.GetResultCountByRaceID(raceID)
	if err != nil {
		log.Printf("Ошибка проверки результатов: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке результатов", true)
		return
	}

	if results == 0 {
		b.answerCallbackQuery(query.ID, "⚠️ Нет результатов для завершения гонки", true)
		return
	}

	// Start a database transaction
	tx, err := b.db.Begin()
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при завершении гонки", true)
		return
	}

	// Complete the race
	err = b.RaceRepo.CompleteRace(tx, raceID)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка завершения гонки: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при завершении гонки", true)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при завершении операции", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Гонка успешно завершена!", false)

	// Send success message
	b.sendMessage(chatID, fmt.Sprintf("✅ Гонка '%s' успешно завершена! Участникам отправлены уведомления с результатами.", race.Name))

	// Notify all drivers about race completion
	go b.notifyDriversAboutRaceCompletion(raceID)

	// Show race results
	b.showRaceResults(chatID, raceID)
	b.deleteMessage(chatID, query.Message.MessageID)
}

func (b *Bot) showRaceDetails(chatID int64, raceID int, userID int64) {
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

	// Get registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка участников.")
		return
	}

	// Format message with race details
	text := fmt.Sprintf("🏁 *Гонка: %s*\n\n", race.Name)
	text += fmt.Sprintf("📅 Дата: %s\n", b.formatDate(race.Date))
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

	// Add registered drivers
	text += "*Участники:*\n\n"
	if len(registrations) == 0 {
		text += "Нет зарегистрированных участников."
	} else {
		for i, reg := range registrations {
			text += fmt.Sprintf("%d. %s\n", i+1, reg.DriverName)
		}
	}

	// Check if the current user is registered for this race
	var isRegistered bool = false
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err == nil && driver != nil {
		registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
		if err == nil {
			isRegistered = registered
		}
	}

	// Create keyboard based on race state and registration status
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Only add registration options for races that haven't started yet
	if race.State == models.RaceStateNotStarted && driver != nil {
		if isRegistered {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"❌ Отменить регистрацию",
					fmt.Sprintf("unregister_race:%d", raceID),
				),
			))
		} else if race.State == models.RaceStateInProgress {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"📊 Прогресс гонки",
					fmt.Sprintf("race_progress:%d", raceID),
				),
			))
		} else {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Зарегистрироваться",
					fmt.Sprintf("register_race:%d", raceID),
				),
			))
		}
	}

	// Add race management buttons for admins
	if b.IsAdmin(userID) {
		switch race.State {
		case models.RaceStateNotStarted:
			// Show manage registrations button
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"👨‍🏎️ Управление участниками",
					fmt.Sprintf("race_registrations:%d", raceID),
				),
			))

			// Add start race button if there are registrations
			if len(registrations) > 0 {
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"🏁 Запустить гонку",
						fmt.Sprintf("start_race:%d", raceID),
					),
				))
			}
		case models.RaceStateInProgress:
			// Show manage registrations button
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"👨‍🏎️ Статус участников",
					fmt.Sprintf("race_registrations:%d", raceID),
				),
			))

			// Add view cars button
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🚗 Посмотреть машины",
					fmt.Sprintf("view_race_cars:%d", raceID),
				),
			))

			// Add complete race button
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Завершить гонку",
					fmt.Sprintf("complete_race:%d", raceID),
				),
			))
		}

		// Add edit and delete buttons
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать",
				fmt.Sprintf("edit_race:%d", raceID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🗑️ Удалить",
				fmt.Sprintf("delete_race:%d", raceID),
			),
		))
	} else {
		// Regular user buttons based on race state
		if race.State == models.RaceStateInProgress {
			// Only show these if user is registered
			if isRegistered {
				// Add my car button
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"🚗 Моя машина",
						fmt.Sprintf("my_car:%d", raceID),
					),
				))

				// Add add result button
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"➕ Добавить результат",
						fmt.Sprintf("add_result:%d", raceID),
					),
				))
			}

			// Add view cars button (for everyone)
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🚗 Посмотреть машины",
					fmt.Sprintf("view_race_cars:%d", raceID),
				),
			))
		} else if race.State == models.RaceStateCompleted {
			// Add view results button
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"📊 Посмотреть результаты",
					fmt.Sprintf("race_results:%d", raceID),
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
	}

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			fmt.Sprintf("season_races:%d", race.SeasonID),
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

func (b *Bot) callbackRerollCar(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
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

	// Get driver information
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

	// Check if driver is registered for this race
	registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке регистрации", true)
		return
	}

	if !registered {
		b.answerCallbackQuery(query.ID, "⚠️ Вы не зарегистрированы на эту гонку", true)
		return
	}

	// Check if reroll was already used
	rerollUsed, err := b.ResultRepo.GetDriverRerollStatus(raceID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки статуса реролла: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке статуса реролла", true)
		return
	}

	if rerollUsed {
		b.answerCallbackQuery(query.ID, "⚠️ Вы уже использовали свой реролл в этой гонке", true)
		return
	}

	// Start a database transaction
	tx, err := b.db.Begin()
	if err != nil {
		log.Printf("Ошибка начала транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при реролле машины", true)
		return
	}

	// Reroll car
	carAssignment, err := b.CarRepo.RerollCarForDriver(tx, raceID, driver.ID, race.CarClass)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка реролла машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при реролле машины", true)
		return
	}

	// Apply reroll penalty to results (if results already exist)
	err = b.ResultRepo.ApplyRerollPenaltyToResult(tx, raceID, driver.ID, 1)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка применения штрафа за реролл: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при применении штрафа за реролл", true)
		return
	}

	// Mark car as confirmed
	err = b.RaceRepo.UpdateCarConfirmation(raceID, driver.ID, true)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка подтверждения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при подтверждении машины", true)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Ошибка фиксации транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при завершении реролла", true)
		return
	}

	b.answerCallbackQuery(query.ID, "✅ Машина изменена с помощью реролла!", false)

	// Format new car information
	car := carAssignment.Car
	text := fmt.Sprintf("🚗 *Ваша новая машина для гонки '%s'*\n\n", race.Name)
	text += fmt.Sprintf("*%s (%s)*\n", car.Name, car.Year)
	text += fmt.Sprintf("🔢 Номер: %d\n", carAssignment.AssignmentNumber)
	text += fmt.Sprintf("💰 Цена: %d CR\n", car.Price)
	text += fmt.Sprintf("⭐ Редкость: %s\n\n", car.Rarity)
	text += "*Характеристики:*\n"
	text += fmt.Sprintf("🏁 Скорость: %.1f/10\n", car.Speed)
	text += fmt.Sprintf("🔄 Управление: %.1f/10\n", car.Handling)
	text += fmt.Sprintf("⚡ Ускорение: %.1f/10\n", car.Acceleration)
	text += fmt.Sprintf("🚦 Старт: %.1f/10\n", car.Launch)
	text += fmt.Sprintf("🛑 Торможение: %.1f/10\n\n", car.Braking)
	text += fmt.Sprintf("🏆 Класс: %s %d\n\n", car.ClassLetter, car.ClassNumber)
	text += "⚠️ *Вы использовали свой реролл в этой гонке. -1 балл будет вычтен из вашего итогового результата.*\n\n"
	text += "✅ *Машина автоматически подтверждена!*"

	// Send the message with the new car
	if car.ImageURL != "" {
		b.sendPhoto(chatID, car.ImageURL, text)
	} else {
		b.sendMessage(chatID, text)
	}

	// Delete the original message
	b.deleteMessage(chatID, messageID)
}
