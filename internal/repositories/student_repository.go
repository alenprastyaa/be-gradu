package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) Create(ctx context.Context, s *models.Student) error {
	schoolID, ok := authcontext.SchoolID(ctx)
	if !ok {
		return errors.New("school_id tidak tersedia")
	}
	s.SchoolID = schoolID
	return r.db.QueryRow(ctx, `
		INSERT INTO students (school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at
	`, *schoolID, s.Name, s.ClassName, s.Major, s.LaneCode, s.SeatNumber, s.WhatsappNumber, s.Email, s.InvitationCode, s.QRPayload, s.AttendanceStatus).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *StudentRepository) Update(ctx context.Context, s *models.Student) error {
	row := r.db.QueryRow(ctx, `
		UPDATE students
		SET name=$1, class_name=$2, major=$3, lane_code=$4, seat_number=$5, whatsapp_number=$6, email=$7, updated_at=CURRENT_TIMESTAMP
		WHERE id=$8 `+tenantSQL(ctx, 9)+`
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`, tenantArgs(ctx, s.Name, s.ClassName, s.Major, s.LaneCode, s.SeatNumber, s.WhatsappNumber, s.Email, s.ID)...)
	updated, err := scanStudent(row)
	if err != nil {
		return err
	}
	*s = *updated
	return nil
}

func (r *StudentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM students WHERE id=$1`+tenantSQL(ctx, 2), tenantArgs(ctx, id)...)
	return err
}

func (r *StudentRepository) DeleteAll(ctx context.Context) (int64, error) {
	result, err := r.db.Exec(ctx, `DELETE FROM students WHERE 1=1`+tenantSQL(ctx, 1), tenantArgs(ctx)...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *StudentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	row := r.db.QueryRow(ctx, studentSelect()+` WHERE id=$1`+tenantSQL(ctx, 2), tenantArgs(ctx, id)...)
	return scanStudent(row)
}

func (r *StudentRepository) FindByInvitationCode(ctx context.Context, code string) (*models.Student, error) {
	row := r.db.QueryRow(ctx, studentSelect()+` WHERE invitation_code=$1`+tenantSQL(ctx, 2), tenantArgs(ctx, code)...)
	return scanStudent(row)
}

func (r *StudentRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Student, error) {
	if len(ids) == 0 {
		return []models.Student{}, nil
	}

	query := studentSelect() + ` WHERE id IN (`
	args := make([]interface{}, 0, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+1)
		args = append(args, id)
	}
	query += `)`
	if schoolID, ok := authcontext.SchoolID(ctx); ok {
		args = append(args, *schoolID)
		query += fmt.Sprintf(` AND school_id=$%d`, len(args))
	}
	query += ` ORDER BY class_name, major, seat_number, name`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		students = append(students, *student)
	}
	return students, rows.Err()
}

func (r *StudentRepository) FindByQRPayload(ctx context.Context, payload string) (*models.Student, error) {
	row := r.db.QueryRow(ctx, studentSelect()+` WHERE qr_payload=$1`+tenantSQL(ctx, 2), tenantArgs(ctx, payload)...)
	return scanStudent(row)
}

func (r *StudentRepository) FindByEmail(ctx context.Context, email string) (*models.Student, error) {
	row := r.db.QueryRow(ctx, studentSelect()+` WHERE LOWER(email)=LOWER($1)`+tenantSQL(ctx, 2)+` ORDER BY updated_at DESC LIMIT 1`, tenantArgs(ctx, strings.TrimSpace(email))...)
	return scanStudent(row)
}

func (r *StudentRepository) FindByBrevoMessageID(ctx context.Context, messageID string) (*models.Student, error) {
	row := r.db.QueryRow(ctx, studentSelect()+` WHERE email_brevo_message_id=$1`+tenantSQL(ctx, 2), tenantArgs(ctx, strings.TrimSpace(messageID))...)
	return scanStudent(row)
}

func (r *StudentRepository) InvitationCodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM students WHERE invitation_code=$1)`, code).Scan(&exists)
	return exists, err
}

func (r *StudentRepository) CountByClassMajorLane(ctx context.Context, className, major, laneCode string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM students WHERE class_name=$1 AND major=$2 AND lane_code=$3
	`+tenantSQL(ctx, 4), tenantArgs(ctx, className, major, laneCode)...).Scan(&count)
	return count, err
}

func (r *StudentRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE 1=1`+tenantSQL(ctx, 1), tenantArgs(ctx)...).Scan(&count)
	return count, err
}

