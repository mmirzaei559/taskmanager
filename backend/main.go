package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmirzaei559/taskmanager/database"
	"github.com/mmirzaei559/taskmanager/handlers"
	"github.com/rs/cors"
)

func main() {
	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Create router
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetTasks(w, r)
		case http.MethodPost:
			handlers.CreateTask(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/tasks/update", handlers.UpdateTask)
	mux.HandleFunc("/api/benchmark", handlers.Benchmark)
	mux.HandleFunc("/api/tasks/bulk", handlers.ProcessTasksConcurrently)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   getOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours in seconds (12*60*60)
	})

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler.Handler(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func getOrigins() []string {
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		return strings.Split(envOrigins, ",")
	}
	return []string{"http://localhost:5173"} // Default to Vite dev server
}
