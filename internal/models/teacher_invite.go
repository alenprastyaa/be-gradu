package models

import (
	"time"

	"github.com/google/uuid"
)

type TeacherInvite struct {
	ID               uuid.UUID  `json:"id"`
	SchoolID         *uuid.UUID `json:"school_id,omitempty"`
	Name             string     `json:"name"`
	Position         string     `json:"position"`
	WhatsappNumber   string     `json:"whatsapp_number"`
	Email            *string    `json:"email"`
	InvitationCode   string     `json:"invitation_code"`
	QRPayload        string     `json:"qr_payload"`
	AttendanceStatus string     `json:"attendance_status"`
	AttendanceTime   *time.Time `json:"attendance_time"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TeacherInviteInput struct {
	Name           string  `json:"name"`
	Position       string  `json:"position"`
	WhatsappNumber string  `json:"whatsapp_number"`
	Email          *string `json:"email"`
}

type TeacherInviteFilter struct {
	Search           string
	AttendanceStatus string
	Page             int
	Limit            int
}

type TeacherInviteListResponse struct {
	Items      []TeacherInvite      `json:"items"`
	Pagination PaginationMeta       `json:"pagination"`
	Filters    TeacherInviteFilters `json:"filters"`
}

type TeacherInviteFilters struct {
	Positions []string `json:"positions"`
}