func (r *StudentRepository) List(ctx context.Context, filter models.StudentFilter) ([]models.Student, error) {
	query := studentSelect()
	clauses, args := studentFilterClauses(filter)
	if schoolID, ok := authcontext.SchoolID(ctx); ok {
		args = append(args, *schoolID)
		clauses = append(clauses, fmt.Sprintf("school_id=$%d", len(args)))
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY class_name, major, seat_number, name"
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

	var students []models.Student
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		students = append(students, *student)
	}
	return students, rows.Err()
}

func (r *StudentRepository) Count(ctx context.Context, filter models.StudentFilter) (int, error) {
	query := `SELECT COUNT(*)::int FROM students`
	clauses, args := studentFilterClauses(filter)
	if schoolID, ok := authcontext.SchoolID(ctx); ok {
		args = append(args, *schoolID)
		clauses = append(clauses, fmt.Sprintf("school_id=$%d", len(args)))
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	var count int
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *StudentRepository) FilterOptions(ctx context.Context) (models.StudentFilterOptions, error) {
	options := models.StudentFilterOptions{Classes: []string{}, Majors: []string{}}

	classRows, err := r.db.Query(ctx, `SELECT DISTINCT class_name FROM students WHERE class_name <> ''`+tenantSQL(ctx, 1)+` ORDER BY class_name`, tenantArgs(ctx)...)
	if err != nil {
		return options, err
	}
	defer classRows.Close()
	for classRows.Next() {
		var className string
		if err := classRows.Scan(&className); err != nil {
			return options, err
		}
		options.Classes = append(options.Classes, className)
	}
	if err := classRows.Err(); err != nil {
		return options, err
	}

	majorRows, err := r.db.Query(ctx, `SELECT DISTINCT major FROM students WHERE major <> ''`+tenantSQL(ctx, 1)+` ORDER BY major`, tenantArgs(ctx)...)
	if err != nil {
		return options, err
	}
	defer majorRows.Close()
	for majorRows.Next() {
		var major string
		if err := majorRows.Scan(&major); err != nil {
			return options, err
		}
		options.Majors = append(options.Majors, major)
	}
	return options, majorRows.Err()
}

func (r *StudentRepository) ListAllOrdered(ctx context.Context) ([]models.Student, error) {
	rows, err := r.db.Query(ctx, studentSelect()+` WHERE 1=1`+tenantSQL(ctx, 1)+` ORDER BY class_name, major, name`, tenantArgs(ctx)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var students []models.Student
	for rows.Next() {
		student, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		students = append(students, *student)
	}
	return students, rows.Err()
}

func (r *StudentRepository) UpdateSeat(ctx context.Context, id uuid.UUID, laneCode, seatNumber string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE students SET lane_code=$1, seat_number=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3
	`+tenantSQL(ctx, 4), tenantArgs(ctx, laneCode, seatNumber, id)...)
	return err
}

func (r *StudentRepository) MarkAttendance(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE students
		SET attendance_status='hadir', attendance_time=CURRENT_TIMESTAMP AT TIME ZONE 'UTC', updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND attendance_status='belum_hadir' `+tenantSQL(ctx, 2)+`
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`, tenantArgs(ctx, id)...)
	return scanStudent(row)
}

func (r *StudentRepository) UpdateAttendanceStatus(ctx context.Context, id uuid.UUID, status string) (*models.Student, error) {
	query := `
		UPDATE students
		SET attendance_status=$2::varchar,
			attendance_time=CASE WHEN $2::text='hadir' THEN CURRENT_TIMESTAMP AT TIME ZONE 'UTC' ELSE NULL END,
			updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 ` + tenantSQL(ctx, 3) + `
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`
	row := r.db.QueryRow(ctx, query, tenantArgs(ctx, id, status)...)
	return scanStudent(row)
}

func (r *StudentRepository) MarkWhatsappSent(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE students
		SET whatsapp_sent_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 `+tenantSQL(ctx, 2)+`
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`, tenantArgs(ctx, id)...)
	return scanStudent(row)
}

func (r *StudentRepository) ResetWhatsappSent(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE students
		SET whatsapp_sent_at=NULL, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 `+tenantSQL(ctx, 2)+`
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`, tenantArgs(ctx, id)...)
	return scanStudent(row)
}

