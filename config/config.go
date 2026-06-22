package config

import (
	"os"
	"strings"
)

type Config struct {
	AppPort                   string
	DatabaseURL               string
	JWTSecret                 string
	FrontendURL               string
	PublicInvitationURL       string
	BrevoAPIKey               string
	BrevoSenderEmail          string
	BrevoSenderName           string
	BrevoWebhookSecret        string
	R2Bucket                  string
	R2AccessKeyID             string
	R2SecretAccessKey         string
	R2PublicBaseURL           string
	R2Endpoint                string
	R2AccountID               string
	DefaultAdminName          string
	DefaultAdminEmail         string
	DefaultAdminPassword      string
	DefaultSchoolName         string
	DefaultSchoolSlug         string
	DefaultSuperAdminName     string
	DefaultSuperAdminEmail    string
	DefaultSuperAdminPassword string
}

func Load() Config {
	frontendURL := env("FRONTEND_URL", "http://localhost:5173")
	publicInvitationURL := strings.TrimSpace(os.Getenv("PUBLIC_INVITATION_URL"))
	if publicInvitationURL == "" {
		publicInvitationURL = strings.TrimRight(frontendURL, "/") + "/invite"
	}

	return Config{
		AppPort:                   env("APP_PORT", "8080"),
		DatabaseURL:               env("DATABASE_URL", "postgres://username:password@localhost:5432/graduation?sslmode=disable"),
		JWTSecret:                 env("JWT_SECRET", "change_this_secret"),
		FrontendURL:               frontendURL,
		PublicInvitationURL:       publicInvitationURL,
		BrevoAPIKey:               env("BREVO_API_KEY", ""),
		BrevoSenderEmail:          env("BREVO_SENDER_EMAIL", ""),
		BrevoSenderName:           env("BREVO_SENDER_NAME", "Graduation Invitation CMS"),
		BrevoWebhookSecret:        env("BREVO_WEBHOOK_SECRET", ""),
		R2Bucket:                  env("R2_BUCKET", ""),
		R2AccessKeyID:             env("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:         env("R2_SECRET_ACCESS_KEY", ""),
		R2PublicBaseURL:           env("R2_PUBLIC_BASE_URL", ""),
		R2Endpoint:                env("R2_ENDPOINT", ""),
		R2AccountID:               env("R2_ACCOUNT_ID", ""),
		DefaultAdminName:          env("DEFAULT_ADMIN_NAME", "Administrator"),
		DefaultAdminEmail:         env("DEFAULT_ADMIN_EMAIL", "admin@graduation.local"),
		DefaultAdminPassword:      env("DEFAULT_ADMIN_PASSWORD", "admin123"),
		DefaultSchoolName:         env("DEFAULT_SCHOOL_NAME", "Sekolah Default"),
		DefaultSchoolSlug:         env("DEFAULT_SCHOOL_SLUG", "default-school"),
		DefaultSuperAdminName:     env("DEFAULT_SUPER_ADMIN_NAME", "Super Admin"),
		DefaultSuperAdminEmail:    env("DEFAULT_SUPER_ADMIN_EMAIL", "superadmin@graduation.local"),
		DefaultSuperAdminPassword: env("DEFAULT_SUPER_ADMIN_PASSWORD", "superadmin123"),
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
