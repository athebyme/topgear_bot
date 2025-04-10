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
			text += fmt.Sprintf("🚗 Машина: %s (%d)\n", assignment.Car.Name, assignment.Car.Year)
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
