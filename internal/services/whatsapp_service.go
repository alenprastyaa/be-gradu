package services

import (
	"context"
	"errors"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"
	"graduation-invitation/internal/utils"

	"github.com/google/uuid"
)

type WhatsappService struct {
	repo      *repositories.StudentRepository
	events    *EventSettingsService
	publicURL string
}

func NewWhatsappService(repo *repositories.StudentRepository, events *EventSettingsService, publicURL string) *WhatsappService {
	return &WhatsappService{repo: repo, events: events, publicURL: publicURL}
}

func (s *WhatsappService) LinkForStudent(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	student, err := s.repo.FindByID(ctx, id)
	if err != nil || student == nil {
		return nil, err
	}
	return s.link(ctx, student)
}

func (s *WhatsappService) MarkSent(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	student, err := s.repo.MarkWhatsappSent(ctx, id)
	if err != nil || student == nil {
		return nil, err
	}
	return s.link(ctx, student)
}

func (s *WhatsappService) ResetSent(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	student, err := s.repo.ResetWhatsappSent(ctx, id)
	if err != nil || student == nil {
		return nil, err
	}
	return s.link(ctx, student)
}

func (s *WhatsappService) Links(ctx context.Context) ([]map[string]interface{}, error) {
	students, err := s.repo.List(ctx, models.StudentFilter{})
	if err != nil {
		return nil, err
	}
	links := make([]map[string]interface{}, 0, len(students))
	for i := range students {
		link, err := s.link(ctx, &students[i])
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (s *WhatsappService) LinksForStudents(ctx context.Context, ids []uuid.UUID) ([]map[string]interface{}, error) {
	students, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	links := make([]map[string]interface{}, 0, len(students))
	for i := range students {
		link, err := s.link(ctx, &students[i])
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (s *WhatsappService) link(ctx context.Context, student *models.Student) (map[string]interface{}, error) {
	if student == nil {
		return nil, errors.New("siswa tidak ditemukan")
	}
	settings, err := s.events.Get(ctx)
	if err != nil {
		return nil, err
	}
	invitationLink := s.publicURL + "/" + student.InvitationCode
	message := RenderInvitationTemplate(settings.WhatsappTemplate, settings, student, invitationLink)
	return map[string]interface{}{
		"student_id":       student.ID,
		"name":             student.Name,
		"whatsapp_number":  student.WhatsappNumber,
		"whatsapp_sent_at": student.WhatsappSentAt,
		"link":             utils.WhatsappLink(student.WhatsappNumber, message),
	}, nil
}
