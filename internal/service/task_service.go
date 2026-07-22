package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/repository"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrInvalidStatus = errors.New("status must be one of: todo, in_progress, done")
)

var validStatuses = map[string]bool{
	"todo":        true,
	"in_progress": true,
	"done":        true,
}

type TaskService struct {
	taskRepo    *repository.TaskRepository
	subtaskRepo *repository.SubtaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository, subtaskRepo *repository.SubtaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo, subtaskRepo: subtaskRepo}
}

func (s *TaskService) CreateTask(ctx context.Context, t *model.Task) error {
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		return ErrTitleRequired
	}

	if t.Status == "" {
		t.Status = "todo"
	}
	if !validStatuses[t.Status] {
		return ErrInvalidStatus
	}

	return s.taskRepo.Create(ctx, t)
}

func (s *TaskService) GetTask(ctx context.Context, id int) (*model.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	return s.taskRepo.List(ctx, filter)
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, id int, status string) error {
	if !validStatuses[status] {
		return ErrInvalidStatus
	}

	var completedAt *string
	if status == "done" {
		now := time.Now().UTC().Format(time.RFC3339)
		completedAt = &now
	}

	if err := s.taskRepo.UpdateStatus(ctx, id, status, completedAt); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	return s.taskRepo.Delete(ctx, id)
}

func (s *TaskService) syncParentTaskStatus(ctx context.Context, taskID int) error {
	total, done, err := s.subtaskRepo.CountByTaskID(ctx, taskID)
	if err != nil {
		return err
	}

	if total == 0 {
		return nil
	}

	if total == done {
		return s.UpdateTaskStatus(ctx, taskID, "done")
	}
	return s.UpdateTaskStatus(ctx, taskID, "in_progress")
}
