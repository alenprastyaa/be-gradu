package services

import (
	"context"
	"strings"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"

	"github.com/google/uuid"
)

type EventSettingsService struct {
	repo *repositories.EventSettingsRepository
}

func NewEventSettingsService(repo *repositories.EventSettingsRepository) *EventSettingsService {
	return &EventSettingsService{repo: repo}
}

func (s *EventSettingsService) List(ctx context.Context) ([]models.EventSettings, error) {
	if err := s.ensureDefault(ctx); err != nil {
		return nil, err
	}
	return s.repo.List(ctx)
}

func (s *EventSettingsService) Get(ctx context.Context) (*models.EventSettings, error) {
	if err := s.ensureDefault(ctx); err != nil {
		return nil, err
	}
	settings, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		defaults := DefaultEventSettings()
		defaults.IsActive = true
		return s.repo.Create(ctx, defaults)
	}
	return settings, nil
}

func (s *EventSettingsService) Create(ctx context.Context, input models.EventSettings) (*models.EventSettings, error) {
	input = normalizeEventSettings(input)
	if input.IsActive {
		input.IsActive = false
		created, err := s.repo.Create(ctx, input)
		if err != nil {
			return nil, err
		}
		return s.repo.Activate(ctx, created.ID)
	}
	return s.repo.Create(ctx, input)
}

func (s *EventSettingsService) Update(ctx context.Context, input models.EventSettings) (*models.EventSettings, error) {
	if input.ID == uuid.Nil {
		active, err := s.Get(ctx)
		if err != nil {
			return nil, err
		}
		input.ID = active.ID
		input.IsActive = active.IsActive
		if input.SeatMapColumns <= 0 {
			input.SeatMapColumns = active.SeatMapColumns
		}
		if input.SeatMapColorMode == "" {
			input.SeatMapColorMode = active.SeatMapColorMode
		}
	} else if input.SeatMapColumns <= 0 || input.SeatMapColorMode == "" {
		existing, err := s.repo.GetByID(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			input.SeatMapColumns = existing.SeatMapColumns
			if input.SeatMapColorMode == "" {
				input.SeatMapColorMode = existing.SeatMapColorMode
			}
		}
	}
	input = normalizeEventSettings(input)
	return s.repo.Update(ctx, input)
}

func (s *EventSettingsService) Activate(ctx context.Context, id uuid.UUID) (*models.EventSettings, error) {
	return s.repo.Activate(ctx, id)
}

