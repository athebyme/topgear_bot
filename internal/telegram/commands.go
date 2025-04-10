package telegram

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// registerCommandHandlers registers all command handlers
func (b *Bot) registerCommandHandlers() {
	b.CommandHandlers = map[string]CommandHandler{
		"start":       b.handleStart,
		"register":    b.handleRegister,
		"driver":      b.handleDriver,
		"seasons":     b.handleSeasons,
		"races":       b.handleRaces,
		"newrace":     b.handleNewRace,
		"results":     b.handleResults,
		"help":        b.handleHelp,
		"addresult":   b.handleAddResult,
		"cancel":      b.handleCancel,
		"joinrace":    b.handleJoinRace,
		"leaverage":   b.handleLeaveRace,
		"mycar":       b.handleMyCar,
		"stats":       b.handleStats,
		"leaderboard": b.handleLeaderboard,
		// Новые команды
		"activerace":  b.handleActiveRace,  // Информация о текущей активной гонке
		"racestatus":  b.handleRaceStatus,  // Подробный статус гонки и участников
		"adminrace":   b.handleAdminRace,   // Админ-панель для управления гонкой
		"editresult":  b.handleEditResult,  // Редактирование результатов (для админов)
		"racedetails": b.handleRaceDetails, // Детали конкретной гонки
	}
}

func (b *Bot) handleLeaderboard(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Show overall leaderboard by default
	b.showLeaderboard(chatID, 0) // 0 means "all seasons"
}

func (b *Bot) showLeaderboard(chatID int64, seasonID int) {
	// Get all drivers
	drivers, err := b.DriverRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения списка гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонщиков.")
		return
	}

	if len(drivers) == 0 {
		b.sendMessage(chatID, "⚠️ Пока нет зарегистрированных гонщиков.")
		return
	}

	// Get all seasons for the keyboard
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения списка сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	// Prepare data structures
	type driverStats struct {
		ID           int
		Name         string
		TotalScore   int
		Races        int
		Wins         int
		SecondPlaces int
		ThirdPlaces  int
		BestRally    string
	}

	var stats []driverStats

	// For each driver, calculate statistics
	for _, driver := range drivers {
		// Get all results for this driver
		results, err := b.ResultRepo.GetByDriverID(driver.ID)
		if err != nil {
			log.Printf("Ошибка получения результатов гонщика %d: %v", driver.ID, err)
			continue
		}

		// Skip if no results
		if len(results) == 0 {
			continue
		}

		// Initialize driver stats
		ds := driverStats{
			ID:   driver.ID,
			Name: driver.Name,
		}

		// Analyze results
		for _, result := range results {
			// Skip if filtering by season and this result is not from that season
			if seasonID > 0 {
				race, err := b.RaceRepo.GetByID(result.RaceID)
				if err != nil || race == nil || race.SeasonID != seasonID {
					continue
				}
			}

			// Accumulate stats
			ds.TotalScore += result.TotalScore
			ds.Races++

			// Count places in disciplines
			for _, place := range result.Results {
				switch place {
				case 1:
					ds.Wins++
				case 2:
					ds.SecondPlaces++
				case 3:
					ds.ThirdPlaces++
				}
			}

			// Check rally discipline if there is one
			race, err := b.RaceRepo.GetByID(result.RaceID)
			if err == nil && race != nil {
				for _, discipline := range race.Disciplines {
					// Check if this is a rally discipline and the driver participated
					if strings.Contains(strings.ToLower(discipline), "ралли") {
						place, exists := result.Results[discipline]
						if exists && place > 0 {
							// In a real implementation, you would get the actual rally time
							rallyTime := "2:34.567" // Example time

							// If no best rally yet, or this one is better
							if ds.BestRally == "" || rallyTime < ds.BestRally {
								ds.BestRally = rallyTime
							}
						}
					}
				}
			}
		}

		// Add to stats array if driver has races after filtering
		if ds.Races > 0 {
			stats = append(stats, ds)
		}
	}

	// Sort by total score (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TotalScore > stats[j].TotalScore
	})

	// Format the message
	var title string
	if seasonID == 0 {
		title = "🏆 *Общий рейтинг гонщиков*"
	} else {
		// Get season name
		season, err := b.SeasonRepo.GetByID(seasonID)
		if err != nil || season == nil {
			title = "🏆 *Рейтинг гонщиков (выбранный сезон)*"
		} else {
			title = fmt.Sprintf("🏆 *Рейтинг гонщиков %s*", season.Name)
		}
	}

	text := title + "\n\n"

	if len(stats) == 0 {
		text += "Нет данных для отображения."
	} else {
		// Add header row
		text += "# | Гонщик | Очки | Гонки | 🥇 | 🥈 | 🥉\n"
		text += "---|--------|------|-------|---|---|---\n"

		// Add driver rows
		for i, s := range stats {
			text += fmt.Sprintf("%d | *%s* | %d | %d | %d | %d | %d\n",
				i+1, s.Name, s.TotalScore, s.Races, s.Wins, s.SecondPlaces, s.ThirdPlaces)
		}

		// Add rally records section if any
		var rallyRecords []string
		for _, s := range stats {
			if s.BestRally != "" {
				rallyRecords = append(rallyRecords, fmt.Sprintf("• *%s*: %s", s.Name, s.BestRally))
			}
		}

		if len(rallyRecords) > 0 {
			text += "\n*Лучшие времена в Ралли:*\n"
			text += strings.Join(rallyRecords, "\n")
		}
	}

	// Create keyboard for season selection
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// All-time button
	allTimeText := "📊 Все сезоны"
	if seasonID == 0 {
		allTimeText = "✅ " + allTimeText
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			allTimeText,
			"leaderboard:0",
		),
	))

	// Season buttons
	for _, season := range seasons {
		seasonText := season.Name
		if season.ID == seasonID {
			seasonText = "✅ " + seasonText
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				seasonText,
				fmt.Sprintf("leaderboard:%d", season.ID),
			),
		))
	}

	// Back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			"back_to_main",
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// Add callback handler for leaderboard
func (b *Bot) callbackLeaderboard(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	// Parse season ID from callback data
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный формат запроса", true)
		return
	}

	seasonID, err := strconv.Atoi(parts[1])
	if err != nil {
		b.answerCallbackQuery(query.ID, "⚠️ Неверный ID сезона", true)
		return
	}

	// Delete the original message
	b.deleteMessage(chatID, query.Message.MessageID)

	// Show leaderboard for selected season
	b.showLeaderboard(chatID, seasonID)
}

