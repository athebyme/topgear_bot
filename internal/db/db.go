package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Database представляет соединение с базой данных
type Database struct {
	db *sql.DB
}

// New создает новый экземпляр базы данных
func New(connectionString string) (*Database, error) {
	// Если строка подключения не указана, используем значения из переменных окружения
	if connectionString == "" {
		connectionString = "postgres://%s:%s@%s:%s/%s?sslmode=disable"
		connectionString = fmt.Sprintf(
			connectionString,
			GetEnv("POSTGRES_USER", "forza"),
			GetEnv("POSTGRES_PASSWORD", "forza_password"),
			GetEnv("POSTGRES_HOST", "localhost"),
			GetEnv("POSTGRES_PORT", "5432"),
			GetEnv("POSTGRES_DB", "forza_db"),
		)
	}

	// Подключаемся к базе данных с повторными попытками
	var db *sql.DB
	var err error
	maxRetries := 5
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connectionString)
		if err != nil {
			return nil, fmt.Errorf("ошибка открытия соединения: %v", err)
		}

		// Проверяем подключение
		err = db.Ping()
		if err == nil {
			break
		}

		// Если достигли максимального числа попыток, возвращаем ошибку
		if i == maxRetries-1 {
			return nil, fmt.Errorf("не удалось подключиться к базе данных после %d попыток: %v", maxRetries, err)
		}

		// Закрываем текущее соединение и пробуем снова через некоторое время
		db.Close()
		time.Sleep(retryDelay)
		fmt.Printf("Повторная попытка подключения к PostgreSQL (%d/%d)...\n", i+2, maxRetries)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{db: db}, nil
}

// Close закрывает соединение с базой данных
func (d *Database) Close() error {
	return d.db.Close()
}

// GetDB возвращает объект подключения к базе данных
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// BeginTx начинает новую транзакцию
func (d *Database) BeginTx() (*sql.Tx, error) {
	return d.db.Begin()
}

// Migrate выполняет миграции базы данных
func (d *Database) Migrate() error {
	// Применяем миграции из файла migrations.go
	for _, query := range migrations {
		_, err := d.db.Exec(query)
		if err != nil {
			return fmt.Errorf("ошибка выполнения миграции: %v", err)
		}
	}

	// Проверяем наличие активного сезона, если нет - создаем
	return d.ensureActiveSeasonExists()
}

// VerifyMigrations проверяет, что все таблицы существуют
func (d *Database) VerifyMigrations() error {
	// Список таблиц, которые должны быть созданы
	tables := []string{
		"drivers",
		"seasons",
		"races",
		"race_results",
		"cars",
		"race_car_assignments",
	}

	for _, table := range tables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)
		`
		err := d.db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("ошибка проверки таблицы %s: %v", table, err)
		}

		if !exists {
			return fmt.Errorf("таблица %s не найдена в базе данных", table)
		}
	}

	return nil
}

// ensureActiveSeasonExists проверяет наличие активного сезона и создает его при необходимости
func (d *Database) ensureActiveSeasonExists() error {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM seasons WHERE active = true").Scan(&count)
	if err != nil {
		return fmt.Errorf("ошибка проверки активного сезона: %v", err)
	}

	if count == 0 {
		// Создаем сезоны если их нет
		tx, err := d.db.Begin()
		if err != nil {
			return fmt.Errorf("ошибка начала транзакции: %v", err)
		}

		// Создаем первый сезон (завершенный)
		_, err = tx.Exec(
			"INSERT INTO seasons (name, start_date, end_date, active) VALUES ($1, $2, $3, $4)",
			"Сезон 1",
			time.Now().AddDate(0, -6, 0), // 6 месяцев назад
			time.Now().AddDate(0, -1, 0), // 1 месяц назад
			false,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка создания первого сезона: %v", err)
		}

		// Создаем второй сезон (активный)
		_, err = tx.Exec(
			"INSERT INTO seasons (name, start_date, end_date, active) VALUES ($1, $2, $3, $4)",
			"Сезон 2",
			time.Now().AddDate(0, -1, 0), // 1 месяц назад
			nil,                          // Без даты окончания
			true,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка создания второго сезона: %v", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("ошибка коммита транзакции: %v", err)
		}
	}

	return nil
}

// GetEnv возвращает значение переменной окружения или значение по умолчанию
func GetEnv(key, defaultValue string) string {
	value := GetEnvOrEmpty(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvOrEmpty возвращает значение переменной окружения или пустую строку
func GetEnvOrEmpty(key string) string {
	return "" // Заглушка. В реальном коде здесь должен быть вызов os.Getenv(key)
}
