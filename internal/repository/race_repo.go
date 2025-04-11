package repository

import (
	"database/sql"
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/models"
	"log"
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
	log.Printf("RaceRepository.GetBySeason(): Запрос гонок для сезона ID=%d", seasonID)

	// Проверяем наличие строки state в запросе - если в таблице нет этой колонки,
	// используем упрощенный запрос
	var hasStateColumn bool
	err := r.db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'races' AND column_name = 'state')").Scan(&hasStateColumn)
	if err != nil {
		log.Printf("Ошибка проверки наличия колонки state: %v", err)
		hasStateColumn = false
	}

	var t string
	if hasStateColumn {
		t = "присутствует"
	} else {
		t = "отсутствует"
	}

	log.Printf("Колонка state %s в таблице races", t)

	var query string
	if hasStateColumn {
		query = `
			SELECT id, season_id, name, date, car_class, disciplines, completed, state
			FROM races
			WHERE season_id = $1
			ORDER BY date DESC
		`
	} else {
		query = `
			SELECT id, season_id, name, date, car_class, disciplines, completed
			FROM races
			WHERE season_id = $1
			ORDER BY date DESC
		`
	}

	rows, err := r.db.Query(query, seasonID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения гонок сезона %d: %v", seasonID, err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var disciplinesJSON string

		var scanErr error
		if hasStateColumn {
			scanErr = rows.Scan(
				&race.ID,
				&race.SeasonID,
				&race.Name,
				&race.Date,
				&race.CarClass,
				&disciplinesJSON,
				&race.Completed,
				&race.State,
			)
		} else {
			scanErr = rows.Scan(
				&race.ID,
				&race.SeasonID,
				&race.Name,
				&race.Date,
				&race.CarClass,
				&disciplinesJSON,
				&race.Completed,
			)
			// Устанавливаем состояние на основе флага Completed
			if race.Completed {
				race.State = models.RaceStateCompleted
			} else {
				race.State = models.RaceStateNotStarted
			}
		}

		if scanErr != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", scanErr)
		}

		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			// Продолжаем работу с другими гонками в случае ошибки с дисциплинами
			log.Printf("Ошибка десериализации дисциплин для гонки ID %d: %v", race.ID, err)
			race.Disciplines = []string{"Неизвестные дисциплины"}
		}

		races = append(races, &race)
		log.Printf("Получена гонка ID=%d, Name='%s', Completed=%v, State='%s'",
			race.ID, race.Name, race.Completed, race.State)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	log.Printf("RaceRepository.GetBySeason(): Найдено %d гонок для сезона ID=%d", len(races), seasonID)
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

// UpdateState updates the state of a race
func (r *RaceRepository) UpdateState(raceID int, state string) error {
	_, err := r.db.Exec(
		"UPDATE races SET state = $1 WHERE id = $2",
		state, raceID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления состояния гонки: %v", err)
	}
	return nil
}

// GetRegisteredDrivers gets all drivers registered for a race
func (r *RaceRepository) GetRegisteredDrivers(raceID int) ([]*models.RaceRegistration, error) {
	query := `
		SELECT rr.id, rr.race_id, rr.driver_id, rr.registered_at, rr.car_confirmed, rr.reroll_used, d.name
		FROM race_registrations rr
		JOIN drivers d ON rr.driver_id = d.id
		WHERE rr.race_id = $1
		ORDER BY rr.registered_at
	`

	rows, err := r.db.Query(query, raceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения зарегистрированных гонщиков: %v", err)
	}
	defer rows.Close()

	var registrations []*models.RaceRegistration

	for rows.Next() {
		var reg models.RaceRegistration
		err := rows.Scan(
			&reg.ID,
			&reg.RaceID,
			&reg.DriverID,
			&reg.RegisteredAt,
			&reg.CarConfirmed,
			&reg.RerollUsed,
			&reg.DriverName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных регистрации: %v", err)
		}
		registrations = append(registrations, &reg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по регистрациям: %v", err)
	}

	return registrations, nil
}

// RegisterDriver registers a driver for a race
func (r *RaceRepository) RegisterDriver(raceID, driverID int) error {
	_, err := r.db.Exec(
		`INSERT INTO race_registrations (race_id, driver_id) 
		 VALUES ($1, $2) 
		 ON CONFLICT (race_id, driver_id) DO NOTHING`,
		raceID, driverID,
	)
	if err != nil {
		return fmt.Errorf("ошибка регистрации гонщика: %v", err)
	}
	return nil
}

// UnregisterDriver unregisters a driver from a race
func (r *RaceRepository) UnregisterDriver(raceID, driverID int) error {
	_, err := r.db.Exec(
		"DELETE FROM race_registrations WHERE race_id = $1 AND driver_id = $2",
		raceID, driverID,
	)
	if err != nil {
		return fmt.Errorf("ошибка отмены регистрации гонщика: %v", err)
	}
	return nil
}

// CheckDriverRegistered checks if a driver is registered for a race
func (r *RaceRepository) CheckDriverRegistered(raceID, driverID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM race_registrations WHERE race_id = $1 AND driver_id = $2)",
		raceID, driverID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки регистрации гонщика: %v", err)
	}
	return exists, nil
}

// UpdateCarConfirmation updates the car confirmation status
func (r *RaceRepository) UpdateCarConfirmation(raceID, driverID int, confirmed bool) error {
	_, err := r.db.Exec(
		"UPDATE race_registrations SET car_confirmed = $1 WHERE race_id = $2 AND driver_id = $3",
		confirmed, raceID, driverID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса подтверждения машины: %v", err)
	}
	return nil
}

