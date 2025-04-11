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

// showRaceCarAssignments показывает назначения машин для гонки
func (b *Bot) showRaceCarAssignments(chatID int64, raceID int, userID int64) {
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

	// Get car assignments
	assignments, err := b.CarRepo.GetRaceCarAssignments(raceID)
	if err != nil {
		log.Printf("Ошибка получения назначений машин: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении назначений машин.")
		return
	}

	// Check if user is admin or a participant with confirmed car
	isAdmin := b.IsAdmin(userID)

	var isParticipant bool
	var hasConfirmedCar bool

	if !isAdmin {
		// Check if user is a participant
		driver, err := b.DriverRepo.GetByTelegramID(userID)
		if err == nil && driver != nil {
			// Check if driver is registered for this race
			isRegistered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
			if err == nil && isRegistered {
				isParticipant = true

				// Check if driver has confirmed their car
				err = b.db.QueryRow(`
                    SELECT car_confirmed FROM race_registrations
                    WHERE race_id = $1 AND driver_id = $2
                `, raceID, driver.ID).Scan(&hasConfirmedCar)

				if err != nil {
					log.Printf("Ошибка проверки подтверждения машины: %v", err)
					hasConfirmedCar = false
				}
			}
		}
	}

	// Format message with assignments
	text := fmt.Sprintf("🏁 *Машины для гонки '%s'*\n\n", race.Name)
	text += fmt.Sprintf("📅 %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s (%s)\n\n", race.CarClass, models.GetCarClassName(race.CarClass))

	if len(assignments) == 0 {
		text += "⚠️ Машины еще не назначены для этой гонки."
	} else if isAdmin || race.State == models.RaceStateCompleted {
		// Show all cars to admins or if race is completed
		for _, assignment := range assignments {
			text += fmt.Sprintf("*%s*\n", assignment.DriverName)
			text += fmt.Sprintf("🔢 Номер: %d\n", assignment.AssignmentNumber)
			text += fmt.Sprintf("🚗 Машина: %s (%s)\n", assignment.Car.Name, assignment.Car.Year)
			text += fmt.Sprintf("⭐ Редкость: %s\n\n", assignment.Car.Rarity)
		}
	} else if isParticipant && race.State == models.RaceStateInProgress {
		// For participants in a race that's in progress

		// First get all confirmations to check if all participants have confirmed
		var allConfirmed bool = true
		var confirmedCount int = 0

		registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
		if err == nil {
			for _, reg := range registrations {
				if reg.CarConfirmed {
					confirmedCount++
				} else {
					allConfirmed = false
				}
			}
		}

		// If the participant has confirmed their car and either all cars are confirmed
		// or the admin has explicitly made cars visible
		if hasConfirmedCar && (allConfirmed || confirmedCount > len(registrations)/2) {
			for _, assignment := range assignments {
				text += fmt.Sprintf("*%s*\n", assignment.DriverName)
				text += fmt.Sprintf("🔢 Номер: %d\n", assignment.AssignmentNumber)
				text += fmt.Sprintf("🚗 Машина: %s (%s)\n", assignment.Car.Name, assignment.Car.Year)
				text += fmt.Sprintf("⭐ Редкость: %s\n\n", assignment.Car.Rarity)
			}
		} else {
			text += "⚠️ Машины других участников будут видны после того, как все гонщики подтвердят свой выбор."

			// Show at least their own car
			driver, err := b.DriverRepo.GetByTelegramID(userID)
			if err == nil && driver != nil {
				for _, assignment := range assignments {
					if assignment.DriverID == driver.ID {
						text += "\n\n*Ваша машина:*\n"
						text += fmt.Sprintf("🔢 Номер: %d\n", assignment.AssignmentNumber)
						text += fmt.Sprintf("🚗 Машина: %s (%s)\n", assignment.Car.Name, assignment.Car.Year)
						text += fmt.Sprintf("⭐ Редкость: %s\n", assignment.Car.Rarity)
						break
					}
				}
			}
		}
	} else {
		text += "⚠️ Машины участников будут видны после начала гонки."
	}

	// Create keyboard with additional actions
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add buttons for admins
	if isAdmin {
		if race.State == models.RaceStateInProgress {
			// Add buttons for forcing confirmation of all cars
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Подтвердить все машины",
					fmt.Sprintf("admin_confirm_all_cars:%d", raceID),
				),
			))
		}

		// Admin can send car notifications again
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📨 Отправить уведомления о машинах",
				fmt.Sprintf("admin_send_notifications:%d:cars", raceID),
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

	// Send message with keyboard
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// callbackRaceDetails обрабатывает переход к деталям гонки
func (b *Bot) callbackRaceDetails(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID

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

	if b.IsAdmin(userID) && race.State == models.RaceStateInProgress {
		b.showAdminRacePanel(chatID, raceID)
		b.deleteMessage(chatID, query.Message.MessageID)
		return
	}

	b.showUniversalRaceCard(chatID, raceID, userID)
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

// callbackStartRace обрабатывает запуск гонки администратором
func (b *Bot) callbackStartRace(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	// Отправляем уведомление, что запрос обрабатывается
	b.answerCallbackQuery(query.ID, "⏳ Запуск гонки...", false)

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.answerCallbackQuery(query.ID, "⛔ У вас нет прав для запуска гонки", true)
		return
	}

	// Извлекаем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.sendMessage(chatID, "⚠️ Неверный формат запроса для запуска гонки")
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный ID гонки")
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
		b.sendMessage(chatID, "⚠️ Гонка не найдена")
		return
	}

	// Получаем всех зарегистрированных гонщиков
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

	// Запускаем гонку: обновляем статус на "в процессе"
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
	successMsg := b.sendMessage(chatID, fmt.Sprintf("✅ Гонка '%s' успешно запущена! Участникам отправлены уведомления с их машинами.", race.Name))

	// Отправляем уведомления участникам в отдельной горутине
	go b.notifyDriversAboutCarAssignments(raceID)

	// Удаляем старое сообщение с кнопками
	b.deleteMessage(chatID, messageID)

	// Важно: автоматически показываем админ-панель после запуска гонки
	b.showAdminRacePanel(chatID, raceID)

	// Удаляем сообщение об успешном запуске через некоторое время
	go func() {
		time.Sleep(5 * time.Second)
		b.deleteMessage(chatID, successMsg.MessageID)
	}()
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

