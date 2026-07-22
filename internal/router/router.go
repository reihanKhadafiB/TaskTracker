package router

import (
	"net/http"

	"github.com/username/task-tracker/internal/handler"
)

func New(taskHandler *handler.TaskHandler, contextHandler *handler.ContextHandler, subtaskHandler *handler.SubtaskHandler, authHandler *handler.AuthHandler, authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /tasks", taskHandler.Create)
	mux.HandleFunc("GET /tasks", taskHandler.List)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.GetByID)
	mux.HandleFunc("PATCH /tasks/{id}/status", taskHandler.UpdateStatus)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.Delete)

	mux.HandleFunc("POST /tasks/{id}/subtasks", subtaskHandler.Create)
	mux.HandleFunc("PATCH /tasks/{id}/subtasks/{subtaskId}", subtaskHandler.UpdateDone)

	mux.HandleFunc("POST /contexts", contextHandler.Create)
	mux.HandleFunc("GET /contexts", contextHandler.List)

	return mux
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
