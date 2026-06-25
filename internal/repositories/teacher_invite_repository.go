package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeacherInviteRepository struct {
	db *pgxpool.Pool

	schemaMu                sync.RWMutex
	hasTeacherInviteSchoolID *bool
}

func NewTeacherInviteRepository(db *pgxpool.Pool) *TeacherInviteRepository {
	return &TeacherInviteRepository{db: db}
}

func (r *TeacherInviteRepository) List(ctx context.Context, filter models.TeacherInviteFilter) ([]models.TeacherInvite, error) {
	query, err := r.teacherInviteSelect(ctx)
	if err != nil {
		return nil, err
	}
	clauses, args := teacherInviteFilterClauses(filter)
	if schoolID, ok := authcontext.SchoolID(ctx); ok {
		hasSchoolID, err := r.hasSchoolIDColumn(ctx)
		if err != nil {
			return nil, err
		}
		if hasSchoolID {
			args = append(args, *schoolID)
			clauses = append(clauses, fmt.Sprintf("school_id=$%d", len(args)))
		}
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY name"
	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		limitArg := fmt.Sprintf("$%d", len(args))
		offset := (filter.Page - 1) * filter.Limit
		if offset < 0 {
			offset = 0
		}
		args = append(args, offset)
		offsetArg := fmt.Sprintf("$%d", len(args))
		query += " LIMIT " + limitArg + " OFFSET " + offsetArg
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.TeacherInvite
	for rows.Next() {
		item, err := scanTeacherInvite(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *TeacherInviteRepository) Count(ctx context.Context, filter models.TeacherInviteFilter) (int, error) {
	query := `SELECT COUNT(*)::int FROM teacher_invites`
	clauses, args := teacherInviteFilterClauses(filter)
	if schoolID, ok := authcontext.SchoolID(ctx); ok {
		hasSchoolID, err := r.hasSchoolIDColumn(ctx)
		if err != nil {
			return 0, err
		}
		if hasSchoolID {
			args = append(args, *schoolID)
			clauses = append(clauses, fmt.Sprintf("school_id=$%d", len(args)))
		}
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *TeacherInviteRepository) FilterOptions(ctx context.Context) (models.TeacherInviteFilters, error) {
	options := models.TeacherInviteFilters{Positions: []string{}}
	tenantClause, tenantArgs, err := r.teacherInviteTenantClause(ctx, 1)
	if err != nil {
		return options, err
	}
	rows, err := r.db.Query(ctx, `SELECT DISTINCT position FROM teacher_invites WHERE position <> ''`+tenantClause+` ORDER BY position`, tenantArgs...)
	if err != nil {
		return options, err
	}
	defer rows.Close()
	for rows.Next() {
		var position string
		if err := rows.Scan(&position); err != nil {
			return options, err
		}
		options.Positions = append(options.Positions, position)
	}
	return options, rows.Err()
}

func (r *TeacherInviteRepository) Create(ctx context.Context, item *models.TeacherInvite) error {
	hasSchoolID, err := r.hasSchoolIDColumn(ctx)
	if err != nil {
		return err
	}
	if hasSchoolID {
		schoolID, ok := authcontext.SchoolID(ctx)
		if !ok {
			return errors.New("school_id tidak tersedia")
		}
		item.SchoolID = schoolID
		return r.db.QueryRow(ctx, `
			INSERT INTO teacher_invites (school_id, name, position, whatsapp_number, email, invitation_code, qr_payload, attendance_status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id, created_at, updated_at
		`, *schoolID, item.Name, item.Position, item.WhatsappNumber, item.Email, item.InvitationCode, item.QRPayload, item.AttendanceStatus).
			Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	}
	item.SchoolID = nil
	return r.db.QueryRow(ctx, `
		INSERT INTO teacher_invites (name, position, whatsapp_number, email, invitation_code, qr_payload, attendance_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at, updated_at
	`, item.Name, item.Position, item.WhatsappNumber, item.Email, item.InvitationCode, item.QRPayload, item.AttendanceStatus).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *TeacherInviteRepository) Update(ctx context.Context, item *models.TeacherInvite) error {
	tenantClause, tenantArgs := "", []interface{}{item.Name, item.Position, item.WhatsappNumber, item.Email, item.ID}
	tenantClause, tenantArgs, err := r.teacherInviteTenantClauseWithArgs(ctx, 6, tenantArgs...)
	if err != nil {
		return err
	}
	row := r.db.QueryRow(ctx, `
		UPDATE teacher_invites
		SET name=$1, position=$2, whatsapp_number=$3, email=$4, updated_at=CURRENT_TIMESTAMP
		WHERE id=$5`+tenantClause+`
		RETURNING `+teacherInviteReturningColumns()+`
	`, tenantArgs...)
	updated, err := scanTeacherInvite(row)
	if err != nil {
		return err
	}
	*item = *updated
	return nil
}

func (r *TeacherInviteRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.TeacherInvite, error) {
	selectSQL, err := r.teacherInviteSelect(ctx)
	if err != nil {
		return nil, err
	}
	tenantClause, args, err := r.teacherInviteTenantClauseWithArgs(ctx, 2, id)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, selectSQL+` WHERE id=$1`+tenantClause, args...)
	return scanTeacherInvite(row)
}

func (r *TeacherInviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantClause, args, err := r.teacherInviteTenantClauseWithArgs(ctx, 2, id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM teacher_invites WHERE id=$1`+tenantClause, args...)
	return err
}

func (r *TeacherInviteRepository) FindByQRPayload(ctx context.Context, payload string) (*models.TeacherInvite, error) {
	selectSQL, err := r.teacherInviteSelect(ctx)
	if err != nil {
		return nil, err
	}
	tenantClause, args, err := r.teacherInviteTenantClauseWithArgs(ctx, 2, payload)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, selectSQL+` WHERE qr_payload=$1`+tenantClause, args...)
	return scanTeacherInvite(row)
}

func (r *TeacherInviteRepository) FindByInvitationCode(ctx context.Context, code string) (*models.TeacherInvite, error) {
	selectSQL, err := r.teacherInviteSelect(ctx)
	if err != nil {
		return nil, err
	}
	tenantClause, args, err := r.teacherInviteTenantClauseWithArgs(ctx, 2, code)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, selectSQL+` WHERE invitation_code=$1`+tenantClause, args...)
	return scanTeacherInvite(row)
}

func (r *TeacherInviteRepository) MarkAttendance(ctx context.Context, id uuid.UUID) (*models.TeacherInvite, error) {
	tenantClause, args, err := r.teacherInviteTenantClauseWithArgs(ctx, 3, models.AttendanceHadir, id)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		UPDATE teacher_invites
		SET attendance_status=$1, attendance_time=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP
		WHERE id=$2`+tenantClause+`
		RETURNING `+teacherInviteReturningColumns()+`
	`, args...)
	return scanTeacherInvite(row)
}

func (r *TeacherInviteRepository) InvitationCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	tenantClause, args, err := r.teacherInviteTenantClauseWithArgs(ctx, 2, code)
	if err != nil {
		return false, err
	}
	err = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teacher_invites WHERE invitation_code=$1`+tenantClause+`)`, args...).Scan(&exists)
	return exists, err
}

func (r *TeacherInviteRepository) teacherInviteSelect(ctx context.Context) (string, error) {
	hasSchoolID, err := r.hasSchoolIDColumn(ctx)
	if err != nil {
		return "", err
	}
	if hasSchoolID {
		return `SELECT ` + teacherInviteReturningColumns() + ` FROM teacher_invites`, nil
	}
	return `SELECT ` + teacherInviteReturningColumnsLegacy() + ` FROM teacher_invites`, nil
}

func (r *TeacherInviteRepository) teacherInviteTenantClause(ctx context.Context, nextArgIndex int) (string, []interface{}, error) {
	return r.teacherInviteTenantClauseWithArgs(ctx, nextArgIndex)
}

func (r *TeacherInviteRepository) teacherInviteTenantClauseWithArgs(ctx context.Context, nextArgIndex int, args ...interface{}) (string, []interface{}, error) {
	hasSchoolID, err := r.hasSchoolIDColumn(ctx)
	if err != nil {
		return "", nil, err
	}
	if !hasSchoolID {
		return "", args, nil
	}
	return tenantSQL(ctx, nextArgIndex), tenantArgs(ctx, args...), nil
}

func (r *TeacherInviteRepository) hasSchoolIDColumn(ctx context.Context) (bool, error) {
	r.schemaMu.RLock()
	if r.hasTeacherInviteSchoolID != nil {
		value := *r.hasTeacherInviteSchoolID
		r.schemaMu.RUnlock()
		return value, nil
	}
	r.schemaMu.RUnlock()

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'teacher_invites'
			  AND column_name = 'school_id'
		)
	`).Scan(&exists)
	if err != nil {
		return false, err
	}

	r.schemaMu.Lock()
	r.hasTeacherInviteSchoolID = &exists
	r.schemaMu.Unlock()
	return exists, nil
}

