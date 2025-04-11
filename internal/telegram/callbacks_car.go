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

	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		b.deleteMessage(chatID, query.Message.MessageID)
		return
	}

	if b.IsAdmin(userID) {
		// Get race state
		if race != nil && race.State == models.RaceStateInProgress {
			b.showAdminRacePanel(chatID, raceID)

			b.deleteMessage(chatID, query.Message.MessageID)

			b.answerCallbackQuery(query.ID, "", false)

			return
		}
	}

	if race != nil && race.State == models.RaceStateInProgress {
		// Перенаправляем в callback активной гонки
		b.callbackActiveRace(query)
		return
	}

	// For non-admins or other race states, proceed with normal race details
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

// Обновленная функция showRaceDetails для использования новой клавиатуры
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

	// Check if the current user is registered for this race
	var isRegistered bool = false
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err == nil && driver != nil {
		registered, err := b.RaceRepo.CheckDriverRegistered(raceID, driver.ID)
		if err == nil {
			isRegistered = registered
		}
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

	// Get registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
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

	// Show entry status for the current user
	if isRegistered {
		text += "\n✅ *Вы зарегистрированы на эту гонку*"
	}

	// Create keyboard using RaceDetailsKeyboard
	keyboard := RaceDetailsKeyboard(raceID, userID, isRegistered, race, b.IsAdmin(userID))

	b.sendMessageWithKeyboard(chatID, text, keyboard)
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

	log.Printf("Обработка реролла машины пользователем: %d", userID)

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

	// Explicitly mark reroll as used in race_registrations table
	_, err = tx.Exec(`
		UPDATE race_registrations
		SET reroll_used = TRUE
		WHERE race_id = $1 AND driver_id = $2
	`, raceID, driver.ID)

	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка установки флага реролла: %v", err)
		b.answerCallbackQuery(query.ID, "⚠️ Произошла ошибка при сохранении информации о реролле", true)
		return
	}

	// Apply reroll penalty to results (if results already exist)
	err = b.ResultRepo.ApplyRerollPenaltyToResult(tx, raceID, driver.ID, 1)
	if err != nil {
		log.Printf("Предупреждение при применении штрафа за реролл: %v (игнорируется, если результаты еще не добавлены)", err)
		// Не делаем rollback, это нормальная ситуация если результатов ещё нет
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

	log.Printf("Успешный реролл машины для гонщика %d (ID: %d) в гонке %d",
		driver.ID, userID, raceID)

	b.answerCallbackQuery(query.ID, "✅ Машина изменена с помощью реролла!", false)

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

	if car.ImageURL != "" {
		b.sendPhoto(chatID, car.ImageURL, text)
	} else {
		b.sendMessage(chatID, text)
	}

	// Delete the original message
	b.deleteMessage(chatID, messageID)

	// Check if all cars are confirmed after this reroll
	b.checkAllCarsConfirmed(raceID)
}
