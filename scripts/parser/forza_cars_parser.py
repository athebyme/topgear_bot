#!/usr/bin/env python3
import os
import time
import requests
import psycopg2
from psycopg2.extras import execute_values
from bs4 import BeautifulSoup
import re
import logging

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger('cars_parser')

# PostgreSQL соединение
DB_USER = os.environ.get('POSTGRES_USER', 'forza')
DB_PASSWORD = os.environ.get('POSTGRES_PASSWORD', 'forza_password')
DB_NAME = os.environ.get('POSTGRES_DB', 'forza_db')
DB_HOST = os.environ.get('POSTGRES_HOST', 'localhost')
DB_PORT = os.environ.get('POSTGRES_PORT', '5432')

# URL страницы с машинами
CARS_URL = "https://forza.fandom.com/wiki/Forza_Horizon_4/Cars"

def get_db_connection():
    """Создает подключение к базе данных."""
    try:
        conn = psycopg2.connect(
            user=DB_USER,
            password=DB_PASSWORD,
            host=DB_HOST,
            port=DB_PORT,
            database=DB_NAME
        )
        return conn
    except Exception as e:
        logger.error(f"Ошибка подключения к базе данных: {e}")
        raise

def create_cars_table(conn):
    """Создает таблицу для хранения данных о машинах."""
    try:
        with conn.cursor() as cursor:
            cursor.execute("""
            CREATE TABLE IF NOT EXISTS cars (
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
                class_letter VARCHAR(50),
                class_number INTEGER,
                source VARCHAR(100)
            )
            """)
        conn.commit()
        logger.info("Таблица cars создана или уже существует")
    except Exception as e:
        logger.error(f"Ошибка при создании таблицы: {e}")
        conn.rollback()
        raise

def parse_cars():
    """Parses car data from the Forza Horizon 4 wiki page."""
    try:
        response = requests.get(CARS_URL, timeout=30)
        response.raise_for_status()

        soup = BeautifulSoup(response.text, 'html.parser')
        cars_data = []

        # Find the table with cars - updated class selector
        car_table = soup.find('table', class_='sortable')
        if not car_table:
            logger.error("Не удалось найти таблицу с машинами")
            return []

        # Get table rows (skip header)
        rows = car_table.find_all('tr')[1:]
        logger.info(f"Найдено {len(rows)} строк с машинами")

        for row in rows:
            try:
                # Get all cells in the row
                cells = row.find_all('td')
                if len(cells) < 11:
                    continue

                # Extract image
                img_cell = cells[0]
                a_tag = img_cell.find('a', class_='mw-file-description image')
                img_url = a_tag['href'] if a_tag else None

                # Extract name and year
                name_cell = cells[1]
                name_link = name_cell.find('a')
                # Updated selector to get the car name div
                name_div = name_cell.find('div', style=lambda s: s and "line-height: 18px" in s)
                full_name = name_div.text.strip() if name_div else name_cell.text.strip()

                # Extract year
                year_match = re.search(r'\b(19|20)\d{2}\b', full_name)
                year = int(year_match.group(0)) if year_match else None
                name = full_name.replace(str(year), '').strip() if year else full_name

                # Extract source (Autoshow, etc)
                source_div = name_cell.find('div', style=lambda s: s and "font-size: smaller" in s)
                source_text = source_div.text.strip() if source_div else "Unknown"

                # Extract price and rarity
                price_cell = cells[5]
                price_div = price_cell.find('div', style=lambda s: s and "line-height: 18px" in s)
                price_text = price_div.text.strip() if price_div else ""
                price = int(re.sub(r'[^\d]', '', price_text)) if re.search(r'\d', price_text) else 0

                rarity_span = price_cell.find('span', style=lambda s: s and "background-color" in s)
                rarity = rarity_span.text.strip() if rarity_span else "Unknown"

                # Extract performance stats
                speed = float(cells[6].text.strip()) if cells[6].text.strip() else 0
                handling = float(cells[7].text.strip()) if cells[7].text.strip() else 0
                acceleration = float(cells[8].text.strip()) if cells[8].text.strip() else 0
                launch = float(cells[9].text.strip()) if cells[9].text.strip() else 0
                braking = float(cells[10].text.strip()) if cells[10].text.strip() else 0

                # Extract class
                class_cell = cells[11] if len(cells) > 11 else None
                class_letter = None
                class_number = None

                if class_cell:
                    # Find the class letter and number from the spans
                    letter_span = class_cell.find('span', style=lambda s: s and "background-color" in s)
                    number_span = class_cell.find('span', style=lambda s: s and "border:" in s)

                    if letter_span:
                        class_letter = letter_span.text.strip()
                    if number_span:
                        class_number_text = number_span.text.strip()
                        class_number = int(re.sub(r'[^\d]', '', class_number_text)) if re.search(r'\d', class_number_text) else 0

                car_data = {
                    'name': name,
                    'year': year,
                    'image_url': img_url,
                    'price': price,
                    'rarity': rarity,
                    'speed': speed,
                    'handling': handling,
                    'acceleration': acceleration,
                    'launch': launch,
                    'braking': braking,
                    'class_letter': class_letter,
                    'class_number': class_number,
                    'source': source_text
                }

                cars_data.append(car_data)
                logger.debug(f"Обработана машина: {name} ({year})")

            except Exception as e:
                logger.error(f"Ошибка при обработке строки: {e}")
                continue

        logger.info(f"Успешно обработано {len(cars_data)} машин")
        return cars_data

    except requests.RequestException as e:
        logger.error(f"Ошибка при запросе к сайту: {e}")
        return []
    except Exception as e:
        logger.error(f"Непредвиденная ошибка при парсинге: {e}")
        return []

