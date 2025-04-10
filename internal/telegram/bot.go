package telegram

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/config"
	"github.com/athebyme/forza-top-gear-bot/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет экземпляр Telegram бота
type Bot struct {
	API              *tgbotapi.BotAPI
	Config           *config.Config
	StateManager     *UserStateManager
	DriverRepo       *repository.DriverRepository
	SeasonRepo       *repository.SeasonRepository
	RaceRepo         *repository.RaceRepository
	ResultRepo       *repository.ResultRepository
	CarRepo          *repository.CarRepository
	CommandHandlers  map[string]CommandHandler
	CallbackHandlers map[string]CallbackHandler
	AdminIDs         map[int64]bool
	db               *sql.DB
}

// CommandHandler обработчик команд бота
type CommandHandler func(message *tgbotapi.Message)

// CallbackHandler обработчик callback запросов
type CallbackHandler func(query *tgbotapi.CallbackQuery)

// New creates a new bot instance with all handlers registered
func New(cfg *config.Config, db *sql.DB) (*Bot, error) {
	// Initialize bot
	botAPI, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = cfg.Bot.Debug

	// Initialize repositories
	driverRepo := repository.NewDriverRepository(db)
	seasonRepo := repository.NewSeasonRepository(db)
	raceRepo := repository.NewRaceRepository(db)
	resultRepo := repository.NewResultRepository(db)
	carRepo := repository.NewCarRepository(db)

	// Create user state manager
	stateManager := NewUserStateManager()

	// Create admin IDs map for quick lookup
	adminIDs := make(map[int64]bool)
	for _, id := range cfg.Admin.Users {
		adminIDs[id] = true
	}

	// Create bot instance
	bot := &Bot{
		API:              botAPI,
		Config:           cfg,
		StateManager:     stateManager,
		DriverRepo:       driverRepo,
		SeasonRepo:       seasonRepo,
		RaceRepo:         raceRepo,
		ResultRepo:       resultRepo,
		CarRepo:          carRepo,
		CommandHandlers:  make(map[string]CommandHandler),
		CallbackHandlers: make(map[string]CallbackHandler),
		AdminIDs:         adminIDs,
		db:               db,
	}

	bot.registerCommandHandlers()
	bot.registerCarCommandHandlers()
	bot.registerCallbackHandlers()

	bot.registerRaceFlowCommandHandlers()
	bot.registerRaceFlowCallbackHandlers()

	// Проверим, что колбэк для регистрации гонки зарегистрирован
	if _, exists := bot.CallbackHandlers["register_race"]; !exists {
		log.Printf("ВНИМАНИЕ: Обработчик register_race не зарегистрирован!")
		bot.CallbackHandlers["register_race"] = bot.callbackRegisterRace
	}

	// Проверим, что колбэк для лидерборда зарегистрирован
	if _, exists := bot.CallbackHandlers["leaderboard"]; !exists {
		log.Printf("ВНИМАНИЕ: Обработчик leaderboard не зарегистрирован!")
		bot.CallbackHandlers["leaderboard"] = bot.callbackLeaderboard
	}

	return bot, nil
}

// Start launches the bot
func (b *Bot) Start() {
	log.Printf("Bot %s successfully started", b.API.Self.UserName)

	// Configure update receiver
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Get updates channel
	updates := b.API.GetUpdatesChan(u)

	// Process updates
	for update := range updates {
		go b.handleUpdate(update)
	}

	// Start a goroutine to check and notify about upcoming races
	go b.startRaceNotifier()
}

// startRaceNotifier periodically checks for upcoming races and sends reminders
func (b *Bot) startRaceNotifier() {
	ticker := time.NewTicker(30 * time.Minute) // Check every 30 minutes
	defer ticker.Stop()

	for {
		<-ticker.C
		b.checkUpcomingRaces()
	}
}

// checkUpcomingRaces checks for races starting soon and sends reminders
func (b *Bot) checkUpcomingRaces() {
	// Get upcoming races
	upcomingRaces, err := b.RaceRepo.GetUpcomingRaces()
	if err != nil {
		log.Printf("Error getting upcoming races for notifications: %v", err)
		return
	}

	now := time.Now()
	for _, race := range upcomingRaces {
		// Check if race starts within the next 24 hours
		if race.Date.Sub(now) < 24*time.Hour && race.Date.After(now) {
			// Get registered drivers
			registrations, err := b.RaceRepo.GetRegisteredDrivers(race.ID)
			if err != nil {
				log.Printf("Error getting registrations for race %d: %v", race.ID, err)
				continue
			}

			// Notify each registered driver
			for _, reg := range registrations {
				// Get driver's Telegram ID
				var telegramID int64
				err = b.db.QueryRow("SELECT telegram_id FROM drivers WHERE id = $1", reg.DriverID).Scan(&telegramID)
				if err != nil {
					log.Printf("Error getting Telegram ID for driver %d: %v", reg.DriverID, err)
					continue
				}

				// Send reminder
				hoursLeft := int(race.Date.Sub(now).Hours())
				reminderText := ""

				if hoursLeft <= 1 {
					reminderText = fmt.Sprintf("🔔 *Напоминание:* Гонка '%s' начнется менее чем через час!", race.Name)
				} else {
					reminderText = fmt.Sprintf("🔔 *Напоминание:* Гонка '%s' начнется через %d часов!", race.Name, hoursLeft)
				}

				b.sendMessage(telegramID, reminderText)
			}
		}
	}
}

