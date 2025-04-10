package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/config"
	"github.com/athebyme/forza-top-gear-bot/internal/db"
	"github.com/athebyme/forza-top-gear-bot/internal/telegram"
)

func main() {
	log.Println("Starting Forza Top Gear Bot")

	// Парсим аргументы командной строки
	configPath := flag.String("config", "configs/config.yml", "Path to config file")
	flag.Parse()

	// Получаем абсолютный путь к файлу конфигурации
	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatalf("Ошибка получения абсолютного пути: %v", err)
	}

	// Загружаем конфигурацию
	cfg, err := config.Load(absConfigPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Формируем строку подключения к PostgreSQL
	connectionString := cfg.GetDatabaseConnectionString()
	log.Printf("Подключение к базе данных PostgreSQL...")

	// Инициализируем базу данных с повторными попытками
	var database *db.Database
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		database, err = db.New(connectionString)
		if err == nil {
			break
		}
		log.Printf("Попытка %d/%d подключения к базе данных: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Повторная попытка через 5 секунд...")
			time.Sleep(5 * time.Second)
		} else {
			log.Fatalf("Не удалось подключиться к базе данных после %d попыток: %v", maxRetries, err)
		}
	}
	defer database.Close()

	log.Println("Успешное подключение к базе данных PostgreSQL")

	// Применяем миграции
	log.Println("Применение миграций базы данных...")
	err = database.Migrate()
	if err != nil {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}

	// Проверяем, что миграции применены корректно
	err = database.VerifyMigrations()
	if err != nil {
		log.Fatalf("Ошибка проверки миграций: %v", err)
	}
	log.Println("Миграции успешно применены")

	// Инициализируем бота
	log.Println("Инициализация Telegram бота...")
	bot, err := telegram.New(cfg, database.GetDB())
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}

	// Отображаем информацию о запуске
	log.Printf("Бот %s успешно инициализирован", bot.GetBotUsername())
	log.Printf("Режим отладки: %t", cfg.Bot.Debug)
	log.Printf("Количество администраторов: %d", len(cfg.Admin.Users))

	// Запускаем бота в отдельной горутине
	go bot.Start()
	log.Println("Бот запущен и готов к работе")

	// Ждем сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("Нажмите Ctrl+C для завершения работы")
	<-stop

	log.Println("Завершение работы бота...")
}