def save_cars_to_db(conn, cars_data):
    """Сохраняет данные о машинах в базу данных."""
    if not cars_data:
        logger.warning("Нет данных для сохранения")
        return

    try:
        with conn.cursor() as cursor:
            # Очищаем таблицу перед вставкой новых данных
            cursor.execute("TRUNCATE TABLE cars RESTART IDENTITY CASCADE")

            # Подготавливаем данные для вставки
            columns = cars_data[0].keys()
            values = [[car[column] for column in columns] for car in cars_data]

            # Вставляем данные
            query = f"""
            INSERT INTO cars ({', '.join(columns)})
            VALUES %s
            """
            execute_values(cursor, query, values)

        conn.commit()
        logger.info(f"Сохранено {len(cars_data)} машин в базу данных")
    except Exception as e:
        logger.error(f"Ошибка при сохранении данных: {e}")
        conn.rollback()
        raise

def main():
    """Основная функция программы."""
    logger.info("Запуск парсера машин Forza Horizon 4")

    # Повторяем попытку подключения к базе данных, пока PostgreSQL не запустится
    max_retries = 5
    retry_delay = 5  # секунды

    for i in range(max_retries):
        try:
            conn = get_db_connection()
            break
        except Exception as e:
            if i < max_retries - 1:
                logger.warning(f"Попытка {i+1}/{max_retries} подключения к БД не удалась. Повторная попытка через {retry_delay} сек...")
                time.sleep(retry_delay)
            else:
                logger.error("Не удалось подключиться к базе данных после нескольких попыток")
                return

    try:
        # Создаем таблицу для машин
        create_cars_table(conn)

        # Парсим и сохраняем данные
        cars_data = parse_cars()
        save_cars_to_db(conn, cars_data)

        logger.info("Парсинг машин успешно завершен")
    except Exception as e:
        logger.error(f"Произошла ошибка: {e}")
    finally:
        if conn:
            conn.close()
            logger.info("Соединение с базой данных закрыто")

if __name__ == "__main__":
    main()