// handleStats shows overall driver statistics with filters
func (b *Bot) handleStats(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Get all drivers
	drivers, err := b.DriverRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения списка гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонщиков.")
		return
	}

	if len(drivers) == 0 {
		b.sendMessage(chatID, "⚠️ Пока нет зарегистрированных гонщиков.")
		return
	}

	// Get all seasons
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения списка сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	// Get active season
	activeSeason, err := b.SeasonRepo.GetActive()
	if err != nil {
		log.Printf("Ошибка получения активного сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении активного сезона.")
		return
	}

	// Default to active season or first season in list
	var defaultSeasonID int
	if activeSeason != nil {
		defaultSeasonID = activeSeason.ID
	} else if len(seasons) > 0 {
		defaultSeasonID = seasons[0].ID
	}

	// Show driver statistics for the default season
	b.showDriverStats(chatID, defaultSeasonID)
}

// showDriverStats displays driver statistics for a specific season
func (b *Bot) showDriverStats(chatID int64, seasonID int) {
	// Get season information
	season, err := b.SeasonRepo.GetByID(seasonID)
	if err != nil {
		log.Printf("Ошибка получения информации о сезоне: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о сезоне.")
		return
	}

	if season == nil {
		b.sendMessage(chatID, "⚠️ Сезон не найден.")
		return
	}

	// Get all seasons for the keyboard
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения списка сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	// Get races for this season
	races, err := b.RaceRepo.GetBySeason(seasonID)
	if err != nil {
		log.Printf("Ошибка получения гонок сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении гонок сезона.")
		return
	}

	// Get all completed races
	var completedRaces []*models.Race
	for _, race := range races {
		if race.State == models.RaceStateCompleted {
			completedRaces = append(completedRaces, race)
		}
	}

	// Get all drivers
	drivers, err := b.DriverRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения списка гонщиков: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка гонщиков.")
		return
	}

	// Prepare driver stats map
	driverStats := make(map[int]map[int]int) // driverID -> raceID -> score
	driverTotalScores := make(map[int]int)   // driverID -> total score

	// Get rally records for each driver
	driverRallyRecords := make(map[int]map[int]string) // driverID -> raceID -> rally time

	// For each completed race, get results
	for _, race := range completedRaces {
		results, err := b.ResultRepo.GetRaceResultsWithRerollPenalty(race.ID)
		if err != nil {
			log.Printf("Ошибка получения результатов гонки %d: %v", race.ID, err)
			continue
		}

		for _, result := range results {
			// Initialize driver's map if needed
			if _, exists := driverStats[result.DriverID]; !exists {
				driverStats[result.DriverID] = make(map[int]int)
			}

			// Store score for this race
			driverStats[result.DriverID][race.ID] = result.TotalScore

			// Add to total score
			driverTotalScores[result.DriverID] += result.TotalScore

			// Check for rally disciplines
			for discipline, place := range result.Results {
				if strings.Contains(strings.ToLower(discipline), "ралли") && place > 0 {
					if _, exists := driverRallyRecords[result.DriverID]; !exists {
						driverRallyRecords[result.DriverID] = make(map[int]string)
					}

					// Store rally time (this is a placeholder - in a real implementation,
					// you would get the actual time from the results)
					driverRallyRecords[result.DriverID][race.ID] = "2:34.567" // Example time
				}
			}
		}
	}

	// Sort drivers by total score (descending)
	type driverScore struct {
		Driver *models.Driver
		Score  int
	}

	var rankedDrivers []driverScore
	for _, driver := range drivers {
		rankedDrivers = append(rankedDrivers, driverScore{
			Driver: driver,
			Score:  driverTotalScores[driver.ID],
		})
	}

	// Sort by score descending
	sort.Slice(rankedDrivers, func(i, j int) bool {
		return rankedDrivers[i].Score > rankedDrivers[j].Score
	})

	// Format the message
	text := fmt.Sprintf("📊 *Статистика гонщиков %s*\n\n", season.Name)

	if len(completedRaces) == 0 {
		text += "Нет завершенных гонок в этом сезоне."
	} else {
		// Table header
		text += "🏎️ | "
		for _, race := range completedRaces {
			text += fmt.Sprintf("%s | ", race.Name[:3]) // First 3 chars of race name
		}
		text += "Всего\n"
		text += strings.Repeat("-", 50) + "\n"

		// Driver rows
		for i, ds := range rankedDrivers {
			driver := ds.Driver
			text += fmt.Sprintf("%d. *%s* | ", i+1, driver.Name)

			// Scores for each race
			for _, race := range completedRaces {
				score, exists := driverStats[driver.ID][race.ID]
				if exists {
					text += fmt.Sprintf("%d | ", score)
				} else {
					text += "- | "
				}
			}

			// Total score
			text += fmt.Sprintf("*%d*\n", driverTotalScores[driver.ID])
		}

		// Best rally times
		text += "\n*Лучшие времена в Ралли:*\n"
		for _, ds := range rankedDrivers {
			driver := ds.Driver
			if records, exists := driverRallyRecords[driver.ID]; exists && len(records) > 0 {
				// Find best time
				var bestRaceID int
				var bestTime string
				for raceID, time := range records {
					if bestTime == "" || time < bestTime {
						bestTime = time
						bestRaceID = raceID
					}
				}

				// Find race name
				var raceName string
				for _, race := range races {
					if race.ID == bestRaceID {
						raceName = race.Name
						break
					}
				}

				text += fmt.Sprintf("• *%s*: %s (%s)\n", driver.Name, bestTime, raceName)
			}
		}
	}

	// Build keyboard for season selection
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Seasons row
	var seasonsRow []tgbotapi.InlineKeyboardButton
	for _, s := range seasons {
		// Add visual indicator for selected season
		name := s.Name
		if s.ID == seasonID {
			name = "✅ " + name
		}

		seasonsRow = append(seasonsRow, tgbotapi.NewInlineKeyboardButtonData(
			name,
			fmt.Sprintf("stats_season:%d", s.ID),
		))

		// Maximum 3 buttons per row
		if len(seasonsRow) == 3 {
			keyboard = append(keyboard, seasonsRow)
			seasonsRow = nil
		}
	}

	// Add remaining season buttons
	if len(seasonsRow) > 0 {
		keyboard = append(keyboard, seasonsRow)
	}

	// Add back button
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад",
			"back_to_main",
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// handleStart provides main menu and starting point
func (b *Bot) handleStart(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is already registered
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
	}

	// Create keyboard with main commands
	keyboard := MainKeyboard()

	var messageText string
	if driver != nil {
		messageText = fmt.Sprintf("🏎️ *Top Gear Racing Club* 🏎️\n\nПривет, %s! Выбери действие:", driver.Name)
	} else {
		messageText = "🏎️ *Top Gear Racing Club* 🏎️\n\nДобро пожаловать! Для начала зарегистрируйтесь как гонщик с помощью команды /register."
	}

	b.sendMessageWithKeyboard(chatID, messageText, keyboard)
}

