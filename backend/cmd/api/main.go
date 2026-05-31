package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/beetrack/backend/internal/config"
	"github.com/beetrack/backend/internal/database"
	"github.com/beetrack/backend/internal/handler"
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/internal/service"
	"github.com/joho/godotenv"
)

//go:embed ../../migrations
var migrations embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Migrate(db, migrations); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)

	log.Printf("Starting BeeTrack API on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
