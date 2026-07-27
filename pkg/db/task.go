package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrTaskNotFound возвращается, если задача отсутствует в базе данных.
var ErrTaskNotFound = errors.New("задача не найдена")

// Task описывает структуру задачи в планировщике.
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет новую задачу в таблицу scheduler и возвращает её ID.
func AddTask(ctx context.Context, database *sql.DB, task *Task) (int64, error) {
	if database == nil {
		return 0, fmt.Errorf("подключение к базе данных отсутствует")
	}

	if task == nil {
		return 0, fmt.Errorf("задача не может быть nil")
	}

	const query = `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	result, err := database.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления задачи: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения идентификатора задачи: %w", err)
	}
	return id, nil
}

// Tasks возвращает список задач с ограничением по количеству.
func Tasks(ctx context.Context, database *sql.DB, limit int) ([]*Task, error) {
	if database == nil {
		return nil, fmt.Errorf("подключение к базе данных отсутствует")
	}

	if limit <= 0 {
		return nil, fmt.Errorf("лимит должен быть больше нуля")
	}

	const query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC, id ASC LIMIT ?`

	rows, err := database.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка задач: %w", err)
	}
	defer rows.Close()

	tasks := make([]*Task, 0)

	for rows.Next() {
		task := &Task{}

		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка чтения задачи из базы данных: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка перебора списка задач: %w", err)
	}
	return tasks, nil
}

// GetTask возвращает задачу по её уникальному идентификатору.
func GetTask(ctx context.Context, database *sql.DB, id string) (*Task, error) {
	if database == nil {
		return nil, fmt.Errorf("подключение к базе данных отсутствует")
	}

	if id == "" {
		return nil, fmt.Errorf("не указан идентификатор")
	}

	const query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	task := &Task{}

	err := database.QueryRowContext(ctx, query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("ошибка получения задачи: %w", err)
	}
	return task, nil
}

// UpdateTask обновляет все поля существующей задачи.
func UpdateTask(ctx context.Context, database *sql.DB, task *Task) error {
	if database == nil {
		return fmt.Errorf("подключение к базе данных отсутствует")
	}
	if task == nil {
		return fmt.Errorf("задача не может быть nil")
	}
	if task.ID == "" {
		return fmt.Errorf("не указан идентификатор")
	}
	const query = `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	result, err := database.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления задачи: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата обновления: %w", err)
	}
	if count == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// UpdateDate обновляет только дату выполнения задачи.
func UpdateDate(ctx context.Context, database *sql.DB, nextDate string, id string) error {
	if database == nil {
		return fmt.Errorf("подключение к базе данных отсутствует")
	}
	if id == "" {
		return fmt.Errorf("не указан идентификатор")
	}
	if nextDate == "" {
		return fmt.Errorf("не указана новая дата")
	}
	const query = `UPDATE scheduler SET date = ? WHERE id = ?`

	result, err := database.ExecContext(ctx, query, nextDate, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления даты задачи: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата обновления даты: %w", err)
	}

	if count == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// DeleteTask удаляет задачу по её идентификатору.
func DeleteTask(ctx context.Context, database *sql.DB, id string) error {
	if database == nil {
		return fmt.Errorf("подключение к базе данных отсутсвует")
	}
	if id == "" {
		return fmt.Errorf("не указан идентификатор")
	}

	const query = `DELETE FROM scheduler WHERE id = ?`

	result, err := database.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления задачи: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения результата удаления: %w", err)
	}

	if count == 0 {
		return ErrTaskNotFound
	}
	return nil
}
