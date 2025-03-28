package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mmirzaei559/taskmanager/backend/database"
	"github.com/mmirzaei559/taskmanager/backend/models"
)

func GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := database.GetAllTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := database.CreateTask(task.Title, task.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.UpdateTaskStatus(task.ID, task.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Benchmark(w http.ResponseWriter, r *http.Request) {
	count := 1000 // default
	if r.URL.Query().Get("count") != "" {
		_, err := fmt.Sscanf(r.URL.Query().Get("count"), "%d", &count)
		if err != nil {
			http.Error(w, "Invalid count parameter", http.StatusBadRequest)
			return
		}
	}

	err := database.BenchmarkTasks(count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Inserted %d tasks for benchmarking", count)))
}
