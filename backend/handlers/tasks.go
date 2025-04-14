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

// Get client IP from request
func getClientIP(r *http.Request) string {
	// Check for forwarded headers
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
		log.Printf("❌ Error fetching tasks: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Fetched %d tasks", len(tasks))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("📨 New task request from IP: %s", clientIP)

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("❌ Invalid task data from %s: %v", clientIP, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := database.CreateTask(task.Title, task.Description, clientIP)
	if err != nil {
		log.Printf("⚠️ Failed to create task from %s: %v", clientIP, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task.ID = int(id)
	task.ClientIP = clientIP // Include IP in response

	log.Printf("➕ Created task #%d from %s", id, clientIP)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("🔄 Update request from IP: %s", clientIP)

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("❌ Invalid update data from %s: %v", clientIP, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.UpdateTaskStatus(task.ID, task.Completed)
	if err != nil {
		log.Printf("⚠️ Failed to update task #%d from %s: %v", task.ID, clientIP, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			log.Printf("❌ Invalid benchmark count from %s: %v", clientIP, err)
			http.Error(w, "Invalid count parameter", http.StatusBadRequest)
			return
		}
	}

	err := database.BenchmarkTasks(count, clientIP)
	if err != nil {
		log.Printf("⚠️ Benchmark failed from %s: %v", clientIP, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("⚡ Completed benchmark of %d tasks from %s", count, clientIP)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Inserted %d tasks for benchmarking", count)))
}

func ProcessTasksConcurrently(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	startTime := time.Now()
	log.Printf("🚀 Bulk processing started from %s at %v", clientIP, startTime.Format("15:04:05.000"))

	if r.Method != http.MethodPost {
		log.Printf("❌ Invalid method from %s", clientIP)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tasks []models.Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		log.Printf("❌ Invalid bulk data from %s: %v", clientIP, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("📦 Received %d tasks from %s", len(tasks), clientIP)

	results := make(chan models.TaskResult, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(taskNum int, t models.Task) {
			start := time.Now()
			taskID := taskNum + 1
			log.Printf("🛫 [%s] Goroutine %d processing '%s'", clientIP, taskID, t.Title)

			defer func() {
				wg.Done()
				log.Printf("🛬 [%s] Goroutine %d completed in %v", clientIP, taskID, time.Since(start))
			}()

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
			} else {
				log.Printf("✅ [%s] Goroutine %d saved task #%d", clientIP, taskID, id)
				result.Success = true
				result.TaskID = id
				result.Task.ClientIP = clientIP
			}

			results <- result
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

	log.Printf("🏁 [%s] Processed %d tasks in %v", clientIP, len(response), time.Since(startTime))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
