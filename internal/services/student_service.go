package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"
	"graduation-invitation/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

type StudentService struct {
	repo      *repositories.StudentRepository
	seat      *SeatService
	db        *pgxpool.Pool
	publicURL string
}

func NewStudentService(repo *repositories.StudentRepository, seat *SeatService, db *pgxpool.Pool, publicURL string) *StudentService {
	return &StudentService{repo: repo, seat: seat, db: db, publicURL: strings.TrimRight(publicURL, "/")}
}

func (s *StudentService) List(ctx context.Context, filter models.StudentFilter) (*models.StudentListResponse, error) {
	filter = normalizeStudentFilter(filter)
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + filter.Limit - 1) / filter.Limit
	}
	if totalPages > 0 && filter.Page > totalPages {
		filter.Page = totalPages
	}
	students, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	options, err := s.repo.FilterOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &models.StudentListResponse{
		Items: students,
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

func (s *StudentService) SeatMap(ctx context.Context, filter models.StudentFilter) (map[string]interface{}, error) {
	filter.Page = 1
	filter.Limit = 0
	students, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"items":       students,
		"total":       len(students),
		"seat_total":  len(students) * 2,
		"description": "Setiap siswa memiliki 2 nomor bangku: siswa dan pendamping.",
	}, nil
}

func (s *StudentService) Get(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *StudentService) GetByInvitationCode(ctx context.Context, code string) (*models.Student, error) {
	return s.repo.FindByInvitationCode(ctx, code)
}

func (s *StudentService) Create(ctx context.Context, input models.StudentInput) (*models.Student, error) {
	clean, err := s.cleanInput(input)
	if err != nil {
		return nil, err
	}
	laneCode, seatNumber, err := s.seat.GenerateNextSeatNumber(ctx, clean.ClassName, clean.Major)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 8; i++ {
		code, err := s.GenerateUniqueInvitationCode(ctx)
		if err != nil {
			return nil, err
		}
		student := &models.Student{
			Name: clean.Name, ClassName: clean.ClassName, Major: clean.Major,
			LaneCode: laneCode, SeatNumber: seatNumber, WhatsappNumber: clean.WhatsappNumber, Email: clean.Email,
			InvitationCode: code, QRPayload: "GRAD-ATTENDANCE:" + code, AttendanceStatus: models.AttendanceBelumHadir,
		}
		err = s.repo.Create(ctx, student)
		if err == nil {
			student.PopulateSeatNumbers()
			return student, nil
		}
		if !repositories.IsUniqueViolation(err) {
			return nil, err
		}
	}
	return nil, errors.New("gagal membuat invitation_code unik")
}

func (s *StudentService) Update(ctx context.Context, id uuid.UUID, input models.StudentInput) (*models.Student, error) {
	student, err := s.repo.FindByID(ctx, id)
	if err != nil || student == nil {
		return student, err
	}
	clean, err := s.cleanInput(input)
	if err != nil {
		return nil, err
	}
	if student.ClassName != clean.ClassName || student.Major != clean.Major {
		laneCode, seatNumber, err := s.seat.GenerateNextSeatNumber(ctx, clean.ClassName, clean.Major)
		if err != nil {
			return nil, err
		}
		student.LaneCode = laneCode
		student.SeatNumber = seatNumber
	}
	student.Name = clean.Name
	student.ClassName = clean.ClassName
	student.Major = clean.Major
	student.WhatsappNumber = clean.WhatsappNumber
	student.Email = clean.Email
	if err := s.repo.Update(ctx, student); err != nil {
		return nil, err
	}
	return student, nil
}

