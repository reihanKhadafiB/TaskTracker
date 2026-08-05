package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

type createTaskRequest struct {
	ContextID   *int    `json:"context_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	DueDate     *string `json:"due_date"`
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t := model.Task{
		ContextID:   req.ContextID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "due_date must be in YYYY-MM-DD format")
			return
		}
		t.DueDate = &model.Date{Time: parsed}
	}

	if err := h.taskService.CreateTask(r.Context(), &t); err != nil {
		switch {
		case errors.Is(err, service.ErrTitleRequired), errors.Is(err, service.ErrInvalidStatus):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to create task")
		}
		return
	}

	respondJSON(w, http.StatusCreated, t)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.taskService.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "task not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get task")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseTaskFilter(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	tasks, err := h.taskService.ListTasks(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	if tasks == nil {
		tasks = []model.Task{}
	}

	var nextCursor *int
	if len(tasks) > 0 && filter.Limit > 0 && len(tasks) == filter.Limit {
		last := tasks[len(tasks)-1].ID
		nextCursor = &last
	}

	respondJSON(w, http.StatusOK, taskListResponse{
		Data:       tasks,
		NextCursor: nextCursor,
	})
}

type taskListResponse struct {
	Data       []model.Task `json:"data"`
	NextCursor *int         `json:"next_cursor"`
}

func parseTaskFilter(r *http.Request) (model.TaskFilter, error) {
	q := r.URL.Query()

	filter := model.TaskFilter{
		Status:  q.Get("status"),
		Overdue: q.Get("overdue") == "true",
		Limit:   20,
	}

	if filter.Status != "" && !validStatusFilter(filter.Status) {
		return filter, errors.New("invalid status filter value")
	}

	if raw := q.Get("context_id"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil {
			return filter, errors.New("invalid context_id value")
		}
		filter.ContextID = &id
	}

	if raw := q.Get("cursor"); raw != "" {
		cursor, err := strconv.Atoi(raw)
		if err != nil {
			return filter, errors.New("invalid cursor value")
		}
		filter.Cursor = &cursor
	}

	if raw := q.Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 || limit > 100 {
			return filter, errors.New("invalid limit value, must be between 1 and 100")
		}
		filter.Limit = limit
	}

	return filter, nil
}

func validStatusFilter(status string) bool {
	return status == "todo" || status == "in_progress" || status == "done"
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t := model.Task{
		ContextID:   req.ContextID,
		Title:       req.Title,
		Description: req.Description,
	}

	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "due_date must be in YYYY-MM-DD format")
			return
		}
		t.DueDate = &model.Date{Time: parsed}
	}

	if err := h.taskService.UpdateTask(r.Context(), id, &t); err != nil {
		if errors.Is(err, service.ErrTitleRequired) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "task updated"})
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.taskService.UpdateTaskStatus(r.Context(), id, body.Status); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStatus):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update task status")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "task updated"})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.taskService.DeleteTask(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
