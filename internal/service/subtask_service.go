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

func (s *SubtaskService) CreateSubtask(ctx context.Context, st *model.Subtask) error {
	st.Title = strings.TrimSpace(st.Title)
	if st.Title == "" {
		return ErrTitleRequired
	}
	return s.subtaskRepo.Create(ctx, st)
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
