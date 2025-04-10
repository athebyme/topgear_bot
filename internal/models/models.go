package models

import (
	"encoding/json"
	"time"
)

// Driver представляет гонщика
type Driver struct {
	ID          int    `json:"id"`
	TelegramID  int64  `json:"telegram_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoURL    string `json:"photo_url"`
}

// Season представляет сезон гонок
type Season struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Active    bool      `json:"active"`
}

// Race представляет гонку
type Race struct {
	ID          int       `json:"id"`
	SeasonID    int       `json:"season_id"`
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	CarClass    string    `json:"car_class"`
	Disciplines []string  `json:"disciplines"`
	Completed   bool      `json:"completed"`
}

// RaceResult представляет результаты гонщика в гонке
type RaceResult struct {
	ID          int            `json:"id"`
	RaceID      int            `json:"race_id"`
	DriverID    int            `json:"driver_id"`
	CarNumber   int            `json:"car_number"`
	CarName     string         `json:"car_name"`
	CarPhotoURL string         `json:"car_photo_url"`
	Results     map[string]int `json:"results"` // discipline -> place
	TotalScore  int            `json:"total_score"`
}

// SerializeDisciplines сериализует список дисциплин в JSON
func SerializeDisciplines(disciplines []string) (string, error) {
	jsonData, err := json.Marshal(disciplines)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// DeserializeDisciplines десериализует список дисциплин из JSON
func DeserializeDisciplines(data string) ([]string, error) {
	var disciplines []string
	err := json.Unmarshal([]byte(data), &disciplines)
	if err != nil {
		return nil, err
	}
	return disciplines, nil
}

// SerializeResults сериализует карту результатов в JSON
func SerializeResults(results map[string]int) (string, error) {
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// DeserializeResults десериализует карту результатов из JSON
func DeserializeResults(data string) (map[string]int, error) {
	var results map[string]int
	err := json.Unmarshal([]byte(data), &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// DriverStats представляет статистику гонщика
type DriverStats struct {
	TotalScore   int             `json:"total_score"`
	RecentRaces  []RaceScorePair `json:"recent_races"`
	TotalRaces   int             `json:"total_races"`
	Achievements []Achievement   `json:"achievements"`
}

// RaceScorePair пара гонка-счет
type RaceScorePair struct {
	RaceName string `json:"race_name"`
	Score    int    `json:"score"`
}

// Achievement достижение гонщика
type Achievement struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DateEarned  string `json:"date_earned"`
}

// UserState представляет состояние пользователя в цепочке диалога
type UserState struct {
	State       string                 `json:"state"`
	ContextData map[string]interface{} `json:"context_data"`
}

// Константы с перечислением всех возможных дисциплин
var DefaultDisciplines = []string{
	"Визуал",
	"Драг",
	"Круговая гонка",
	"Офроад",
	"Гонка от А к Б",
	"Ралли",
}
