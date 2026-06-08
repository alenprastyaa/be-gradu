package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"graduation-invitation/config"
	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo    *repositories.AdminRepository
	schools *repositories.SchoolRepository
	config  config.Config
}

func NewAuthService(repo *repositories.AdminRepository, schools *repositories.SchoolRepository, cfg config.Config) *AuthService {
	return &AuthService{repo: repo, schools: schools, config: cfg}
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
		"sub":       admin.ID.String(),
		"email":     admin.Email,
		"role":      admin.Role,
		"school_id": schoolIDClaim(admin),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
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
	school, err := s.ensureDefaultSchool(ctx)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.config.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &models.Admin{
		Name:         s.config.DefaultAdminName,
		Email:        s.config.DefaultAdminEmail,
		PasswordHash: string(hash),
		Role:         authcontext.RoleSchoolAdmin,
		SchoolID:     &school.ID,
	})
}

func (s *AuthService) EnsureDefaultSuperAdmin(ctx context.Context) error {
	email := strings.TrimSpace(s.config.DefaultSuperAdminEmail)
	if email == "" {
		return nil
	}
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil || existing != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.config.DefaultSuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &models.Admin{
		Name:         s.config.DefaultSuperAdminName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         authcontext.RoleSuperAdmin,
	})
}

func (s *AuthService) ensureDefaultSchool(ctx context.Context) (*models.School, error) {
	slug := strings.TrimSpace(s.config.DefaultSchoolSlug)
	if slug == "" {
		slug = "default-school"
	}
	school, err := s.schools.FindBySlug(ctx, slug)
	if err != nil || school != nil {
		return school, err
	}
	school = &models.School{
		Name:     strings.TrimSpace(s.config.DefaultSchoolName),
		Slug:     slug,
		IsActive: true,
	}
	if school.Name == "" {
		school.Name = "Sekolah Default"
	}
	if err := s.schools.Create(ctx, school); err != nil {
		return nil, err
	}
	return school, nil
}

func schoolIDClaim(admin *models.Admin) string {
	if admin.SchoolID == nil {
		return ""
	}
	return admin.SchoolID.String()
}
