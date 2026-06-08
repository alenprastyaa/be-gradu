package models

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	AttendanceBelumHadir = "belum_hadir"
	AttendanceHadir      = "hadir"
)

var seatNumberPattern = regexp.MustCompile(`\d+`)

type Student struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	ClassName           string     `json:"class_name"`
	Major               string     `json:"major"`
	LaneCode            string     `json:"lane_code"`
	SeatNumber          string     `json:"seat_number"`
	StudentSeatNumber   string     `json:"student_seat_number"`
	CompanionSeatNumber string     `json:"companion_seat_number"`
	WhatsappNumber      string     `json:"whatsapp_number"`
	Email               *string    `json:"email"`
	InvitationCode      string     `json:"invitation_code"`
	QRPayload           string     `json:"qr_payload"`
	AttendanceStatus    string     `json:"attendance_status"`
	AttendanceTime      *time.Time `json:"attendance_time"`
	WhatsappSentAt      *time.Time `json:"whatsapp_sent_at"`
	EmailSentAt         *time.Time `json:"email_sent_at"`
	EmailBrevoMessageID *string    `json:"email_brevo_message_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (s *Student) PopulateSeatNumbers() {
	numbers := seatNumberPattern.FindAllString(s.SeatNumber, -1)
	if len(numbers) > 0 {
		s.StudentSeatNumber = numbers[0]
	}
	if len(numbers) > 1 {
		s.CompanionSeatNumber = numbers[1]
	}
}

type StudentInput struct {
	Name           string  `json:"name"`
	ClassName      string  `json:"class_name"`
	Major          string  `json:"major"`
	WhatsappNumber string  `json:"whatsapp_number"`
	Email          *string `json:"email"`
}

type StudentFilter struct {
	ClassName        string
	Major            string
	AttendanceStatus string
	Search           string
	Page             int
	Limit            int
}

type PaginationMeta struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

type StudentFilterOptions struct {
	Classes []string `json:"classes"`
	Majors  []string `json:"majors"`
}

type StudentListResponse struct {
	Items      []Student            `json:"items"`
	Pagination PaginationMeta       `json:"pagination"`
	Filters    StudentFilterOptions `json:"filters"`
}
