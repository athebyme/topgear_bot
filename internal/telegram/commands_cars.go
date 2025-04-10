package telegram

import (
	"fmt"
	"log"
	"strings"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Новые команды для работы с машинами

// registerCarCommandHandlers регистрирует обработчики команд для работы с машинами
func (b *Bot) registerCarCommandHandlers() {
	// Добавляем новые команды в основной обработчик
	b.CommandHandlers["cars"] = b.handleCars
	b.CommandHandlers["carclass"] = b.handleCarClass
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
