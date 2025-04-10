package repository

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/athebyme/forza-top-gear-bot/internal/models"
	"github.com/lib/pq"
)

// CarRepository представляет репозиторий для работы с машинами
type CarRepository struct {
	db *sql.DB
}

// NewCarRepository создает новый репозиторий машин
func NewCarRepository(db *sql.DB) *CarRepository {
	return &CarRepository{db: db}
}

// GetByID получает машину по ID
func (r *CarRepository) GetByID(id int) (*models.Car, error) {
	query := `
		SELECT id, name, year, image_url, price, rarity, speed, handling, 
		       acceleration, launch, braking, class_letter, class_number, source
		FROM cars 
		WHERE id = $1
	`

	var car models.Car

	var yearRaw sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&car.ID,
		&car.Name,
		&yearRaw,
		&car.ImageURL,
		&car.Price,
		&car.Rarity,
		&car.Speed,
		&car.Handling,
		&car.Acceleration,
		&car.Launch,
		&car.Braking,
		&car.ClassLetter,
		&car.ClassNumber,
		&car.Source,
	)

	if yearRaw.Valid {
		car.Year = fmt.Sprintf("%d", yearRaw.Int64)
	} else {
		car.Year = "нет информации"
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Машина не найдена
		}
		return nil, fmt.Errorf("ошибка получения машины: %v", err)
	}

	return &car, nil
}

// GetByClass получает машины определенного класса
func (r *CarRepository) GetByClass(classLetter string) ([]*models.Car, error) {
	query := `
		SELECT id, name, year, image_url, price, rarity, speed, handling, 
		       acceleration, launch, braking, class_letter, class_number, source
		FROM cars 
		WHERE class_letter = $1
		ORDER BY name, year
	`

	rows, err := r.db.Query(query, classLetter)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения машин класса %s: %v", classLetter, err)
	}
	defer rows.Close()

	var cars []*models.Car
	var yearRaw sql.NullInt64

	for rows.Next() {
		yearRaw = sql.NullInt64{}
		var car models.Car

		err := rows.Scan(
			&car.ID,
			&car.Name,
			&yearRaw,
			&car.ImageURL,
			&car.Price,
			&car.Rarity,
			&car.Speed,
			&car.Handling,
			&car.Acceleration,
			&car.Launch,
			&car.Braking,
			&car.ClassLetter,
			&car.ClassNumber,
			&car.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных машины: %v", err)
		}
		if yearRaw.Valid {
			car.Year = fmt.Sprintf("%d", yearRaw.Int64)
		} else {
			car.Year = "нет информации"
		}

		cars = append(cars, &car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по машинам: %v", err)
	}

	return cars, nil
}

