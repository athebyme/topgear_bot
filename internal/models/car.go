package models

import "time"

// Car представляет автомобиль из Forza Horizon 4
type Car struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Year         int     `json:"year"`
	ImageURL     string  `json:"image_url"`
	Price        int     `json:"price"`
	Rarity       string  `json:"rarity"`
	Speed        float64 `json:"speed"`
	Handling     float64 `json:"handling"`
	Acceleration float64 `json:"acceleration"`
	Launch       float64 `json:"launch"`
	Braking      float64 `json:"braking"`
	ClassLetter  string  `json:"class_letter"`
	ClassNumber  int     `json:"class_number"`
	Source       string  `json:"source"`
}

// CarClass представляет класс автомобиля
type CarClass struct {
	Letter string `json:"letter"`
	Name   string `json:"name"`
}

// Список классов автомобилей
var CarClasses = []CarClass{
	{Letter: "D", Name: "D класс (500-599)"},
	{Letter: "C", Name: "C класс (600-699)"},
	{Letter: "B", Name: "B класс (700-799)"},
	{Letter: "A", Name: "A класс (800-899)"},
	{Letter: "S1", Name: "S1 класс (900-999)"},
	{Letter: "S2", Name: "S2 класс (1000-1099)"},
	{Letter: "X", Name: "X класс (1100+)"},
}

// RaceCarAssignment представляет назначение машины для гонщика
type RaceCarAssignment struct {
	ID               int       `json:"id"`
	RaceID           int       `json:"race_id"`
	DriverID         int       `json:"driver_id"`
	CarID            int       `json:"car_id"`
	AssignmentNumber int       `json:"assignment_number"`
	CreatedAt        time.Time `json:"created_at"`

	// Вложенные данные
	Car        *Car   `json:"car,omitempty"`
	DriverName string `json:"driver_name,omitempty"`
}

// CarAssignmentResult представляет результат случайного назначения машины
type CarAssignmentResult struct {
	DriverID         int    `json:"driver_id"`
	DriverName       string `json:"driver_name"`
	AssignmentNumber int    `json:"assignment_number"`
	Car              *Car   `json:"car,omitempty"`
}

// GetCarClassByLetter возвращает класс автомобиля по букве
func GetCarClassByLetter(letter string) *CarClass {
	for _, class := range CarClasses {
		if class.Letter == letter {
			return &class
		}
	}
	return nil
}

// GetCarClassName возвращает название класса автомобиля
func GetCarClassName(letter string) string {
	class := GetCarClassByLetter(letter)
	if class != nil {
		return class.Name
	}
	return "Неизвестный класс"
}
