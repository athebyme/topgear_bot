package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config представляет структуру конфигурации приложения
type Config struct {
	Bot struct {
		Token string `yaml:"token"`
		Debug bool   `yaml:"debug"`
	} `yaml:"bot"`

	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Admin struct {
		Users []int64 `yaml:"users"`
	} `yaml:"admin"`

	// Добавлено для работы с Docker
	IsDockerized bool `yaml:"is_dockerized"`
}

// Load загружает конфигурацию из файла
func Load(path string) (*Config, error) {
	// Создаем новый экземпляр конфигурации
	config := &Config{}

	// Читаем содержимое файла
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %v", err)
	}

	// Заменяем переменные окружения в файле конфигурации
	content := string(data)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		content = strings.ReplaceAll(content, "${"+key+"}", value)
	}

	// Разбираем YAML
	if err := yaml.Unmarshal([]byte(content), config); err != nil {
		return nil, fmt.Errorf("ошибка разбора файла конфигурации: %v", err)
	}

	// Проверяем обязательные поля
	if config.Bot.Token == "" {
		// Пробуем получить из переменной окружения
		config.Bot.Token = os.Getenv("TELEGRAM_BOT_TOKEN")
		if config.Bot.Token == "" {
			return nil, fmt.Errorf("не указан токен Telegram бота")
		}
	}

	// Переопределяем настройки базы данных, если мы в Docker
	if config.IsDockerized || os.Getenv("IS_DOCKERIZED") == "true" {
		config.Database.Host = getEnvOrDefault("POSTGRES_HOST", config.Database.Host, "postgres")
		config.Database.Port = getEnvOrDefault("POSTGRES_PORT", config.Database.Port, "5432")
		config.Database.User = getEnvOrDefault("POSTGRES_USER", config.Database.User, "forza")
		config.Database.Password = getEnvOrDefault("POSTGRES_PASSWORD", config.Database.Password, "forza_password")
		config.Database.DBName = getEnvOrDefault("POSTGRES_DB", config.Database.DBName, "forza_db")
	}

	// Проверяем настройки базы данных
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == "" {
		config.Database.Port = "5432"
	}
	if config.Database.User == "" {
		config.Database.User = "forza"
	}
	if config.Database.DBName == "" {
		config.Database.DBName = "forza_db"
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}

	return config, nil
}

// GetDatabaseConnectionString возвращает строку подключения к базе данных
func (c *Config) GetDatabaseConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, configValue, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	if configValue != "" {
		return configValue
	}
	return defaultValue
}
