package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

// SeasonRepository представляет репозиторий для работы с сезонами
type SeasonRepository struct {
	db *sql.DB
}

// NewSeasonRepository создает новый репозиторий сезонов
func NewSeasonRepository(db *sql.DB) *SeasonRepository {
	return &SeasonRepository{db: db}
}

// Create создает новый сезон
func (r *SeasonRepository) Create(season *models.Season) (int, error) {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	// Отключаем все активные сезоны, если новый сезон активен
	if season.Active {
		_, err = tx.Exec("UPDATE seasons SET active = 0 WHERE active = 1")
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("ошибка деактивации текущих сезонов: %v", err)
		}
	}

	// Форматируем даты для SQLite
	startDate := season.StartDate.Format("2006-01-02")
	var endDate sql.NullString
	if !season.EndDate.IsZero() {
		endDate = sql.NullString{
			String: season.EndDate.Format("2006-01-02"),
			Valid:  true,
		}
	}

	// Вставляем новый сезон
	result, err := tx.Exec(
		"INSERT INTO seasons (name, start_date, end_date, active) VALUES (?, ?, ?, ?)",
		season.Name, startDate, endDate, season.Active,
	)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("ошибка создания сезона: %v", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("ошибка подтверждения транзакции: %v", err)
	}

	// Получаем ID нового сезона
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID сезона: %v", err)
	}

	return int(id), nil
}

// GetByID получает сезон по ID
func (r *SeasonRepository) GetByID(id int) (*models.Season, error) {
	query := `
		SELECT id, name, start_date, end_date, active 
		FROM seasons 
		WHERE id = ?
	`

	var season models.Season
	var startDateStr string
	var endDateStr sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&season.ID,
		&season.Name,
		&startDateStr,
		&endDateStr,
		&season.Active,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Сезон не найден
		}
		return nil, fmt.Errorf("ошибка получения сезона: %v", err)
	}

	// Преобразуем строки в даты
	season.StartDate, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора даты начала: %v", err)
	}

	if endDateStr.Valid {
		season.EndDate, err = time.Parse("2006-01-02", endDateStr.String)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты окончания: %v", err)
		}
	}

	return &season, nil
}

// GetAll возвращает все сезоны
func (r *SeasonRepository) GetAll() ([]*models.Season, error) {
	query := `
		SELECT id, name, start_date, end_date, active 
		FROM seasons 
		ORDER BY start_date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка сезонов: %v", err)
	}
	defer rows.Close()

	var seasons []*models.Season

	for rows.Next() {
		var season models.Season
		var startDateStr string
		var endDateStr sql.NullString

		err := rows.Scan(
			&season.ID,
			&season.Name,
			&startDateStr,
			&endDateStr,
			&season.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных сезона: %v", err)
		}

		// Преобразуем строки в даты
		season.StartDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты начала: %v", err)
		}

		if endDateStr.Valid {
			season.EndDate, err = time.Parse("2006-01-02", endDateStr.String)
			if err != nil {
				return nil, fmt.Errorf("ошибка разбора даты окончания: %v", err)
			}
		}

		seasons = append(seasons, &season)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по сезонам: %v", err)
	}

	return seasons, nil
}

// GetActive возвращает активный сезон
func (r *SeasonRepository) GetActive() (*models.Season, error) {
	query := `
		SELECT id, name, start_date, end_date, active 
		FROM seasons 
		WHERE active = 1 
		LIMIT 1
	`

	var season models.Season
	var startDateStr string
	var endDateStr sql.NullString

	err := r.db.QueryRow(query).Scan(
		&season.ID,
		&season.Name,
		&startDateStr,
		&endDateStr,
		&season.Active,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Активный сезон не найден
		}
		return nil, fmt.Errorf("ошибка получения активного сезона: %v", err)
	}

	// Преобразуем строки в даты
	season.StartDate, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора даты начала: %v", err)
	}

	if endDateStr.Valid {
		season.EndDate, err = time.Parse("2006-01-02", endDateStr.String)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты окончания: %v", err)
		}
	}

	return &season, nil
}

// Update обновляет сезон
func (r *SeasonRepository) Update(season *models.Season) error {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	// Отключаем все активные сезоны, если обновляемый сезон активен
	if season.Active {
		_, err = tx.Exec("UPDATE seasons SET active = 0 WHERE active = 1 AND id != ?", season.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка деактивации текущих сезонов: %v", err)
		}
	}

	// Форматируем даты для SQLite
	startDate := season.StartDate.Format("2006-01-02")
	var endDate sql.NullString
	if !season.EndDate.IsZero() {
		endDate = sql.NullString{
			String: season.EndDate.Format("2006-01-02"),
			Valid:  true,
		}
	}

	// Обновляем сезон
	_, err = tx.Exec(
		"UPDATE seasons SET name = ?, start_date = ?, end_date = ?, active = ? WHERE id = ?",
		season.Name, startDate, endDate, season.Active, season.ID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка обновления сезона: %v", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка подтверждения транзакции: %v", err)
	}

	return nil
}

// Complete завершает сезон с указанной датой окончания
func (r *SeasonRepository) Complete(id int, endDate time.Time) error {
	endDateStr := endDate.Format("2006-01-02")

	_, err := r.db.Exec(
		"UPDATE seasons SET end_date = ?, active = 0 WHERE id = ?",
		endDateStr, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка завершения сезона: %v", err)
	}

	return nil
}

// Delete удаляет сезон
func (r *SeasonRepository) Delete(id int) error {
	// Проверяем, есть ли гонки в этом сезоне
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM races WHERE season_id = ?", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("ошибка проверки гонок сезона: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("нельзя удалить сезон с гонками")
	}

	_, err = r.db.Exec("DELETE FROM seasons WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления сезона: %v", err)
	}

	return nil
}

// Activate активирует сезон и деактивирует все остальные
func (r *SeasonRepository) Activate(id int) error {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	// Деактивируем все сезоны
	_, err = tx.Exec("UPDATE seasons SET active = 0")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка деактивации сезонов: %v", err)
	}

	// Активируем указанный сезон
	_, err = tx.Exec("UPDATE seasons SET active = 1 WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка активации сезона: %v", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка подтверждения транзакции: %v", err)
	}

	return nil
}