// handleDriver shows driver profile with enhanced statistics
func (b *Bot) handleDriver(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Get driver data
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении вашей карточки гонщика.")
		return
	}

	if driver == nil {
		// FIXED: Changed from "/start" to "/register"
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Get basic driver stats
	stats, err := b.DriverRepo.GetStats(driver.ID)
	if err != nil {
		log.Printf("Ошибка получения статистики гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении вашей статистики.")
		return
	}

	// Get all race results for this driver to find rally time record
	results, err := b.ResultRepo.GetByDriverID(driver.ID)
	if err != nil {
		log.Printf("Ошибка получения результатов гонщика: %v", err)
		// Continue anyway, this is not critical
	}

	// Find best rally time
	type rallyRecord struct {
		Time     string
		RaceName string
		SeasonID int
	}

	var bestRally *rallyRecord

	// Look through all results for rally disciplines
	for _, result := range results {
		// Get race to find disciplines and season
		race, err := b.RaceRepo.GetByID(result.RaceID)
		if err != nil {
			log.Printf("Ошибка получения гонки %d: %v", result.RaceID, err)
			continue
		}

		if race == nil {
			continue
		}

		// Look for rally discipline in this race
		for _, discipline := range race.Disciplines {
			if strings.Contains(strings.ToLower(discipline), "ралли") {
				// Check if driver has a result for this discipline
				place, exists := result.Results[discipline]
				if exists && place > 0 {
					// In a real implementation, you would get the time from the result
					// This is a placeholder
					rallyTime := "2:34.567" // Example time

					// If this is the first or better record, save it
					if bestRally == nil || rallyTime < bestRally.Time {
						bestRally = &rallyRecord{
							Time:     rallyTime,
							RaceName: race.Name,
							SeasonID: race.SeasonID,
						}
					}
				}
			}
		}
	}

	// Get season name for the rally record
	var rallySeasonName string
	if bestRally != nil {
		season, err := b.SeasonRepo.GetByID(bestRally.SeasonID)
		if err == nil && season != nil {
			rallySeasonName = season.Name
		}
	}

	// Format driver profile
	text := fmt.Sprintf("👨‍🏎️ *Карточка гонщика*\n\n*%s*\n", driver.Name)

	if driver.Description != "" {
		text += fmt.Sprintf("📋 *Описание:* %s\n\n", driver.Description)
	}

	text += fmt.Sprintf("🏆 *Всего очков:* %d\n", stats.TotalScore)
	text += fmt.Sprintf("🏁 *Гонок:* %d\n", stats.TotalRaces)

	// Add rally record if available
	if bestRally != nil {
		text += fmt.Sprintf("⏱️ *Рекорд в Ралли:* %s (%s, %s)\n", bestRally.Time, bestRally.RaceName, rallySeasonName)
	}

	text += "\n"

	if len(stats.RecentRaces) > 0 {
		text += "*Последние гонки:*\n"
		for _, race := range stats.RecentRaces {
			text += fmt.Sprintf("• %s: %d очков\n", race.RaceName, race.Score)
		}
	} else {
		text += "*Пока нет завершенных гонок*"
	}

	// Клавиатура для редактирования профиля
	keyboard := DriverProfileKeyboard()

	if driver.PhotoURL != "" {
		// Отправляем фото с подписью
		b.sendPhotoWithKeyboard(chatID, driver.PhotoURL, text, keyboard)
	} else {
		// Отправляем только текст
		b.sendMessageWithKeyboard(chatID, text, keyboard)
	}
}

// handleSeasons обрабатывает команду /seasons
func (b *Bot) handleSeasons(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Получаем все сезоны
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	if len(seasons) == 0 {
		b.sendMessage(chatID, "🏁 Сезоны еще не созданы.")
		return
	}

	// Отображаем список сезонов с кнопками для просмотра деталей
	text := "🏆 *Сезоны Top Gear Racing Club*\n\n"

	for _, season := range seasons {
		var status string
		if season.Active {
			status = "🟢 Активный"
		} else {
			status = "🔴 Завершен"
		}

		dateRange := fmt.Sprintf("%s", b.formatDate(season.StartDate))
		if !season.EndDate.IsZero() {
			dateRange += fmt.Sprintf(" - %s", b.formatDate(season.EndDate))
		} else {
			dateRange += " - н.в."
		}

		text += fmt.Sprintf("*%s* (%s)\n%s\n\n", season.Name, status, dateRange)
	}

	// Создаем клавиатуру с сезонами
	keyboard := SeasonsKeyboard(seasons, b.IsAdmin(userID))

	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleNewRace обрабатывает команду /newrace
func (b *Bot) handleNewRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав для создания новых гонок")
		return
	}

	// Получаем активный сезон
	activeSeason, err := b.SeasonRepo.GetActive()
	if err != nil {
		log.Printf("Ошибка получения активного сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении активного сезона.")
		return
	}

	if activeSeason == nil {
		b.sendMessage(chatID, "⚠️ Не найден активный сезон. Создайте сезон перед добавлением гонки.")
		return
	}

	// Создаем состояние для пользователя
	b.StateManager.SetState(userID, "new_race_name", map[string]interface{}{
		"season_id": activeSeason.ID,
	})

	b.sendMessage(chatID, fmt.Sprintf("🏁 Создание новой гонки для *%s*\n\nВведите название гонки:", activeSeason.Name))
}

