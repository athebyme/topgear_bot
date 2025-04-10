package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/athebyme/forza-top-gear-bot/internal/models"
)

// TxFn представляет функцию, которая выполняется в транзакции
type TxFn func(*sql.Tx) error

// WithTransaction выполняет функцию в транзакции
func WithTransaction(db *sql.DB, fn TxFn) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// В случае паники откатываем транзакцию
			tx.Rollback()
			panic(p) // Re-panic после отката
		} else if err != nil {
			// В случае ошибки откатываем транзакцию
			tx.Rollback()
		} else {
			// Если все хорошо, подтверждаем транзакцию
			err = tx.Commit()
		}
	}()

	// Выполняем функцию в транзакции
	err = fn(tx)
	return err
}

// DriverRepository методы для работы с транзакциями

// CreateWithTx создает нового гонщика в рамках транзакции
func (r *DriverRepository) CreateWithTx(tx *sql.Tx, driver *models.Driver) (int, error) {
	query := `
		INSERT INTO drivers (telegram_id, name, description, photo_url) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int
	err := tx.QueryRow(query, driver.TelegramID, driver.Name, driver.Description, driver.PhotoURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания гонщика: %v", err)
	}

	return id, nil
}

// UpdateWithTx обновляет данные гонщика в рамках транзакции
func (r *DriverRepository) UpdateWithTx(tx *sql.Tx, driver *models.Driver) error {
	query := `
		UPDATE drivers 
		SET name = $1, description = $2, photo_url = $3 
		WHERE id = $4
	`

	_, err := tx.Exec(query, driver.Name, driver.Description, driver.PhotoURL, driver.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления гонщика: %v", err)
	}

	return nil
}

// SeasonRepository методы для работы с транзакциями

// CreateWithTx создает новый сезон в рамках транзакции
func (r *SeasonRepository) CreateWithTx(tx *sql.Tx, season *models.Season) (int, error) {
	// Если новый сезон активен, деактивируем все остальные
	if season.Active {
		_, err := tx.Exec("UPDATE seasons SET active = false WHERE active = true")
		if err != nil {
			return 0, fmt.Errorf("ошибка деактивации текущих сезонов: %v", err)
		}
	}

	// Форматируем даты для PostgreSQL
	var endDate sql.NullTime
	if !season.EndDate.IsZero() {
		endDate = sql.NullTime{
			Time:  season.EndDate,
			Valid: true,
		}
	}

	// Вставляем новый сезон
	var id int
	err := tx.QueryRow(
		"INSERT INTO seasons (name, start_date, end_date, active) VALUES ($1, $2, $3, $4) RETURNING id",
		season.Name, season.StartDate, endDate, season.Active,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания сезона: %v", err)
	}

	return id, nil
}

// UpdateWithTx обновляет сезон в рамках транзакции
func (r *SeasonRepository) UpdateWithTx(tx *sql.Tx, season *models.Season) error {
	// Если обновляемый сезон активен, деактивируем все остальные
	if season.Active {
		_, err := tx.Exec("UPDATE seasons SET active = false WHERE active = true AND id != $1", season.ID)
		if err != nil {
			return fmt.Errorf("ошибка деактивации текущих сезонов: %v", err)
		}
	}

	// Форматируем даты для PostgreSQL
	var endDate sql.NullTime
	if !season.EndDate.IsZero() {
		endDate = sql.NullTime{
			Time:  season.EndDate,
			Valid: true,
		}
	}

	// Обновляем сезон
	_, err := tx.Exec(
		"UPDATE seasons SET name = $1, start_date = $2, end_date = $3, active = $4 WHERE id = $5",
		season.Name, season.StartDate, endDate, season.Active, season.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления сезона: %v", err)
	}

	return nil
}

// RaceRepository методы для работы с транзакциями

// CreateWithTx создает новую гонку в рамках транзакции
func (r *RaceRepository) CreateWithTx(tx *sql.Tx, race *models.Race) (int, error) {
	// Сериализуем дисциплины в JSON
	disciplinesJSON, err := json.Marshal(race.Disciplines)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации дисциплин: %v", err)
	}

	// Вставляем новую гонку
	var id int
	err = tx.QueryRow(
		`INSERT INTO races (season_id, name, date, car_class, disciplines, completed) 
		 VALUES ($1, $2, $3, $4, $5, $6) 
		 RETURNING id`,
		race.SeasonID, race.Name, race.Date, race.CarClass, disciplinesJSON, race.Completed,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания гонки: %v", err)
	}

	return id, nil
}

// UpdateWithTx обновляет гонку в рамках транзакции
func (r *RaceRepository) UpdateWithTx(tx *sql.Tx, race *models.Race) error {
	// Сериализуем дисциплины в JSON
	disciplinesJSON, err := json.Marshal(race.Disciplines)
	if err != nil {
		return fmt.Errorf("ошибка сериализации дисциплин: %v", err)
	}

	// Обновляем гонку
	_, err = tx.Exec(
		`UPDATE races 
		 SET season_id = $1, name = $2, date = $3, car_class = $4, disciplines = $5, completed = $6 
		 WHERE id = $7`,
		race.SeasonID, race.Name, race.Date, race.CarClass, disciplinesJSON, race.Completed, race.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления гонки: %v", err)
	}

	return nil
}

// UpdateCompletedWithTx изменяет статус завершенности гонки в рамках транзакции
func (r *RaceRepository) UpdateCompletedWithTx(tx *sql.Tx, id int, completed bool) error {
	_, err := tx.Exec(
		"UPDATE races SET completed = $1 WHERE id = $2",
		completed, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса гонки: %v", err)
	}

	return nil
}

// DeleteWithTx удаляет гонку в рамках транзакции
func (r *RaceRepository) DeleteWithTx(tx *sql.Tx, id int) error {
	// Удаляем связанные результаты
	_, err := tx.Exec("DELETE FROM race_results WHERE race_id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления результатов гонки: %v", err)
	}

	// Удаляем связанные назначения машин
	_, err = tx.Exec("DELETE FROM race_car_assignments WHERE race_id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления назначений машин: %v", err)
	}

	// Удаляем гонку
	_, err = tx.Exec("DELETE FROM races WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления гонки: %v", err)
	}

	return nil
}

// ResultRepository методы для работы с транзакциями

// CreateWithTx создает новый результат гонки в рамках транзакции
func (r *ResultRepository) CreateWithTx(tx *sql.Tx, result *models.RaceResult) (int, error) {
	// Сериализуем результаты в JSON
	resultsJSON, err := json.Marshal(result.Results)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации результатов: %v", err)
	}

	// Вставляем новый результат
	var id int
	err = tx.QueryRow(
		`INSERT INTO race_results 
		(race_id, driver_id, car_number, car_name, car_photo_url, results, total_score) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		result.RaceID, result.DriverID, result.CarNumber, result.CarName,
		result.CarPhotoURL, resultsJSON, result.TotalScore,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания результата: %v", err)
	}

	return id, nil
}

