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
	"github.com/beetrack/backend/pkg/mailer"
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
	emailTokenRepo := repository.NewEmailTokenRepository(db)
	apiaryRepo := repository.NewApiaryRepository(db)
	hiveRepo := repository.NewHiveRepository(db)
	inspectionRepo := repository.NewInspectionRepository(db)
	inspectionImageRepo := repository.NewInspectionImageRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)

	mail := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	authSvc := service.NewAuthService(userRepo, tokenRepo, emailTokenRepo, mail, cfg.APIURL, cfg.AppURL, cfg.JWTSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLDays)
	apiarySvc := service.NewApiaryService(apiaryRepo, hiveRepo)
	invitationSvc := service.NewInvitationService(apiaryRepo, invitationRepo, userRepo)
	hiveSvc := service.NewHiveService(apiaryRepo, hiveRepo)
	inspectionSvc := service.NewInspectionService(apiaryRepo, hiveRepo, inspectionRepo)
	inspectionImageSvc := service.NewInspectionImageService(apiaryRepo, hiveRepo, inspectionRepo, inspectionImageRepo, cfg.ImageStoragePath)
	userSvc := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	apiaryHandler := handler.NewApiaryHandler(apiarySvc)
	invitationHandler := handler.NewInvitationHandler(invitationSvc)
	hiveHandler := handler.NewHiveHandler(hiveSvc, inspectionSvc)
	inspectionHandler := handler.NewInspectionHandler(inspectionSvc, inspectionImageSvc)
	inspectionImageHandler := handler.NewInspectionImageHandler(inspectionImageSvc)
	userHandler := handler.NewUserHandler(userSvc)

	auth := middleware.Auth(cfg.JWTSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/resend-verification", authHandler.ResendVerification)
	mux.HandleFunc("POST /api/v1/auth/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("GET /api/v1/auth/reset-password-form", authHandler.ResetPasswordForm)
	mux.HandleFunc("POST /api/v1/auth/reset-password-form", authHandler.ResetPasswordForm)
	mux.HandleFunc("GET /api/v1/auth/verify-email", authHandler.VerifyEmail)

	mux.Handle("PATCH /api/v1/users/me/name", auth(http.HandlerFunc(userHandler.UpdateName)))

	mux.Handle("POST /api/v1/apiaries", auth(http.HandlerFunc(apiaryHandler.Create)))
	mux.Handle("GET /api/v1/apiaries", auth(http.HandlerFunc(apiaryHandler.List)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives", auth(http.HandlerFunc(hiveHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives", auth(http.HandlerFunc(hiveHandler.Create)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}", auth(http.HandlerFunc(hiveHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}", auth(http.HandlerFunc(hiveHandler.Get)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}", auth(http.HandlerFunc(hiveHandler.Update)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/frames", auth(http.HandlerFunc(hiveHandler.AddFrames)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/position", auth(http.HandlerFunc(hiveHandler.Move)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/diseases", auth(http.HandlerFunc(hiveHandler.AddDisease)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/diseases/{diseaseId}", auth(http.HandlerFunc(hiveHandler.RemoveDisease)))
	mux.Handle("POST /api/v1/apiaries/{id}/copy", auth(http.HandlerFunc(apiaryHandler.Copy)))
	mux.Handle("DELETE /api/v1/apiaries/{id}", auth(http.HandlerFunc(apiaryHandler.Delete)))
	mux.Handle("PATCH /api/v1/apiaries/{id}", auth(http.HandlerFunc(apiaryHandler.Update)))
	mux.Handle("POST /api/v1/apiaries/{id}/invitations", auth(http.HandlerFunc(invitationHandler.Invite)))
	mux.Handle("GET /api/v1/apiaries/{id}/invitations", auth(http.HandlerFunc(invitationHandler.ListForApiary)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/invitations/{invitationId}", auth(http.HandlerFunc(invitationHandler.CancelInvitation)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/members/{userId}", auth(http.HandlerFunc(invitationHandler.RemoveMember)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/leave", auth(http.HandlerFunc(invitationHandler.Leave)))
	mux.Handle("GET /api/v1/invitations", auth(http.HandlerFunc(invitationHandler.ListMine)))
	mux.Handle("GET /api/v1/invitations/count", auth(http.HandlerFunc(invitationHandler.CountMine)))
	mux.Handle("POST /api/v1/invitations/{id}/accept", auth(http.HandlerFunc(invitationHandler.Accept)))
	mux.Handle("POST /api/v1/invitations/{id}/decline", auth(http.HandlerFunc(invitationHandler.Decline)))

	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections", auth(http.HandlerFunc(inspectionHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/inspections", auth(http.HandlerFunc(inspectionHandler.Create)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}", auth(http.HandlerFunc(inspectionHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}", auth(http.HandlerFunc(inspectionHandler.Get)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}", auth(http.HandlerFunc(inspectionHandler.Update)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases", auth(http.HandlerFunc(inspectionHandler.AddDisease)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases/{diseaseId}", auth(http.HandlerFunc(inspectionHandler.RemoveDisease)))

	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images", auth(http.HandlerFunc(inspectionImageHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images", auth(http.HandlerFunc(inspectionImageHandler.Upload)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}", auth(http.HandlerFunc(inspectionImageHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}/file", auth(http.HandlerFunc(inspectionImageHandler.ServeFile)))

	cors := middleware.CORS(cfg.AllowedOrigins)

	log.Printf("Starting BeeTrack API on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, cors(mux)); err != nil {
		log.Fatal(err)
	}
}