// handleResults обрабатывает команду /results
func (b *Bot) handleResults(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Получаем все сезоны
	seasons, err := b.SeasonRepo.GetAll()
	if err != nil {
		log.Printf("Ошибка получения сезонов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении списка сезонов.")
		return
	}

	text := "📊 *Результаты гонок*\n\nВыберите сезон для просмотра результатов:"

	if len(seasons) == 0 {
		text += "\n\nПока нет созданных сезонов."
		b.sendMessage(chatID, text)
		return
	}

	// Создаем клавиатуру с сезонами
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, season := range seasons {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				season.Name,
				fmt.Sprintf("season_results:%d", season.ID),
			),
		))
	}

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// handleHelp provides documentation for all commands
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	isAdmin := b.IsAdmin(message.From.ID)

	text := `🏎️ *Top Gear Racing Club Бот* 🏎️

*Основные команды:*

/start - Главное меню
/register - Регистрация нового гонщика
/driver - Просмотр карточки гонщика
/seasons - Просмотр сезонов
/races - Просмотр гонок текущего сезона
/results - Просмотр результатов гонок
/leaderboard - Рейтинг гонщиков
/stats - Детальная статистика гонщиков
/help - Эта справка
/cancel - Отмена текущего действия

*Команды для гонок:*
/activerace - Информация о текущей активной гонке
/racestatus - Подробный статус текущей гонки
/joinrace - Регистрация на предстоящую гонку
/leaverage - Отмена регистрации на гонку
/mycar - Просмотр назначенной машины для гонки
/addresult - Добавить свой результат в гонке
/racedetails [ID] - Подробная информация о гонке

*Просмотр машин:*
/cars - Просмотр всех машин, доступных в игре
/carclass [класс] - Просмотр машин определенного класса`

	// Добавляем админские команды, если пользователь - администратор
	if isAdmin {
		text += `

*Команды администратора:*
/adminrace - Панель управления текущей гонкой
/editresult [ID] - Редактирование результатов участников
/newrace - Создание новой гонки`
	}

	text += `

*Как проходит гонка:*
1. Гонщик регистрируется на предстоящую гонку через /joinrace
2. Администратор запускает гонку, и всем участникам выдаются случайные машины
3. Гонщик может принять машину или использовать реролл (штраф -1 очко)
4. Гонщики проводят заезды в каждой дисциплине и вводят свои результаты
5. Администратор завершает гонку и публикует результаты

*Система подсчета очков:*
🥇 1 место - 3 очка
🥈 2 место - 2 очка
🥉 3 место - 1 очко
⚠️ Реролл машины - штраф -1 очко`

	// Create helpful keyboard for main commands
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 Гонки", "races"),
			tgbotapi.NewInlineKeyboardButtonData("👨‍🏎️ Мой профиль", "driver_command"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚗 Машины", "cars"),
			tgbotapi.NewInlineKeyboardButtonData("🏆 Рейтинг", "leaderboard"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Текущая гонка", "activerace"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Главное меню", "back_to_main"),
		),
	)

	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleCancel обрабатывает команду /cancel
func (b *Bot) handleCancel(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	if b.StateManager.HasState(userID) {
		b.StateManager.ClearState(userID)
		b.sendMessage(chatID, "🚫 Текущее действие отменено.")
	} else {
		b.sendMessage(chatID, "🤔 Нет активных действий для отмены.")
	}
}

func (b *Bot) callbackDriverCommand(query *tgbotapi.CallbackQuery) {
	message := tgbotapi.Message{
		From: query.From,
		Chat: query.Message.Chat,
	}

	b.handleDriver(&message)

	b.deleteMessage(query.Message.Chat.ID, query.Message.MessageID)
}

// handleStateInput routes input to appropriate handler based on state
func (b *Bot) handleStateInput(message *tgbotapi.Message, state models.UserState) {
	switch state.State {
	case "register_name":
		b.handleRegisterName(message, state)
	case "register_description":
		b.handleRegisterDescription(message, state)
	case "register_photo":
		b.handleRegisterPhoto(message, state)
	case "new_race_name":
		b.handleNewRaceName(message, state)
	case "new_race_date":
		b.handleNewRaceDate(message, state)
	case "new_race_car_class":
		b.handleNewRaceCarClass(message, state)
	case "edit_driver_name":
		b.handleEditDriverName(message, state)
	case "edit_driver_description":
		b.handleEditDriverDescription(message, state)
	case "edit_driver_photo":
		b.handleEditDriverPhoto(message, state)
	case "add_result_car_number":
		b.handleResultCarNumber(message, state)
	case "add_result_car_name":
		b.handleResultCarName(message, state)
	case "add_result_car_photo":
		b.handleResultCarPhoto(message, state)
	case "add_result_discipline":
		b.handleResultDiscipline(message, state)
	case "new_season_name":
		b.handleNewSeasonName(message, state)
	case "new_season_start_date":
		b.handleNewSeasonStartDate(message, state)
	default:
		b.sendMessage(message.Chat.ID, "⚠️ Неизвестное состояние. Используйте /cancel для отмены текущего действия.")
	}
}

// Обработчики состояний пользователя