func (b *Bot) callbackConfirmCar(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	log.Printf("Обработка подтверждения машины пользователем: %d", userID)

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

	// Проверяем, не подтвердил ли уже гонщик свою машину
	var alreadyConfirmed bool
	err = b.db.QueryRow(`
		SELECT car_confirmed FROM race_registrations
		WHERE race_id = $1 AND driver_id = $2
	`, raceID, driver.ID).Scan(&alreadyConfirmed)

	if err != nil {
		log.Printf("Ошибка проверки статуса подтверждения: %v", err)
	} else if alreadyConfirmed {
		log.Printf("Гонщик %d (ID: %d) пытается повторно подтвердить машину в гонке %d",
			driver.ID, userID, raceID)
		b.answerCallbackQuery(query.ID, "Машина уже подтверждена", true)
		return
	}

	// Обновляем статус подтверждения машины
	err = b.RaceRepo.UpdateCarConfirmation(raceID, driver.ID, true)
	if err != nil {
		log.Printf("Ошибка подтверждения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при подтверждении машины", true)
		return
	}

	log.Printf("Гонщик %d (ID: %d) подтвердил машину в гонке %d",
		driver.ID, userID, raceID)

	b.answerCallbackQuery(query.ID, "✅ Машина подтверждена!", false)

	// Получаем данные о машине для отображения
	car, err := b.CarRepo.GetDriverCarAssignment(raceID, driver.ID)
	if err == nil && car != nil {
		race, err := b.RaceRepo.GetByID(raceID)
		raceName := "текущей гонки"
		if err == nil && race != nil {
			raceName = race.Name
		}

		text := fmt.Sprintf("🚗 *Ваша машина для гонки '%s'*\n\n", raceName)
		text += fmt.Sprintf("*%s (%s)*\n", car.Car.Name, car.Car.Year)
		text += fmt.Sprintf("🔢 Номер: %d\n", car.AssignmentNumber)
		text += fmt.Sprintf("✅ *Машина подтверждена!*\n\n")

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

		if car.Car.ImageURL != "" {
			b.sendPhotoWithKeyboard(chatID, car.Car.ImageURL, text, keyboard)
		} else {
			b.sendMessageWithKeyboard(chatID, text, keyboard)
		}
	}

	b.deleteMessage(chatID, messageID)

	b.checkAllCarsConfirmed(raceID)

	b.notifyAdminsAboutCarConfirmation(raceID, driver.ID)
}