// UpdateRerollUsed marks that a driver has used their reroll
func (r *RaceRepository) UpdateRerollUsed(raceID, driverID int, used bool) error {
	_, err := r.db.Exec(
		"UPDATE race_registrations SET reroll_used = $1 WHERE race_id = $2 AND driver_id = $3",
		used, raceID, driverID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса использования реролла: %v", err)
	}
	return nil
}

// GetActiveRace returns the currently active race (in progress)
func (r *RaceRepository) GetActiveRace() (*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed, state
		FROM races
		WHERE state = $1
		ORDER BY date DESC
		LIMIT 1
	`

	var race models.Race
	var disciplinesJSON string

	err := r.db.QueryRow(query, models.RaceStateInProgress).Scan(
		&race.ID,
		&race.SeasonID,
		&race.Name,
		&race.Date,
		&race.CarClass,
		&disciplinesJSON,
		&race.Completed,
		&race.State,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active race
		}
		return nil, fmt.Errorf("ошибка получения активной гонки: %v", err)
	}

	race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации дисциплин: %v", err)
	}

	return &race, nil
}

// GetUpcomingRaces returns races that haven't started yet
func (r *RaceRepository) GetUpcomingRaces() ([]*models.Race, error) {
	query := `
		SELECT id, season_id, name, date, car_class, disciplines, completed, state
		FROM races
		WHERE state = $1
		ORDER BY date ASC
	`

	rows, err := r.db.Query(query, models.RaceStateNotStarted)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения предстоящих гонок: %v", err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var disciplinesJSON string

		err := rows.Scan(
			&race.ID,
			&race.SeasonID,
			&race.Name,
			&race.Date,
			&race.CarClass,
			&disciplinesJSON,
			&race.Completed,
			&race.State,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", err)
		}

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

// StartRace changes race state to in_progress and assigns cars to registered drivers
func (r *RaceRepository) StartRace(tx *sql.Tx, raceID int) error {
	// Update race state
	_, err := tx.Exec("UPDATE races SET state = $1 WHERE id = $2", models.RaceStateInProgress, raceID)
	if err != nil {
		return fmt.Errorf("ошибка обновления состояния гонки: %v", err)
	}

	return nil
}

// CompleteRace changes race state to completed
func (r *RaceRepository) CompleteRace(tx *sql.Tx, raceID int) error {
	// Update race state and mark as completed
	_, err := tx.Exec("UPDATE races SET state = $1, completed = true WHERE id = $2",
		models.RaceStateCompleted, raceID)
	if err != nil {
		return fmt.Errorf("ошибка завершения гонки: %v", err)
	}

	return nil
}

// GetAll возвращает все гонки
func (r *RaceRepository) GetAll() ([]*models.Race, error) {
	log.Printf("RaceRepository.GetAll(): Запрос на получение всех гонок")

	// Проверяем наличие строки state в запросе - если в таблице нет этой колонки,
	// используем упрощенный запрос
	var hasStateColumn bool
	err := r.db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'races' AND column_name = 'state')").Scan(&hasStateColumn)
	if err != nil {
		log.Printf("Ошибка проверки наличия колонки state: %v", err)
		hasStateColumn = false
	}

	var t string
	if hasStateColumn {
		t = "присутствует"
	} else {
		t = "отсутствует"
	}

	log.Printf("Колонка state %s в таблице races", t)

	var query string
	if hasStateColumn {
		query = `
			SELECT id, season_id, name, date, car_class, disciplines, completed, state
			FROM races
			ORDER BY date DESC
		`
	} else {
		query = `
			SELECT id, season_id, name, date, car_class, disciplines, completed
			FROM races
			ORDER BY date DESC
		`
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех гонок: %v", err)
	}
	defer rows.Close()

	var races []*models.Race

	for rows.Next() {
		var race models.Race
		var disciplinesJSON string

		var scanErr error
		if hasStateColumn {
			scanErr = rows.Scan(
				&race.ID,
				&race.SeasonID,
				&race.Name,
				&race.Date,
				&race.CarClass,
				&disciplinesJSON,
				&race.Completed,
				&race.State,
			)
		} else {
			scanErr = rows.Scan(
				&race.ID,
				&race.SeasonID,
				&race.Name,
				&race.Date,
				&race.CarClass,
				&disciplinesJSON,
				&race.Completed,
			)
			// Устанавливаем состояние на основе флага Completed
			if race.Completed {
				race.State = models.RaceStateCompleted
			} else {
				race.State = models.RaceStateNotStarted
			}
		}

		if scanErr != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонки: %v", scanErr)
		}

		race.Disciplines, err = models.DeserializeDisciplines(disciplinesJSON)
		if err != nil {
			// Продолжаем работу с другими гонками в случае ошибки с дисциплинами
			log.Printf("Ошибка десериализации дисциплин для гонки ID %d: %v", race.ID, err)
			race.Disciplines = []string{"Неизвестные дисциплины"}
		}

		races = append(races, &race)
		log.Printf("Получена гонка ID=%d, Name='%s', Completed=%v, State='%s'",
			race.ID, race.Name, race.Completed, race.State)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонкам: %v", err)
	}

	log.Printf("RaceRepository.GetAll(): Найдено %d гонок", len(races))
	return races, nil
}