// handleRegisterDescription обрабатывает ввод описания при регистрации
func (b *Bot) handleRegisterDescription(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	description := message.Text
	if description == "-" {
		description = ""
	}

	// Сохраняем описание в контексте и запрашиваем фото
	b.StateManager.SetState(userID, "register_photo", map[string]interface{}{
		"name":        state.ContextData["name"],
		"description": description,
	})

	b.sendMessage(chatID, "Отлично! Теперь отправьте фото для вашей карточки гонщика (или отправьте '-' чтобы пропустить):")
}

// handleRegisterPhoto обрабатывает отправку фото при регистрации
func (b *Bot) handleRegisterPhoto(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	var photoURL string

	if message.Text == "-" {
		photoURL = ""
	} else if message.Photo != nil && len(message.Photo) > 0 {
		// Получаем ID фото для сохранения
		photo := message.Photo[len(message.Photo)-1]
		photoURL = photo.FileID
	} else {
		b.sendMessage(chatID, "⚠️ Пожалуйста, отправьте фото или '-' для пропуска.")
		return
	}

	// Создаем нового гонщика
	driver := &models.Driver{
		TelegramID:  userID,
		Name:        state.ContextData["name"].(string),
		Description: state.ContextData["description"].(string),
		PhotoURL:    photoURL,
	}

	// Сохраняем в БД
	_, err := b.DriverRepo.Create(driver)
	if err != nil {
		log.Printf("Ошибка сохранения гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при регистрации. Пожалуйста, попробуйте еще раз.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Регистрация успешно завершена! Используйте /driver чтобы увидеть свою карточку гонщика.")

	// Показываем главное меню
	b.handleStart(message)
}

// handleNewRaceName обрабатывает ввод названия гонки
func (b *Bot) handleNewRaceName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность названия
	name := strings.TrimSpace(message.Text)
	if len(name) < 3 || len(name) > 50 {
		b.sendMessage(chatID, "⚠️ Название должно содержать от 3 до 50 символов. Пожалуйста, введите корректное название:")
		return
	}

	// Сохраняем название в контексте и запрашиваем дату
	b.StateManager.SetState(userID, "new_race_date", map[string]interface{}{
		"season_id": state.ContextData["season_id"],
		"name":      name,
	})

	b.sendMessage(chatID, "Введите дату гонки в формате ДД.ММ.ГГГГ:")
}

// handleNewRaceDate обрабатывает ввод даты гонки
func (b *Bot) handleNewRaceDate(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем формат даты
	dateStr := message.Text
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный формат даты. Используйте формат ДД.ММ.ГГГГ (например, 15.04.2025):")
		return
	}

	// Сохраняем дату в контексте и запрашиваем класс автомобилей
	b.StateManager.SetState(userID, "new_race_car_class", map[string]interface{}{
		"season_id": state.ContextData["season_id"],
		"name":      state.ContextData["name"],
		"date":      date.Format("2006-01-02"),
	})

	b.sendMessage(chatID, "Введите класс автомобилей для гонки:")
}

// handleNewRaceCarClass обрабатывает ввод класса автомобилей
func (b *Bot) handleNewRaceCarClass(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность класса
	carClass := strings.TrimSpace(message.Text)
	if len(carClass) < 1 || len(carClass) > 30 {
		b.sendMessage(chatID, "⚠️ Класс автомобилей должен содержать от 1 до 30 символов. Пожалуйста, введите корректный класс:")
		return
	}

	// Сохраняем класс в контексте и предлагаем выбрать дисциплины
	b.StateManager.SetState(userID, "new_race_disciplines", map[string]interface{}{
		"season_id":   state.ContextData["season_id"],
		"name":        state.ContextData["name"],
		"date":        state.ContextData["date"],
		"car_class":   carClass,
		"disciplines": []string{}, // Пустой массив для выбранных дисциплин
	})

	// Создаем клавиатуру для выбора дисциплин
	keyboard := DisciplinesKeyboard([]string{})

	b.sendMessageWithKeyboard(chatID, "Выберите дисциплины для гонки (можно выбрать несколько):", keyboard)
}

// handleEditDriverName обрабатывает изменение имени гонщика
func (b *Bot) handleEditDriverName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность имени
	name := strings.TrimSpace(message.Text)
	if len(name) < 2 || len(name) > 30 {
		b.sendMessage(chatID, "⚠️ Имя должно содержать от 2 до 30 символов. Пожалуйста, введите корректное имя:")
		return
	}

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
		b.StateManager.ClearState(userID)
		return
	}

	// Обновляем имя в БД
	err = b.DriverRepo.UpdateName(driver.ID, name)
	if err != nil {
		log.Printf("Ошибка обновления имени гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении имени.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Имя успешно обновлено!")

	// Перезагружаем карточку гонщика
	b.handleDriver(message)
}

// handleEditDriverDescription обрабатывает изменение описания гонщика
func (b *Bot) handleEditDriverDescription(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	description := message.Text

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
		b.StateManager.ClearState(userID)
		return
	}

	// Обновляем описание в БД
	err = b.DriverRepo.UpdateDescription(driver.ID, description)
	if err != nil {
		log.Printf("Ошибка обновления описания гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении описания.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Описание успешно обновлено!")

	// Перезагружаем карточку гонщика
	b.handleDriver(message)
}

// handleEditDriverPhoto обрабатывает изменение фото гонщика
func (b *Bot) handleEditDriverPhoto(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, что отправлено фото
	if message.Photo == nil || len(message.Photo) == 0 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, отправьте фото.")
		return
	}

	// Получаем ID фото
	photo := message.Photo[len(message.Photo)-1]
	photoURL := photo.FileID

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
		b.StateManager.ClearState(userID)
		return
	}

	// Обновляем фото в БД
	err = b.DriverRepo.UpdatePhoto(driver.ID, photoURL)
	if err != nil {
		log.Printf("Ошибка обновления фото гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при обновлении фото.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Фото успешно обновлено!")

	// Перезагружаем карточку гонщика
	b.handleDriver(message)
}

