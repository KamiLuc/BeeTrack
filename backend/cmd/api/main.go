package main

import (
	"log"
	"net/http"

	"github.com/beetrack/backend/internal/config"
	"github.com/beetrack/backend/internal/database"
	"github.com/beetrack/backend/internal/handler"
	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/repository"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/migrations"
	"github.com/joho/godotenv"
)

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

	if err := database.Migrate(db, migrations.FS); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	apiaryRepo := repository.NewApiaryRepository(db)
	hiveRepo := repository.NewHiveRepository(db)

	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLDays)
	apiarySvc := service.NewApiaryService(apiaryRepo, hiveRepo)
	hiveSvc := service.NewHiveService(apiaryRepo, hiveRepo)
	userSvc := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	apiaryHandler := handler.NewApiaryHandler(apiarySvc)
	hiveHandler := handler.NewHiveHandler(hiveSvc)
	userHandler := handler.NewUserHandler(userSvc)

	auth := middleware.Auth(cfg.JWTSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)

	mux.Handle("PATCH /api/v1/users/me/name", auth(http.HandlerFunc(userHandler.UpdateName)))

	mux.Handle("POST /api/v1/apiaries", auth(http.HandlerFunc(apiaryHandler.Create)))
	mux.Handle("GET /api/v1/apiaries", auth(http.HandlerFunc(apiaryHandler.List)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives", auth(http.HandlerFunc(hiveHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives", auth(http.HandlerFunc(hiveHandler.Create)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}", auth(http.HandlerFunc(hiveHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}", auth(http.HandlerFunc(hiveHandler.Get)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}", auth(http.HandlerFunc(hiveHandler.Update)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/position", auth(http.HandlerFunc(hiveHandler.Move)))
	mux.Handle("DELETE /api/v1/apiaries/{id}", auth(http.HandlerFunc(apiaryHandler.Delete)))
	mux.Handle("PATCH /api/v1/apiaries/{id}", auth(http.HandlerFunc(apiaryHandler.Update)))

	cors := middleware.CORS(cfg.AllowedOrigins)

	log.Printf("Starting BeeTrack API on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, cors(mux)); err != nil {
		log.Fatal(err)
	}
}