// GetAll возвращает все машины
func (r *CarRepository) GetAll() ([]*models.Car, error) {
	query := `
		SELECT id, name, year, image_url, price, rarity, speed, handling, 
		       acceleration, launch, braking, class_letter, class_number, source
		FROM cars 
		ORDER BY name, year
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех машин: %v", err)
	}
	defer rows.Close()

	var cars []*models.Car

	for rows.Next() {
		var yearRaw sql.NullInt64
		var car models.Car

		err := rows.Scan(
			&car.ID,
			&car.Name,
			&yearRaw,
			&car.ImageURL,
			&car.Price,
			&car.Rarity,
			&car.Speed,
			&car.Handling,
			&car.Acceleration,
			&car.Launch,
			&car.Braking,
			&car.ClassLetter,
			&car.ClassNumber,
			&car.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных машины: %v", err)
		}

		if yearRaw.Valid {
			car.Year = fmt.Sprintf("%d", yearRaw.Int64)
		} else {
			car.Year = "нет информации"
		}

		cars = append(cars, &car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по машинам: %v", err)
	}

	return cars, nil
}

// CountByClass подсчитывает количество машин определенного класса
func (r *CarRepository) CountByClass(classLetter string) (int, error) {
	query := `SELECT COUNT(*) FROM cars WHERE class_letter = $1`

	var count int
	err := r.db.QueryRow(query, classLetter).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка подсчета машин класса %s: %v", classLetter, err)
	}

	return count, nil
}

// GetClassCounts возвращает количество машин по каждому классу
func (r *CarRepository) GetClassCounts() (map[string]int, error) {
	query := `SELECT class_letter, COUNT(*) FROM cars GROUP BY class_letter ORDER BY class_letter`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения количества машин по классам: %v", err)
	}
	defer rows.Close()

	counts := make(map[string]int)

	for rows.Next() {
		var classLetter string
		var count int

		err := rows.Scan(&classLetter, &count)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных: %v", err)
		}

		counts[classLetter] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по данным: %v", err)
	}

	return counts, nil
}

// AssignRandomCars назначает случайные машины для гонки
func (r *CarRepository) AssignRandomCars(tx *sql.Tx, raceID int, driverIDs []int, carClass string) ([]*models.CarAssignmentResult, error) {
	// Получаем все машины указанного класса
	cars, err := r.GetByClass(carClass)
	if err != nil {
		return nil, err
	}

	if len(cars) == 0 {
		return nil, fmt.Errorf("нет машин класса %s", carClass)
	}

	// Получаем количество машин в классе
	carCount := len(cars)

	// Определяем максимальный номер машины (количество машин * 1.7)
	maxCarNumber := int(float64(carCount) * 1.7)

	// Устанавливаем сид для генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Генерируем уникальные случайные номера для каждого гонщика
	usedNumbers := make(map[int]bool)
	var results []*models.CarAssignmentResult

	// Получаем имена гонщиков
	var driverNames = make(map[int]string)

	// Если есть транзакция, используем её для запроса
	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query("SELECT id, name FROM drivers WHERE id = ANY($1)", pq.Array(driverIDs))
	} else {
		rows, err = r.db.Query("SELECT id, name FROM drivers WHERE id = ANY($1)", pq.Array(driverIDs))
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка получения имен гонщиков: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string

		err := rows.Scan(&id, &name)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных гонщика: %v", err)
		}

		driverNames[id] = name
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по гонщикам: %v", err)
	}

	// Назначаем случайные номера и машины
	for _, driverID := range driverIDs {
		var assignmentNumber int

		// Генерируем уникальный случайный номер
		for {
			assignmentNumber = rand.Intn(maxCarNumber) + 1 // От 1 до maxCarNumber
			if !usedNumbers[assignmentNumber] {
				usedNumbers[assignmentNumber] = true
				break
			}
		}

		// Определяем индекс машины, соответствующий номеру (по модулю)
		carIndex := (assignmentNumber - 1) % carCount
		car := cars[carIndex]

		// Создаем запись о назначении машины
		var assignmentID int
		var insertErr error

		if tx != nil {
			insertErr = tx.QueryRow(
				`INSERT INTO race_car_assignments (race_id, driver_id, car_id, assignment_number)
				 VALUES ($1, $2, $3, $4)
				 RETURNING id`,
				raceID, driverID, car.ID, assignmentNumber,
			).Scan(&assignmentID)
		} else {
			insertErr = r.db.QueryRow(
				`INSERT INTO race_car_assignments (race_id, driver_id, car_id, assignment_number)
				 VALUES ($1, $2, $3, $4)
				 RETURNING id`,
				raceID, driverID, car.ID, assignmentNumber,
			).Scan(&assignmentID)
		}

		if insertErr != nil {
			return nil, fmt.Errorf("ошибка создания назначения машины: %v", insertErr)
		}

		// Добавляем результат
		results = append(results, &models.CarAssignmentResult{
			DriverID:         driverID,
			DriverName:       driverNames[driverID],
			AssignmentNumber: assignmentNumber,
			Car:              car,
		})
	}

	return results, nil
}

// GetRaceCarAssignments получает назначения машин для гонки
func (r *CarRepository) GetRaceCarAssignments(raceID int) ([]*models.RaceCarAssignment, error) {
	query := `
		SELECT rca.id, rca.race_id, rca.driver_id, rca.car_id, rca.assignment_number, rca.created_at,
		       c.id, c.name, c.year, c.image_url, c.price, c.rarity, c.speed, c.handling, c.acceleration,
		       c.launch, c.braking, c.class_letter, c.class_number, c.source,
		       d.name
		FROM race_car_assignments rca
		JOIN cars c ON rca.car_id = c.id
		JOIN drivers d ON rca.driver_id = d.id
		WHERE rca.race_id = $1
		ORDER BY rca.assignment_number
	`

	rows, err := r.db.Query(query, raceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения назначений машин: %v", err)
	}
	defer rows.Close()

	var assignments []*models.RaceCarAssignment

	for rows.Next() {
		var assignment models.RaceCarAssignment
		var car models.Car
		var yearRaw sql.NullInt64

		err := rows.Scan(
			&assignment.ID,
			&assignment.RaceID,
			&assignment.DriverID,
			&assignment.CarID,
			&assignment.AssignmentNumber,
			&assignment.CreatedAt,
			&car.ID,
			&car.Name,
			&yearRaw,
			&car.ImageURL,
			&car.Price,
			&car.Rarity,
			&car.Speed,
			&car.Handling,
			&car.Acceleration,
			&car.Launch,
			&car.Braking,
			&car.ClassLetter,
			&car.ClassNumber,
			&car.Source,
			&assignment.DriverName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных назначения: %v", err)
		}

		if yearRaw.Valid {
			car.Year = fmt.Sprintf("%d", yearRaw.Int64)
		} else {
			car.Year = "нет информации"
		}

		assignment.Car = &car
		assignments = append(assignments, &assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по назначениям: %v", err)
	}

	return assignments, nil
}

// DeleteRaceCarAssignments удаляет все назначения машин для гонки
func (r *CarRepository) DeleteRaceCarAssignments(tx *sql.Tx, raceID int) error {
	var err error

	if tx != nil {
		_, err = tx.Exec("DELETE FROM race_car_assignments WHERE race_id = $1", raceID)
	} else {
		_, err = r.db.Exec("DELETE FROM race_car_assignments WHERE race_id = $1", raceID)
	}

	if err != nil {
		return fmt.Errorf("ошибка удаления назначений машин: %v", err)
	}

	return nil
}