// UpdateWithTx обновляет результат в рамках транзакции
func (r *ResultRepository) UpdateWithTx(tx *sql.Tx, result *models.RaceResult) error {
	// Сериализуем результаты в JSON
	resultsJSON, err := json.Marshal(result.Results)
	if err != nil {
		return fmt.Errorf("ошибка сериализации результатов: %v", err)
	}

	// Обновляем результат
	_, err = tx.Exec(
		`UPDATE race_results 
		SET race_id = $1, driver_id = $2, car_number = $3, car_name = $4, 
			car_photo_url = $5, results = $6, total_score = $7 
		WHERE id = $8`,
		result.RaceID, result.DriverID, result.CarNumber, result.CarName,
		result.CarPhotoURL, resultsJSON, result.TotalScore, result.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления результата: %v", err)
	}

	return nil
}

// DeleteWithTx удаляет результат в рамках транзакции
func (r *ResultRepository) DeleteWithTx(tx *sql.Tx, id int) error {
	_, err := tx.Exec("DELETE FROM race_results WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления результата: %v", err)
	}

	return nil
}

// DeleteByRaceIDWithTx удаляет все результаты указанной гонки в рамках транзакции
func (r *ResultRepository) DeleteByRaceIDWithTx(tx *sql.Tx, raceID int) error {
	_, err := tx.Exec("DELETE FROM race_results WHERE race_id = $1", raceID)
	if err != nil {
		return fmt.Errorf("ошибка удаления результатов гонки: %v", err)
	}

	return nil
}
