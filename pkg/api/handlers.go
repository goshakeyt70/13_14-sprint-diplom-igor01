package api

import (
	"database/sql"
	"net/http"
	"time"
)

// TaskServer представляет НТТР-сервер для управления задачами.
type TaskServer struct {
	DB *sql.DB
}

// NewTaskServer создает и возвращает новый экземпляр TaskServer.
func NewTaskServer(database *sql.DB) *TaskServer {
	return &TaskServer{
		DB: database,
	}
}

// NextDateHandler обрабатывает запрос на вычисление следующей даты задачи.
func (ts *TaskServer) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	if nowValue := r.FormValue("now"); nowValue != "" {
		parsedNow, err := time.Parse(dateFormat, nowValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		now = parsedNow
	}
	nextDate, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(nextDate))
}
