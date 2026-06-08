package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

type SuperAdminService struct {
	schools *repositories.SchoolRepository
	admins  *repositories.AdminRepository
}

func NewSuperAdminService(schools *repositories.SchoolRepository, admins *repositories.AdminRepository) *SuperAdminService {
	return &SuperAdminService{schools: schools, admins: admins}
}

func (s *SuperAdminService) ListSchools(ctx context.Context) ([]models.School, error) {
	return s.schools.List(ctx)
}

func (s *SuperAdminService) CreateSchool(ctx context.Context, input models.SchoolInput) (*models.School, error) {
	school := schoolFromInput(input)
	if school.Name == "" {
		return nil, errors.New("nama sekolah wajib diisi")
	}
	if school.Slug == "" {
		school.Slug = slugify(school.Name)
	}
	if school.Slug == "" {
		return nil, errors.New("slug sekolah tidak valid")
	}
	if err := s.schools.Create(ctx, school); err != nil {
		return nil, err
	}
	return school, nil
}

func (s *SuperAdminService) UpdateSchool(ctx context.Context, id uuid.UUID, input models.SchoolInput) (*models.School, error) {
	school, err := s.schools.FindByID(ctx, id)
	if err != nil || school == nil {
		return school, err
	}
	next := schoolFromInput(input)
	if next.Name == "" {
		return nil, errors.New("nama sekolah wajib diisi")
	}
	if next.Slug == "" {
		next.Slug = slugify(next.Name)
	}
	school.Name = next.Name
	school.Slug = next.Slug
	school.Address = next.Address
	school.LogoURL = next.LogoURL
	school.IsActive = next.IsActive
	if err := s.schools.Update(ctx, school); err != nil {
		return nil, err
	}
	return school, nil
}

func (s *SuperAdminService) DeleteSchool(ctx context.Context, id uuid.UUID) error {
	return s.schools.Delete(ctx, id)
}

func (s *SuperAdminService) ListAdmins(ctx context.Context) ([]models.Admin, error) {
	return s.admins.List(ctx)
}

func (s *SuperAdminService) CreateAdmin(ctx context.Context, input models.AdminInput) (*models.Admin, error) {
	admin, err := s.adminFromInput(ctx, input, true)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(input.Password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	admin.PasswordHash = string(hash)
	if err := s.admins.Create(ctx, admin); err != nil {
		return nil, err
	}
	created, err := s.admins.FindByID(ctx, admin.ID)
	if err != nil || created == nil {
		return admin, err
	}
	return created, nil
}

func (s *SuperAdminService) UpdateAdmin(ctx context.Context, id uuid.UUID, input models.AdminInput) (*models.Admin, error) {
	admin, err := s.admins.FindByID(ctx, id)
	if err != nil || admin == nil {
		return admin, err
	}
	next, err := s.adminFromInput(ctx, input, false)
	if err != nil {
		return nil, err
	}
	admin.Name = next.Name
	admin.Email = next.Email
	admin.Role = next.Role
	admin.SchoolID = next.SchoolID
	if err := s.admins.Update(ctx, admin); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(input.Password)), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		if err := s.admins.UpdatePassword(ctx, id, string(hash)); err != nil {
			return nil, err
		}
	}
	return s.admins.FindByID(ctx, id)
}

func (s *SuperAdminService) DeleteAdmin(ctx context.Context, id uuid.UUID) error {
	info, _ := authcontext.FromContext(ctx)
	if info.AdminID == id {
		return errors.New("super admin tidak bisa menghapus akun sendiri")
	}
	return s.admins.Delete(ctx, id)
}

func (s *SuperAdminService) adminFromInput(ctx context.Context, input models.AdminInput, requirePassword bool) (*models.Admin, error) {
	admin := &models.Admin{
		Name:  strings.TrimSpace(input.Name),
		Email: strings.ToLower(strings.TrimSpace(input.Email)),
		Role:  strings.TrimSpace(input.Role),
	}
	if admin.Role == "" {
		admin.Role = authcontext.RoleSchoolAdmin
	}
	if admin.Name == "" || admin.Email == "" {
		return nil, errors.New("nama dan email admin wajib diisi")
	}
	if requirePassword && strings.TrimSpace(input.Password) == "" {
		return nil, errors.New("password admin wajib diisi")
	}
	if admin.Role != authcontext.RoleSuperAdmin && admin.Role != authcontext.RoleSchoolAdmin {
		return nil, errors.New("role admin tidak valid")
	}
	if admin.Role == authcontext.RoleSchoolAdmin {
		if input.SchoolID == nil || *input.SchoolID == uuid.Nil {
			return nil, errors.New("admin sekolah wajib memilih sekolah")
		}
		school, err := s.schools.FindByID(ctx, *input.SchoolID)
		if err != nil {
			return nil, err
		}
		if school == nil {
			return nil, errors.New("sekolah tidak ditemukan")
		}
		admin.SchoolID = input.SchoolID
	}
	return admin, nil
}

func schoolFromInput(input models.SchoolInput) *models.School {
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	return &models.School{
		Name:     strings.TrimSpace(input.Name),
		Slug:     slugify(input.Slug),
		Address:  strings.TrimSpace(input.Address),
		LogoURL:  strings.TrimSpace(input.LogoURL),
		IsActive: active,
	}
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = nonSlugChars.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}
