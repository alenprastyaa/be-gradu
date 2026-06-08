package repositories

import (
	"context"
	"encoding/json"
	"strings"

	"graduation-invitation/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailLogRepository struct {
	db *pgxpool.Pool
}

func NewEmailLogRepository(db *pgxpool.Pool) *EmailLogRepository {
	return &EmailLogRepository{db: db}
}

func (r *EmailLogRepository) Create(ctx context.Context, input models.EmailLogInput) error {
	raw := input.RawPayload
	if len(raw) == 0 || !json.Valid(raw) {
		raw = json.RawMessage(`{}`)
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO email_logs (student_id, email, message_id, subject, event, event_time, reason, link, raw_payload)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb
		WHERE NOT EXISTS (
			SELECT 1 FROM email_logs
			WHERE COALESCE(message_id, '') = COALESCE($3, '')
				AND email = $2
				AND event = $5
				AND COALESCE(event_time, TIMESTAMP 'epoch') = COALESCE($6::timestamp, TIMESTAMP 'epoch')
		)
	`, input.StudentID, strings.TrimSpace(input.Email), strings.TrimSpace(input.MessageID), strings.TrimSpace(input.Subject),
		strings.TrimSpace(input.Event), input.EventTime, strings.TrimSpace(input.Reason), strings.TrimSpace(input.Link), []byte(raw))
	return err
}
