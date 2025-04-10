package telegram

import (
	"database/sql"
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

// New создает новый экземпляр бота
func New(cfg *config.Config, db *sql.DB) (*Bot, error) {
	// Инициализируем бота
	botAPI, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = cfg.Bot.Debug

	// Инициализируем репозитории
	driverRepo := repository.NewDriverRepository(db)
	seasonRepo := repository.NewSeasonRepository(db)
	raceRepo := repository.NewRaceRepository(db)
	resultRepo := repository.NewResultRepository(db)
	carRepo := repository.NewCarRepository(db)

	// Создаем менеджер состояний пользователей
	stateManager := NewUserStateManager()

	// Создаем карту ID администраторов для быстрого поиска
	adminIDs := make(map[int64]bool)
	for _, id := range cfg.Admin.Users {
		adminIDs[id] = true
	}

	// Создаем экземпляр бота
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

	// Регистрируем обработчики команд
	bot.registerCommandHandlers()

	// Регистрируем обработчики команд для работы с машинами
	bot.registerCarCommandHandlers()

	// Регистрируем обработчики callback-запросов
	bot.registerCallbackHandlers()

	return bot, nil
}

// Start запускает бота
func (b *Bot) Start() {
	log.Printf("Бот %s успешно запущен", b.API.Self.UserName)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Получаем канал обновлений
	updates := b.API.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		go b.handleUpdate(update)
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
