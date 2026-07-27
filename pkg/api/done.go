package api

import (
	"errors"
	"first-version/pkg/db"
	"log"
	"net/http"
	"strings"
	"time"
)

// DoneHandler обрабатывает запрос на выполнение задачи, удаляя её или перенося на
// следующую дату.
func (ts *TaskServer) DoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)

		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{
			Error: "метод не поддерживается"})
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))

	if id == "" {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: "не указан идентификатор"})
		return
	}

	task, err := db.GetTask(r.Context(), ts.DB, id)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			writeJSON(w, http.StatusNotFound, jsonResponse{
				Error: "задача не найдена"})
			return
		}
		log.Printf("DoneHandler GetTask error: %v", err)

		writeJSON(w, http.StatusInternalServerError, jsonResponse{
			Error: err.Error()})
		return
	}

	if strings.TrimSpace(task.Repeat) == "" {
		if err := db.DeleteTask(r.Context(), ts.DB, id); err != nil {
			writeTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, struct{}{})
		return
	}

	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, jsonResponse{
			Error: err.Error()})
		return
	}

	if err := db.UpdateDate(r.Context(), ts.DB, nextDate, id); err != nil {
		writeTaskError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
