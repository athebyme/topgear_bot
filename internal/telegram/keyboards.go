package telegram

import (
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
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

// RaceResultsKeyboard создает клавиатуру для просмотра результатов гонки
func RaceResultsKeyboard(raceID int, completed bool, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Добавляем кнопку для завершения гонки, если она не завершена и пользователь - админ
	if !completed && isAdmin {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Завершить гонку",
				fmt.Sprintf("complete_race:%d", raceID),
			),
		))
	}

	// Добавляем кнопку для добавления результата
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"➕ Добавить свой результат",
			fmt.Sprintf("add_result:%d", raceID),
		),
	))

	// Добавляем кнопку для просмотра назначенных машин
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🚗 Посмотреть машины",
			fmt.Sprintf("view_race_cars:%d", raceID),
		),
	))

	// Для админов добавляем кнопку редактирования гонки
	if isAdmin {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать гонку",
				fmt.Sprintf("edit_race:%d", raceID),
			),
		))

		// Добавляем кнопку удаления гонки
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🗑️ Удалить гонку",
				fmt.Sprintf("delete_race:%d", raceID),
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

// IncompleteRacesKeyboard создает клавиатуру для выбора незавершенной гонки
func IncompleteRacesKeyboard(races []*models.Race) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, race := range races {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				race.Name,
				fmt.Sprintf("add_result:%d", race.ID),
			),
		))
	}

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

// NumberKeyboard создает цифровую клавиатуру для выбора числа
func NumberKeyboard(prefix string, min, max int) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	// Создаем кнопки с цифрами
	for i := min; i <= max; i++ {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			strconv.Itoa(i),
			fmt.Sprintf("%s:%d", prefix, i),
		))

		// Максимум 5 кнопок в ряду
		if len(row) == 5 || i == max {
			keyboard = append(keyboard, row)
			row = nil
		}
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// BackToMainKeyboard создает клавиатуру для возврата в главное меню
func BackToMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад в главное меню",
				"back_to_main",
			),
		),
	)
}

// BackKeyboard создает клавиатуру для возврата назад
func BackKeyboard(action string, id int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад",
				fmt.Sprintf("%s:%d", action, id),
			),
		),
	)
}

// CancelKeyboard создает клавиатуру для отмены действия
func CancelKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Отмена",
				"cancel",
			),
		),
	)
}

// RaceStateKeyboard creates a keyboard for changing race state
func RaceStateKeyboard(raceID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏁 Запустить гонку",
				fmt.Sprintf("start_race:%d", raceID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✅ Завершить гонку",
				fmt.Sprintf("complete_race:%d", raceID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Отмена",
				"cancel",
			),
		),
	)
}

// Updated RacesKeyboard to use race_details callback
func RacesKeyboard(races []*models.Race, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Group races by state
	var notStartedRaces []*models.Race
	var inProgressRaces []*models.Race
	var completedRaces []*models.Race

	for _, race := range races {
		switch race.State {
		case models.RaceStateNotStarted:
			notStartedRaces = append(notStartedRaces, race)
		case models.RaceStateInProgress:
			inProgressRaces = append(inProgressRaces, race)
		case models.RaceStateCompleted:
			completedRaces = append(completedRaces, race)
		}
	}

	// Add in-progress races first
	if len(inProgressRaces) > 0 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏎️ Текущие гонки:", "no_action"),
		))

		for _, race := range inProgressRaces {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🏎️ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
			))
		}
	}

	// Add upcoming races
	if len(notStartedRaces) > 0 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏳ Предстоящие гонки:", "no_action"),
		))

		for _, race := range notStartedRaces {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("⏳ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
			))
		}
	}

	// Add completed races
	if len(completedRaces) > 0 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Завершенные гонки:", "no_action"),
		))

		for _, race := range completedRaces {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("✅ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
			))
		}
	}

	// Add create race button for admins
	if isAdmin {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новую гонку",
				"new_race",
			),
		))
	}

	// Изменено: Используем back_to_main вместо seasons для возврата в главное меню
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад в главное меню",
			"back_to_main",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// RaceRegistrationsKeyboard creates a keyboard for managing race registrations
