package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/mmirzaei559/taskmanager/database"
	"github.com/mmirzaei559/taskmanager/models"
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

// handlers/tasks.go
func ProcessTasksConcurrently(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tasks []models.Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Channel for results
	results := make(chan models.TaskResult, len(tasks))
	var wg sync.WaitGroup

	// Process tasks concurrently
	for _, task := range tasks {
		wg.Add(1)
		go processTask(task, results, &wg)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var response []models.TaskResult
	for result := range results {
		response = append(response, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// chan<- (send-only)
func processTask(task models.Task, results chan<- models.TaskResult, wg *sync.WaitGroup) {
	defer wg.Done()

	result := models.TaskResult{Task: task}

	// Simulate processing delay
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	// Save to database
	id, err := database.CreateTask(task.Title, task.Description)
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Success = true
		result.TaskID = id
	}

	results <- result
}
