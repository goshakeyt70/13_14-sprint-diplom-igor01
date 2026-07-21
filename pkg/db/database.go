package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// tableSchema содержит SQL-запрос для инициализации таблицы задач scheduler.
const tableSchema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title VARCHAR(255) NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat VARCHAR(128) NOT NULL DEFAULT ''
);`

// indexSchema содержит SQL-запрос для создания индекса по полю date.
const indexSchema = `
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);`

// InitDatabase инициализирует подключение к SQLite и создает структуру таблиц.
func InitDatabase(dbFile string) (*sql.DB, error) {
	if dbFile == "" {
		return nil, fmt.Errorf("путь к файлу БД не может быть пустым")
	}

	dir := filepath.Dir(dbFile)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("не удалось создать директорию для БД: %w", err)
		}
	}

	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия SQLite: %w", err)
	}

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("база данных недоступна: %w", err)
	}

	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(10 * time.Minute)

	if _, err := database.Exec(tableSchema); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ошибка создания таблицы scheduler: %w", err)
	}

	if _, err := database.Exec(indexSchema); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ошибка создания индекса scheduler: %w", err)
	}

	log.Printf("[DATABASE] База данных успешно инициализирована по пути: %s", dbFile)
	return database, nil
}