// Исправленная функция проверки подтверждения всех машин
func (b *Bot) checkAllCarsConfirmed(raceID int) {
	// Получаем все регистрации
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения регистраций: %v", err)
		return
	}

	if len(registrations) == 0 {
		return
	}

	log.Printf("Проверка подтверждения машин: гонка ID=%d, всего участников: %d", raceID, len(registrations))

	// Проверяем все ли машины подтверждены
	allConfirmed := true
	confirmedCount := 0

	for _, reg := range registrations {
		if reg.CarConfirmed {
			confirmedCount++
		} else {
			allConfirmed = false
		}
	}

	log.Printf("Подтверждено машин: %d из %d, все подтверждены: %v",
		confirmedCount, len(registrations), allConfirmed)

	if allConfirmed && confirmedCount > 0 {
		race, err := b.RaceRepo.GetByID(raceID)
		if err != nil || race == nil {
			log.Printf("Ошибка получения гонки: %v", err)
			return
		}

		if race.State == models.RaceStateInProgress {
			log.Printf("Все машины подтверждены для гонки %d (%s). Отправка уведомлений участникам.",
				raceID, race.Name)

			for _, reg := range registrations {
				var telegramID int64
				err := b.db.QueryRow("SELECT telegram_id FROM drivers WHERE id = $1", reg.DriverID).Scan(&telegramID)
				if err != nil {
					log.Printf("Ошибка получения Telegram ID гонщика %d: %v", reg.DriverID, err)
					continue
				}

				log.Printf("Отправка уведомления о подтверждении всех машин гонщику %d (Telegram ID: %d)",
					reg.DriverID, telegramID)

				message := fmt.Sprintf("🏁 *Все участники подтвердили свои машины!*\n\nГонка '%s' официально началась. Теперь вы можете видеть машины всех участников.", race.Name)
				b.sendMessage(telegramID, message)
			}

			// Отправляем уведомления администраторам
			for adminID := range b.AdminIDs {
				log.Printf("Отправка уведомления администратору: %d", adminID)
				b.sendMessage(adminID, fmt.Sprintf("🏁 *Все участники подтвердили свои машины в гонке '%s'!*", race.Name))
			}
		}
	}
}

func (b *Bot) notifyAdminsAboutCarConfirmation(raceID int, driverID int) {
	// Get driver information
	var driverName string
	err := b.db.QueryRow("SELECT name FROM drivers WHERE id = $1", driverID).Scan(&driverName)
	if err != nil {
		log.Printf("Ошибка получения имени гонщика: %v", err)
		return
	}

	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		return
	}

	if race == nil {
		return
	}

	// Get admins
	for adminID := range b.AdminIDs {
		b.sendMessage(adminID, fmt.Sprintf("✅ Гонщик *%s* подтвердил выбор машины в гонке '%s'",
			driverName, race.Name))
	}
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

