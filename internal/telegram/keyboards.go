package telegram

import (
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SeasonsKeyboard создает клавиатуру для просмотра сезонов
func SeasonsKeyboard(seasons []*models.Season, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, season := range seasons {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🏁 Гонки %s", season.Name),
				fmt.Sprintf("season_races:%d", season.ID),
			),
		))
	}

	// Добавляем кнопку создания нового сезона для админов
	if isAdmin {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новый сезон",
				"new_season",
			),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// DriversKeyboard создает клавиатуру для просмотра гонщиков
func DriversKeyboard(drivers []*models.Driver) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, driver := range drivers {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("👤 %s", driver.Name),
				fmt.Sprintf("driver_card:%d", driver.ID),
			),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// DisciplinesKeyboard создает клавиатуру для выбора дисциплин
func DisciplinesKeyboard(selectedDisciplines []string) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Создаем карту выбранных дисциплин для быстрого поиска
	selected := make(map[string]bool)
	for _, discipline := range selectedDisciplines {
		selected[discipline] = true
	}

	// Добавляем кнопки для всех стандартных дисциплин
	for i, discipline := range models.DefaultDisciplines {
		var buttonText string
		if selected[discipline] {
			buttonText = "✅ " + discipline
		} else {
			buttonText = discipline
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				buttonText,
				fmt.Sprintf("discipline:%d", i),
			),
		))
	}

	// Добавляем кнопку "Готово"
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"✅ Готово",
			"disciplines_done",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// PlacesKeyboard создает клавиатуру для выбора места
func PlacesKeyboard(discipline string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🥇 1 место",
				fmt.Sprintf("place:%s:1", discipline),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🥈 2 место",
				fmt.Sprintf("place:%s:2", discipline),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🥉 3 место",
				fmt.Sprintf("place:%s:3", discipline),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Не участвовал",
				fmt.Sprintf("place:%s:0", discipline),
			),
		),
	)
}

// ConfirmationKeyboard создает клавиатуру для подтверждения действия
func ConfirmationKeyboard(action string, id int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Да",
				fmt.Sprintf("confirm_%s:%d", action, id),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Нет",
				fmt.Sprintf("cancel_%s:%d", action, id),
			),
		),
	)
}

func MainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 Гонки", "races"),
			tgbotapi.NewInlineKeyboardButtonData("📝 Регистрация", "register_driver"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨‍🏎️ Гонщики", "drivers"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Результаты", "results"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚗 Машины", "cars"),
			tgbotapi.NewInlineKeyboardButtonData("🏆 Сезоны", "seasons"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏆 Рейтинг", "leaderboard"),
		),
	)
}

func DriverProfileKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Изменить имя",
				"edit_driver_name",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"📝 Изменить описание",
				"edit_driver_desc",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🖼️ Изменить фото",
				"edit_driver_photo",
			),
		),
		// Добавлена кнопка "Назад в главное меню"
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад в главное меню",
				"back_to_main",
			),
		),
	)
}

