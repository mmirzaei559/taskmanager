package database

import (
	"fmt"
	"log"

	"github.com/mmirzaei559/taskmanager/models"
)

func GetAllTasks() ([]models.Task, error) {
	query := "SELECT id, title, description, completed, created_at FROM tasks"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func CreateTask(title, description string) (int64, error) {
	result, err := DB.Exec("INSERT INTO tasks (title, description) VALUES (?, ?)", title, description)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateTaskStatus(id int, completed bool) error {
	_, err := DB.Exec("UPDATE tasks SET completed = ? WHERE id = ?", completed, id)
	return err
}

// Benchmark function
func BenchmarkTasks(count int) error {
	// Start transaction
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	// Prepare statement for faster inserts
	stmt, err := tx.Prepare("INSERT INTO tasks (title, description, completed) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := 0; i < count; i++ {
		title := fmt.Sprintf("Task %d", i)
		description := fmt.Sprintf("Description for task %d", i)
		completed := i%2 == 0

		_, err := stmt.Exec(title, description, completed)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
