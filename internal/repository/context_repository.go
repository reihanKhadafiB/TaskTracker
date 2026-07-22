package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/task-tracker/internal/model"
)

type ContextRepository struct {
	pool *pgxpool.Pool
}

func NewContextRepository(pool *pgxpool.Pool) *ContextRepository {
	return &ContextRepository{pool: pool}
}

func (r *ContextRepository) Create(ctx context.Context, c *model.Context) error {
	query := `
		INSERT INTO contexts (name, color)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, c.Name, c.Color).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}
	return nil
}

func (r *ContextRepository) List(ctx context.Context) ([]model.Context, error) {
	query := `SELECT id, name, color, created_at FROM contexts ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list contexts: %w", err)
	}
	defer rows.Close()

	var contexts []model.Context
	for rows.Next() {
		var c model.Context
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan context row: %w", err)
		}
		contexts = append(contexts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating context rows: %w", err)
	}

	return contexts, nil
}