func teacherInviteReturningColumns() string {
	return `id, school_id, name, position, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, created_at, updated_at`
}

func teacherInviteReturningColumnsLegacy() string {
	return `id, NULL::uuid AS school_id, name, position, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, created_at, updated_at`
}

func scanTeacherInvite(row pgx.Row) (*models.TeacherInvite, error) {
	var item models.TeacherInvite
	if err := row.Scan(
		&item.ID,
		&item.SchoolID,
		&item.Name,
		&item.Position,
		&item.WhatsappNumber,
		&item.Email,
		&item.InvitationCode,
		&item.QRPayload,
		&item.AttendanceStatus,
		&item.AttendanceTime,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func teacherInviteFilterClauses(filter models.TeacherInviteFilter) ([]string, []interface{}) {
	var clauses []string
	var args []interface{}
	add := func(sql string, value interface{}) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(sql, len(args)))
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		search = "%" + strings.ToLower(search) + "%"
		add("(LOWER(name) LIKE $%d OR LOWER(position) LIKE $%d)", search)
		args = append(args, search)
		clauses[len(clauses)-1] = fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(position) LIKE $%d)", len(args)-1, len(args))
	}
	if status := strings.TrimSpace(filter.AttendanceStatus); status != "" {
		add("attendance_status=$%d", status)
	}
	return clauses, args
}
