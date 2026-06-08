package main

import (
	"context"
	"log"
	"strings"

	"graduation-invitation/config"
	"graduation-invitation/database"
	"graduation-invitation/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(context.Background(), db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	app := fiber.New(fiber.Config{AppName: "Graduation Invitation CMS"})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins(cfg.FrontendURL),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	authService := routes.Register(app, db, cfg)
	if err := authService.EnsureDefaultAdmin(context.Background()); err != nil {
		log.Fatalf("default admin setup failed: %v", err)
	}
	if err := authService.EnsureDefaultSuperAdmin(context.Background()); err != nil {
		log.Fatalf("default super admin setup failed: %v", err)
	}

	log.Printf("backend running on :%s", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}

func allowedOrigins(frontendURL string) string {
	defaults := []string{
		frontendURL,
		"http://localhost:5173",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:8081",
	}
	seen := map[string]bool{}
	origins := []string{}
	for _, item := range defaults {
		for _, origin := range strings.Split(item, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" && !seen[origin] {
				origins = append(origins, origin)
				seen[origin] = true
			}
		}
	}
	return strings.Join(origins, ",")
}
