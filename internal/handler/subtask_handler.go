package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/service"
)

type SubtaskHandler struct {
	subtaskService *service.SubtaskService
}

func NewSubtaskHandler(subtaskService *service.SubtaskService) *SubtaskHandler {
	return &SubtaskHandler{subtaskService: subtaskService}
}

func (h *SubtaskHandler) List(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	subtasks, err := h.subtaskService.ListSubtasks(r.Context(), taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list subtasks")
		return
	}
	if subtasks == nil {
		subtasks = []model.Subtask{}
	}

	respondJSON(w, http.StatusOK, subtasks)
}

func (h *SubtaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var s model.Subtask
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.TaskID = taskID

	if err := h.subtaskService.CreateSubtask(r.Context(), &s); err != nil {
		if errors.Is(err, service.ErrTitleRequired) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create subtask")
		return
	}

	respondJSON(w, http.StatusCreated, s)
}

func (h *SubtaskHandler) UpdateDone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("subtaskId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid subtask id")
		return
	}

	var body struct {
		IsDone bool `json:"is_done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.subtaskService.UpdateSubtaskDone(r.Context(), id, body.IsDone); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update subtask")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "subtask updated"})
}

func (h *SubtaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("subtaskId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid subtask id")
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.subtaskService.UpdateSubtask(r.Context(), id, body.Title); err != nil {
		if errors.Is(err, service.ErrTitleRequired) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update subtask")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "subtask updated"})
}

func (h *SubtaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("subtaskId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid subtask id")
		return
	}
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.subtaskService.DeleteSubtask(r.Context(), id, taskID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete subtask")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
