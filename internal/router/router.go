package router

import (
	"net/http"

	"github.com/username/task-tracker/internal/handler"
)

func New(
	taskHandler *handler.TaskHandler,
	contextHandler *handler.ContextHandler,
	subtaskHandler *handler.SubtaskHandler,
	authHandler *handler.AuthHandler,
	exportHandler *handler.ExportHandler,
	authMiddleware func(http.Handler) http.Handler,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.Handle("POST /tasks", authMiddleware(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /tasks", authMiddleware(http.HandlerFunc(taskHandler.List)))
	mux.Handle("GET /tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.GetByID)))
	mux.Handle("PUT /tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("PATCH /tasks/{id}/status", authMiddleware(http.HandlerFunc(taskHandler.UpdateStatus)))
	mux.Handle("DELETE /tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.Delete)))

	mux.Handle("GET /tasks/{id}/subtasks", authMiddleware(http.HandlerFunc(subtaskHandler.List)))
	mux.Handle("POST /tasks/{id}/subtasks", authMiddleware(http.HandlerFunc(subtaskHandler.Create)))
	mux.Handle("PATCH /tasks/{id}/subtasks/{subtaskId}", authMiddleware(http.HandlerFunc(subtaskHandler.UpdateDone)))
	mux.Handle("PUT /tasks/{id}/subtasks/{subtaskId}", authMiddleware(http.HandlerFunc(subtaskHandler.Update)))
	mux.Handle("DELETE /tasks/{id}/subtasks/{subtaskId}", authMiddleware(http.HandlerFunc(subtaskHandler.Delete)))

	mux.Handle("POST /contexts", authMiddleware(http.HandlerFunc(contextHandler.Create)))
	mux.Handle("GET /contexts", authMiddleware(http.HandlerFunc(contextHandler.List)))
	mux.Handle("PUT /contexts/{id}", authMiddleware(http.HandlerFunc(contextHandler.Update)))
	mux.Handle("DELETE /contexts/{id}", authMiddleware(http.HandlerFunc(contextHandler.Delete)))
	mux.Handle("GET /contexts/{id}/pdf", authMiddleware(http.HandlerFunc(exportHandler.ExportContextPDF)))

	return mux
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
