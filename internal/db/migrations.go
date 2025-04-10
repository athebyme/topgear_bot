package db

// migrations содержит SQL-запросы для создания таблиц в базе данных
var migrations = []string{
	// Таблица гонщиков
	`CREATE TABLE IF NOT EXISTS drivers (
		id SERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		photo_url TEXT
	)`,

	// Таблица сезонов
	`CREATE TABLE IF NOT EXISTS seasons (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		start_date TIMESTAMP NOT NULL,
		end_date TIMESTAMP,
		active BOOLEAN DEFAULT FALSE
	)`,

	// Таблица гонок
	`CREATE TABLE IF NOT EXISTS races (
		id SERIAL PRIMARY KEY,
		season_id INTEGER REFERENCES seasons(id),
		name VARCHAR(255) NOT NULL,
		date TIMESTAMP NOT NULL,
		car_class VARCHAR(10) NOT NULL,
		disciplines JSONB NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,

	// Таблица результатов гонок
	`CREATE TABLE IF NOT EXISTS race_results (
		id SERIAL PRIMARY KEY,
		race_id INTEGER REFERENCES races(id) ON DELETE CASCADE,
		driver_id INTEGER REFERENCES drivers(id),
		car_number INTEGER,
		car_name VARCHAR(255) NOT NULL,
		car_photo_url TEXT,
		results JSONB NOT NULL,
		total_score INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,

	// Таблица машин из игры Forza Horizon 4
	`CREATE TABLE IF NOT EXISTS cars (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		year INTEGER,
		image_url TEXT,
		price INTEGER,
		rarity VARCHAR(50),
		speed FLOAT,
		handling FLOAT,
		acceleration FLOAT,
		launch FLOAT,
		braking FLOAT,
		class_letter VARCHAR(5),
		class_number INTEGER,
		source VARCHAR(100)
	)`,

	// Таблица назначений машин для гонок
	`CREATE TABLE IF NOT EXISTS race_car_assignments (
		id SERIAL PRIMARY KEY,
		race_id INTEGER REFERENCES races(id) ON DELETE CASCADE,
		driver_id INTEGER REFERENCES drivers(id),
		car_id INTEGER REFERENCES cars(id),
		assignment_number INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,

	// Индексы
	`CREATE INDEX IF NOT EXISTS idx_races_season_id ON races(season_id)`,
	`CREATE INDEX IF NOT EXISTS idx_race_results_race_id ON race_results(race_id)`,
	`CREATE INDEX IF NOT EXISTS idx_race_results_driver_id ON race_results(driver_id)`,
	`CREATE INDEX IF NOT EXISTS idx_cars_class_letter ON cars(class_letter)`,
	`CREATE INDEX IF NOT EXISTS idx_race_car_assignments_race_id ON race_car_assignments(race_id)`,
	`CREATE INDEX IF NOT EXISTS idx_race_car_assignments_driver_id ON race_car_assignments(driver_id)`,
}
