package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/username/task-tracker/internal/handler"
	"github.com/username/task-tracker/internal/middleware"
	"github.com/username/task-tracker/internal/repository"
	"github.com/username/task-tracker/internal/router"
	"github.com/username/task-tracker/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system env vars")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	pool, err := repository.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("database connected successfully")

	taskRepo := repository.NewTaskRepository(pool)
	subtaskRepo := repository.NewSubtaskRepository(pool)
	taskSvc := service.NewTaskService(taskRepo, subtaskRepo)
	taskHandler := handler.NewTaskHandler(taskSvc)

	subTaskSvc := service.NewSubtaskService(subtaskRepo, taskSvc)
	subtaskHandler := handler.NewSubtaskHandler(subTaskSvc)

	contextRepo := repository.NewContextRepository(pool)
	contextSvc := service.NewContextService(contextRepo)
	contextHandler := handler.NewContextHandler(contextSvc)

	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)
	authMw := middleware.Auth(authSvc)

	exportHandler := handler.NewExportHandler(contextSvc, taskSvc, subTaskSvc)

	mux := router.New(taskHandler, contextHandler, subtaskHandler, authHandler, exportHandler, authMw)

	var h http.Handler = mux
	h = middleware.CORS(h)

	log.Printf("server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, h); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
