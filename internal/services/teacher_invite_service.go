package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"
	"graduation-invitation/internal/utils"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type TeacherInviteService struct {
	repo *repositories.TeacherInviteRepository
}

func NewTeacherInviteService(repo *repositories.TeacherInviteRepository) *TeacherInviteService {
	return &TeacherInviteService{repo: repo}
}

func (s *TeacherInviteService) List(ctx context.Context, filter models.TeacherInviteFilter) (*models.TeacherInviteListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + filter.Limit - 1) / filter.Limit
	}
	items, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	options, err := s.repo.FilterOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &models.TeacherInviteListResponse{
		Items: items,
		Pagination: models.PaginationMeta{
			Page:        filter.Page,
			Limit:       filter.Limit,
			Total:       total,
			TotalPages:  totalPages,
			HasNext:     filter.Page < totalPages,
			HasPrevious: filter.Page > 1 && totalPages > 0,
		},
		Filters: options,
	}, nil
}

func (s *TeacherInviteService) Create(ctx context.Context, input models.TeacherInviteInput) (*models.TeacherInvite, error) {
	clean, err := s.cleanInput(input)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 8; i++ {
		code, err := s.generateUniqueInvitationCode(ctx)
		if err != nil {
			return nil, err
		}
		item := &models.TeacherInvite{
			Name:             clean.Name,
			Position:         clean.Position,
			WhatsappNumber:   clean.WhatsappNumber,
			Email:            clean.Email,
			InvitationCode:   code,
			QRPayload:        "GRAD-TEACHER:" + code,
			AttendanceStatus: models.AttendanceBelumHadir,
		}
		if err := s.repo.Create(ctx, item); err == nil {
			return item, nil
		} else if !repositories.IsUniqueViolation(err) {
			return nil, err
		}
	}
	return nil, errors.New("gagal membuat invitation_code guru unik")
}

func (s *TeacherInviteService) Update(ctx context.Context, id uuid.UUID, input models.TeacherInviteInput) (*models.TeacherInvite, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil || item == nil {
		return item, err
	}
	clean, err := s.cleanInput(input)
	if err != nil {
		return nil, err
	}
	item.Name = clean.Name
	item.Position = clean.Position
	item.WhatsappNumber = clean.WhatsappNumber
	item.Email = clean.Email
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *TeacherInviteService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *TeacherInviteService) Get(ctx context.Context, id uuid.UUID) (*models.TeacherInvite, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TeacherInviteService) GetByInvitationCode(ctx context.Context, code string) (*models.TeacherInvite, error) {
	return s.repo.FindByInvitationCode(ctx, code)
}

func (s *TeacherInviteService) FindByQRPayload(ctx context.Context, payload string) (*models.TeacherInvite, error) {
	return s.repo.FindByQRPayload(ctx, payload)
}

func (s *TeacherInviteService) MarkAttendance(ctx context.Context, id uuid.UUID) (*models.TeacherInvite, error) {
	return s.repo.MarkAttendance(ctx, id)
}

func (s *TeacherInviteService) UpdateAttendanceStatus(ctx context.Context, id uuid.UUID, status string) (*models.TeacherInvite, error) {
	status = strings.TrimSpace(status)
	if status != models.AttendanceHadir && status != models.AttendanceBelumHadir {
		return nil, errors.New("status kehadiran tidak valid")
	}
	return s.repo.UpdateAttendanceStatus(ctx, id, status)
}

func (s *TeacherInviteService) Import(ctx context.Context, rows []utils.TeacherImportRow) (map[string]interface{}, error) {
	var created []models.TeacherInvite
	var rowErrors []string
	for _, row := range rows {
		email := strings.TrimSpace(row.Email)
		var emailPtr *string
		if email != "" {
			emailPtr = &email
		}
		item, err := s.Create(ctx, models.TeacherInviteInput{
			Name:           row.Name,
			Position:       row.Position,
			WhatsappNumber: row.WhatsappNumber,
			Email:          emailPtr,
		})
		if err != nil {
			rowErrors = append(rowErrors, fmt.Sprintf("Baris %d: %s", row.RowNumber, err.Error()))
			continue
		}
		created = append(created, *item)
	}
	if len(rowErrors) > 0 {
		return map[string]interface{}{"created": created, "errors": rowErrors}, errors.New("sebagian data import tidak valid")
	}
	return map[string]interface{}{"created": created, "errors": []string{}}, nil
}

func (s *TeacherInviteService) ImportTemplateXLSX() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Template Import Guru"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"name", "position", "whatsapp_number", "email"}
	descriptions := []string{
		"Nama lengkap guru",
		"Jabatan, contoh: Guru / Wali Kelas / Kepala Sekolah",
		"Nomor WhatsApp aktif (opsional)",
		"Email opsional",
	}
	sample := []string{"Ibu Siti Aminah", "Guru", "081234567890", "siti@example.com"}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"334155"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	noteStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Color: "64748B"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F8FAFC"}, Pattern: 1},
	})
	if err != nil {
		return nil, err
	}
	textStyle, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return nil, err
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		noteCell, _ := excelize.CoordinatesToCellName(i+1, 2)
		sampleCell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sheet, cell, header)
		f.SetCellValue(sheet, noteCell, descriptions[i])
		if header == "whatsapp_number" {
			f.SetCellStr(sheet, sampleCell, sample[i])
		} else {
			f.SetCellValue(sheet, sampleCell, sample[i])
		}
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		f.SetCellStyle(sheet, noteCell, noteCell, noteStyle)
	}

	_ = f.SetColWidth(sheet, "A", "A", 28)
	_ = f.SetColWidth(sheet, "B", "B", 28)
	_ = f.SetColWidth(sheet, "C", "C", 20)
	_ = f.SetColWidth(sheet, "D", "D", 28)
	_ = f.SetRowHeight(sheet, 1, 24)
	_ = f.SetRowHeight(sheet, 2, 34)
	_ = f.SetCellStyle(sheet, "A3", "D3", textStyle)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *TeacherInviteService) cleanInput(input models.TeacherInviteInput) (models.TeacherInviteInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Position = strings.TrimSpace(input.Position)
	input.WhatsappNumber = strings.TrimSpace(input.WhatsappNumber)
	if err := utils.ValidateRequired(input.Name, "name"); err != nil {
		return input, err
	}
	if input.Position == "" {
		input.Position = "Guru"
	}
	return input, nil
}

func (s *TeacherInviteService) generateUniqueInvitationCode(ctx context.Context) (string, error) {
	for i := 0; i < 20; i++ {
		random, err := utils.RandomString(8)
		if err != nil {
			return "", err
		}
		code := "GRAD-TEACHER-" + random
		exists, err := s.repo.InvitationCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", errors.New("tidak dapat membuat invitation_code guru unik")
}
