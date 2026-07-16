package api

import (
	"database/sql"
	"fmt"
	"net/http"
)

// TaskServer - это структура-контейнер (сервер задач)
// Она инкапсулирует в себя пул соединений с БД, чтобы методы-обработчики
// имели к нему прямой доступ без использования глобальных переменных.
type TaskServer struct {
	DB *sql.DB
}

// NewTaskServer - функция-конструктор. Main создаёт базу, server передаёт её сюда,
// а мы возвращаем готовый объект для обслуживания НТТР-запросов.
func NewTaskServer(db *sql.DB) *TaskServer {
	return &TaskServer{
		DB: db,
	}
}

// AddHandler обрабатывает POST-запросы на добавление задач (/api/task).
// Пока это заглушка, возвращающая текстовый маркер.
func (ts *TaskServer) AddHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "[STUB] Обработчик AddHandler успешно вызван. Логика добавления задачи будет здесь.")
}

// NextDateHandler обрабатывает GET-запросы для вычисления дат (/api/nextdate).
func (ts *TaskServer) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "[STUB] Обработчик NextDateHandler успешно вызван. Логика расчета дат будет здесь.")
}