func (s *EventSettingsService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *EventSettingsService) UpdateSeatMap(ctx context.Context, columns int, colorMode, layout string) (*models.EventSettings, error) {
	if err := s.ensureDefault(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(colorMode) == "" {
		active, err := s.Get(ctx)
		if err != nil {
			return nil, err
		}
		if active != nil {
			colorMode = active.SeatMapColorMode
		}
	}
	return s.repo.UpdateActiveSeatMap(ctx, normalizeSeatMapColumns(columns), normalizeSeatMapColorMode(colorMode), strings.TrimSpace(layout))
}

func (s *EventSettingsService) UpdateSchoolLogo(ctx context.Context, url string, key string) (*models.EventSettings, error) {
	if err := s.ensureDefault(ctx); err != nil {
		return nil, err
	}
	return s.repo.UpdateActiveSchoolLogo(ctx, strings.TrimSpace(url), strings.TrimSpace(key))
}

func (s *EventSettingsService) ensureDefault(ctx context.Context) error {
	templates, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	if len(templates) > 0 {
		return nil
	}
	defaults := DefaultEventSettings()
	defaults.IsActive = true
	_, err = s.repo.Create(ctx, defaults)
	return err
}

func normalizeEventSettings(input models.EventSettings) models.EventSettings {
	input.TemplateName = strings.TrimSpace(input.TemplateName)
	input.EventTitle = strings.TrimSpace(input.EventTitle)
	input.SchoolName = strings.TrimSpace(input.SchoolName)
	input.GraduationYear = strings.TrimSpace(input.GraduationYear)
	input.RecipientGreeting = strings.TrimSpace(input.RecipientGreeting)
	input.OpeningText = strings.TrimSpace(input.OpeningText)
	input.EventDate = strings.TrimSpace(input.EventDate)
	input.EventTime = strings.TrimSpace(input.EventTime)
	input.VenueName = strings.TrimSpace(input.VenueName)
	input.VenueAddress = strings.TrimSpace(input.VenueAddress)
	input.MapsURL = strings.TrimSpace(input.MapsURL)
	input.DressCode = strings.TrimSpace(input.DressCode)
	input.AdditionalNote = strings.TrimSpace(input.AdditionalNote)
	input.ScheduleTitle = strings.TrimSpace(input.ScheduleTitle)
	input.ScheduleHeaders = normalizeJSONString(input.ScheduleHeaders)
	input.ScheduleRows = normalizeJSONString(input.ScheduleRows)
	input.SchoolLogoURL = strings.TrimSpace(input.SchoolLogoURL)
	input.SchoolLogoKey = strings.TrimSpace(input.SchoolLogoKey)
	input.WhatsappTemplate = normalizeSeatPlaceholder(stripLanePlaceholderLine(input.WhatsappTemplate))
	input.EmailSubject = strings.TrimSpace(input.EmailSubject)
	input.EmailTemplate = normalizeSeatPlaceholder(stripLanePlaceholderLine(input.EmailTemplate))
	input.AudioURL = strings.TrimSpace(input.AudioURL)
	input.AudioKey = strings.TrimSpace(input.AudioKey)
	input.AudioTitle = strings.TrimSpace(input.AudioTitle)
	input.ThemePrimary = normalizeColor(input.ThemePrimary)
	input.ThemeSecondary = normalizeColor(input.ThemeSecondary)
	input.ThemeAccent = normalizeColor(input.ThemeAccent)
	input.ThemeBackground = normalizeColor(input.ThemeBackground)
	input.ThemeSurface = normalizeColor(input.ThemeSurface)
	input.ThemeText = normalizeColor(input.ThemeText)
	input.EventDatetime = strings.TrimSpace(input.EventDatetime)
	input.LayoutVariant = strings.TrimSpace(input.LayoutVariant)
	input.LayoutSections = strings.TrimSpace(input.LayoutSections)
	input.SeatMapColumns = normalizeSeatMapColumns(input.SeatMapColumns)
	input.SeatMapColorMode = normalizeSeatMapColorMode(input.SeatMapColorMode)
	input.SeatMapLayout = strings.TrimSpace(input.SeatMapLayout)

	defaults := DefaultEventSettings()
	if input.LayoutVariant == "" {
		input.LayoutVariant = defaults.LayoutVariant
	}
	if input.TemplateName == "" {
		input.TemplateName = defaults.TemplateName
	}
	if input.EventTitle == "" {
		input.EventTitle = defaults.EventTitle
	}
	if input.WhatsappTemplate == "" {
		input.WhatsappTemplate = defaults.WhatsappTemplate
	}
	if input.EmailSubject == "" {
		input.EmailSubject = defaults.EmailSubject
	}
	if input.EmailTemplate == "" {
		input.EmailTemplate = defaults.EmailTemplate
	}
	if input.ScheduleTitle == "" {
		input.ScheduleTitle = defaults.ScheduleTitle
	}
	if input.ScheduleHeaders == "" {
		input.ScheduleHeaders = defaults.ScheduleHeaders
	}
	if input.ScheduleRows == "" {
		input.ScheduleRows = defaults.ScheduleRows
	}
	if input.ThemePrimary == "" {
		input.ThemePrimary = defaults.ThemePrimary
	}
	if input.ThemeSecondary == "" {
		input.ThemeSecondary = defaults.ThemeSecondary
	}
	if input.ThemeAccent == "" {
		input.ThemeAccent = defaults.ThemeAccent
	}
	if input.ThemeBackground == "" {
		input.ThemeBackground = defaults.ThemeBackground
	}
	if input.ThemeSurface == "" {
		input.ThemeSurface = defaults.ThemeSurface
	}
	if input.ThemeText == "" {
		input.ThemeText = defaults.ThemeText
	}
	return input
}

func normalizeSeatMapColumns(value int) int {
	if value <= 0 {
		return 20
	}
	if value < 4 {
		return 4
	}
	if value > 40 {
		return 40
	}
	return value
}

func normalizeSeatMapColorMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "class" {
		return "class"
	}
	return "attendance"
}

func normalizeColor(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "#") {
		value = "#" + value
	}
	if len(value) != 7 {
		return ""
	}
	return value
}

func stripLanePlaceholderLine(value string) string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, "{{lane_code}}") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func normalizeSeatPlaceholder(value string) string {
	return strings.ReplaceAll(
		value,
		"Bangku Siswa & Pendamping: {{seat_number}}",
		"Nomor Siswa: {{student_seat_number}}\nNomor Pendamping: {{companion_seat_number}}",
	)
}

