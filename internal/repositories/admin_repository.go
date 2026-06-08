package repositories

import (
	"context"
	"errors"

	"graduation-invitation/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	db *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM admins WHERE email=$1`, email)
	return scanAdmin(row)
}

func (r *AdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM admins WHERE id=$1`, id)
	return scanAdmin(row)
}

func (r *AdminRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM admins`).Scan(&count)
	return count, err
}

func (r *AdminRepository) Create(ctx context.Context, admin *models.Admin) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO admins (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, admin.Name, admin.Email, admin.PasswordHash).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
}

func scanAdmin(row pgx.Row) (*models.Admin, error) {
	var admin models.Admin
	if err := row.Scan(&admin.ID, &admin.Name, &admin.Email, &admin.PasswordHash, &admin.CreatedAt, &admin.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}
