package api

import (
	"first-version/pkg/db"
	"log"
	"net/http"
)

const tasksLimit = 50

// TasksResp определяет структуру JSON-ответа со списком задач.
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// TasksHandler обрабатывает НТТР-запрос на получение списка ближайших задач.
func (ts *TaskServer) TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{
			Error: "метод не поддерживается"})
		return
	}

	tasks, err := db.Tasks(r.Context(), ts.DB, tasksLimit)
	if err != nil {
		log.Printf("TasksHandler error: %v", err)
		writeJSON(w, http.StatusInternalServerError, jsonResponse{
			Error: err.Error(),
		})
		return
	}
	
	if tasks == nil {
		tasks = []*db.Task{}
	}	

	writeJSON(w, http.StatusOK, TasksResp{
		Tasks: tasks,
	})
}