func (r *StudentRepository) MarkEmailSent(ctx context.Context, id uuid.UUID, messageID string) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE students
		SET email_sent_at=COALESCE(email_sent_at, CURRENT_TIMESTAMP),
			email_brevo_message_id=COALESCE(NULLIF($2, ''), email_brevo_message_id),
			updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 `+tenantSQL(ctx, 3)+`
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`, tenantArgs(ctx, id, strings.TrimSpace(messageID))...)
	return scanStudent(row)
}

func (r *StudentRepository) ResetEmailSent(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE students
		SET email_sent_at=NULL, email_brevo_message_id=NULL, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 `+tenantSQL(ctx, 2)+`
		RETURNING id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at
	`, tenantArgs(ctx, id)...)
	return scanStudent(row)
}

func (r *StudentRepository) Summary(ctx context.Context) (map[string]interface{}, error) {
	var total, hadir, belum, classes, majors, lanes int
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE attendance_status='hadir')::int,
			COUNT(*) FILTER (WHERE attendance_status='belum_hadir')::int,
			COUNT(DISTINCT class_name)::int,
			COUNT(DISTINCT major)::int,
			COUNT(DISTINCT lane_code)::int
		FROM students
		WHERE 1=1 `+tenantSQL(ctx, 1), tenantArgs(ctx)...).Scan(&total, &hadir, &belum, &classes, &majors, &lanes)
	if err != nil {
		return nil, err
	}
	percentage := 0.0
	if total > 0 {
		percentage = float64(hadir) / float64(total) * 100
	}
	return map[string]interface{}{
		"total_students":        total,
		"total_hadir":           hadir,
		"total_belum_hadir":     belum,
		"total_classes":         classes,
		"total_majors":          majors,
		"total_lanes":           lanes,
		"attendance_percentage": percentage,
	}, nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func studentSelect() string {
	return `SELECT id, school_id, name, class_name, major, lane_code, seat_number, whatsapp_number, email, invitation_code, qr_payload, attendance_status, attendance_time, whatsapp_sent_at, email_sent_at, email_brevo_message_id, created_at, updated_at FROM students`
}

func studentFilterClauses(filter models.StudentFilter) ([]string, []interface{}) {
	var clauses []string
	var args []interface{}
	add := func(clause string, value interface{}) {
		args = append(args, value)
		clauses = append(clauses, strings.ReplaceAll(clause, "?", fmt.Sprintf("$%d", len(args))))
	}
	if filter.ClassName != "" {
		add("class_name=?", filter.ClassName)
	}
	if filter.Major != "" {
		add("major=?", filter.Major)
	}
	if filter.AttendanceStatus != "" {
		add("attendance_status=?", filter.AttendanceStatus)
	}
	if filter.Search != "" {
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		idx := fmt.Sprintf("$%d", len(args))
		clauses = append(clauses, "(LOWER(name) LIKE "+idx+" OR LOWER(seat_number) LIKE "+idx+")")
	}
	return clauses, args
}

func scanStudent(row pgx.Row) (*models.Student, error) {
	var student models.Student
	if err := row.Scan(
		&student.ID,
		&student.SchoolID,
		&student.Name,
		&student.ClassName,
		&student.Major,
		&student.LaneCode,
		&student.SeatNumber,
		&student.WhatsappNumber,
		&student.Email,
		&student.InvitationCode,
		&student.QRPayload,
		&student.AttendanceStatus,
		&student.AttendanceTime,
		&student.WhatsappSentAt,
		&student.EmailSentAt,
		&student.EmailBrevoMessageID,
		&student.CreatedAt,
		&student.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	student.PopulateSeatNumbers()
	return &student, nil
}

func tenantSQL(ctx context.Context, argIndex int) string {
	if _, ok := authcontext.SchoolID(ctx); !ok {
		return ""
	}
	return fmt.Sprintf(" AND school_id=$%d", argIndex)
}

func tenantArgs(ctx context.Context, args ...interface{}) []interface{} {
	if schoolID, ok := authcontext.SchoolID(ctx); ok {
		args = append(args, *schoolID)
	}
	return args
}