// handleAddResultCarName обрабатывает ввод названия машины
func (b *Bot) handleAddResultCarName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность названия
	carName := strings.TrimSpace(message.Text)
	if len(carName) < 2 || len(carName) > 50 {
		b.sendMessage(chatID, "⚠️ Название машины должно содержать от 2 до 50 символов. Пожалуйста, введите корректное название:")
		return
	}

	// Сохраняем название машины и запрашиваем фото
	b.StateManager.SetState(userID, "add_result_car_photo", map[string]interface{}{
		"race_id":    state.ContextData["race_id"],
		"car_number": state.ContextData["car_number"],
		"car_name":   carName,
	})

	b.sendMessage(chatID, "Отправьте фото вашей машины (или '-' чтобы пропустить):")
}

// handleAddResultCarPhoto обрабатывает отправку фото машины
func (b *Bot) handleAddResultCarPhoto(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	var photoURL string

	if message.Text == "-" {
		photoURL = ""
	} else if message.Photo != nil && len(message.Photo) > 0 {
		// Получаем ID фото для сохранения
		photo := message.Photo[len(message.Photo)-1]
		photoURL = photo.FileID
	} else {
		b.sendMessage(chatID, "⚠️ Пожалуйста, отправьте фото или '-' для пропуска.")
		return
	}

	// Получаем гонку для определения дисциплин
	raceID := state.ContextData["race_id"].(int)
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонки.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка не найдена.")
		b.StateManager.ClearState(userID)
		return
	}

	// Сохраняем данные и переходим к вводу результатов первой дисциплины
	b.StateManager.SetState(userID, "add_result_discipline", map[string]interface{}{
		"race_id":     raceID,
		"car_number":  state.ContextData["car_number"],
		"car_name":    state.ContextData["car_name"],
		"car_photo":   photoURL,
		"disciplines": race.Disciplines,
		"current_idx": 0,
		"results":     make(map[string]int),
	})

	// Запрашиваем результат первой дисциплины
	b.sendMessage(
		chatID,
		fmt.Sprintf("Введите ваше место в дисциплине '%s' (1-3 или 0 если не участвовали):",
			race.Disciplines[0]),
	)
}

// handleAddResultDiscipline обрабатывает ввод результата дисциплины
func (b *Bot) handleAddResultDiscipline(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, что введено корректное число
	place, err := strconv.Atoi(message.Text)
	if err != nil || place < 0 || place > 3 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, введите число от 0 до 3 (0 - не участвовал, 1-3 - место).")
		return
	}

	// Получаем текущие данные
	disciplines := state.ContextData["disciplines"].([]string)
	currentIdx := state.ContextData["current_idx"].(int)
	results := state.ContextData["results"].(map[string]int)

	// Сохраняем результат текущей дисциплины
	currentDiscipline := disciplines[currentIdx]
	results[currentDiscipline] = place

	// Переходим к следующей дисциплине или завершаем
	currentIdx++

	if currentIdx < len(disciplines) {
		// Еще есть дисциплины
		b.StateManager.SetState(userID, "add_result_discipline", map[string]interface{}{
			"race_id":     state.ContextData["race_id"],
			"car_number":  state.ContextData["car_number"],
			"car_name":    state.ContextData["car_name"],
			"car_photo":   state.ContextData["car_photo"],
			"disciplines": disciplines,
			"current_idx": currentIdx,
			"results":     results,
		})

		// Запрашиваем результат следующей дисциплины
		b.sendMessage(
			chatID,
			fmt.Sprintf("Введите ваше место в дисциплине '%s' (1-3 или 0 если не участвовали):",
				disciplines[currentIdx]),
		)
	} else {
		// Все дисциплины заполнены, сохраняем результат
		// Получаем ID гонщика
		driver, err := b.DriverRepo.GetByTelegramID(userID)
		if err != nil {
			log.Printf("Ошибка получения гонщика: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
			return
		}

		if driver == nil {
			b.sendMessage(chatID, "⚠️ Гонщик не найден. Используйте /register для регистрации.")
			b.StateManager.ClearState(userID)
			return
		}

		// Вычисляем общий счет
		totalScore := 0
		for _, place := range results {
			switch place {
			case 1:
				totalScore += 3
			case 2:
				totalScore += 2
			case 3:
				totalScore += 1
			}
		}

		// Создаем результат гонки
		result := &models.RaceResult{
			RaceID:      state.ContextData["race_id"].(int),
			DriverID:    driver.ID,
			CarNumber:   state.ContextData["car_number"].(int),
			CarName:     state.ContextData["car_name"].(string),
			CarPhotoURL: state.ContextData["car_photo"].(string),
			Results:     results,
			TotalScore:  totalScore,
		}

		// Сохраняем результат в БД
		_, err = b.ResultRepo.Create(result)
		if err != nil {
			log.Printf("Ошибка сохранения результата: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при сохранении результатов.")
			return
		}

		// Очищаем состояние
		b.StateManager.ClearState(userID)

		b.sendMessage(
			chatID,
			fmt.Sprintf("✅ Результаты успешно сохранены! Вы набрали %d очков в этой гонке.", totalScore),
		)

		// Показываем результаты гонки
		b.showRaceResults(chatID, result.RaceID)
	}
}

// handleNewSeasonName обрабатывает ввод названия сезона
func (b *Bot) handleNewSeasonName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем валидность названия
	name := strings.TrimSpace(message.Text)
	if len(name) < 3 || len(name) > 30 {
		b.sendMessage(chatID, "⚠️ Название должно содержать от 3 до 30 символов. Пожалуйста, введите корректное название:")
		return
	}

	// Сохраняем название в контексте и запрашиваем дату начала
	b.StateManager.SetState(userID, "new_season_start_date", map[string]interface{}{
		"name": name,
	})

	b.sendMessage(chatID, "Введите дату начала сезона в формате ДД.ММ.ГГГГ:")
}

