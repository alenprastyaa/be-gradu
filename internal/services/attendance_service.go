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
	teachers *repositories.TeacherInviteRepository
}

func NewAttendanceService(students *repositories.StudentRepository, teachers *repositories.TeacherInviteRepository) *AttendanceService {
	return &AttendanceService{students: students, teachers: teachers}
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
		teacher, teacherErr := s.teachers.FindByQRPayload(ctx, payload)
		if teacherErr != nil {
			return nil, teacherErr
		}
		if teacher == nil {
			code := strings.TrimPrefix(payload, "GRAD-TEACHER:")
			if code != payload {
				teacher, teacherErr = s.teachers.FindByInvitationCode(ctx, code)
				if teacherErr != nil {
					return nil, teacherErr
				}
			}
		}
		if teacher != nil {
			generic := mapTeacherInviteAsStudentLike(teacher)
			if teacher.AttendanceStatus == models.AttendanceHadir {
				return map[string]interface{}{"status": "already_attended", "student": generic}, nil
			}
			updated, err := s.teachers.MarkAttendance(ctx, teacher.ID)
			if err != nil || updated == nil {
				return nil, err
			}
			return map[string]interface{}{"status": "success", "student": mapTeacherInviteAsStudentLike(updated)}, nil
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

func mapTeacherInviteAsStudentLike(item *models.TeacherInvite) map[string]interface{} {
	return map[string]interface{}{
		"id":                    item.ID,
		"name":                  item.Name,
		"class_name":            "Guru",
		"major":                 item.Position,
		"student_seat_number":   "",
		"companion_seat_number": "",
		"attendance_status":     item.AttendanceStatus,
		"attendance_time":       item.AttendanceTime,
		"qr_payload":            item.QRPayload,
		"invitation_code":       item.InvitationCode,
		"invite_type":           "teacher",
	}
}
