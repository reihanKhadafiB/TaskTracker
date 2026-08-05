package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/username/task-tracker/internal/model"
	"github.com/username/task-tracker/internal/service"
)

type ContextHandler struct {
	contextService *service.ContextService
}

func NewContextHandler(contextService *service.ContextService) *ContextHandler {
	return &ContextHandler{contextService: contextService}
}

func (h *ContextHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c model.Context
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.contextService.CreateContext(r.Context(), &c); err != nil {
		if errors.Is(err, service.ErrContextNameRequired) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create context")
		return
	}

	respondJSON(w, http.StatusCreated, c)
}

func (h *ContextHandler) List(w http.ResponseWriter, r *http.Request) {
	contexts, err := h.contextService.ListContexts(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list contexts")
		return
	}
	if contexts == nil {
		contexts = []model.Context{}
	}
	respondJSON(w, http.StatusOK, contexts)
}

func (h *ContextHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid context id")
		return
	}

	var c model.Context
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.contextService.UpdateContext(r.Context(), id, c.Name, c.Color); err != nil {
		if errors.Is(err, service.ErrContextNameRequired) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update context")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "context updated"})
}

func (h *ContextHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid context id")
		return
	}

	if err := h.contextService.DeleteContext(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete context")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
