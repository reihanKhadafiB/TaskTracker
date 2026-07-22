package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/task-tracker/internal/model"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, t *model.Task) error {
	query := `
		INSERT INTO tasks (context_id, title, description, status, due_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		t.ContextID, t.Title, t.Description, t.Status, t.DueDate,
	).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int) (*model.Task, error) {
	query := `
		SELECT id, context_id, title, description, status, due_date, created_at, completed_at
		FROM tasks
		WHERE id = $1
	`
	var t model.Task
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.ContextID, &t.Title, &t.Description,
		&t.Status, &t.DueDate, &t.CreatedAt, &t.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task by id %d: %w", id, err)
	}
	return &t, nil
}

func (r *TaskRepository) List(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	query := `
		SELECT id, context_id, title, description, status, due_date, created_at, completed_at
		FROM tasks
		WHERE 1=1
	`
	args := []any{}
	argPos := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, filter.Status)
		argPos++
	} else {
		query += " AND status != 'done'"
	}

	if filter.ContextID != nil {
		query += fmt.Sprintf(" AND context_id = $%d", argPos)
		args = append(args, *filter.ContextID)
		argPos++
	}

	if filter.Overdue {
		query += fmt.Sprintf(" AND due_date < CURRENT_DATE AND status != $%d", argPos)
		args = append(args, "done")
		argPos++
	}

	if filter.Cursor != nil {
		query += fmt.Sprintf(" AND id < $%d", argPos)
		args = append(args, *filter.Cursor)
		argPos++
	}

	query += " ORDER BY id DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks with filter: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.ContextID, &t.Title, &t.Description,
			&t.Status, &t.DueDate, &t.CreatedAt, &t.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id int, status string, completedAt *string) error {
	query := `
		UPDATE tasks
		SET status = $1, completed_at = $2
		WHERE id = $3
	`
	tag, err := r.pool.Exec(ctx, query, status, completedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task with id %d not found", id)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM tasks WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task with id %d not found", id)
	}
	return nil
}
