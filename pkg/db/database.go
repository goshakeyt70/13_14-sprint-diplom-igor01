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

// Константы со схемами БД приведены к стандартам SQLite
const tableSchema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	data TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat TEXT NOT NULL DEFAULT ''
);`

const indexSchema = `
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(data);`

// InitDatabase инициализирует подключение к СУБД SQLite, проверяет связь
// автоматически создает необходимую структуру таблиц и настраивает пул соединений.
func InitDatabase(dbFile string) (*sql.DB, error) {
	if dbFile == "" {
		return nil, fmt.Errorf("путь к файлу БД не может быть пустым")
	}

	// Автоматически создаем директорию для базы данных, если её нет.
	dir := filepath.Dir(dbFile)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("не удалось создать директорию для БД: %w", err)
		}
	}

	// Открываем дескриптер базы данных.
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации драйвера sqlite: %w", err)
	}

	// Проверяем фактическую доступность файла.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("база данных недоступна: %w", err)
	}

	// Настройка пула соединений (Connection Pool).
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(10 * time.Minute)

	// Декларативное создание таблиц.
	if _, err := db.Exec(tableSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка создания таблицы scheduler: %w", err)
	}

	// Деклативное создание индексов.
	if _, err := db.Exec(indexSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка создания индексов для таблиц scheduler: %w", err)
	}

	log.Printf("[DATABASE] База данных успешно инициализирована по пути: %s", dbFile)
	return db, nil
}
