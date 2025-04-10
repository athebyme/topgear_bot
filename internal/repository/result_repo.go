package repository

import (
	"database/sql"
	"fmt"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

// ResultRepository представляет репозиторий для работы с результатами гонок
type ResultRepository struct {
	db *sql.DB
}

// NewResultRepository создает новый репозиторий результатов
func NewResultRepository(db *sql.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

// Create создает новый результат гонки
func (r *ResultRepository) Create(result *models.RaceResult) (int, error) {
	// Сериализуем результаты в JSON
	resultsJSON, err := models.SerializeResults(result.Results)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации результатов: %v", err)
	}

	// Вставляем новый результат
	newResult, err := r.db.Exec(
		`INSERT INTO race_results 
		(race_id, driver_id, car_number, car_name, car_photo_url, results, total_score) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		result.RaceID, result.DriverID, result.CarNumber, result.CarName,
		result.CarPhotoURL, resultsJSON, result.TotalScore,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания результата: %v", err)
	}

	// Получаем ID нового результата
	id, err := newResult.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID результата: %v", err)
	}

	return int(id), nil
}

// GetByID получает результат по ID
func (r *ResultRepository) GetByID(id int) (*models.RaceResult, error) {
	query := `
		SELECT id, race_id, driver_id, car_number, car_name, car_photo_url, results, total_score 
		FROM race_results 
		WHERE id = $1
	`

	var result models.RaceResult
	var resultsJSON string

	err := r.db.QueryRow(query, id).Scan(
		&result.ID,
		&result.RaceID,
		&result.DriverID,
		&result.CarNumber,
		&result.CarName,
		&result.CarPhotoURL,
		&resultsJSON,
		&result.TotalScore,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Результат не найден
		}
		return nil, fmt.Errorf("ошибка получения результата: %v", err)
	}

	// Десериализуем результаты из JSON
	result.Results, err = models.DeserializeResults(resultsJSON)
	if err != nil {
		return nil, fmt.Errorf("ошибка десериализации результатов: %v", err)
	}

	return &result, nil
}

// GetByRaceID получает все результаты указанной гонки
func (r *ResultRepository) GetByRaceID(raceID int) ([]*models.RaceResult, error) {
	query := `
		SELECT id, race_id, driver_id, car_number, car_name, car_photo_url, results, total_score 
		FROM race_results 
		WHERE race_id = $1 
		ORDER BY total_score DESC
	`

	rows, err := r.db.Query(query, raceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения результатов гонки: %v", err)
	}
	defer rows.Close()

	var results []*models.RaceResult

	for rows.Next() {
		var result models.RaceResult
		var resultsJSON string

		err := rows.Scan(
			&result.ID,
			&result.RaceID,
			&result.DriverID,
			&result.CarNumber,
			&result.CarName,
			&result.CarPhotoURL,
			&resultsJSON,
			&result.TotalScore,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных результата: %v", err)
		}

		// Десериализуем результаты из JSON
		result.Results, err = models.DeserializeResults(resultsJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации результатов: %v", err)
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по результатам: %v", err)
	}

	return results, nil
}

// GetByDriverID получает все результаты указанного гонщика
func (r *ResultRepository) GetByDriverID(driverID int) ([]*models.RaceResult, error) {
	query := `
		SELECT id, race_id, driver_id, car_number, car_name, car_photo_url, results, total_score 
		FROM race_results 
		WHERE driver_id = $1
		ORDER BY id DESC
	`

	rows, err := r.db.Query(query, &driverID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения результатов гонщика: %v", err)
	}
	defer rows.Close()

	var results []*models.RaceResult

	for rows.Next() {
		var result models.RaceResult
		var resultsJSON string

		err := rows.Scan(
			&result.ID,
			&result.RaceID,
			&result.DriverID,
			&result.CarNumber,
			&result.CarName,
			&result.CarPhotoURL,
			&resultsJSON,
			&result.TotalScore,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных результата: %v", err)
		}

		// Десериализуем результаты из JSON
		result.Results, err = models.DeserializeResults(resultsJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации результатов: %v", err)
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по результатам: %v", err)
	}

	return results, nil
}

// Update обновляет результат
func (r *ResultRepository) Update(result *models.RaceResult) error {
	// Сериализуем результаты в JSON
	resultsJSON, err := models.SerializeResults(result.Results)
	if err != nil {
		return fmt.Errorf("ошибка сериализации результатов: %v", err)
	}

	// Обновляем результат
	_, err = r.db.Exec(
		`UPDATE race_results 
		SET race_id = $1, driver_id = $2, car_number = $3, car_name = $4, 
			car_photo_url = $5, results = $6, total_score = $7 
		WHERE id = $7`,
		&result.RaceID, &result.DriverID, &result.CarNumber, &result.CarName,
		&result.CarPhotoURL, &resultsJSON, &result.TotalScore, &result.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления результата: %v", err)
	}

	return nil
}

// Delete удаляет результат
func (r *ResultRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM race_results WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления результата: %v", err)
	}

	return nil
}

// DeleteByRaceID удаляет все результаты указанной гонки
func (r *ResultRepository) DeleteByRaceID(raceID int) error {
	_, err := r.db.Exec("DELETE FROM race_results WHERE race_id = $1", raceID)
	if err != nil {
		return fmt.Errorf("ошибка удаления результатов гонки: %v", err)
	}

	return nil
}

// CheckDriverResultExists проверяет, существует ли результат гонщика в указанной гонке
func (r *ResultRepository) CheckDriverResultExists(raceID, driverID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM race_results 
			WHERE race_id = $1 AND driver_id = $2
		)
	`, &raceID, &driverID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("ошибка проверки результата гонщика: %v", err)
	}

	return exists, nil
}

// GetResultCountByRaceID получает количество результатов для указанной гонки
func (r *ResultRepository) GetResultCountByRaceID(raceID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM race_results WHERE race_id = $1", raceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка подсчета результатов гонки: %v", err)
	}

	return count, nil
}

