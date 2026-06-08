package routes

import (
	"graduation-invitation/config"
	"graduation-invitation/internal/handlers"
	"graduation-invitation/internal/middlewares"
	"graduation-invitation/internal/repositories"
	"graduation-invitation/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(app *fiber.App, db *pgxpool.Pool, cfg config.Config) *services.AuthService {
	adminRepo := repositories.NewAdminRepository(db)
	studentRepo := repositories.NewStudentRepository(db)
	emailLogRepo := repositories.NewEmailLogRepository(db)
	eventSettingsRepo := repositories.NewEventSettingsRepository(db)

	seatService := services.NewSeatService(studentRepo)
	studentService := services.NewStudentService(studentRepo, seatService, db, cfg.PublicInvitationURL)
	eventSettingsService := services.NewEventSettingsService(eventSettingsRepo)
	authService := services.NewAuthService(adminRepo, cfg)
	invitationService := services.NewInvitationService(studentService, eventSettingsService)
	attendanceService := services.NewAttendanceService(studentRepo)
	whatsappService := services.NewWhatsappService(studentRepo, eventSettingsService, cfg.PublicInvitationURL)
	emailService := services.NewEmailService(studentRepo, emailLogRepo, eventSettingsService, cfg.PublicInvitationURL, cfg.BrevoAPIKey, cfg.BrevoSenderEmail, cfg.BrevoSenderName, cfg.BrevoWebhookSecret)
	r2StorageService := services.NewR2StorageService(cfg)

	authHandler := handlers.NewAuthHandler(authService)
	studentHandler := handlers.NewStudentHandler(studentService, seatService)
	eventSettingsHandler := handlers.NewEventSettingsHandler(eventSettingsService, r2StorageService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceService)
	whatsappHandler := handlers.NewWhatsappHandler(whatsappService)
	emailHandler := handlers.NewEmailHandler(emailService)

	api := app.Group("/api")
	api.Post("/auth/login", authHandler.Login)
	api.Get("/invite/:code", invitationHandler.Get)
	api.Post("/webhooks/brevo/email", emailHandler.BrevoWebhook)

	protected := api.Group("", middlewares.AuthMiddleware(cfg.JWTSecret))
	protected.Get("/auth/me", authHandler.Me)
	protected.Post("/auth/logout", authHandler.Logout)

	protected.Get("/event-settings", eventSettingsHandler.Get)
	protected.Put("/event-settings", eventSettingsHandler.Update)
	protected.Put("/event-settings/seat-map", eventSettingsHandler.UpdateSeatMap)
	protected.Get("/event-templates", eventSettingsHandler.List)
	protected.Post("/event-templates", eventSettingsHandler.Create)
	protected.Put("/event-templates/:id", eventSettingsHandler.Update)
	protected.Post("/event-templates/:id/activate", eventSettingsHandler.Activate)
	protected.Delete("/event-templates/:id", eventSettingsHandler.Delete)
	protected.Post("/event-templates/logo", eventSettingsHandler.UploadLogo)
	protected.Post("/event-templates/audio", eventSettingsHandler.UploadAudio)

	protected.Get("/students", studentHandler.List)
	protected.Post("/students", studentHandler.Create)
	protected.Get("/students/seat-map", studentHandler.SeatMap)
	protected.Get("/students/export", studentHandler.Export)
	protected.Get("/students/import-template", studentHandler.ImportTemplate)
	protected.Post("/students/import", studentHandler.Import)
	protected.Post("/students/regenerate-seat-numbers", studentHandler.RegenerateSeats)
	protected.Post("/students/reset-all", studentHandler.ResetAll)
	protected.Get("/students/whatsapp-links", whatsappHandler.All)
	protected.Post("/students/whatsapp-links", whatsappHandler.All)
	protected.Post("/students/email-send", emailHandler.All)
	protected.Post("/students/email-history/sync", emailHandler.SyncBrevoHistory)
	protected.Get("/students/:id", studentHandler.Get)
	protected.Put("/students/:id", studentHandler.Update)
	protected.Put("/students/:id/attendance", studentHandler.UpdateAttendanceStatus)
	protected.Delete("/students/:id", studentHandler.Delete)
	protected.Get("/students/:id/whatsapp-link", whatsappHandler.One)
	protected.Post("/students/:id/whatsapp-sent", whatsappHandler.MarkSent)
	protected.Delete("/students/:id/whatsapp-sent", whatsappHandler.ResetSent)
	protected.Post("/students/:id/send-email", emailHandler.One)
	protected.Delete("/students/:id/email-sent", emailHandler.ResetSent)

	protected.Post("/attendance/scan", attendanceHandler.Scan)
	protected.Get("/attendance/summary", attendanceHandler.Summary)

	return authService
}
