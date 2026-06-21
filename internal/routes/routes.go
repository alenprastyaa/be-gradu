package routes

import (
	"graduation-invitation/config"
	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/handlers"
	"graduation-invitation/internal/middlewares"
	"graduation-invitation/internal/repositories"
	"graduation-invitation/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(app *fiber.App, db *pgxpool.Pool, cfg config.Config) *services.AuthService {
	adminRepo := repositories.NewAdminRepository(db)
	schoolRepo := repositories.NewSchoolRepository(db)
	studentRepo := repositories.NewStudentRepository(db)
	emailLogRepo := repositories.NewEmailLogRepository(db)
	eventSettingsRepo := repositories.NewEventSettingsRepository(db)

	seatService := services.NewSeatService(studentRepo, eventSettingsRepo)
	studentService := services.NewStudentService(studentRepo, seatService, db, cfg.PublicInvitationURL)
	eventSettingsService := services.NewEventSettingsService(eventSettingsRepo)
	authService := services.NewAuthService(adminRepo, schoolRepo, cfg)
	superAdminService := services.NewSuperAdminService(schoolRepo, adminRepo)
	invitationService := services.NewInvitationService(studentService, eventSettingsService)
	attendanceService := services.NewAttendanceService(studentRepo)
	whatsappService := services.NewWhatsappService(studentRepo, eventSettingsService, cfg.PublicInvitationURL)
	emailService := services.NewEmailService(studentRepo, emailLogRepo, eventSettingsService, cfg.PublicInvitationURL, cfg.BrevoAPIKey, cfg.BrevoSenderEmail, cfg.BrevoSenderName, cfg.BrevoWebhookSecret)
	r2StorageService := services.NewR2StorageService(cfg)

	authHandler := handlers.NewAuthHandler(authService)
	superAdminHandler := handlers.NewSuperAdminHandler(superAdminService)
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

	super := protected.Group("/super", middlewares.RequireRole(authcontext.RoleSuperAdmin))
	super.Get("/schools", superAdminHandler.ListSchools)
	super.Post("/schools", superAdminHandler.CreateSchool)
	super.Put("/schools/:id", superAdminHandler.UpdateSchool)
	super.Delete("/schools/:id", superAdminHandler.DeleteSchool)
	super.Get("/admins", superAdminHandler.ListAdmins)
	super.Post("/admins", superAdminHandler.CreateAdmin)
	super.Put("/admins/:id", superAdminHandler.UpdateAdmin)
	super.Delete("/admins/:id", superAdminHandler.DeleteAdmin)

	schoolAdmin := protected.Group("", middlewares.RequireRole(authcontext.RoleSchoolAdmin))

	schoolAdmin.Get("/event-settings", eventSettingsHandler.Get)
	schoolAdmin.Put("/event-settings", eventSettingsHandler.Update)
	schoolAdmin.Put("/event-settings/seat-map", eventSettingsHandler.UpdateSeatMap)
	schoolAdmin.Get("/event-templates", eventSettingsHandler.List)
	schoolAdmin.Post("/event-templates", eventSettingsHandler.Create)
	schoolAdmin.Put("/event-templates/:id", eventSettingsHandler.Update)
	schoolAdmin.Post("/event-templates/:id/activate", eventSettingsHandler.Activate)
	schoolAdmin.Delete("/event-templates/:id", eventSettingsHandler.Delete)
	schoolAdmin.Post("/event-templates/logo", eventSettingsHandler.UploadLogo)
	schoolAdmin.Post("/event-templates/audio", eventSettingsHandler.UploadAudio)

	schoolAdmin.Get("/students", studentHandler.List)
	schoolAdmin.Post("/students", studentHandler.Create)
	schoolAdmin.Get("/students/seat-map", studentHandler.SeatMap)
	schoolAdmin.Get("/students/export", studentHandler.Export)
	schoolAdmin.Get("/students/import-template", studentHandler.ImportTemplate)
	schoolAdmin.Post("/students/import", studentHandler.Import)
	schoolAdmin.Post("/students/regenerate-seat-numbers", studentHandler.RegenerateSeats)
	schoolAdmin.Post("/students/reset-all", studentHandler.ResetAll)
	schoolAdmin.Get("/students/whatsapp-links", whatsappHandler.All)
	schoolAdmin.Post("/students/whatsapp-links", whatsappHandler.All)
	schoolAdmin.Post("/students/email-send", emailHandler.All)
	schoolAdmin.Post("/students/email-history/sync", emailHandler.SyncBrevoHistory)
	schoolAdmin.Get("/students/:id", studentHandler.Get)
	schoolAdmin.Put("/students/:id", studentHandler.Update)
	schoolAdmin.Put("/students/:id/attendance", studentHandler.UpdateAttendanceStatus)
	schoolAdmin.Delete("/students/:id", studentHandler.Delete)
	schoolAdmin.Get("/students/:id/whatsapp-link", whatsappHandler.One)
	schoolAdmin.Post("/students/:id/whatsapp-sent", whatsappHandler.MarkSent)
	schoolAdmin.Delete("/students/:id/whatsapp-sent", whatsappHandler.ResetSent)
	schoolAdmin.Post("/students/:id/send-email", emailHandler.One)
	schoolAdmin.Delete("/students/:id/email-sent", emailHandler.ResetSent)

	schoolAdmin.Post("/attendance/scan", attendanceHandler.Scan)
	schoolAdmin.Get("/attendance/summary", attendanceHandler.Summary)

	return authService
}