// GetRaceResults получает результаты гонки с именами гонщиков
type RaceResultWithDriver struct {
	models.RaceResult
	DriverName string
}

// GetRaceResultsWithDriverNames получает результаты гонки с именами гонщиков
func (r *ResultRepository) GetRaceResultsWithDriverNames(raceID int) ([]*RaceResultWithDriver, error) {
	query := `
		SELECT rr.id, rr.race_id, rr.driver_id, rr.car_number, rr.car_name, 
			   rr.car_photo_url, rr.results, rr.total_score, d.name 
		FROM race_results rr
		JOIN drivers d ON rr.driver_id = d.id
		WHERE rr.race_id = $1
		ORDER BY rr.total_score DESC
	`

	rows, err := r.db.Query(query, raceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения результатов гонки: %v", err)
	}
	defer rows.Close()

	var results []*RaceResultWithDriver

	for rows.Next() {
		var result RaceResultWithDriver
		var resultsJSON string

		err := rows.Scan(
			&result.ID,
			&result.RaceID,
			&result.DriverID,
			&result.CarNumber,
			&result.CarName,
			&result.CarPhotoURL,
			&resultsJSON,
			&result.TotalScore,
			&result.DriverName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных результата: %v", err)
		}

		// Десериализуем результаты из JSON
		result.Results, err = models.DeserializeResults(resultsJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации результатов: %v", err)
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по результатам: %v", err)
	}

	return results, nil
}

// CreateWithRerollPenalty creates a new race result with reroll penalty
func (r *ResultRepository) CreateWithRerollPenalty(result *models.RaceResult) (int, error) {
	// Serialize results to JSON
	resultsJSON, err := models.SerializeResults(result.Results)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации результатов: %v", err)
	}

	// Insert the new result with reroll penalty
	var id int
	err = r.db.QueryRow(
		`INSERT INTO race_results 
		(race_id, driver_id, car_number, car_name, car_photo_url, results, total_score, reroll_penalty) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		result.RaceID, result.DriverID, result.CarNumber, result.CarName,
		result.CarPhotoURL, resultsJSON, result.TotalScore, result.RerollPenalty,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания результата: %v", err)
	}

	return id, nil
}

// GetDriverRerollStatus checks if a driver has used their reroll
func (r *ResultRepository) GetDriverRerollStatus(raceID, driverID int) (bool, error) {
	var rerollUsed bool
	err := r.db.QueryRow(`
		SELECT reroll_used FROM race_registrations
		WHERE race_id = $1 AND driver_id = $2
	`, raceID, driverID).Scan(&rerollUsed)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Not registered, so reroll not used
		}
		return false, fmt.Errorf("ошибка получения статуса реролла: %v", err)
	}

	return rerollUsed, nil
}

// GetRaceResultsWithRerollPenalty gets race results including reroll penalties
func (r *ResultRepository) GetRaceResultsWithRerollPenalty(raceID int) ([]*RaceResultWithDriver, error) {
	query := `
		SELECT rr.id, rr.race_id, rr.driver_id, rr.car_number, rr.car_name, 
			   rr.car_photo_url, rr.results, rr.total_score, rr.reroll_penalty, d.name 
		FROM race_results rr
		JOIN drivers d ON rr.driver_id = d.id
		WHERE rr.race_id = $1
		ORDER BY rr.total_score DESC
	`

	rows, err := r.db.Query(query, raceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения результатов гонки: %v", err)
	}
	defer rows.Close()

	var results []*RaceResultWithDriver

	for rows.Next() {
		var result RaceResultWithDriver
		var resultsJSON string

		err := rows.Scan(
			&result.ID,
			&result.RaceID,
			&result.DriverID,
			&result.CarNumber,
			&result.CarName,
			&result.CarPhotoURL,
			&resultsJSON,
			&result.TotalScore,
			&result.RerollPenalty,
			&result.DriverName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных результата: %v", err)
		}

		// Deserialize results from JSON
		result.Results, err = models.DeserializeResults(resultsJSON)
		if err != nil {
			return nil, fmt.Errorf("ошибка десериализации результатов: %v", err)
		}

		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по результатам: %v", err)
	}

	return results, nil
}

// ApplyRerollPenaltyToResult applies a reroll penalty to a result
func (r *ResultRepository) ApplyRerollPenaltyToResult(tx *sql.Tx, raceID, driverID int, penalty int) error {
	// First check if the result already exists
	var resultID int
	var currentScore int

	err := tx.QueryRow(`
		SELECT id, total_score FROM race_results
		WHERE race_id = $1 AND driver_id = $2
	`, raceID, driverID).Scan(&resultID, &currentScore)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("ошибка проверки существования результата: %v", err)
	}

	if err == sql.ErrNoRows {
		// No result exists yet, nothing to update
		return nil
	}

	// Update the existing result with the penalty
	_, err = tx.Exec(`
		UPDATE race_results
		SET reroll_penalty = $1, total_score = total_score - $1
		WHERE id = $2
	`, penalty, resultID)

	if err != nil {
		return fmt.Errorf("ошибка применения штрафа за реролл: %v", err)
	}

	return nil
}
