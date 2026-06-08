package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EmailLog struct {
	ID         uuid.UUID       `json:"id"`
	StudentID  *uuid.UUID      `json:"student_id"`
	Email      string          `json:"email"`
	MessageID  string          `json:"message_id"`
	Subject    string          `json:"subject"`
	Event      string          `json:"event"`
	EventTime  *time.Time      `json:"event_time"`
	Reason     string          `json:"reason"`
	Link       string          `json:"link"`
	RawPayload json.RawMessage `json:"raw_payload"`
	CreatedAt  time.Time       `json:"created_at"`
}

type EmailLogInput struct {
	StudentID  *uuid.UUID
	Email      string
	MessageID  string
	Subject    string
	Event      string
	EventTime  *time.Time
	Reason     string
	Link       string
	RawPayload json.RawMessage
}