// handleUpdate обрабатывает обновления от Telegram
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	// Определяем тип обновления и вызываем соответствующий обработчик
	if update.Message != nil {
		b.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallbackQuery(update.CallbackQuery)
	}
}

// handleMessage обрабатывает входящие сообщения
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Проверяем, является ли сообщение командой
	if message.IsCommand() {
		command := message.Command()
		if handler, exists := b.CommandHandlers[command]; exists {
			handler(message)
		} else {
			// Обрабатываем неизвестную команду
			b.sendMessage(chatID, "🤔 Неизвестная команда. Используйте /help для получения справки.")
		}

		// Удаляем сообщение с командой для чистоты чата
		b.deleteMessage(chatID, message.MessageID)
		return
	}

	// Проверяем состояние пользователя для обработки диалогов
	if state, exists := b.StateManager.GetState(userID); exists {
		b.handleStateInput(message, state)

		// Удаляем сообщение пользователя для чистоты чата (если это не фото)
		if message.Photo == nil {
			b.deleteMessage(chatID, message.MessageID)
		}
		return
	}

	// Если у пользователя нет состояния и это не команда, игнорируем сообщение
	// или можно ответить подсказкой
	b.sendMessage(chatID, "Используйте /help для получения списка доступных команд.")
}

// IsAdmin проверяет, является ли пользователь администратором
func (b *Bot) IsAdmin(userID int64) bool {
	return b.AdminIDs[userID]
}

// sendMessage отправляет сообщение
func (b *Bot) sendMessage(chatID int64, text string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	message, err := b.API.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}

	return message
}

// sendMessageWithKeyboard отправляет сообщение с клавиатурой
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) tgbotapi.Message {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	message, err := b.API.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения с клавиатурой: %v", err)
	}

	return message
}

// deleteMessage удаляет сообщение
func (b *Bot) deleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := b.API.Request(deleteMsg)
	if err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}
}

// editMessage редактирует сообщение
func (b *Bot) editMessage(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"

	_, err := b.API.Request(edit)
	if err != nil {
		log.Printf("Ошибка редактирования сообщения: %v", err)
	}
}

// editMessageWithKeyboard редактирует сообщение с клавиатурой
func (b *Bot) editMessageWithKeyboard(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard

	_, err := b.API.Request(edit)
	if err != nil {
		log.Printf("Ошибка редактирования сообщения с клавиатурой: %v", err)
	}
}

// sendPhoto отправляет фото с подписью
func (b *Bot) sendPhoto(chatID int64, photoURL, caption string) tgbotapi.Message {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))
	photo.Caption = caption
	photo.ParseMode = "Markdown"

	message, err := b.API.Send(photo)
	if err != nil {
		log.Printf("Ошибка отправки фото: %v", err)
	}

	return message
}

// sendPhotoWithKeyboard отправляет фото с подписью и клавиатурой
func (b *Bot) sendPhotoWithKeyboard(chatID int64, photoURL, caption string, keyboard tgbotapi.InlineKeyboardMarkup) tgbotapi.Message {
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))
	photo.Caption = caption
	photo.ParseMode = "Markdown"
	photo.ReplyMarkup = keyboard

	message, err := b.API.Send(photo)
	if err != nil {
		log.Printf("Ошибка отправки фото с клавиатурой: %v", err)
	}

	return message
}

// editMessageKeyboard редактирует только клавиатуру сообщения
func (b *Bot) editMessageKeyboard(chatID int64, messageID int, keyboard tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)

	_, err := b.API.Request(edit)
	if err != nil {
		log.Printf("Ошибка редактирования клавиатуры: %v", err)
	}
}

// answerCallbackQuery отвечает на callback-запрос
func (b *Bot) answerCallbackQuery(queryID string, text string, showAlert bool) {
	callback := tgbotapi.NewCallback(queryID, text)
	callback.ShowAlert = showAlert

	_, err := b.API.Request(callback)
	if err != nil {
		log.Printf("Ошибка ответа на callback-запрос: %v", err)
	}
}

// formatDate форматирует дату в удобочитаемый вид
func (b *Bot) formatDate(date time.Time) string {
	return date.Format("02.01.2006")
}

// waitAndDelete ожидает указанное время и удаляет сообщение
func (b *Bot) waitAndDelete(chatID int64, messageID int, duration time.Duration) {
	time.Sleep(duration)
	b.deleteMessage(chatID, messageID)
}

// GetBotUsername возвращает имя пользователя бота
func (b *Bot) GetBotUsername() string {
	return b.API.Self.UserName
}
