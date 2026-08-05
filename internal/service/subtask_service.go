package service

import (
	"context"
	"strings"

	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/repository"
)

type SubtaskService struct {
	subtaskRepo *repository.SubtaskRepository
	taskService *TaskService
}

func NewSubtaskService(subtaskRepo *repository.SubtaskRepository, taskService *TaskService) *SubtaskService {
	return &SubtaskService{subtaskRepo: subtaskRepo, taskService: taskService}
}

func (s *SubtaskService) ListSubtasks(ctx context.Context, taskID int) ([]model.Subtask, error) {
	return s.subtaskRepo.ListByTaskID(ctx, taskID)
}

func (s *SubtaskService) CreateSubtask(ctx context.Context, st *model.Subtask) error {
	st.Title = strings.TrimSpace(st.Title)
	if st.Title == "" {
		return ErrTitleRequired
	}
	return s.subtaskRepo.Create(ctx, st)
}

func (s *SubtaskService) UpdateSubtask(ctx context.Context, id int, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleRequired
	}
	return s.subtaskRepo.Update(ctx, id, title)
}

func (s *SubtaskService) DeleteSubtask(ctx context.Context, id int, taskID int) error {
	if err := s.subtaskRepo.Delete(ctx, id); err != nil {
		return err
	}
	return s.syncParentTaskStatus(ctx, taskID)
}

func (s *SubtaskService) UpdateSubtaskDone(ctx context.Context, id int, isDone bool) error {
	taskID, err := s.subtaskRepo.UpdateDone(ctx, id, isDone)
	if err != nil {
		return err
	}
	return s.syncParentTaskStatus(ctx, taskID)
}

func (s *SubtaskService) syncParentTaskStatus(ctx context.Context, taskID int) error {
	return s.taskService.syncParentTaskStatus(ctx, taskID)
}
