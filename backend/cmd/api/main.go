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
	treatmentRepo := repository.NewTreatmentRepository(db)
	feedingRepo := repository.NewFeedingRepository(db)
	harvestRepo := repository.NewHarvestRepository(db)
	listingRepo := repository.NewListingRepository(db)
	listingFavoriteRepo := repository.NewListingFavoriteRepository(db)

	mail := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	authSvc := service.NewAuthService(userRepo, tokenRepo, emailTokenRepo, mail, cfg.APIURL, cfg.AppURL, cfg.JWTSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLDays)
	apiarySvc := service.NewApiaryService(apiaryRepo, hiveRepo)
	invitationSvc := service.NewInvitationService(apiaryRepo, invitationRepo, userRepo)
	hiveSvc := service.NewHiveService(apiaryRepo, hiveRepo)
	inspectionSvc := service.NewInspectionService(apiaryRepo, hiveRepo, inspectionRepo)
	inspectionImageSvc := service.NewInspectionImageService(apiaryRepo, hiveRepo, inspectionRepo, inspectionImageRepo, cfg.ImageStoragePath)
	treatmentSvc := service.NewTreatmentService(apiaryRepo, hiveRepo, hiveRepo, treatmentRepo)
	feedingSvc := service.NewFeedingService(apiaryRepo, hiveRepo, hiveRepo, feedingRepo)
	harvestSvc := service.NewHarvestService(apiaryRepo, hiveRepo, harvestRepo)
	listingSvc := service.NewListingService(listingRepo, apiaryRepo)
	listingImageSvc := service.NewListingImageService(listingRepo, listingRepo, cfg.ImageStoragePath)
	listingFavoriteSvc := service.NewListingFavoriteService(listingFavoriteRepo, listingRepo)
	userSvc := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	apiaryHandler := handler.NewApiaryHandler(apiarySvc)
	invitationHandler := handler.NewInvitationHandler(invitationSvc)
	hiveHandler := handler.NewHiveHandler(hiveSvc, inspectionSvc)
	inspectionHandler := handler.NewInspectionHandler(inspectionSvc, inspectionImageSvc)
	inspectionImageHandler := handler.NewInspectionImageHandler(inspectionImageSvc)
	treatmentHandler := handler.NewTreatmentHandler(treatmentSvc)
	feedingHandler := handler.NewFeedingHandler(feedingSvc)
	harvestHandler := handler.NewHarvestHandler(harvestSvc)
	listingHandler := handler.NewListingHandler(listingSvc)
	listingImageHandler := handler.NewListingImageHandler(listingImageSvc)
	listingFavoriteHandler := handler.NewListingFavoriteHandler(listingFavoriteSvc)
	userHandler := handler.NewUserHandler(userSvc)

	auth := middleware.Auth(cfg.JWTSecret)
	optionalAuth := middleware.OptionalAuth(cfg.JWTSecret)

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
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/position", auth(http.HandlerFunc(hiveHandler.Move)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/diseases", auth(http.HandlerFunc(hiveHandler.AddDisease)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/transfer", auth(http.HandlerFunc(hiveHandler.ChangeApiary)))
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

	mux.Handle("GET /api/v1/medicines", auth(http.HandlerFunc(treatmentHandler.Medicines)))
	mux.Handle("GET /api/v1/doses", auth(http.HandlerFunc(treatmentHandler.Doses)))

	mux.Handle("POST /api/v1/apiaries/{id}/treatments/bulk", auth(http.HandlerFunc(treatmentHandler.BulkCreate)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/treatments", auth(http.HandlerFunc(treatmentHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/treatments", auth(http.HandlerFunc(treatmentHandler.Create)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId}", auth(http.HandlerFunc(treatmentHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId}", auth(http.HandlerFunc(treatmentHandler.Get)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId}", auth(http.HandlerFunc(treatmentHandler.Update)))

	mux.Handle("GET /api/v1/feed-types", auth(http.HandlerFunc(feedingHandler.FeedTypes)))
	mux.Handle("GET /api/v1/feed-amounts", auth(http.HandlerFunc(feedingHandler.Amounts)))

	mux.Handle("POST /api/v1/apiaries/{id}/feedings/bulk", auth(http.HandlerFunc(feedingHandler.BulkCreate)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/feedings", auth(http.HandlerFunc(feedingHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/feedings", auth(http.HandlerFunc(feedingHandler.Create)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/feedings/{feedingId}", auth(http.HandlerFunc(feedingHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/feedings/{feedingId}", auth(http.HandlerFunc(feedingHandler.Get)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/feedings/{feedingId}", auth(http.HandlerFunc(feedingHandler.Update)))

	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/harvests", auth(http.HandlerFunc(harvestHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/harvests", auth(http.HandlerFunc(harvestHandler.Create)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId}", auth(http.HandlerFunc(harvestHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId}", auth(http.HandlerFunc(harvestHandler.Get)))
	mux.Handle("PATCH /api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId}", auth(http.HandlerFunc(harvestHandler.Update)))

	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images", auth(http.HandlerFunc(inspectionImageHandler.List)))
	mux.Handle("POST /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images", auth(http.HandlerFunc(inspectionImageHandler.Upload)))
	mux.Handle("DELETE /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}", auth(http.HandlerFunc(inspectionImageHandler.Delete)))
	mux.Handle("GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}/file", auth(http.HandlerFunc(inspectionImageHandler.ServeFile)))

	mux.Handle("POST /api/v1/listings", auth(http.HandlerFunc(listingHandler.Create)))
	mux.Handle("GET /api/v1/listings", optionalAuth(http.HandlerFunc(listingHandler.Search)))
	mux.Handle("GET /api/v1/listings/{id}", optionalAuth(http.HandlerFunc(listingHandler.Get)))
	mux.Handle("PATCH /api/v1/listings/{id}", auth(http.HandlerFunc(listingHandler.Update)))
	mux.Handle("PATCH /api/v1/listings/{id}/hide", auth(http.HandlerFunc(listingHandler.Hide)))
	mux.Handle("DELETE /api/v1/listings/{id}", auth(http.HandlerFunc(listingHandler.Delete)))
	mux.Handle("POST /api/v1/listings/{id}/images", auth(http.HandlerFunc(listingImageHandler.Upload)))
	mux.HandleFunc("GET /api/v1/listings/{id}/images/{imageId}/file", listingImageHandler.ServeFile)
	mux.Handle("DELETE /api/v1/listings/{id}/images/{imageId}", auth(http.HandlerFunc(listingImageHandler.Delete)))
	mux.Handle("GET /api/v1/favorites", auth(http.HandlerFunc(listingFavoriteHandler.List)))
	mux.Handle("POST /api/v1/listings/{id}/favorite", auth(http.HandlerFunc(listingFavoriteHandler.Add)))
	mux.Handle("DELETE /api/v1/listings/{id}/favorite", auth(http.HandlerFunc(listingFavoriteHandler.Remove)))
	mux.Handle("GET /api/v1/listings/{id}/favorite", auth(http.HandlerFunc(listingFavoriteHandler.Check)))

	cors := middleware.CORS(cfg.AllowedOrigins)

	log.Printf("Starting BeeTrack API on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, cors(mux)); err != nil {
		log.Fatal(err)
	}
}