// showRaceDetails показывает детальную информацию о гонке
func (b *Bot) showRaceDetails(chatID int64, raceID int, userID int64) {
	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке %d: %v", raceID, err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке.")
		return
	}

	if race == nil {
		log.Printf("Гонка с ID %d не найдена", raceID)
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		return
	}

	// Проверяем, зарегистрирован ли пользователь на эту гонку
	var isRegistered bool
	var driver *models.Driver

	if driverObj, err := b.DriverRepo.GetByTelegramID(userID); err == nil && driverObj != nil {
		driver = driverObj
		registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
		if err == nil {
			isRegistered = registered
		}
	}

	// Получаем зарегистрированных гонщиков
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков для гонки %d: %v", raceID, err)
		// Продолжаем выполнение без списка регистраций
	}

	// Получаем информацию о сезоне
	season, err := b.SeasonRepo.GetByID(race.SeasonID)
	if err != nil {
		log.Printf("Ошибка получения информации о сезоне %d: %v", race.SeasonID, err)
		// Продолжаем выполнение без информации о сезоне
	}

	// Формируем заголовок и основную информацию
	var text string

	// Заголовок с учетом статуса гонки
	switch race.State {
	case models.RaceStateInProgress:
		text = fmt.Sprintf("🏎️ *АКТИВНАЯ ГОНКА: %s*\n\n", race.Name)
	case models.RaceStateNotStarted:
		text = fmt.Sprintf("⏳ *ПРЕДСТОЯЩАЯ ГОНКА: %s*\n\n", race.Name)
	case models.RaceStateCompleted:
		text = fmt.Sprintf("✅ *ЗАВЕРШЕННАЯ ГОНКА: %s*\n\n", race.Name)
	default:
		text = fmt.Sprintf("🏁 *ГОНКА: %s*\n\n", race.Name)
	}

	// Добавляем информацию о сезоне, если доступна
	if season != nil {
		text += fmt.Sprintf("🏆 Сезон: %s\n", season.Name)
	}

	// Основная информация о гонке
	text += fmt.Sprintf("📅 Дата: %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс машин: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n\n", strings.Join(race.Disciplines, ", "))

	// Информация о регистрации текущего пользователя
	if driver != nil {
		if isRegistered {
			text += "✅ *Вы зарегистрированы на эту гонку*\n\n"

			// Если гонка активна, добавляем информацию о машине
			if race.State == models.RaceStateInProgress {
				assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, driver.ID)
				if err == nil && assignment != nil {
					text += "*Ваша машина:*\n"
					text += fmt.Sprintf("🚗 %s\n", assignment.Car.Name)
					text += fmt.Sprintf("🔢 Номер: %d\n\n", assignment.AssignmentNumber)

					// Проверяем статус подтверждения машины
					var confirmed bool
					err = b.db.QueryRow(`
						SELECT car_confirmed FROM race_registrations 
						WHERE race_id = $1 AND driver_id = $2
					`, raceID, driver.ID).Scan(&confirmed)

					if err == nil {
						if confirmed {
							text += "✅ Машина подтверждена\n\n"
						} else {
							text += "⚠️ *Машина не подтверждена.* Используйте кнопку 'Моя машина' для подтверждения\n\n"
						}
					}
				}
			}
		} else if race.State == models.RaceStateNotStarted {
			text += "❌ *Вы не зарегистрированы на эту гонку*\n"
			text += "Нажмите кнопку 'Зарегистрироваться' ниже, чтобы принять участие\n\n"
		}
	}

	// Информация об участниках
	if len(registrations) > 0 {
		text += fmt.Sprintf("👨‍🏎️ *Участники (%d):*\n", len(registrations))

		// Ограничиваем список до 10 участников, чтобы не перегружать интерфейс
		showLimit := 10
		showAll := len(registrations) <= showLimit

		for i, reg := range registrations {
			if !showAll && i >= showLimit {
				break
			}

			// Добавляем статус для активных гонок
			if race.State == models.RaceStateInProgress {
				var carConfirmed bool
				err = b.db.QueryRow(`
					SELECT car_confirmed FROM race_registrations 
					WHERE race_id = $1 AND driver_id = $2
				`, raceID, reg.DriverID).Scan(&carConfirmed)

				if err == nil && carConfirmed {
					text += fmt.Sprintf("• %s ✅\n", reg.DriverName)
				} else {
					text += fmt.Sprintf("• %s ⏳\n", reg.DriverName)
				}
			} else {
				text += fmt.Sprintf("• %s\n", reg.DriverName)
			}
		}

		if !showAll {
			text += fmt.Sprintf("...и еще %d участников\n", len(registrations)-showLimit)
		}

		text += "\n"
	} else {
		text += "👨‍🏎️ *Пока нет зарегистрированных участников*\n\n"
	}

	// Дополнительная информация в зависимости от статуса гонки
	switch race.State {
	case models.RaceStateInProgress:
		// Для активных гонок показываем информацию о подтверждении машин
		var confirmedCount int
		for _, reg := range registrations {
			var carConfirmed bool
			err = b.db.QueryRow(`
				SELECT car_confirmed FROM race_registrations 
				WHERE race_id = $1 AND driver_id = $2
			`, raceID, reg.DriverID).Scan(&carConfirmed)

			if err == nil && carConfirmed {
				confirmedCount++
			}
		}

		text += fmt.Sprintf("✅ *Подтверждено машин:* %d из %d\n", confirmedCount, len(registrations))

		// Информация о поданных результатах
		resultCount, _ := b.ResultRepo.GetResultCountByRaceID(raceID)
		text += fmt.Sprintf("📊 *Подано результатов:* %d из %d\n", resultCount, len(registrations))

	case models.RaceStateNotStarted:
		// Для предстоящих гонок показываем ожидаемое время до начала
		timeDiff := race.Date.Sub(time.Now())
		if timeDiff > 0 {
			days := int(timeDiff.Hours() / 24)
			hours := int(timeDiff.Hours()) % 24

			if days > 0 {
				text += fmt.Sprintf("⏱️ *До начала:* %d дней %d часов\n", days, hours)
			} else {
				text += fmt.Sprintf("⏱️ *До начала:* %d часов %d минут\n", hours, int(timeDiff.Minutes())%60)
			}
		}
	case models.RaceStateCompleted:
		// Для завершенных гонок показываем победителей
		results, err := b.ResultRepo.GetRaceResultsWithRerollPenalty(raceID)
		if err == nil && len(results) > 0 {
			text += "🏆 *Топ-3 победителя:*\n"

			count := len(results)
			if count > 3 {
				count = 3
			}

			for i := 0; i < count; i++ {
				text += fmt.Sprintf("%d. *%s* - %d очков\n", i+1, results[i].DriverName, results[i].TotalScore)
			}

			text += "\nИспользуйте кнопку 'Результаты' для просмотра полных результатов\n"
		}
	}

	// Создаем клавиатуру в зависимости от статуса гонки
	var keyboard [][]tgbotapi.InlineKeyboardButton

	switch race.State {
	case models.RaceStateInProgress:
		// Для активных гонок
		if driver != nil && isRegistered {
			// Кнопки для зарегистрированного участника
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

		// Общие кнопки для активной гонки
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📊 Прогресс гонки",
				fmt.Sprintf("race_progress:%d", raceID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Машины участников",
				fmt.Sprintf("view_race_cars:%d", raceID),
			),
		))

	case models.RaceStateNotStarted:
		// Для предстоящих гонок
		if driver != nil {
			if isRegistered {
				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"❌ Отменить регистрацию",
						fmt.Sprintf("unregister_race:%d", raceID),
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

		// Кнопка для просмотра списка участников
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"👨‍🏎️ Список участников",
				fmt.Sprintf("race_registrations:%d", raceID),
			),
		))

	case models.RaceStateCompleted:
		// Для завершенных гонок
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏆 Результаты",
				fmt.Sprintf("race_results:%d", raceID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Машины участников",
				fmt.Sprintf("view_race_cars:%d", raceID),
			),
		))
	}

	// Специальные кнопки для администраторов
	if b.IsAdmin(userID) {
		var adminButtons []tgbotapi.InlineKeyboardButton

		switch race.State {
		case models.RaceStateNotStarted:
			adminButtons = append(adminButtons,
				tgbotapi.NewInlineKeyboardButtonData(
					"🏁 Запустить гонку",
					fmt.Sprintf("start_race:%d", raceID),
				),
			)
		case models.RaceStateInProgress:
			adminButtons = append(adminButtons,
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Завершить гонку",
					fmt.Sprintf("complete_race:%d", raceID),
				),
			)
		}

		// Если есть кнопки для админа, добавляем их
		if len(adminButtons) > 0 {
			keyboard = append(keyboard, adminButtons)
		}

		// Добавляем общую админ-панель
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⚙️ Админ-панель",
				fmt.Sprintf("admin_race_panel:%d", raceID),
			),
		))
	}

	// Кнопка возврата к списку гонок
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонкам",
			"races",
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// Обновленная функция showAdminRacePanel для использования новой клавиатуры
func (b *Bot) showAdminRacePanel(chatID int64, raceID int) {
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

	// Get registered drivers with car confirmation status
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка участников.")
		return
	}

	// Get results count
	resultsCount, err := b.ResultRepo.GetResultCountByRaceID(raceID)
	if err != nil {
		log.Printf("Ошибка получения количества результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении количества результатов.")
		return
	}

	// Format message with admin panel
	text := fmt.Sprintf("⚙️ *Админ-панель гонки: %s*\n\n", race.Name)
	text += fmt.Sprintf("📅 Дата: %s\n", b.formatDate(race.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", race.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n", strings.Join(race.Disciplines, ", "))
	text += fmt.Sprintf("🏆 Статус: %s\n\n", getStatusText(race.State))

	text += fmt.Sprintf("👨‍🏎️ Участников: %d\n", len(registrations))
	text += fmt.Sprintf("📊 Подано результатов: %d\n\n", resultsCount)

	// Add driver statuses
	text += "*Статусы участников:*\n"

	var (
		confirmedCount     int
		unconfirmedDrivers []int
	)

	for i, reg := range registrations {
		var statusText string

		if reg.CarConfirmed {
			statusText = "✅ подтвердил"
			confirmedCount++
		} else {
			statusText = "⏳ ожидает"
			unconfirmedDrivers = append(unconfirmedDrivers, reg.DriverID)
		}

		if reg.RerollUsed {
			statusText += ", 🎲 реролл"
		}

		text += fmt.Sprintf("%d. %s - %s\n", i+1, reg.DriverName, statusText)
	}

	// Create keyboard using AdminRacePanelKeyboard
	keyboard := AdminRacePanelKeyboard(raceID, race.State)

	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) callbackRerollCar(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	log.Printf("⭐ callbackRerollCar: Начало обработки реролла машины для пользователя: %d", userID)

	// Разбираем ID гонки из данных запроса
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		log.Printf("❌ callbackRerollCar: Ошибка - неверный формат данных колбэка: %s", query.Data)
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	raceID, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка - не удалось преобразовать ID гонки: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID гонки", true)
		return
	}

	log.Printf("📌 callbackRerollCar: Получен ID гонки: %d", raceID)

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка получения данных гонщика: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении данных гонщика", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика. Пожалуйста, попробуйте снова.")
		return
	}

	if driver == nil {
		log.Printf("❌ callbackRerollCar: Гонщик не найден для пользователя %d", userID)
		b.answerCallbackQuery(query.ID, "⚠️ Вы не зарегистрированы как гонщик", true)
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	log.Printf("📌 callbackRerollCar: Гонщик найден: ID=%d, Name=%s", driver.ID, driver.Name)

	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка получения информации о гонке: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при получении информации о гонке", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке. Пожалуйста, попробуйте снова.")
		return
	}

	if race == nil {
		log.Printf("❌ callbackRerollCar: Гонка с ID %d не найдена", raceID)
		b.answerCallbackQuery(query.ID, "⚠️ Гонка не найдена", true)
		b.sendMessage(chatID, "⚠️ Гонка не найдена. Пожалуйста, выберите другую гонку.")
		return
	}

	log.Printf("📌 callbackRerollCar: Гонка найдена: ID=%d, Name=%s, State=%s", race.ID, race.Name, race.State)

	// Проверяем, зарегистрирован ли гонщик на эту гонку
	registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка проверки регистрации: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при проверке регистрации", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке регистрации. Пожалуйста, попробуйте снова.")
		return
	}

	if !registered {
		log.Printf("❌ callbackRerollCar: Гонщик %d не зарегистрирован на гонку %d", driver.ID, raceID)
		b.answerCallbackQuery(query.ID, "⚠️ Вы не зарегистрированы на эту гонку", true)
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы на эту гонку.")
		return
	}

	// Проверяем, был ли уже использован реролл
	rerollUsed, err := b.ResultRepo.GetDriverRerollStatus(raceID, driver.ID)
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка проверки статуса реролла: %v", err)
		// Продолжаем, предполагая, что реролл не использован
		rerollUsed = false
	}

	log.Printf("📌 callbackRerollCar: Статус реролла для гонщика %d в гонке %d: %v", driver.ID, raceID, rerollUsed)

	if rerollUsed {
		log.Printf("❌ callbackRerollCar: Гонщик %d уже использовал реролл в гонке %d", driver.ID, raceID)
		b.answerCallbackQuery(query.ID, "⚠️ Вы уже использовали свой реролл в этой гонке", true)
		b.sendMessage(chatID, "⚠️ Вы уже использовали реролл в этой гонке. Каждому гонщику разрешен только один реролл.")
		return
	}

	// Начинаем транзакцию
	tx, err := b.db.Begin()
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка начала транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при реролле машины", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при выполнении операции. Пожалуйста, попробуйте снова.")
		return
	}

	log.Printf("📌 callbackRerollCar: Транзакция начата")

	// Реролл машины
	carAssignment, err := b.CarRepo.RerollCarForDriver(tx, raceID, driver.ID, race.CarClass)
	if err != nil {
		tx.Rollback()
		log.Printf("❌ callbackRerollCar: Ошибка реролла машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при реролле машины", true)
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Ошибка при реролле машины: %v", err))
		return
	}

	log.Printf("📌 callbackRerollCar: Новая машина назначена: %s", carAssignment.Car.Name)

	// Отмечаем, что реролл был использован
	_, err = tx.Exec(`
		UPDATE race_registrations
		SET reroll_used = TRUE
		WHERE race_id = $1 AND driver_id = $2
	`, raceID, driver.ID)

	if err != nil {
		tx.Rollback()
		log.Printf("❌ callbackRerollCar: Ошибка установки флага реролла: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при сохранении информации о реролле", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении информации о реролле. Пожалуйста, попробуйте снова.")
		return
	}

	log.Printf("📌 callbackRerollCar: Флаг реролла установлен для гонщика %d в гонке %d", driver.ID, raceID)

	// Применяем штраф реролла к результатам (если результаты уже существуют)
	err = b.ResultRepo.ApplyRerollPenaltyToResult(tx, raceID, driver.ID, 1)
	if err != nil {
		log.Printf("⚠️ callbackRerollCar: Предупреждение при применении штрафа: %v (игнорируется, если результаты еще не добавлены)", err)
		// Не делаем rollback, это нормальная ситуация если результатов еще нет
	}

	// Отмечаем машину как подтвержденную
	err = b.RaceRepo.UpdateCarConfirmation(raceID, driver.ID, true)
	if err != nil {
		tx.Rollback()
		log.Printf("❌ callbackRerollCar: Ошибка подтверждения машины: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при подтверждении машины", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при подтверждении машины. Пожалуйста, попробуйте снова.")
		return
	}

	log.Printf("📌 callbackRerollCar: Машина отмечена как подтвержденная")

	// Завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		log.Printf("❌ callbackRerollCar: Ошибка подтверждения транзакции: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при завершении реролла", true)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при завершении операции реролла.")
		return
	}

	log.Printf("✅ callbackRerollCar: Успешный реролл машины для гонщика %d (ID: %d) в гонке %d",
		driver.ID, userID, raceID)

	b.answerCallbackQuery(query.ID, "✅ Машина изменена с помощью реролла!", false)

	// Отображаем информацию о новой машине
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

	// Добавляем клавиатуру для возврата к гонке
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

	// Отправляем информацию о новой машине
	if car.ImageURL != "" {
		b.sendPhotoWithKeyboard(chatID, car.ImageURL, text, keyboard)
	} else {
		b.sendMessageWithKeyboard(chatID, text, keyboard)
	}

	// Удаляем исходное сообщение
	b.deleteMessage(chatID, messageID)

	// Проверяем, все ли машины подтверждены после этого реролла
	b.checkAllCarsConfirmed(raceID)

	// Уведомляем администраторов о реролле
	//b.notifyAdminsAboutReroll(raceID, driver.ID, car.Name)
}
