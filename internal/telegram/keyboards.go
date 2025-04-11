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

	// Add in-progress races first with direct buttons to activerace
	if len(inProgressRaces) > 0 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏎️ Текущие гонки:", "no_action"),
		))

		for _, race := range inProgressRaces {
			// Добавляем ряд с двумя кнопками для каждой активной гонки
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🏎️ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"▶️ Перейти",
					fmt.Sprintf("activerace:%d", race.ID),
				),
			))
		}
	}

	// Add upcoming races with registration buttons
	if len(notStartedRaces) > 0 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏳ Предстоящие гонки:", "no_action"),
		))

		for _, race := range notStartedRaces {
			// Добавляем ряд с двумя кнопками для каждой предстоящей гонки
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("⏳ %s", race.Name),
					fmt.Sprintf("race_details:%d", race.ID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Регистрация",
					fmt.Sprintf("register_race:%d", race.ID),
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
				tgbotapi.NewInlineKeyboardButtonData(
					"🏆 Результаты",
					fmt.Sprintf("race_results:%d", race.ID),
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

	// Используем back_to_main для возврата в главное меню
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад в главное меню",
			"back_to_main",
		),
	))

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
