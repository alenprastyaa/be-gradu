package services

import (
	"context"
	"errors"
	"strings"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"
)

type AttendanceService struct {
	students *repositories.StudentRepository
}

func NewAttendanceService(students *repositories.StudentRepository) *AttendanceService {
	return &AttendanceService{students: students}
}

func (s *AttendanceService) Scan(ctx context.Context, payload string) (map[string]interface{}, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, errors.New("QR payload wajib diisi")
	}
	student, err := s.students.FindByQRPayload(ctx, payload)
	if err != nil {
		return nil, err
	}
	if student == nil {
		code := strings.TrimPrefix(payload, "GRAD-ATTENDANCE:")
		if code != payload {
			student, err = s.students.FindByInvitationCode(ctx, code)
			if err != nil {
				return nil, err
			}
		}
	}
	if student == nil {
		return nil, errors.New("QR Code tidak valid")
	}
	if student.AttendanceStatus == models.AttendanceHadir {
		return map[string]interface{}{"status": "already_attended", "student": student}, nil
	}
	updated, err := s.students.MarkAttendance(ctx, student.ID)
	if err != nil || updated == nil {
		return nil, err
	}
	return map[string]interface{}{"status": "success", "student": updated}, nil
}

func (s *AttendanceService) Summary(ctx context.Context) (map[string]interface{}, error) {
	return s.students.Summary(ctx)
}
