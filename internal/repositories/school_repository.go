package repositories

import (
	"context"
	"errors"
	"strings"

	"graduation-invitation/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SchoolRepository struct {
	db *pgxpool.Pool
}

func NewSchoolRepository(db *pgxpool.Pool) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) List(ctx context.Context) ([]models.School, error) {
	rows, err := r.db.Query(ctx, schoolSelect()+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schools []models.School
	for rows.Next() {
		school, err := scanSchool(rows)
		if err != nil {
			return nil, err
		}
		schools = append(schools, *school)
	}
	return schools, rows.Err()
}

func (r *SchoolRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.School, error) {
	row := r.db.QueryRow(ctx, schoolSelect()+` WHERE id=$1`, id)
	return scanSchool(row)
}

func (r *SchoolRepository) FindBySlug(ctx context.Context, slug string) (*models.School, error) {
	row := r.db.QueryRow(ctx, schoolSelect()+` WHERE slug=$1`, strings.TrimSpace(slug))
	return scanSchool(row)
}

func (r *SchoolRepository) Create(ctx context.Context, school *models.School) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO schools (name, slug, address, logo_url, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, school.Name, school.Slug, school.Address, school.LogoURL, school.IsActive).Scan(&school.ID, &school.CreatedAt, &school.UpdatedAt)
}

func (r *SchoolRepository) Update(ctx context.Context, school *models.School) error {
	row := r.db.QueryRow(ctx, `
		UPDATE schools
		SET name=$2, slug=$3, address=$4, logo_url=$5, is_active=$6, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
		RETURNING id, name, slug, address, logo_url, is_active, created_at, updated_at
	`, school.ID, school.Name, school.Slug, school.Address, school.LogoURL, school.IsActive)
	updated, err := scanSchool(row)
	if err != nil {
		return err
	}
	*school = *updated
	return nil
}

func (r *SchoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM schools WHERE id=$1`, id)
	return err
}

func scanSchool(row pgx.Row) (*models.School, error) {
	var school models.School
	if err := row.Scan(&school.ID, &school.Name, &school.Slug, &school.Address, &school.LogoURL, &school.IsActive, &school.CreatedAt, &school.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &school, nil
}

func schoolSelect() string {
	return `SELECT id, name, slug, address, logo_url, is_active, created_at, updated_at FROM schools`
}
