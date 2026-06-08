package services

import (
	"context"
	"errors"
	"time"

	"graduation-invitation/config"
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   *repositories.AdminRepository
	config config.Config
}

func NewAuthService(repo *repositories.AdminRepository, cfg config.Config) *AuthService {
	return &AuthService{repo: repo, config: cfg}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.Admin, error) {
	admin, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if admin == nil || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return "", nil, errors.New("email atau password salah")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   admin.ID.String(),
		"email": admin.Email,
		"role":  "admin",
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(s.config.JWTSecret))
	return signed, admin, err
}

func (s *AuthService) Me(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *AuthService) EnsureDefaultAdmin(ctx context.Context) error {
	count, err := s.repo.Count(ctx)
	if err != nil || count > 0 {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.config.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &models.Admin{
		Name: s.config.DefaultAdminName, Email: s.config.DefaultAdminEmail, PasswordHash: string(hash),
	})
}
