package service

import (
	"context"
	"errors"
	"strings"

	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/repository"
)

var ErrContextNameRequired = errors.New("context name is required")

type ContextService struct {
	contextRepo *repository.ContextRepository
}

func NewContextService(contextRepo *repository.ContextRepository) *ContextService {
	return &ContextService{contextRepo: contextRepo}
}

func (s *ContextService) CreateContext(ctx context.Context, c *model.Context) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return ErrContextNameRequired
	}
	return s.contextRepo.Create(ctx, c)
}

func (s *ContextService) ListContexts(ctx context.Context) ([]model.Context, error) {
	return s.contextRepo.List(ctx)
}

func (s *ContextService) UpdateContext(ctx context.Context, id int, name string, color string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrContextNameRequired
	}
	return s.contextRepo.Update(ctx, id, name, color)
}

func (s *ContextService) DeleteContext(ctx context.Context, id int) error {
	return s.contextRepo.Delete(ctx, id)
}