// AdminRacePanelKeyboard создает клавиатуру для админ-панели гонки
func AdminRacePanelKeyboard(raceID int, state string) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	switch state {
	case models.RaceStateNotStarted:
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏁 Запустить гонку",
				fmt.Sprintf("start_race:%d", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📨 Отправить напоминание",
				fmt.Sprintf("admin_send_notifications:%d:reminder", raceID),
			),
		))

	case models.RaceStateInProgress:
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Управление результатами",
				fmt.Sprintf("admin_edit_results_menu:%d", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"👨‍🏎️ Управление участниками",
				fmt.Sprintf("race_registrations:%d", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📨 Отправить машины",
				fmt.Sprintf("admin_send_notifications:%d:cars", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Завершить гонку",
				fmt.Sprintf("complete_race:%d", raceID),
			),
		))

	case models.RaceStateCompleted:
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📨 Отправить результаты",
				fmt.Sprintf("admin_send_notifications:%d:results", raceID),
			),
		))
	}

	// Общие кнопки для всех статусов
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_details:%d", raceID),
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// ImprovedRacesKeyboard создает улучшенную клавиатуру для гонок
func ImprovedRacesKeyboard(races []*models.Race, userID int64, b *Bot) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Получаем данные о водителе для проверки регистрации
	var driver *models.Driver
	if driverObj, err := b.DriverRepo.GetByTelegramID(userID); err == nil {
		driver = driverObj
	}

	// Группируем гонки по статусу
	var activeRaces []*models.Race
	var upcomingRaces []*models.Race
	var completedRaces []*models.Race

	for _, race := range races {
		switch race.State {
		case models.RaceStateInProgress:
			activeRaces = append(activeRaces, race)
		case models.RaceStateNotStarted:
			upcomingRaces = append(upcomingRaces, race)
		case models.RaceStateCompleted:
			completedRaces = append(completedRaces, race)
		}
	}

	// Секция активных гонок (приоритет)
	if len(activeRaces) > 0 {
		// Заголовок секции
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏎️ АКТИВНЫЕ ГОНКИ",
				"no_action", // Это просто заголовок, без действия
			),
		))

		for _, race := range activeRaces {
			var isRegistered bool
			if driver != nil {
				if registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID); err == nil {
					isRegistered = registered
				}
			}

			// Добавляем две кнопки для каждой активной гонки в одном ряду
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🏎️ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
			}

			// Добавляем разные кнопки действий в зависимости от регистрации
			if isRegistered {
				row = append(row,
					tgbotapi.NewInlineKeyboardButtonData(
						"Моя машина",
						fmt.Sprintf("my_car:%d", race.ID),
					),
				)
			} else {
				row = append(row,
					tgbotapi.NewInlineKeyboardButtonData(
						"Статус",
						fmt.Sprintf("race_progress:%d", race.ID),
					),
				)
			}

			keyboard = append(keyboard, row)
		}
	}

	// Секция предстоящих гонок
	if len(upcomingRaces) > 0 {

		for _, race := range upcomingRaces {
			var isRegistered bool
			if driver != nil {
				if registered, err := b.RaceRepo.CheckDriverRegistered(race.ID, driver.ID); err == nil {
					isRegistered = registered
				}
			}

			// Создаем название кнопки с или без индикатора регистрации
			var buttonText string
			if isRegistered {
				buttonText = fmt.Sprintf("⏳ %s ✅", race.Name)
			} else {
				buttonText = fmt.Sprintf("⏳ %s", race.Name)
			}

			// Создаем ряд с двумя кнопками для каждой предстоящей гонки
			row := []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					buttonText,
					fmt.Sprintf("race_details:%d", race.ID),
				),
			}

			// Добавляем кнопку регистрации или отмены регистрации
			if isRegistered {
				row = append(row,
					tgbotapi.NewInlineKeyboardButtonData(
						"Отменить",
						fmt.Sprintf("unregister_race:%d", race.ID),
					),
				)
			} else {
				row = append(row,
					tgbotapi.NewInlineKeyboardButtonData(
						"Регистрация",
						fmt.Sprintf("register_race:%d", race.ID),
					),
				)
			}

			keyboard = append(keyboard, row)
		}
	}

	// Секция завершенных гонок
	if len(completedRaces) > 0 {
		// Добавляем пустую строку для разделения, если были другие гонки
		if len(activeRaces) > 0 || len(upcomingRaces) > 0 {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯",
					"no_action",
				),
			))
		}

		// Заголовок секции
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ ЗАВЕРШЕННЫЕ ГОНКИ",
				"no_action",
			),
		))

		for _, race := range completedRaces {
			// Создаем ряд с двумя кнопками для каждой завершенной гонки
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✅ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"Результаты",
					fmt.Sprintf("race_results:%d", race.ID),
				),
			))
		}
	}

	// Дополнительные кнопки
	// Добавляем разделитель
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯",
			"no_action",
		),
	))

	// Кнопка создания новой гонки для админов
	if b.IsAdmin(userID) {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новую гонку",
				"new_race",
			),
		))
	}

	// Кнопка возврата в главное меню
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Главное меню",
			"back_to_main",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}
