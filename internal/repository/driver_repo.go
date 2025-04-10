package repository

import (
	"database/sql"
	"fmt"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

// DriverRepository представляет репозиторий для работы с гонщиками
type DriverRepository struct {
	db *sql.DB
}

// NewDriverRepository создает новый репозиторий гонщиков
func NewDriverRepository(db *sql.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

// Create создает нового гонщика
func (r *DriverRepository) Create(driver *models.Driver) (int, error) {
	query := `
		INSERT INTO drivers (telegram_id, name, description, photo_url) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := r.db.QueryRow(query, driver.TelegramID, driver.Name, driver.Description, driver.PhotoURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания гонщика: %v", err)
	}

	return id, nil
}

// GetByID получает гонщика по ID
func (r *DriverRepository) GetByID(id int) (*models.Driver, error) {
	query := `
		SELECT id, telegram_id, name, description, photo_url 
		FROM drivers 
		WHERE id = $1
	`

	var driver models.Driver
	err := r.db.QueryRow(query, id).Scan(
		&driver.ID,
		&driver.TelegramID,
		&driver.Name,
		&driver.Description,
		&driver.PhotoURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Гонщик не найден
		}
		return nil, fmt.Errorf("ошибка получения гонщика: %v", err)
	}

	return &driver, nil
}

// GetByTelegramID получает гонщика по ID пользователя Telegram
func (r *DriverRepository) GetByTelegramID(telegramID int64) (*models.Driver, error) {
	query := `
		SELECT id, telegram_id, name, description, photo_url 
		FROM drivers 
		WHERE telegram_id = $1
	`

	var driver models.Driver
	err := r.db.QueryRow(query, telegramID).Scan(
		&driver.ID,
		&driver.TelegramID,
		&driver.Name,
		&driver.Description,
		&driver.PhotoURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Гонщик не найден
		}
		return nil, fmt.Errorf("ошибка получения гонщика: %v", err)
	}

	return &driver, nil
}

// Update обновляет данные гонщика
func (r *DriverRepository) Update(driver *models.Driver) error {
	query := `
		UPDATE drivers 
		SET name = $1, description = $2, photo_url = $3 
		WHERE id = $4
	`

	_, err := r.db.Exec(query, driver.Name, driver.Description, driver.PhotoURL, driver.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления гонщика: %v", err)
	}

	return nil
}

// UpdateName обновляет имя гонщика
func (r *DriverRepository) UpdateName(id int, name string) error {
	query := `UPDATE drivers SET name = $1 WHERE id = $2`

	_, err := r.db.Exec(query, name, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления имени гонщика: %v", err)
	}

	return nil
}

// UpdateDescription обновляет описание гонщика
func (r *DriverRepository) UpdateDescription(id int, description string) error {
	query := `UPDATE drivers SET description = $1 WHERE id = $2`

	_, err := r.db.Exec(query, description, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления описания гонщика: %v", err)
	}

	return nil
}

// UpdatePhoto обновляет фото гонщика
func (r *DriverRepository) UpdatePhoto(id int, photoURL string) error {
	query := `UPDATE drivers SET photo_url = $1 WHERE id = $2`

	_, err := r.db.Exec(query, photoURL, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления фото гонщика: %v", err)
	}

	return nil
}

// Delete удаляет гонщика
func (r *DriverRepository) Delete(id int) error {
	query := `DELETE FROM drivers WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления гонщика: %v", err)
	}

	return nil
}

// GetAll возвращает всех гонщиков
func (r *DriverRepository) GetAll() ([]*models.Driver, error) {
	query := `
		SELECT id, telegram_id, name, description, photo_url 
		FROM drivers 
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка гонщиков: %v", err)
	}
	defer rows.Close()

	var drivers []*models.Driver

	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID,
			&driver.TelegramID,
			&driver.Name,
			&driver.Description,
			&driver.PhotoURL,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонщика: %v", err)
		}

		drivers = append(drivers, &driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонщикам: %v", err)
	}

	return drivers, nil
}

// GetStats возвращает статистику гонщика
func (r *DriverRepository) GetStats(driverID int) (*models.DriverStats, error) {
	// Получаем общий счет
	var totalScore int
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(total_score), 0) 
		FROM race_results 
		WHERE driver_id = $1
	`, driverID).Scan(&totalScore)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения общего счета: %v", err)
	}

	// Получаем количество гонок
	var totalRaces int
	err = r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM race_results 
		WHERE driver_id = $1
	`, driverID).Scan(&totalRaces)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения количества гонок: %v", err)
	}

	// Получаем последние гонки
	rows, err := r.db.Query(`
		SELECT r.name, rr.total_score 
		FROM race_results rr 
		JOIN races r ON rr.race_id = r.id 
		WHERE rr.driver_id = $1 
		ORDER BY r.date DESC LIMIT 5
	`, driverID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения последних гонок: %v", err)
	}
	defer rows.Close()

	var recentRaces []models.RaceScorePair

	for rows.Next() {
		var raceName string
		var score int
		err := rows.Scan(&raceName, &score)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", err)
		}

		recentRaces = append(recentRaces, models.RaceScorePair{
			RaceName: raceName,
			Score:    score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	// Создаем статистику
	stats := &models.DriverStats{
		TotalScore:  totalScore,
		RecentRaces: recentRaces,
		TotalRaces:  totalRaces,
		// Достижения можно добавить позже
	}

	return stats, nil
}

// GetAllWithStats возвращает всех гонщиков с их статистикой
func (r *DriverRepository) GetAllWithStats() ([]*models.Driver, map[int]*models.DriverStats, error) {
	// Получаем всех гонщиков
	drivers, err := r.GetAll()
	if err != nil {
		return nil, nil, err
	}

	// Получаем статистику для каждого гонщика
	statsMap := make(map[int]*models.DriverStats)

	for _, driver := range drivers {
		stats, err := r.GetStats(driver.ID)
		if err != nil {
			return nil, nil, err
		}

		statsMap[driver.ID] = stats
	}

	return drivers, statsMap, nil
}

// CheckExists проверяет существование гонщика с указанным Telegram ID
func (r *DriverRepository) CheckExists(telegramID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM drivers WHERE telegram_id = $1)", telegramID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки существования гонщика: %v", err)
	}

	return exists, nil
}
