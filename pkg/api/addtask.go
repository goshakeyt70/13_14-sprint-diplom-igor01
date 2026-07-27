package api

import (
	"encoding/json"
	"errors"
	"first-version/pkg/db"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

// jsonResponse предоставляет стандартную структуру ответа API с идентификатором
// или сообщением об ошибке.
type jsonResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// AddHandler обрабатывает запрос на добавление новой задачи.
func (ts *TaskServer) AddHandler(w http.ResponseWriter, r *http.Request) {	
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{Error: "Метод не поддерживается"})
		return
	}
	
	var task db.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonResponse{Error: fmt.Sprintf("ошибка чтения JSON: %v", err)})
		return
	}

	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, jsonResponse{Error: "не указан заголовок задачи"})
		return
	}

	task.Date = strings.TrimSpace(task.Date)
	task.Repeat = strings.TrimSpace(task.Repeat)

	if err := checkDate(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonResponse{Error: err.Error()})
		return
	}

	id, err := db.AddTask(r.Context(), ts.DB, &task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, jsonResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, jsonResponse{ID: strconv.FormatInt(id, 10)})
}

// checkDate проверяет корректность даты и правила повторения, приводя прошедшие даты
// к текущим или будущим.
func checkDate(task *db.Task) error {
	if task == nil {
		return errors.New("задача не может быть nil")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if task.Date == "" {
		task.Date = today.Format(dateFormat)
	}

	parsedDate, err := time.ParseInLocation(dateFormat, task.Date, today.Location())
	if err != nil {
		return fmt.Errorf("дата должна быть указана в формате %s: %w", dateFormat, err)
	}

	var nextDate string
	if task.Repeat != "" {
		nextDate, err = NextDate(today, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("некорректное правило повторения: %w", err)
		}
	}

	if parsedDate.Before(today) {
		if task.Repeat == "" {
			task.Date = today.Format(dateFormat)
		} else {
			task.Date = nextDate
		}
	}

	return nil
}

// writeJSON сериализует данные в формате JSON, устанавливает заголовок Content-Type
// и отправляет НТТР-статус.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