// handleNewSeasonStartDate обрабатывает ввод даты начала сезона
func (b *Bot) handleNewSeasonStartDate(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем формат даты
	dateStr := message.Text
	startDate, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		b.sendMessage(chatID, "⚠️ Неверный формат даты. Используйте формат ДД.ММ.ГГГГ (например, 15.04.2025):")
		return
	}

	// Создаем новый сезон
	season := &models.Season{
		Name:      state.ContextData["name"].(string),
		StartDate: startDate,
		Active:    true, // Новый сезон сразу активен
	}

	// Сохраняем сезон в БД
	_, err = b.SeasonRepo.Create(season)
	if err != nil {
		log.Printf("Ошибка создания сезона: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при создании нового сезона.")
		return
	}

	// Очищаем состояние
	b.StateManager.ClearState(userID)

	b.sendMessage(chatID, "✅ Новый сезон успешно создан и активирован!")

	// Показываем список сезонов
	b.handleSeasons(message)
}

// Updated handleRegisterName to properly handle driver registration
func (b *Bot) handleRegisterName(message *tgbotapi.Message, state models.UserState) {
	userID := message.From.ID
	chatID := message.Chat.ID

	log.Printf("Processing driver name for user ID: %d", userID)

	// Check name validity
	name := strings.TrimSpace(message.Text)
	log.Printf("Provided name: '%s', length: %d", name, len(name))

	if len(name) < 2 || len(name) > 30 {
		b.sendMessage(chatID, "⚠️ Имя должно содержать от 2 до 30 символов. Пожалуйста, введите корректное имя:")
		return
	}

	// Save name in context and request description
	log.Printf("Setting state to register_description with name: %s", name)
	b.StateManager.SetState(userID, "register_description", map[string]interface{}{
		"name": name,
	})

	b.sendMessage(chatID, fmt.Sprintf("Отлично, %s! Теперь введите краткое описание о себе как о гонщике (или отправьте '-' чтобы пропустить):", name))
}

// handleActiveRace показывает информацию о текущей активной гонке
func (b *Bot) handleActiveRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Получаем данные гонщика
	driver, err := b.DriverRepo.GetByTelegramID(userID)
	if err != nil {
		log.Printf("Ошибка получения данных гонщика: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении данных гонщика.")
		return
	}

	if driver == nil {
		b.sendMessage(chatID, "⚠️ Вы не зарегистрированы как гонщик. Используйте /register чтобы зарегистрироваться.")
		return
	}

	// Получаем активную гонку
	activeRace, err := b.RaceRepo.GetActiveRace()
	if err != nil {
		log.Printf("Ошибка получения активной гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации об активной гонке.")
		return
	}

	if activeRace == nil {
		b.sendMessage(chatID, "ℹ️ В данный момент нет активных гонок.")
		return
	}

	// Проверяем, зарегистрирован ли пользователь на эту гонку
	registered, err := b.RaceRepo.CheckDriverRegistered(activeRace.ID, driver.ID)
	if err != nil {
		log.Printf("Ошибка проверки регистрации: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при проверке вашей регистрации.")
		return
	}

	// Получаем данные о машине гонщика (если зарегистрирован)
	var carInfo string
	if registered {
		assignment, err := b.CarRepo.GetDriverCarAssignment(activeRace.ID, driver.ID)
		if err != nil {
			log.Printf("Ошибка получения назначения машины: %v", err)
		} else if assignment != nil {
			car := assignment.Car
			carInfo = fmt.Sprintf("\n\n*Ваша машина:*\n🚗 %s (%s)\n🔢 Номер: %d",
				car.Name, car.Year, assignment.AssignmentNumber)

			// Проверяем статус подтверждения машины
			var confirmed bool
			err = b.db.QueryRow(
				"SELECT car_confirmed FROM race_registrations WHERE race_id = $1 AND driver_id = $2",
				activeRace.ID, driver.ID,
			).Scan(&confirmed)

			if err == nil {
				if confirmed {
					carInfo += "\n✅ Машина подтверждена"
				} else {
					carInfo += "\n⚠️ Машина не подтверждена. Используйте /mycar чтобы подтвердить"
				}
			}
		}
	}

	// Формируем сообщение о текущей гонке
	text := fmt.Sprintf("🏁 *Активная гонка: %s*\n\n", activeRace.Name)
	text += fmt.Sprintf("📅 Дата: %s\n", b.formatDate(activeRace.Date))
	text += fmt.Sprintf("🚗 Класс: %s\n", activeRace.CarClass)
	text += fmt.Sprintf("🏎️ Дисциплины: %s\n", strings.Join(activeRace.Disciplines, ", "))
	text += fmt.Sprintf("🏆 Статус: %s\n", getStatusText(activeRace.State))

	if registered {
		text += "\n✅ Вы зарегистрированы на эту гонку" + carInfo
	} else {
		text += "\n❌ Вы не зарегистрированы на эту гонку"
	}

	// Создаем клавиатуру с действиями для гонки
	keyboard := ActiveRaceKeyboard(activeRace.ID, registered, activeRace.State, b.IsAdmin(userID))
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// getStatusText возвращает текстовое описание статуса гонки
func getStatusText(state string) string {
	switch state {
	case models.RaceStateNotStarted:
		return "Регистрация участников"
	case models.RaceStateInProgress:
		return "Гонка идет"
	case models.RaceStateCompleted:
		return "Гонка завершена"
	default:
		return "Неизвестно"
	}
}

// ActiveRaceKeyboard создает клавиатуру для активной гонки
func ActiveRaceKeyboard(raceID int, registered bool, state string, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Кнопки для обычных пользователей
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
	}

	if state == models.RaceStateInProgress && registered {
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
			"📊 Статус участников",
			fmt.Sprintf("race_progress:%d", raceID),
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"🚗 Машины участников",
			fmt.Sprintf("view_race_cars:%d", raceID),
		),
	))

	// Админские кнопки
	if isAdmin {
		if state == models.RaceStateNotStarted {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"🏁 Запустить гонку",
					fmt.Sprintf("start_race:%d", raceID),
				),
			))
		} else if state == models.RaceStateInProgress {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Завершить гонку",
					fmt.Sprintf("complete_race:%d", raceID),
				),
			))
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⚙️ Управление гонкой",
				fmt.Sprintf("admin_race_panel:%d", raceID),
			),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(keyboard...)
}

