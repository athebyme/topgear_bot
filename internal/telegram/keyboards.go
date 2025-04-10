package telegram

import (
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
)

// MainKeyboard создает основную клавиатуру для главного меню
func MainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 Гонки", "races"),
			tgbotapi.NewInlineKeyboardButtonData("🏆 Сезоны", "seasons"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨‍🏎️ Гонщики", "drivers"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Результаты", "results"),
		),
	)
}

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

// RacesKeyboard создает клавиатуру для просмотра гонок
func RacesKeyboard(races []*models.Race, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, race := range races {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📊 %s", race.Name),
				fmt.Sprintf("race_results:%d", race.ID),
			),
		))
	}

	// Добавляем кнопку создания новой гонки для админов
	if isAdmin {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новую гонку",
				"new_race",
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

// DriverProfileKeyboard создает клавиатуру для профиля гонщика
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
	)
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
