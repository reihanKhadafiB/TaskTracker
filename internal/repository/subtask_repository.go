package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/task-tracker/internal/model"
)

type SubtaskRepository struct {
	pool *pgxpool.Pool
}

func NewSubtaskRepository(pool *pgxpool.Pool) *SubtaskRepository {
	return &SubtaskRepository{pool: pool}
}

func (r *SubtaskRepository) Create(ctx context.Context, s *model.Subtask) error {
	query := `
		INSERT INTO subtasks (task_id, title, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, s.TaskID, s.Title, s.SortOrder).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create subtask: %w", err)
	}
	return nil
}

func (r *SubtaskRepository) ListByTaskID(ctx context.Context, taskID int) ([]model.Subtask, error) {
	query := `
		SELECT id, task_id, title, is_done, sort_order, created_at
		FROM subtasks
		WHERE task_id = $1
		ORDER BY sort_order ASC, id ASC
	`
	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subtasks: %w", err)
	}
	defer rows.Close()

	var subtasks []model.Subtask
	for rows.Next() {
		var s model.Subtask
		if err := rows.Scan(&s.ID, &s.TaskID, &s.Title, &s.IsDone, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan subtask row: %w", err)
		}
		subtasks = append(subtasks, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subtask rows: %w", err)
	}
	return subtasks, nil
}

func (r *SubtaskRepository) UpdateDone(ctx context.Context, id int, isDone bool) (taskID int, err error) {
	query := `UPDATE subtasks SET is_done = $1 WHERE id = $2 RETURNING task_id`
	err = r.pool.QueryRow(ctx, query, isDone, id).Scan(&taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to update subtask: %w", err)
	}
	return taskID, nil
}

func (r *SubtaskRepository) CountByTaskID(ctx context.Context, taskID int) (total int, done int, err error) {
	query := `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_done = true)
		FROM subtasks
		WHERE task_id = $1
	`
	err = r.pool.QueryRow(ctx, query, taskID).Scan(&total, &done)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count subtasks: %w", err)
	}
	return total, done, nil
}
