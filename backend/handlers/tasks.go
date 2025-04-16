package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mmirzaei559/taskmanager/database"
	"github.com/mmirzaei559/taskmanager/models"
)

// writeErrorResponse handles error responses consistently
func writeErrorResponse(w http.ResponseWriter, err error) {
	var apiErr *models.APIError

	// Convert to APIError if not already one
	switch e := err.(type) {
	case *models.APIError:
		apiErr = e
	default:
		apiErr = &models.APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal Server Error",
			Details:    e.Error(),
		}
	}

	log.Printf("Error: %v", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.StatusCode)
	json.NewEncoder(w).Encode(apiErr)
}

// Get client IP from request
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	// Remove port if present
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := database.GetAllTasks()
	if err != nil {
		writeErrorResponse(w, fmt.Errorf("error fetching tasks: %w", err))
		return
	}

	log.Printf("✅ Fetched %d tasks", len(tasks))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		writeErrorResponse(w, fmt.Errorf("error encoding response: %w", err))
	}
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("📨 New task request from IP: %s", clientIP)

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request body",
			Details:    err.Error(),
		})
		return
	}

	if task.Title == "" {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Title is required",
		})
		return
	}

	id, err := database.CreateTask(task.Title, task.Description, clientIP)
	if err != nil {
		writeErrorResponse(w, fmt.Errorf("failed to create task: %w", err))
		return
	}

	task.ID = int(id)
	task.ClientIP = clientIP

	log.Printf("➕ Created task #%d from %s", id, clientIP)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		writeErrorResponse(w, fmt.Errorf("error encoding response: %w", err))
	}
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("🔄 Update request from IP: %s", clientIP)

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request body",
			Details:    err.Error(),
		})
		return
	}

	if task.ID == 0 {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Task ID is required",
		})
		return
	}

	err := database.UpdateTaskStatus(task.ID, task.Completed)
	if err != nil {
		writeErrorResponse(w, fmt.Errorf("failed to update task #%d: %w", task.ID, err))
		return
	}

	log.Printf("🆙 Updated task #%d from %s", task.ID, clientIP)
	w.WriteHeader(http.StatusOK)
}

func Benchmark(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("🏋️ Benchmark request from IP: %s", clientIP)

	count := 1000 // default
	if r.URL.Query().Get("count") != "" {
		_, err := fmt.Sscanf(r.URL.Query().Get("count"), "%d", &count)
		if err != nil {
			writeErrorResponse(w, &models.APIError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid count parameter",
				Details:    err.Error(),
			})
			return
		}
	}

	if count <= 0 || count > 10000 {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Count must be between 1 and 10000",
		})
		return
	}

	err := database.BenchmarkTasks(count, clientIP)
	if err != nil {
		writeErrorResponse(w, fmt.Errorf("benchmark failed: %w", err))
		return
	}

	log.Printf("⚡ Completed benchmark of %d tasks from %s", count, clientIP)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(fmt.Sprintf(`{"message":"Inserted %d tasks for benchmarking"}`, count))); err != nil {
		writeErrorResponse(w, fmt.Errorf("error writing response: %w", err))
	}
}

func ProcessTasksConcurrently(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	startTime := time.Now()
	log.Printf("🚀 Bulk processing started from %s at %v", clientIP, startTime.Format("15:04:05.000"))

	if r.Method != http.MethodPost {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}

	var tasks []models.Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request body",
			Details:    err.Error(),
		})
		return
	}

	if len(tasks) == 0 {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    "No tasks provided",
		})
		return
	}

	log.Printf("📦 Received %d tasks from %s", len(tasks), clientIP)

	results := make(chan models.TaskResult, len(tasks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var encounteredError bool

	for i, task := range tasks {
		// Check if we've already encountered an error in another goroutine
		mu.Lock()
		if encounteredError {
			mu.Unlock()
			break
		}
		mu.Unlock()

		wg.Add(1)
		go func(taskNum int, t models.Task) {
			defer wg.Done()

			start := time.Now()
			taskID := taskNum + 1
			log.Printf("🛫 [%s] Goroutine %d processing '%s'", clientIP, taskID, t.Title)

			result := models.TaskResult{Task: t}

			// Simulate processing
			delay := time.Duration(rand.Intn(1000)) * time.Millisecond
			log.Printf("⏳ [%s] Goroutine %d working for %v", clientIP, taskID, delay)
			time.Sleep(delay)

			// Save with client IP
			id, err := database.CreateTask(t.Title, t.Description, clientIP)
			if err != nil {
				log.Printf("⚠️ [%s] Goroutine %d failed: %v", clientIP, taskID, err)
				result.Error = err.Error()

				mu.Lock()
				encounteredError = true
				mu.Unlock()
			} else {
				log.Printf("✅ [%s] Goroutine %d saved task #%d", clientIP, taskID, id)
				result.Success = true
				result.TaskID = id
				result.Task.ClientIP = clientIP
			}

			results <- result
			log.Printf("🛬 [%s] Goroutine %d completed in %v", clientIP, taskID, time.Since(start))
		}(i, task)
	}

	// Close channel when done
	go func() {
		wg.Wait()
		close(results)
		log.Printf("🔌 [%s] All goroutines completed", clientIP)
	}()

	// Collect results
	var response []models.TaskResult
	for result := range results {
		response = append(response, result)
	}

	// Check if any errors occurred
	if encounteredError {
		writeErrorResponse(w, &models.APIError{
			StatusCode: http.StatusPartialContent,
			Message:    "Some tasks failed to process",
		})
		return
	}

	log.Printf("🏁 [%s] Processed %d tasks in %v", clientIP, len(response), time.Since(startTime))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeErrorResponse(w, fmt.Errorf("error encoding response: %w", err))
	}
}