// handleRaceStatus показывает подробный ход гонки
func (b *Bot) handleRaceStatus(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Получаем активную гонку
	activeRace, err := b.RaceRepo.GetActiveRace()
	if err != nil {
		log.Printf("Ошибка получения активной гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации об активной гонке.")
		return
	}

	if activeRace == nil {
		b.sendMessage(chatID, "ℹ️ В данный момент нет активных гонок.")
		return
	}

	// Показываем ход гонки
	b.showRaceProgress(chatID, activeRace.ID)
}

// handleAdminRace предоставляет админ-панель для управления гонкой
func (b *Bot) handleAdminRace(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав администратора для выполнения этой команды.")
		return
	}

	// Получаем активную гонку
	activeRace, err := b.RaceRepo.GetActiveRace()
	if err != nil {
		log.Printf("Ошибка получения активной гонки: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации об активной гонке.")
		return
	}

	if activeRace == nil {
		// Если нет активной гонки, показываем список ближайших для возможности запуска
		upcomingRaces, err := b.RaceRepo.GetUpcomingRaces()
		if err != nil {
			log.Printf("Ошибка получения предстоящих гонок: %v", err)
			b.sendMessage(chatID, "⚠️ Произошла ошибка при получении предстоящих гонок.")
			return
		}

		if len(upcomingRaces) == 0 {
			b.sendMessage(chatID, "ℹ️ Нет активных или предстоящих гонок.\n\nИспользуйте /newrace чтобы создать новую гонку.")
			return
		}

		// Формируем сообщение со списком предстоящих гонок
		text := "⚙️ *Админ-панель гонок*\n\nАктивная гонка отсутствует. Выберите гонку для запуска:\n\n"

		// Создаем клавиатуру для выбора гонки
		var keyboard [][]tgbotapi.InlineKeyboardButton

		for _, race := range upcomingRaces {
			text += fmt.Sprintf("• *%s* (📅 %s)\n", race.Name, b.formatDate(race.Date))

			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🏁 Запустить: %s", race.Name),
					fmt.Sprintf("start_race:%d", race.ID),
				),
			))
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Создать новую гонку",
				"new_race",
			),
		))

		b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
		return
	}

	// Если есть активная гонка, показываем админ-панель
	b.showAdminRacePanel(chatID, activeRace.ID)
}

// showAdminRacePanel показывает панель администратора для конкретной гонки
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

	// Create keyboard for race management
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Buttons based on race state
	switch race.State {
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
		// If there are any unconfirmed cars, show button to force confirmation
		if len(unconfirmedDrivers) > 0 {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Подтвердить все машины",
					fmt.Sprintf("admin_confirm_all_cars:%d", raceID),
				),
			))

			// Add individual confirm buttons for each unconfirmed driver
			for _, driverID := range unconfirmedDrivers {
				// Get driver name
				var driverName string
				err := b.db.QueryRow("SELECT name FROM drivers WHERE id = $1", driverID).Scan(&driverName)
				if err != nil {
					continue
				}

				keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						fmt.Sprintf("✅ Подтвердить машину: %s", driverName),
						fmt.Sprintf("admin_force_confirm_car:%d:%d", raceID, driverID),
					),
				))
			}
		}

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📊 Просмотр машин",
				fmt.Sprintf("view_race_cars:%d", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать результаты",
				fmt.Sprintf("admin_edit_results_menu:%d", raceID),
			),
		))

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📨 Отправить уведомления о машинах",
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

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏆 Просмотр результатов",
				fmt.Sprintf("race_results:%d", raceID),
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

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// handleEditResult позволяет администратору редактировать результаты гонщиков
func (b *Bot) handleEditResult(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь администратором
	if !b.IsAdmin(userID) {
		b.sendMessage(chatID, "⛔ У вас нет прав администратора для выполнения этой команды.")
		return
	}

	// Получаем аргументы команды
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "⚠️ Пожалуйста, укажите ID гонки. Пример: /editresult 42")
		return
	}

	// Парсим ID гонки
	raceID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Некорректный ID гонки. Пожалуйста, укажите число.")
		return
	}

	// Получаем информацию о гонке
	race, err := b.RaceRepo.GetByID(raceID)
	if err != nil {
		log.Printf("Ошибка получения информации о гонке: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении информации о гонке.")
		return
	}

	if race == nil {
		b.sendMessage(chatID, "⚠️ Гонка с указанным ID не найдена.")
		return
	}

	// Получаем результаты гонки
	results, err := b.ResultRepo.GetRaceResultsWithDriverNames(raceID)
	if err != nil {
		log.Printf("Ошибка получения результатов: %v", err)
		b.sendMessage(chatID, "⚠️ Произошла ошибка при получении результатов гонки.")
		return
	}

	if len(results) == 0 {
		b.sendMessage(chatID, "ℹ️ Для этой гонки еще нет результатов.")
		return
	}

	// Формируем сообщение с результатами для редактирования
	text := fmt.Sprintf("✏️ *Редактирование результатов гонки: %s*\n\n", race.Name)
	text += "Выберите гонщика для редактирования результатов:\n\n"

	// Создаем клавиатуру для выбора гонщика
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, result := range results {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s - %d очков", result.DriverName, result.TotalScore),
				fmt.Sprintf("admin_edit_result:%d", result.ID),
			),
		))
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"🔙 Назад к гонке",
			fmt.Sprintf("race_results:%d", raceID),
		),
	))

	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(keyboard...))
}

// handleRaceDetails показывает детальную информацию о гонке
func (b *Bot) handleRaceDetails(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	// Получаем аргументы команды
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Если ID не указан, пытаемся показать активную гонку
		b.handleActiveRace(message)
		return
	}

	// Парсим ID гонки
	raceID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "⚠️ Некорректный ID гонки. Пожалуйста, укажите число.")
		return
	}

	// Показываем информацию о гонке
	b.showRaceDetails(chatID, raceID, userID)
}
