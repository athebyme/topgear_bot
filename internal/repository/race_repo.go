package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

// RaceRepository представляет репозиторий для работы с гонками
type RaceRepository struct {
	db *sql.DB
}

// NewRaceRepository создает новый репозиторий гонок
func NewRaceRepository(db *sql.DB) *RaceRepository {
	return &RaceRepository{db: db}
}

// Create создает новую гонку
func (r *RaceRepository) Create(race *models.Race) (int, error) {
	// Сериализуем дисциплины в JSON
	disciplinesJSON, err := models.SerializeDisciplines(race.Disciplines)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации дисциплин: %v", err)
	}

	// Форматируем дату для SQLite
	dateStr := race.Date.Format("2006-01-02")

	// Вставляем новую гонку
	result, err := r.db.Exec(
		"INSERT INTO races (season_id, name, date, car_class, disciplines, completed) VALUES (?, ?, ?, ?, ?, ?)",
		race.SeasonID, race.Name, dateStr, race.CarClass, disciplinesJSON, race.Completed,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания гонки: %v", err)
	}

	// Получаем ID новой гонки
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID гонки: %v", err)
	}

	return int(id), nil
}

// GetByID получает гонку по ID
func (r *RaceRepository) GetByID(id int) (*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed 
		FROM races 
		WHERE id = ?
	`

	var race models.Race
	var dateStr string
	var disciplinesJSON string

	err := r.db.QueryRow(query, id).Scan(
		&race.ID,
		&race.SeasonID,
		&race.Name,
		&dateStr,
		&race.CarClass,
		&disciplinesJSON,
		&race.Completed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Гонка не найдена
		}
		return nil, fmt.Errorf("ошибка получения гонки: %v", err)
	}

	// Преобразуем строку в дату
	race.Date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора даты: %v", err)
	}

	// Десериализуем дисциплины из JSON
	race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации дисциплин: %v", err)
	}

	return &race, nil
}

// GetBySeason получает все гонки указанного сезона
func (r *RaceRepository) GetBySeason(seasonID int) ([]*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed 
		FROM races 
		WHERE season_id = ? 
		ORDER BY date DESC
	`

	rows, err := r.db.Query(query, seasonID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения гонок сезона: %v", err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var dateStr string
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&dateStr,
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", err)
		}

		// Преобразуем строку в дату
		race.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты: %v", err)
		}

		// Десериализуем дисциплины из JSON
		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации дисциплин: %v", err)
		}

		races = append(races, &race)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	return races, nil
}

// GetIncompleteRaces получает все незавершенные гонки
func (r *RaceRepository) GetIncompleteRaces() ([]*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed 
		FROM races 
		WHERE completed = 0 
		ORDER BY date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения незавершенных гонок: %v", err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var dateStr string
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&dateStr,
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", err)
		}

		// Преобразуем строку в дату
		race.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты: %v", err)
		}

		// Десериализуем дисциплины из JSON
		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации дисциплин: %v", err)
		}

		races = append(races, &race)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	return races, nil
}

// Update обновляет гонку
func (r *RaceRepository) Update(tx *sql.Tx, race *models.Race) error {
	// Сериализуем дисциплины в JSON
	disciplinesJSON, err := json.Marshal(race.Disciplines)
	if err != nil {
		return fmt.Errorf("ошибка сериализации дисциплин: %v", err)
	}

	// Проверяем, используем ли мы транзакцию
	if tx != nil {
		// Обновляем гонку в рамках транзакции
		_, err = tx.Exec(
			`UPDATE races 
			 SET season_id = $1, name = $2, date = $3, car_class = $4, disciplines = $5, completed = $6 
			 WHERE id = $7`,
			race.SeasonID, race.Name, race.Date, race.CarClass, disciplinesJSON, race.Completed, race.ID,
		)
	} else {
		// Обновляем гонку без транзакции
		_, err = r.db.Exec(
			`UPDATE races 
			 SET season_id = $1, name = $2, date = $3, car_class = $4, disciplines = $5, completed = $6 
			 WHERE id = $7`,
			race.SeasonID, race.Name, race.Date, race.CarClass, disciplinesJSON, race.Completed, race.ID,
		)
	}

	if err != nil {
		return fmt.Errorf("ошибка обновления гонки: %v", err)
	}

	return nil
}

// UpdateCompleted изменяет статус завершенности гонки
func (r *RaceRepository) UpdateCompleted(id int, completed bool) error {
	_, err := r.db.Exec(
		"UPDATE races SET completed = ? WHERE id = ?",
		completed, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса гонки: %v", err)
	}

	return nil
}

// Delete удаляет гонку
func (r *RaceRepository) Delete(id int) error {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	// Удаляем связанные результаты
	_, err = tx.Exec("DELETE FROM race_results WHERE race_id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка удаления результатов гонки: %v", err)
	}

	// Удаляем гонку
	_, err = tx.Exec("DELETE FROM races WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка удаления гонки: %v", err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка подтверждения транзакции: %v", err)
	}

	return nil
}

// GetAll возвращает все гонки
func (r *RaceRepository) GetAll() ([]*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed 
		FROM races 
		ORDER BY date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех гонок: %v", err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var dateStr string
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&dateStr,
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", err)
		}

		// Преобразуем строку в дату
		race.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты: %v", err)
		}

		// Десериализуем дисциплины из JSON
		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации дисциплин: %v", err)
		}

		races = append(races, &race)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	return races, nil
}

// GetActiveSeasonRaces получает все гонки активного сезона
func (r *RaceRepository) GetActiveSeasonRaces() ([]*models.Race, error) {
	query := `
		SELECT r.id, r.season_id, r.name, r.date, r.car_class, r.disciplines, r.completed 
		FROM races r
		JOIN seasons s ON r.season_id = s.id
		WHERE s.active = 1
		ORDER BY r.date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения гонок активного сезона: %v", err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var dateStr string
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&dateStr,
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", err)
		}

		// Преобразуем строку в дату
		race.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("ошибка разбора даты: %v", err)
		}

		// Десериализуем дисциплины из JSON
		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации дисциплин: %v", err)
		}

		races = append(races, &race)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	return races, nil
}