func DefaultEventSettings() models.EventSettings {
	return models.EventSettings{
		TemplateName:      "Formal Navy Gold",
		EventTitle:        "Graduation Ceremony 2026",
		SchoolName:        "Nama Sekolah",
		GraduationYear:    "2026",
		RecipientGreeting: "Yth. Siswa/i dan Orang Tua/Wali",
		OpeningText:       "Dengan hormat, kami mengundang Anda untuk menghadiri acara wisuda sebagai bentuk apresiasi atas pencapaian dan perjalanan pendidikan siswa/i.",
		EventDate:         "Sabtu, 20 Juni 2026",
		EventTime:         "08.00 WIB - selesai",
		VenueName:         "Aula Utama Sekolah",
		VenueAddress:      "Jl. Pendidikan No. 1",
		MapsURL:           "",
		DressCode:         "Formal rapi",
		AdditionalNote:    "Mohon hadir 30 menit sebelum acara dimulai dan tunjukkan QR Code kepada petugas registrasi.",
		ScheduleTitle:     "Susunan Acara",
		ScheduleHeaders:   `["Waktu","Kegiatan"]`,
		ScheduleRows:      `[["07.00","Registrasi"],["08.00","Pembukaan"],["09.00","Sambutan"]]`,
		SchoolLogoURL:     "",
		SchoolLogoKey:     "",
		AudioURL:          "",
		AudioKey:          "",
		AudioTitle:        "",
		AudioAutoplay:     false,
		ThemePrimary:      "#0f172a",
		ThemeSecondary:    "#1e293b",
		ThemeAccent:       "#facc15",
		ThemeBackground:   "#020617",
		ThemeSurface:      "#ffffff",
		ThemeText:         "#0f172a",
		EventDatetime:     "",
		LayoutVariant:     "classic",
		ShowCountdown:     true,
		ShowMap:           true,
		ShowQR:            true,
		ShowNote:          true,
		SeatMapColumns:    20,
		SeatMapColorMode:  "attendance",
		SeatMapLayout:     "",
		WhatsappTemplate: `Assalamu'alaikum Wr. Wb.

{{recipient_greeting}},

Dengan hormat, kami mengundang {{student_name}} untuk menghadiri:

*{{event_title}}*
{{school_name}}

Hari/Tanggal: {{event_date}}
Waktu: {{event_time}}
Tempat: {{venue_name}}
Alamat: {{venue_address}}
Dress Code: {{dress_code}}

Data Undangan:
Nama: {{student_name}}
Kelas: {{class_name}}
Jurusan: {{major}}
Nomor Siswa: {{student_seat_number}}
Nomor Pendamping: {{companion_seat_number}}

Undangan digital dan QR Code:
{{invitation_link}}

{{additional_note}}

Wassalamu'alaikum Wr. Wb.`,
		EmailSubject: "Undangan Resmi {{event_title}} untuk {{student_name}}",
		EmailTemplate: `{{recipient_greeting}},

Dengan hormat,

Kami mengundang {{student_name}} untuk menghadiri {{event_title}} yang diselenggarakan oleh {{school_name}}.

Undangan digital, QR Code registrasi, detail acara, dan nomor bangku tersedia melalui tautan berikut:
{{invitation_link}}

Ringkasan undangan:
Nama: {{student_name}} - {{class_name}} {{major}}
Nomor Siswa: {{student_seat_number}}
Nomor Pendamping: {{companion_seat_number}}

{{additional_note}}

Hormat kami,
Panitia {{event_title}}`,
	}
}

func normalizeJSONString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return value
}

func RenderInvitationTemplate(template string, settings *models.EventSettings, student *models.Student, invitationLink string) string {
	if settings == nil {
		defaults := DefaultEventSettings()
		settings = &defaults
	}
	student.PopulateSeatNumbers()
	replacer := strings.NewReplacer(
		"{{recipient_greeting}}", settings.RecipientGreeting,
		"{{event_title}}", settings.EventTitle,
		"{{school_name}}", settings.SchoolName,
		"{{graduation_year}}", settings.GraduationYear,
		"{{opening_text}}", settings.OpeningText,
		"{{event_date}}", settings.EventDate,
		"{{event_time}}", settings.EventTime,
		"{{venue_name}}", settings.VenueName,
		"{{venue_address}}", settings.VenueAddress,
		"{{maps_url}}", settings.MapsURL,
		"{{dress_code}}", settings.DressCode,
		"{{additional_note}}", settings.AdditionalNote,
		"{{student_name}}", student.Name,
		"{{class_name}}", student.ClassName,
		"{{major}}", student.Major,
		"{{lane_code}}", student.LaneCode,
		"{{seat_number}}", student.SeatNumber,
		"{{student_seat_number}}", student.StudentSeatNumber,
		"{{companion_seat_number}}", student.CompanionSeatNumber,
		"{{invitation_code}}", student.InvitationCode,
		"{{invitation_link}}", invitationLink,
	)
	return replacer.Replace(template)
}
