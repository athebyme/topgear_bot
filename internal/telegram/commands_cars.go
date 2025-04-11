package telegram

import (
	"fmt"
	"log"
	"strings"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) registerCarCommandHandlers() {
	b.CommandHandlers["cars"] = b.handleCars
	b.CommandHandlers["carclass"] = b.handleCarClass
	b.CommandHandlers["joinrace"] = b.handleJoinRace
	b.CommandHandlers["leaverace"] = b.handleUnregisterFromRace
	b.CommandHandlers["unregister"] = b.handleUnregisterFromRace
	b.CommandHandlers["mycar"] = b.handleMyCar
	b.CommandHandlers["raceregister"] = b.handleRegisterForRace
}

// handleCars обрабатывает команду /cars
func (b *Bot) handleCars(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Получаем общую статистику по машинам
	classCounts, err := b.CarRepo.GetClassCounts()
	if err != nil {
		log.Printf("Ошибка получения количества машин по классам: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении статистики по машинам.")
		return
	}

	// Проверяем, есть ли машины в базе
	if len(classCounts) == 0 {
		b.sendMessage(chatID, "⚠️ База данных машин пуста. Необходимо запустить парсер.")
		return
	}

	// Формируем сообщение со статистикой
	text := "🚗 *Машины Forza Horizon 4*\n\n"
	text += "Выберите класс машин для просмотра:\n\n"

	// Добавляем статистику по каждому классу
	for _, class := range models.CarClasses {
		count := classCounts[class.Letter]
		if count > 0 {
			text += fmt.Sprintf("*%s* - %d машин\n", class.Name, count)
		}
	}

	// Создаем клавиатуру для выбора класса
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, class := range models.CarClasses {
		count := classCounts[class.Letter]
		if count > 0 {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s (%d)", class.Name, count),
					fmt.Sprintf("car_class:%s", class.Letter),
				),
			))
		}
	}

	// Добавляем кнопку обновления базы машин для админов
	if b.IsAdmin(message.From.ID) {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔄 Обновить базу машин",
				"update_cars_db",
			),
		))
	}

	// Отправляем сообщение с клавиатурой
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// handleCarClass обрабатывает команду /carclass
func (b *Bot) handleCarClass(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	args := strings.Split(message.Text, " ")

	if len(args) < 2 {
		b.sendMessage(chatID, "⚠️ Укажите класс машины. Пример: /carclass A")
		return
	}

	// Получаем класс из аргумента
	classLetter := strings.ToUpper(args[1])

	// Проверяем корректность класса
	class := models.GetCarClassByLetter(classLetter)
	if class == nil {
		b.sendMessage(chatID, "⚠️ Указан некорректный класс машины. Доступные классы: D, C, B, A, S1, S2, X")
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

	// Формируем сообщение со списком машин
	text := fmt.Sprintf("🚗 *Машины класса %s*\n\n", class.Name)
	text += fmt.Sprintf("Всего машин: %d\n\n", len(cars))

	// Ограничиваем количество машин в сообщении
	maxCars := 20
	showingAll := len(cars) <= maxCars

	// Добавляем информацию о машинах
	for i, car := range cars {
		if i >= maxCars {
			break
		}

		text += fmt.Sprintf("%d. *%s (%d)* - %d CR\n", i+1, car.Name, car.Year, car.Price)
	}

	// Добавляем примечание, если показаны не все машины
	if !showingAll {
		text += fmt.Sprintf("\n...и еще %d машин. Используйте инлайн кнопки для просмотра всех машин.", len(cars)-maxCars)
	}

	// Создаем клавиатуру для пагинации и просмотра машин
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Добавляем кнопки для просмотра случайной машины этого класса
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🎲 Случайная машина",
			fmt.Sprintf("random_car:%s", classLetter),
		),
	))

	// Добавляем кнопки пагинации, если машин много
	if !showingAll {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Показать все машины",
				fmt.Sprintf("car_class_all:%s", classLetter),
			),
		))
	}

	// Добавляем кнопку возврата к списку классов
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к классам",
			"cars",
		),
	))

	// Отправляем сообщение с клавиатурой
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// assignRandomCarsForRace назначает случайные машины для гонки
func (b *Bot) assignRandomCarsForRace(raceID int, carClass string) error {
	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о гонке: %v", err)
	}

	if race == nil {
		return fmt.Errorf("гонка не найдена")
	}

	// Получаем всех зарегистрированных гонщиков
	drivers, err := b.DriverRepo.GetAll()
	if err != nil {
		return fmt.Errorf("ошибка получения списка гонщиков: %v", err)
	}

	if len(drivers) == 0 {
		return fmt.Errorf("нет зарегистрированных гонщиков")
	}

	// Собираем ID гонщиков
	var driverIDs []int
	for _, driver := range drivers {
		driverIDs = append(driverIDs, driver.ID)
	}

	// Начинаем транзакцию
	tx, err := b.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	// Удаляем предыдущие назначения машин для этой гонки
	err = b.CarRepo.DeleteRaceCarAssignments(tx, raceID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка удаления предыдущих назначений: %v", err)
	}

	// Назначаем случайные машины
	_, err = b.CarRepo.AssignRandomCars(tx, raceID, driverIDs, carClass)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка назначения случайных машин: %v", err)
	}

	// Обновляем класс машин для гонки
	race.CarClass = carClass
	err = b.RaceRepo.Update(tx, race)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка обновления класса машин для гонки: %v", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка подтверждения транзакции: %v", err)
	}

	return nil
}

// handleRegisterForRace handles the /register command to register for an upcoming race
func (b *Bot) handleRegisterForRace(message *tgbotapi.Message) {
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
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /start чтобы начать регистрацию.")
		return
	}

	// Get upcoming races
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

	// Create keyboard with upcoming races
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

