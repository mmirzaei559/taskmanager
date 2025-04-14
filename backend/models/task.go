package models

import "time"

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	ClientIP    string    `json:"client_ip"`
}

type TaskResult struct {
	Task    Task   `json:"task"`
	TaskID  int64  `json:"task_id,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
