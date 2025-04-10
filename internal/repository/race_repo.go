package repository

import (
	"database/sql"
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

type RaceRepository struct {
	db *sql.DB
}

func NewRaceRepository(db *sql.DB) *RaceRepository {
	return &RaceRepository{db: db}
}

func (r *RaceRepository) Create(race *models.Race) (int, error) {
	disciplinesJSON, err := models.SerializeDisciplines(race.Disciplines)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации дисциплин: %v", err)
	}

	dateStr := race.Date.Format("2006-01-02")

	var id int
	err = r.db.QueryRow(
		`INSERT INTO races 
        (season_id, name, date, car_class, disciplines, completed) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id`,
		race.SeasonID,
		race.Name,
		dateStr,
		race.CarClass,
		disciplinesJSON,
		race.Completed,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания гонки: %v", err)
	}

	return id, nil
}

func (r *RaceRepository) GetByID(id int) (*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed
		FROM races
		WHERE id = $1
	`

	var race models.Race
	// var dateStr string // REMOVE
	var disciplinesJSON string

	err := r.db.QueryRow(query, id).Scan(
		&race.ID,
		&race.SeasonID,
		&race.Name,
		&race.Date, // SCAN DIRECTLY INTO race.Date
		&race.CarClass,
		&disciplinesJSON,
		&race.Completed,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil for not found
		}
		return nil, fmt.Errorf("ошибка получения гонки: %v", err)
	}

	// race.Date, err = time.Parse("2006-01-02", dateStr) // REMOVE
	// if err != nil {
	// 	return nil, fmt.Errorf("ошибка разбора даты: %v", err)
	// }

	race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
	if err != nil {
		// Wrap error for context
		return nil, fmt.Errorf("ошибка десериализации дисциплин для гонки ID %d: %v", id, err)
	}

	return &race, nil
}

func (r *RaceRepository) GetBySeason(seasonID int) ([]*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed
		FROM races
		WHERE season_id = $1
		ORDER BY date DESC
	`

	rows, err := r.db.Query(query, seasonID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения гонок сезона %d: %v", seasonID, err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		// var dateStr string // REMOVE
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&race.Date, // SCAN DIRECTLY INTO race.Date
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки в сезоне %d: %v", seasonID, err)
		}

		// race.Date, err = time.Parse("2006-01-02", dateStr) // REMOVE
		// if err != nil {
		// 	return nil, fmt.Errorf("ошибка разбора даты для гонки ID %d: %v", race.ID, err)
		// }

		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			// Wrap error for context
			return nil, fmt.Errorf("ошибка десериализации дисциплин для гонки ID %d: %v", race.ID, err)
		}

		races = append(races, &race)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам сезона %d: %v", seasonID, err)
	}

	return races, nil
}

func (r *RaceRepository) GetIncompleteRaces() ([]*models.Race, error) {
	// WHERE completed = false is correct for boolean
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed
		FROM races
		WHERE completed = false
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
		// var dateStr string // REMOVE
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&race.Date, // SCAN DIRECTLY INTO race.Date
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных незавершенной гонки: %v", err)
		}

		// race.Date, err = time.Parse("2006-01-02", dateStr) // REMOVE
		// if err != nil {
		// 	return nil, fmt.Errorf("ошибка разбора даты для гонки ID %d: %v", race.ID, err)
		// }

		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			// Wrap error for context
			return nil, fmt.Errorf("ошибка десериализации дисциплин для гонки ID %d: %v", race.ID, err)
		}

		races = append(races, &race)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по незавершенным гонкам: %v", err)
	}

	return races, nil
}

// Suggestion: Make Update consistent with Create regarding discipline serialization
func (r *RaceRepository) Update(tx *sql.Tx, race *models.Race) error {
	// Use the same serialization function as in Create for consistency
	disciplinesJSON, err := models.SerializeDisciplines(race.Disciplines)
	if err != nil {
		return fmt.Errorf("ошибка сериализации дисциплин: %v", err)
	}

	// Date needs to be passed as time.Time, the driver handles formatting
	updateQuery := `
		UPDATE races
		SET season_id = $1, name = $2, date = $3, car_class = $4, disciplines = $5, completed = $6
		WHERE id = $7
	`
	// Pass race.Date directly, the driver should handle it.
	// If your DB column is DATE, the time part will be truncated by the DB.
	// If your DB column is TIMESTAMP/TIMESTAMPTZ, the full time will be stored.
	args := []interface{}{
		race.SeasonID, race.Name, race.Date, race.CarClass, disciplinesJSON, race.Completed, race.ID,
	}

	if tx != nil {
		_, err = tx.Exec(updateQuery, args...)
	} else {
		_, err = r.db.Exec(updateQuery, args...)
	}

	if err != nil {
		return fmt.Errorf("ошибка обновления гонки ID %d: %v", race.ID, err)
	}

	return nil
}

// Add GetResultCountByRaceID if it's missing (used in callbackCompleteRace and callbackDeleteRace)
// NOTE: This should ideally be in ResultRepository, but adding here for completeness based on usage.
// Move it to ResultRepository if possible.
func (r *RaceRepository) GetResultCountByRaceID(raceID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM results WHERE race_id = $1"
	err := r.db.QueryRow(query, raceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества результатов для гонки ID %d: %v", raceID, err)
	}
	return count, nil
}

func (r *RaceRepository) UpdateCompleted(id int, completed bool) error {
	// Исправлено для использования boolean значения
	_, err := r.db.Exec(
		"UPDATE races SET completed = $1 WHERE id = $2",
		completed, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса гонки: %v", err)
	}

	return nil
}
