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
	row := r.db.QueryRow(ctx, adminSelect()+` WHERE admins.email=$1`, email)
	return scanAdmin(row)
}

func (r *AdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	row := r.db.QueryRow(ctx, adminSelect()+` WHERE admins.id=$1`, id)
	return scanAdmin(row)
}

func (r *AdminRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM admins`).Scan(&count)
	return count, err
}

func (r *AdminRepository) Create(ctx context.Context, admin *models.Admin) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO admins (name, email, password_hash, role, school_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, admin.Name, admin.Email, admin.PasswordHash, admin.Role, admin.SchoolID).Scan(&admin.ID, &admin.CreatedAt, &admin.UpdatedAt)
}

func (r *AdminRepository) List(ctx context.Context) ([]models.Admin, error) {
	rows, err := r.db.Query(ctx, adminSelect()+` ORDER BY admins.role, schools.name NULLS FIRST, admins.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []models.Admin
	for rows.Next() {
		admin, err := scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		admins = append(admins, *admin)
	}
	return admins, rows.Err()
}

func (r *AdminRepository) Update(ctx context.Context, admin *models.Admin) error {
	row := r.db.QueryRow(ctx, `
		UPDATE admins
		SET name=$2, email=$3, role=$4, school_id=$5, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
		RETURNING id, name, email, password_hash, role, school_id, NULL::text AS school_name, created_at, updated_at
	`, admin.ID, admin.Name, admin.Email, admin.Role, admin.SchoolID)
	updated, err := scanAdmin(row)
	if err != nil {
		return err
	}
	*admin = *updated
	return nil
}

func (r *AdminRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE admins SET password_hash=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, hash)
	return err
}

func (r *AdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM admins WHERE id=$1`, id)
	return err
}

func scanAdmin(row pgx.Row) (*models.Admin, error) {
	var admin models.Admin
	if err := row.Scan(&admin.ID, &admin.Name, &admin.Email, &admin.PasswordHash, &admin.Role, &admin.SchoolID, &admin.SchoolName, &admin.CreatedAt, &admin.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func adminSelect() string {
	return `
		SELECT admins.id, admins.name, admins.email, admins.password_hash, admins.role, admins.school_id, schools.name AS school_name, admins.created_at, admins.updated_at
		FROM admins
		LEFT JOIN schools ON schools.id = admins.school_id
	`
}