func RaceRegistrationsKeyboard(raceID int, registrations []*models.RaceRegistration) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add buttons for each driver
	for _, reg := range registrations {
		var status string
		if reg.CarConfirmed {
			status = "✅"
		} else {
			status = "⏳"
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", status, reg.DriverName),
				fmt.Sprintf("driver_registration:%d:%d", raceID, reg.DriverID),
			),
		))
	}

	// Add race management buttons
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🏁 Запустить гонку",
			fmt.Sprintf("start_race:%d", raceID),
		),
	))

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_details:%d", raceID),
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// RaceCarConfirmationKeyboard creates a keyboard for car confirmation options
func RaceCarConfirmationKeyboard(raceID int, rerollAvailable bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Add confirm button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"✅ Подтвердить выбор машины",
			fmt.Sprintf("confirm_car:%d", raceID),
		),
	))

	// Add reroll button if available
	if rerollAvailable {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🎲 Реролл (-1 балл)",
				fmt.Sprintf("reroll_car:%d", raceID),
			),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
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

// DriverRaceOptionsKeyboard creates a keyboard for driver options in a race
func DriverRaceOptionsKeyboard(raceID int, registered bool, state string) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Options depend on race state and registration status
	if state == models.RaceStateNotStarted {
		if registered {
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
	} else if state == models.RaceStateInProgress && registered {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Моя машина",
				fmt.Sprintf("my_car:%d", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Добавить результат",
				fmt.Sprintf("add_result:%d", raceID),
			),
		))
	}

	// Common buttons for all states
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🚗 Посмотреть машины",
			fmt.Sprintf("view_race_cars:%d", raceID),
		),
	))

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"📊 Результаты",
			fmt.Sprintf("race_results:%d", raceID),
		),
	))

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			"races",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
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

// RaceDetailsKeyboard создает клавиатуру для просмотра деталей гонки
func RaceDetailsKeyboard(raceID int, userID int64, registered bool, race *models.Race, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	switch race.State {
	case models.RaceStateNotStarted:
		if registered {
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

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"👨‍🏎️ Участники",
				fmt.Sprintf("race_registrations:%d", raceID),
			),
		))

	case models.RaceStateInProgress:
		if registered {
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
				"📊 Прогресс гонки",
				fmt.Sprintf("race_progress:%d", raceID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🚗 Машины участников",
				fmt.Sprintf("view_race_cars:%d", raceID),
			),
		))

	case models.RaceStateCompleted:
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

	// Кнопки для администраторов
	if isAdmin {
		if race.State == models.RaceStateNotStarted {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🏁 Запустить гонку",
					fmt.Sprintf("start_race:%d", raceID),
				),
			))
		} else if race.State == models.RaceStateInProgress {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Завершить гонку",
					fmt.Sprintf("complete_race:%d", raceID),
				),
			))
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⚙️ Панель администратора",
				fmt.Sprintf("admin_race_panel:%d", raceID),
			),
		))
	}

	// Кнопка назад
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к списку гонок",
			"races",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// RaceProgressKeyboard создает клавиатуру для просмотра прогресса гонки
func RaceProgressKeyboard(raceID int, registered bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Действия для зарегистрированных участников
	if registered {
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

	// Общие кнопки
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🚗 Машины участников",
			fmt.Sprintf("view_race_cars:%d", raceID),
		),
	))

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_details:%d", raceID),
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// EditResultKeyboard создает клавиатуру для редактирования результата
func EditResultKeyboard(resultID int, raceID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать места",
				fmt.Sprintf("admin_edit_result_places:%d", resultID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🎲 Переключить реролл",
				fmt.Sprintf("admin_toggle_reroll:%d", resultID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔙 Назад к списку",
				fmt.Sprintf("admin_edit_results_menu:%d", raceID),
			),
		),
	)
}
