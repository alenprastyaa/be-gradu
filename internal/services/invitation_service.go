package services

import (
	"context"

	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"
)

type InvitationService struct {
	students *StudentService
	events   *EventSettingsService
}

func NewInvitationService(students *StudentService, events *EventSettingsService) *InvitationService {
	return &InvitationService{students: students, events: events}
}

func (s *InvitationService) GetByCode(ctx context.Context, code string) (map[string]interface{}, error) {
	student, err := s.students.GetByInvitationCode(ctx, code)
	if err != nil || student == nil {
		return nil, err
	}
	if student.SchoolID != nil {
		ctx = authcontext.WithSchoolID(ctx, *student.SchoolID)
	}
	event, err := s.events.Get(ctx)
	if err != nil {
		return nil, err
	}
	seatMap, err := s.students.SeatMap(ctx, models.StudentFilter{})
	if err != nil {
		return nil, err
	}
	publicSeatMap := sanitizePublicSeatMap(seatMap)
	return map[string]interface{}{
		"student":  student,
		"event":    event,
		"seat_map": publicSeatMap,
	}, nil
}

func sanitizePublicSeatMap(seatMap map[string]interface{}) map[string]interface{} {
	items, _ := seatMap["items"].([]models.Student)
	publicItems := make([]map[string]interface{}, 0, len(items))
	for _, student := range items {
		student.PopulateSeatNumbers()
		publicItems = append(publicItems, map[string]interface{}{
			"id":                    student.ID,
			"name":                  student.Name,
			"class_name":            student.ClassName,
			"major":                 student.Major,
			"student_seat_number":   student.StudentSeatNumber,
			"companion_seat_number": student.CompanionSeatNumber,
			"attendance_status":     student.AttendanceStatus,
		})
	}
	return map[string]interface{}{
		"items":      publicItems,
		"total":      seatMap["total"],
		"seat_total": seatMap["seat_total"],
	}
}
