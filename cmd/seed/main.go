package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/username/task-tracker/internal/repository"
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

	ctx := context.Background()
	pool, err := repository.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("database connected successfully")

	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo)

	email := "user@gmail.com"
	password := "password"

	// Register user (hashing password automatically via auth service)
	err = authSvc.Register(ctx, email, password)
	if err != nil {
		log.Fatalf("Failed to seed user: %v\n(If it says email already exists, then the seeder has already been run)", err)
	}

	log.Printf("Successfully seeded user! Email: %s, Password: %s\n", email, password)
}