func (s *StudentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *StudentService) ResetAll(ctx context.Context) (int64, error) {
	schoolID, ok := authcontext.SchoolID(ctx)
	if !ok {
		return 0, errors.New("school_id tidak tersedia")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var deletedStudents, deletedLanes, deletedTemplates int64
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE school_id=$1`, *schoolID).Scan(&deletedStudents); err != nil {
		return 0, err
	}
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM seating_lanes WHERE school_id=$1`, *schoolID).Scan(&deletedLanes); err != nil {
		return 0, err
	}
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM event_templates WHERE school_id=$1`, *schoolID).Scan(&deletedTemplates); err != nil {
		return 0, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM email_logs WHERE school_id=$1`, *schoolID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM students WHERE school_id=$1`, *schoolID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM seating_lanes WHERE school_id=$1`, *schoolID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_templates WHERE school_id=$1`, *schoolID); err != nil {
		return 0, err
	}

	defaults := DefaultEventSettings()
	if _, err := tx.Exec(ctx, `
		INSERT INTO event_templates (
			school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
			event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
			whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
			theme_primary, theme_secondary, theme_accent,
			theme_background, theme_surface, theme_text,
			event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout
		)
		VALUES (
			$1, $2, TRUE, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23,
			$24, $25, $26,
			$27, $28, $29,
			$30, $31, $32, $33, $34, $35, $36, $37, $38, $39
		)
	`, *schoolID, defaults.TemplateName, defaults.EventTitle, defaults.EventTitleSecond, defaults.SchoolName, defaults.GraduationYear, defaults.RecipientGreeting, defaults.OpeningText,
		defaults.EventDate, defaults.EventTime, defaults.VenueName, defaults.VenueAddress, defaults.MapsURL, defaults.DressCodeStudent, defaults.DressCodeParent, defaults.AdditionalNote,
		defaults.WhatsappTemplate, defaults.EmailSubject, defaults.EmailTemplate, defaults.AudioURL, defaults.AudioKey, defaults.AudioTitle, defaults.AudioAutoplay,
		defaults.ThemePrimary, defaults.ThemeSecondary, defaults.ThemeAccent,
		defaults.ThemeBackground, defaults.ThemeSurface, defaults.ThemeText,
		defaults.EventDatetime, defaults.LayoutVariant, defaults.ShowCountdown, defaults.ShowMap, defaults.ShowQR, defaults.ShowNote, defaults.LayoutSections, defaults.SeatMapColumns, defaults.SeatMapColorMode, defaults.SeatMapLayout); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return deletedStudents + deletedLanes + deletedTemplates, nil
}

func (s *StudentService) UpdateAttendanceStatus(ctx context.Context, id uuid.UUID, status string) (*models.Student, error) {
	status = strings.TrimSpace(status)
	if status != models.AttendanceHadir && status != models.AttendanceBelumHadir {
		return nil, errors.New("status absensi tidak valid")
	}
	student, err := s.repo.FindByID(ctx, id)
	if err != nil || student == nil {
		return student, err
	}
	updated, err := s.repo.UpdateAttendanceStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *StudentService) GenerateUniqueInvitationCode(ctx context.Context) (string, error) {
	for i := 0; i < 20; i++ {
		random, err := utils.RandomString(8)
		if err != nil {
			return "", err
		}
		code := "GRAD-2026-" + random
		exists, err := s.repo.InvitationCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", errors.New("tidak dapat membuat invitation_code unik")
}

func (s *StudentService) Import(ctx context.Context, rows []utils.ImportRow) (map[string]interface{}, error) {
	var created []models.Student
	var rowErrors []string
	for _, row := range rows {
		email := strings.TrimSpace(row.Email)
		var emailPtr *string
		if email != "" {
			emailPtr = &email
		}
		student, err := s.Create(ctx, models.StudentInput{
			Name: row.Name, ClassName: row.ClassName, Major: row.Major, WhatsappNumber: row.WhatsappNumber, Email: emailPtr,
		})
		if err != nil {
			rowErrors = append(rowErrors, fmt.Sprintf("Baris %d: %s", row.RowNumber, err.Error()))
			continue
		}
		created = append(created, *student)
	}
	if len(rowErrors) > 0 {
		return map[string]interface{}{"created": created, "errors": rowErrors}, errors.New("sebagian data import tidak valid")
	}
	return map[string]interface{}{"created": created, "errors": []string{}}, nil
}

func (s *StudentService) ExportAttendanceXLSX(ctx context.Context) ([]byte, error) {
	students, err := s.repo.List(ctx, models.StudentFilter{})
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Rekap Absensi"
	f.SetSheetName("Sheet1", sheet)

	total := len(students)
	totalHadir := 0
	for _, student := range students {
		if student.AttendanceStatus == models.AttendanceHadir {
			totalHadir++
		}
	}
	totalBelumHadir := total - totalHadir
	attendancePercentage := 0.0
	if total > 0 {
		attendancePercentage = float64(totalHadir) / float64(total) * 100
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	subtitleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Color: "64748B"},
	})
	if err != nil {
		return nil, err
	}
	summaryLabelStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "64748B"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8FAFC"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "E2E8F0", Style: 1}, {Type: "right", Color: "E2E8F0", Style: 1}, {Type: "top", Color: "E2E8F0", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	summaryValueStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "0F172A"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8FAFC"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "E2E8F0", Style: 1}, {Type: "right", Color: "E2E8F0", Style: 1}, {Type: "bottom", Color: "E2E8F0", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"334155"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "CBD5E1", Style: 1}, {Type: "right", Color: "CBD5E1", Style: 1}, {Type: "top", Color: "CBD5E1", Style: 1}, {Type: "bottom", Color: "CBD5E1", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, err
	}
	rowStyle, err := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "E2E8F0", Style: 1}, {Type: "right", Color: "E2E8F0", Style: 1}, {Type: "top", Color: "E2E8F0", Style: 1}, {Type: "bottom", Color: "E2E8F0", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, err
	}
	altRowStyle, err := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F8FAFC"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "E2E8F0", Style: 1}, {Type: "right", Color: "E2E8F0", Style: 1}, {Type: "top", Color: "E2E8F0", Style: 1}, {Type: "bottom", Color: "E2E8F0", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, err
	}
	centerStyle, err := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "E2E8F0", Style: 1}, {Type: "right", Color: "E2E8F0", Style: 1}, {Type: "top", Color: "E2E8F0", Style: 1}, {Type: "bottom", Color: "E2E8F0", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	statusHadirStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "047857"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D1FAE5"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "A7F3D0", Style: 1}, {Type: "right", Color: "A7F3D0", Style: 1}, {Type: "top", Color: "A7F3D0", Style: 1}, {Type: "bottom", Color: "A7F3D0", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	statusBelumStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "B45309"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"FEF3C7"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "FDE68A", Style: 1}, {Type: "right", Color: "FDE68A", Style: 1}, {Type: "top", Color: "FDE68A", Style: 1}, {Type: "bottom", Color: "FDE68A", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	linkStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "2563EB", Underline: "single"},
		Border:    []excelize.Border{{Type: "left", Color: "E2E8F0", Style: 1}, {Type: "right", Color: "E2E8F0", Style: 1}, {Type: "top", Color: "E2E8F0", Style: 1}, {Type: "bottom", Color: "E2E8F0", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, err
	}

	_ = f.MergeCell(sheet, "A1", "K1")
	_ = f.MergeCell(sheet, "A2", "K2")
	f.SetCellValue(sheet, "A1", "Rekap Absensi Graduation")
	f.SetCellValue(sheet, "A2", "Dibuat otomatis pada "+time.Now().Format("02 Jan 2006 15:04"))
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetCellStyle(sheet, "A2", "A2", subtitleStyle)

	summary := []struct {
		label string
		value interface{}
		cell  string
	}{
		{"Total Siswa", total, "A4"},
		{"Total Hadir", totalHadir, "C4"},
		{"Belum Hadir", totalBelumHadir, "E4"},
		{"Kehadiran", fmt.Sprintf("%.1f%%", attendancePercentage), "G4"},
	}
	for _, item := range summary {
		col, row, _ := excelize.CellNameToCoordinates(item.cell)
		labelCell, _ := excelize.CoordinatesToCellName(col, row)
		valueCell, _ := excelize.CoordinatesToCellName(col, row+1)
		endLabelCell, _ := excelize.CoordinatesToCellName(col+1, row)
		endValueCell, _ := excelize.CoordinatesToCellName(col+1, row+1)
		_ = f.MergeCell(sheet, labelCell, endLabelCell)
		_ = f.MergeCell(sheet, valueCell, endValueCell)
		f.SetCellValue(sheet, labelCell, item.label)
		f.SetCellValue(sheet, valueCell, item.value)
		f.SetCellStyle(sheet, labelCell, endLabelCell, summaryLabelStyle)
		f.SetCellStyle(sheet, valueCell, endValueCell, summaryValueStyle)
	}

	headers := []string{"No", "Nama", "Kelas", "Jurusan", "Nomor Siswa", "Nomor Pendamping", "Nomor WhatsApp", "Email", "Status Absensi", "Waktu Hadir", "Link Undangan"}
	headerRow := 8
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for index, student := range students {
		student.PopulateSeatNumbers()
		row := headerRow + index + 1
		email := ""
		if student.Email != nil {
			email = *student.Email
		}
		attendanceTime := ""
		if student.AttendanceTime != nil {
			attendanceTime = student.AttendanceTime.Format("2006-01-02 15:04:05")
		}
		status := "Belum Hadir"
		if student.AttendanceStatus == models.AttendanceHadir {
			status = "Hadir"
		}
		link := s.InvitationLink(student.InvitationCode)
		values := []interface{}{
			index + 1,
			student.Name,
			student.ClassName,
			student.Major,
			student.StudentSeatNumber,
			student.CompanionSeatNumber,
			student.WhatsappNumber,
			email,
			status,
			attendanceTime,
			link,
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, value)
			style := rowStyle
			if index%2 == 1 {
				style = altRowStyle
			}
			if col == 0 || col == 4 || col == 5 || col == 8 || col == 9 {
				style = centerStyle
			}
			if col == 8 {
				style = statusBelumStyle
				if student.AttendanceStatus == models.AttendanceHadir {
					style = statusHadirStyle
				}
			}
			if col == 10 {
				style = linkStyle
				_ = f.SetCellHyperLink(sheet, cell, link, "External")
			}
			f.SetCellStyle(sheet, cell, cell, style)
		}
		_ = f.SetRowHeight(sheet, row, 24)
	}

	_ = f.SetColWidth(sheet, "A", "A", 8)
	_ = f.SetColWidth(sheet, "B", "B", 26)
	_ = f.SetColWidth(sheet, "C", "C", 18)
	_ = f.SetColWidth(sheet, "D", "D", 18)
	_ = f.SetColWidth(sheet, "E", "E", 12)
	_ = f.SetColWidth(sheet, "F", "F", 16)
	_ = f.SetColWidth(sheet, "G", "G", 20)
	_ = f.SetColWidth(sheet, "H", "H", 28)
	_ = f.SetColWidth(sheet, "I", "I", 18)
	_ = f.SetColWidth(sheet, "J", "J", 22)
	_ = f.SetColWidth(sheet, "K", "K", 48)
	_ = f.SetRowHeight(sheet, 1, 30)
	_ = f.SetRowHeight(sheet, headerRow, 28)
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      headerRow,
		TopLeftCell: "A9",
		ActivePane:  "bottomLeft",
	})

	if len(students) > 0 {
		_ = f.AutoFilter(sheet, fmt.Sprintf("A%d:K%d", headerRow, headerRow+len(students)), nil)
	}

	classGroups := make(map[string][]models.Student)
	for _, student := range students {
		className := strings.TrimSpace(student.ClassName)
		if className == "" {
			className = "Tanpa Kelas"
		}
		classGroups[className] = append(classGroups[className], student)
	}
	classNames := make([]string, 0, len(classGroups))
	for className := range classGroups {
		classNames = append(classNames, className)
	}
	sort.Strings(classNames)

	usedSheetNames := map[string]bool{sheet: true}
	writeClassSheet := func(sheetName string, className string, data []models.Student) {
		classHadir := 0
		for _, student := range data {
			if student.AttendanceStatus == models.AttendanceHadir {
				classHadir++
			}
		}
		classBelumHadir := len(data) - classHadir
		classPercentage := 0.0
		if len(data) > 0 {
			classPercentage = float64(classHadir) / float64(len(data)) * 100
		}

		_ = f.MergeCell(sheetName, "A1", "K1")
		_ = f.MergeCell(sheetName, "A2", "K2")
		f.SetCellValue(sheetName, "A1", "Data Siswa - "+className)
		f.SetCellValue(sheetName, "A2", "Dibuat otomatis pada "+time.Now().Format("02 Jan 2006 15:04"))
		f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
		f.SetCellStyle(sheetName, "A2", "A2", subtitleStyle)

		classSummary := []struct {
			label string
			value interface{}
			cell  string
		}{
			{"Total Siswa", len(data), "A4"},
			{"Total Hadir", classHadir, "C4"},
			{"Belum Hadir", classBelumHadir, "E4"},
			{"Kehadiran", fmt.Sprintf("%.1f%%", classPercentage), "G4"},
		}
		for _, item := range classSummary {
			col, row, _ := excelize.CellNameToCoordinates(item.cell)
			labelCell, _ := excelize.CoordinatesToCellName(col, row)
			valueCell, _ := excelize.CoordinatesToCellName(col, row+1)
			endLabelCell, _ := excelize.CoordinatesToCellName(col+1, row)
			endValueCell, _ := excelize.CoordinatesToCellName(col+1, row+1)
			_ = f.MergeCell(sheetName, labelCell, endLabelCell)
			_ = f.MergeCell(sheetName, valueCell, endValueCell)
			f.SetCellValue(sheetName, labelCell, item.label)
			f.SetCellValue(sheetName, valueCell, item.value)
			f.SetCellStyle(sheetName, labelCell, endLabelCell, summaryLabelStyle)
			f.SetCellStyle(sheetName, valueCell, endValueCell, summaryValueStyle)
		}

		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
			f.SetCellValue(sheetName, cell, header)
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}

		for index, student := range data {
			student.PopulateSeatNumbers()
			row := headerRow + index + 1
			email := ""
			if student.Email != nil {
				email = *student.Email
			}
			attendanceTime := ""
			if student.AttendanceTime != nil {
				attendanceTime = student.AttendanceTime.Format("2006-01-02 15:04:05")
			}
			status := "Belum Hadir"
			if student.AttendanceStatus == models.AttendanceHadir {
				status = "Hadir"
			}
			link := s.InvitationLink(student.InvitationCode)
			values := []interface{}{
				index + 1,
				student.Name,
				student.ClassName,
				student.Major,
				student.StudentSeatNumber,
				student.CompanionSeatNumber,
				student.WhatsappNumber,
				email,
				status,
				attendanceTime,
				link,
			}
			for col, value := range values {
				cell, _ := excelize.CoordinatesToCellName(col+1, row)
				f.SetCellValue(sheetName, cell, value)
				style := rowStyle
				if index%2 == 1 {
					style = altRowStyle
				}
				if col == 0 || col == 4 || col == 5 || col == 8 || col == 9 {
					style = centerStyle
				}
				if col == 8 {
					style = statusBelumStyle
					if student.AttendanceStatus == models.AttendanceHadir {
						style = statusHadirStyle
					}
				}
				if col == 10 {
					style = linkStyle
					_ = f.SetCellHyperLink(sheetName, cell, link, "External")
				}
				f.SetCellStyle(sheetName, cell, cell, style)
			}
			_ = f.SetRowHeight(sheetName, row, 24)
		}

		_ = f.SetColWidth(sheetName, "A", "A", 8)
		_ = f.SetColWidth(sheetName, "B", "B", 26)
		_ = f.SetColWidth(sheetName, "C", "C", 18)
		_ = f.SetColWidth(sheetName, "D", "D", 18)
		_ = f.SetColWidth(sheetName, "E", "E", 12)
		_ = f.SetColWidth(sheetName, "F", "F", 16)
		_ = f.SetColWidth(sheetName, "G", "G", 20)
		_ = f.SetColWidth(sheetName, "H", "H", 28)
		_ = f.SetColWidth(sheetName, "I", "I", 18)
		_ = f.SetColWidth(sheetName, "J", "J", 22)
		_ = f.SetColWidth(sheetName, "K", "K", 48)
		_ = f.SetRowHeight(sheetName, 1, 30)
		_ = f.SetRowHeight(sheetName, headerRow, 28)
		_ = f.SetPanes(sheetName, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      headerRow,
			TopLeftCell: "A9",
			ActivePane:  "bottomLeft",
		})
		if len(data) > 0 {
			_ = f.AutoFilter(sheetName, fmt.Sprintf("A%d:K%d", headerRow, headerRow+len(data)), nil)
		}
	}

	for _, className := range classNames {
		classSheet := uniqueExcelSheetName(className, usedSheetNames)
		if _, err := f.NewSheet(classSheet); err != nil {
			return nil, err
		}
		writeClassSheet(classSheet, className, classGroups[className])
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *StudentService) ImportTemplateXLSX() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Template Import Siswa"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"name", "class_name", "major", "whatsapp_number", "email"}
	descriptions := []string{
		"Nama lengkap siswa",
		"Kelas, contoh: 12 DKV 1",
		"Jurusan, contoh: DKV",
		"Nomor WhatsApp aktif",
		"Email opsional",
	}
	sample := []string{"Alya Putri", "12 DKV 1", "DKV", "081234567890", "alya@example.com"}

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
	textStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 49,
	})
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
	_ = f.SetColWidth(sheet, "B", "B", 18)
	_ = f.SetColWidth(sheet, "C", "C", 16)
	_ = f.SetColWidth(sheet, "D", "D", 22)
	_ = f.SetColWidth(sheet, "E", "E", 28)
	_ = f.SetColStyle(sheet, "D", textStyle)
	_ = f.SetRowHeight(sheet, 1, 24)
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *StudentService) InvitationLink(code string) string {
	return s.publicURL + "/" + code
}

func normalizeStudentFilter(filter models.StudentFilter) models.StudentFilter {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return filter
}

func uniqueExcelSheetName(name string, used map[string]bool) string {
	base := strings.TrimSpace(name)
	if base == "" {
		base = "Sheet"
	}
	replacer := strings.NewReplacer(
		"[", "-",
		"]", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"/", "-",
		"\\", "-",
	)
	base = strings.TrimSpace(replacer.Replace(base))
	if base == "" {
		base = "Sheet"
	}
	base = trimRunes(base, 31)

	candidate := base
	for i := 2; used[candidate]; i++ {
		suffix := fmt.Sprintf(" %d", i)
		candidate = trimRunes(base, 31-len([]rune(suffix))) + suffix
	}
	used[candidate] = true
	return candidate
}

func trimRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func (s *StudentService) cleanInput(input models.StudentInput) (models.StudentInput, error) {
	input.Name = utils.CleanText(input.Name)
	input.ClassName = utils.CleanText(input.ClassName)
	input.Major = utils.CleanText(input.Major)
	if err := utils.ValidateRequired(input.Name, "name"); err != nil {
		return input, err
	}
	if err := utils.ValidateRequired(input.ClassName, "class_name"); err != nil {
		return input, err
	}
	if err := utils.ValidateRequired(input.Major, "major"); err != nil {
		return input, err
	}
	wa, err := utils.NormalizeWhatsapp(input.WhatsappNumber)
	if err != nil {
		return input, err
	}
	input.WhatsappNumber = wa
	email, err := utils.NormalizeEmail(input.Email)
	if err != nil {
		return input, err
	}
	input.Email = email
	return input, nil
}
