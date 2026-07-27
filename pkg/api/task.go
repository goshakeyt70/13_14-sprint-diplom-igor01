package api

import (
	"encoding/json"
	"errors"
	"first-version/pkg/db"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// TaskHandler маршрутизирует НТТР-запросы к задаче в зависимости от метода запроса.
func (ts *TaskServer) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ts.AddHandler(w, r)
	case http.MethodGet:
		ts.GetTaskHandler(w, r)
	case http.MethodPut:
		ts.UpdateTaskHandler(w, r)
	case http.MethodDelete:
		ts.DeleteTaskHandler(w, r)
	default:
		w.Header().Set("Allow", "GET, POST, PUT, DELETE")
		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{
			Error: "метод не поддерживается"})
	}
}

// GetTaskHandler обрабатывает запрос на получение задачи по её идентификатору.
func (ts *TaskServer) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: "не указан идентификатор"})
		return
	}

	task, err := db.GetTask(r.Context(), ts.DB, id)
	if err != nil {
		writeTaskError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// UpdateTaskHandler обрабатывает запрос на обновление параметров существующей задачи.
func (ts *TaskServer) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: fmt.Sprintf("ошибка чтения JSON: %v", err)})
		return
	}
	task.ID = strings.TrimSpace(task.ID)
	if task.ID == "" {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: "не указан идентификатор"})
		return
	}
	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: "не указан заголовок задади"})
		return
	}
	task.Date = strings.TrimSpace(task.Date)
	task.Repeat = strings.TrimSpace(task.Repeat)

	if err := checkDate(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: err.Error()})
		return
	}
	if err := db.UpdateTask(r.Context(), ts.DB, &task); err != nil {
		writeTaskError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

// DeleteTaskHandler обрабатывает запрос на удаление задачи по её идентификатору.
func (ts *TaskServer) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: "не указан идентификатор"})
		return
	}
	if err := db.DeleteTask(r.Context(), ts.DB, id); err != nil {
		writeTaskError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

// writeTaskError преобразует внутренние ошибки слоя базы данных в ответы API.
func writeTaskError(w http.ResponseWriter, err error) {
	if errors.Is(err, db.ErrTaskNotFound) {
		writeJSON(w, http.StatusNotFound, jsonResponse{
			Error: "задачи не найдена"})
		return
	}
	log.Printf("task API error: %v", err)
	writeJSON(w, http.StatusInternalServerError, jsonResponse{
		Error: err.Error()})
}
