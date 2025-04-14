package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/mmirzaei559/taskmanager/database"
	"github.com/mmirzaei559/taskmanager/middleware"
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
	clientIP := middleware.GetClientIP(r)
	log.Printf("Bulk tasks from IP: %s", clientIP)

	startTime := time.Now()
	log.Printf("🚀 Bulk processing started at %v", startTime.Format("15:04:05.000"))

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tasks []models.Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("📦 Received %d tasks for processing", len(tasks))

	results := make(chan models.TaskResult, len(tasks))
	var wg sync.WaitGroup

	// Process tasks concurrently
	for i, task := range tasks {
		wg.Add(1)
		go func(taskNum int, t models.Task) {
			start := time.Now()
			taskID := taskNum + 1 // Just for logging
			log.Printf("🛫 Goroutine %d started processing task '%s'", taskID, t.Title)

			defer func() {
				wg.Done()
				log.Printf("🛬 Goroutine %d completed after %v", taskID, time.Since(start))
			}()

			result := models.TaskResult{Task: t}

			// Simulate variable processing time
			delay := time.Duration(rand.Intn(1000)) * time.Millisecond
			log.Printf("⏳ Goroutine %d working for %v", taskID, delay)
			time.Sleep(delay)

			// Save to database
			id, err := database.CreateTask(t.Title, t.Description)
			if err != nil {
				log.Printf("⚠️ Goroutine %d failed: %v", taskID, err)
				result.Error = err.Error()
			} else {
				log.Printf("✅ Goroutine %d saved task ID %d", taskID, id)
				result.Success = true
				result.TaskID = id
			}

			results <- result
		}(i, task) // Pass current index and task
	}

	// Close channel when done
	go func() {
		wg.Wait()
		close(results)
		log.Printf("🔌 All goroutines completed, closing results channel")
	}()

	// Collect results
	var response []models.TaskResult
	for result := range results {
		response = append(response, result)
	}

	log.Printf("🏁 Completed processing %d tasks in %v", len(response), time.Since(startTime))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
