package main

import (
	"log"
	"net/http"

	"github.com/mmirzaei559/taskmanager/database"
	"github.com/mmirzaei559/taskmanager/handlers"
)

func main() {
	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Set up routes
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTasks(w, r)
		case http.MethodPost:
			handlers.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/tasks/update", handlers.UpdateTask)
	http.HandleFunc("/api/benchmark", handlers.Benchmark)

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