// handleUnregisterFromRace handles the /unregister command to unregister from an upcoming race
func (b *Bot) handleUnregisterFromRace(message *tgbotapi.Message) {
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
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /start чтобы начать регистрацию.")
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

// notifyDriversAboutCarAssignments sends car assignments to all drivers in a race
func (b *Bot) notifyDriversAboutCarAssignments(raceID int) {
	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		return
	}

	if race == nil {
		log.Println("Гонка не найдена для отправки уведомлений")
		return
	}

	// Get all registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		return
	}

	// Logging for debugging
	log.Printf("Отправка уведомлений о машинах для %d гонщиков", len(registrations))

	for _, reg := range registrations {
		// Get driver's Telegram ID
		var telegramID int64
		err := b.db.QueryRow("SELECT telegram_id FROM drivers WHERE id = $1", reg.DriverID).Scan(&telegramID)
		if err != nil {
			log.Printf("Ошибка получения Telegram ID гонщика %d: %v", reg.DriverID, err)
			continue
		}

		// Get car assignment
		assignment, err := b.CarRepo.GetDriverCarAssignment(raceID, reg.DriverID)
		if err != nil {
			log.Printf("Ошибка получения назначения машины для гонщика %d: %v", reg.DriverID, err)
			continue
		}

		if assignment == nil {
			log.Printf("Машина не назначена для гонщика %d в гонке %d", reg.DriverID, raceID)
			continue
		}

		// Format car information
		car := assignment.Car
		text := fmt.Sprintf("🏁 *Гонка началась: %s*\n\n", race.Name)
		text += fmt.Sprintf("🚗 *Ваша машина для этой гонки:*\n\n")
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
		text += fmt.Sprintf("🏆 Класс: %s %d\n\n", car.ClassLetter, car.ClassNumber)
		text += "*У вас есть возможность сделать реролл машины (получить другую), но это будет стоить -1 балл в итоговом зачете гонки.*"

		// Create keyboard for confirmation or reroll
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Подтвердить выбор машины",
					fmt.Sprintf("confirm_car:%d", raceID),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🎲 Реролл (-1 балл)",
					fmt.Sprintf("reroll_car:%d", raceID),
				),
			),
		)

		// Explicitly log before sending to debug issues
		log.Printf("Отправка уведомления о машине гонщику %d (telegramID: %d)", reg.DriverID, telegramID)

		// Send message with keyboard and car image if available
		var sentMsg tgbotapi.Message
		if car.ImageURL != "" {
			sentMsg = b.sendPhotoWithKeyboard(telegramID, car.ImageURL, text, keyboard)
		} else {
			sentMsg = b.sendMessageWithKeyboard(telegramID, text, keyboard)
		}

		log.Printf("Уведомление отправлено гонщику %d, ID сообщения: %d", reg.DriverID, sentMsg.MessageID)
	}
}

// notifyDriversAboutRaceCompletion sends race results to all participants
func (b *Bot) notifyDriversAboutRaceCompletion(raceID int) {
	// Get race information
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		return
	}

	if race == nil {
		log.Println("Гонка не найдена для отправки уведомлений")
		return
	}

	// Get race results
	results, err := b.ResultRepo.GetRaceResultsWithRerollPenalty(raceID)
	if err != nil {
		log.Printf("Ошибка получения результатов гонки: %v", err)
		return
	}

	if len(results) == 0 {
		log.Println("Нет результатов для отправки уведомлений")
		return
	}

	// Format results message
	text := fmt.Sprintf("🏁 *Гонка завершена: %s*\n\n", race.Name)
	text += "*Итоговые результаты:*\n\n"

	for i, result := range results {
		text += fmt.Sprintf("%d. *%s* (%s)\n", i+1, result.DriverName, result.CarName)
		text += fmt.Sprintf("   🔢 Номер: %d\n", result.CarNumber)

		// Add disciplinе results
		var placesText []string
		for _, discipline := range race.Disciplines {
			place := result.Results[discipline]
			emoji := getPlaceEmoji(place)
			placesText = append(placesText, fmt.Sprintf("%s %s: %s", emoji, discipline, getPlaceText(place)))
		}

		text += fmt.Sprintf("   📊 %s\n", strings.Join(placesText, " | "))

		// Add penalty if any
		if result.RerollPenalty > 0 {
			text += fmt.Sprintf("   ⚠️ Штраф за реролл: -%d\n", result.RerollPenalty)
		}

		text += fmt.Sprintf("   🏆 Всего очков: %d\n\n", result.TotalScore)
	}

	// Get all registered drivers
	registrations, err := b.RaceRepo.GetRegisteredDrivers(raceID)
	if err != nil {
		log.Printf("Ошибка получения зарегистрированных гонщиков: %v", err)
		return
	}

	// Send results to all participants
	for _, reg := range registrations {
		// Get driver's Telegram ID
		var telegramID int64
		err := b.db.QueryRow("SELECT telegram_id FROM drivers WHERE id = $1", reg.DriverID).Scan(&telegramID)
		if err != nil {
			log.Printf("Ошибка получения Telegram ID гонщика %d: %v", reg.DriverID, err)
			continue
		}

		// Send message
		b.sendMessage(telegramID, text)
	}
}

// getPlaceEmoji returns emoji for a place
func getPlaceEmoji(place int) string {
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

// getPlaceText returns text for a place
func getPlaceText(place int) string {
	switch place {
	case 0:
		return "не участвовал"
	case 1:
		return "1 место"
	case 2:
		return "2 место"
	case 3:
		return "3 место"
	default:
		return "не участвовал"
	}
}
