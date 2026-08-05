package model

import "time"

type Context struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID          int        `json:"id"`
	ContextID   *int       `json:"context_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	DueDate     *Date      `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type Subtask struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Title     string    `json:"title"`
	IsDone    bool      `json:"is_done"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}
