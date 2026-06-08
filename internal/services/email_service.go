package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"graduation-invitation/internal/models"
	"graduation-invitation/internal/repositories"

	"github.com/google/uuid"
)

type EmailService struct {
	repo          *repositories.StudentRepository
	logs          *repositories.EmailLogRepository
	events        *EventSettingsService
	publicURL     string
	apiKey        string
	senderEmail   string
	senderName    string
	webhookSecret string
	client        *http.Client
}

type brevoAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoEmailPayload struct {
	Sender      brevoAddress   `json:"sender"`
	To          []brevoAddress `json:"to"`
	Subject     string         `json:"subject"`
	TextContent string         `json:"textContent"`
	HTMLContent string         `json:"htmlContent"`
	Tags        []string       `json:"tags,omitempty"`
}

type brevoEmailResponse struct {
	MessageID string `json:"messageId"`
}

type brevoEventsResponse struct {
	Events []json.RawMessage `json:"events"`
}

func NewEmailService(repo *repositories.StudentRepository, logs *repositories.EmailLogRepository, events *EventSettingsService, publicURL string, apiKey string, senderEmail string, senderName string, webhookSecret string) *EmailService {
	return &EmailService{
		repo:          repo,
		logs:          logs,
		events:        events,
		publicURL:     publicURL,
		apiKey:        strings.TrimSpace(apiKey),
		senderEmail:   strings.TrimSpace(senderEmail),
		senderName:    strings.TrimSpace(senderName),
		webhookSecret: strings.TrimSpace(webhookSecret),
		client:        &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *EmailService) SendForStudent(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	student, err := s.repo.FindByID(ctx, id)
	if err != nil || student == nil {
		return nil, err
	}
	return s.send(ctx, student)
}

func (s *EmailService) SendAll(ctx context.Context) ([]map[string]interface{}, error) {
	students, err := s.repo.List(ctx, models.StudentFilter{})
	if err != nil {
		return nil, err
	}
	results := []map[string]interface{}{}
	for i := range students {
		result, err := s.send(ctx, &students[i])
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *EmailService) SendForStudents(ctx context.Context, ids []uuid.UUID) ([]map[string]interface{}, error) {
	students, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0, len(students))
	for i := range students {
		result, err := s.send(ctx, &students[i])
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *EmailService) send(ctx context.Context, student *models.Student) (map[string]interface{}, error) {
	if student == nil {
		return nil, errors.New("siswa tidak ditemukan")
	}
	if student.Email == nil || strings.TrimSpace(*student.Email) == "" {
		return map[string]interface{}{"student_id": student.ID, "name": student.Name, "email": nil, "sent": false, "message": "Email belum tersedia"}, nil
	}
	if student.EmailSentAt != nil {
		return map[string]interface{}{
			"student_id":    student.ID,
			"name":          student.Name,
			"email":         strings.TrimSpace(*student.Email),
			"sent":          false,
			"skipped":       true,
			"already_sent":  true,
			"email_sent_at": student.EmailSentAt,
			"brevo_id":      student.EmailBrevoMessageID,
			"message":       "Email undangan sudah pernah dikirim. Pengiriman ulang dilewati.",
		}, nil
	}
	if s.apiKey == "" || s.senderEmail == "" {
		return nil, errors.New("konfigurasi Brevo belum lengkap")
	}

	settings, err := s.events.Get(ctx)
	if err != nil {
		return nil, err
	}

	recipientEmail := strings.TrimSpace(*student.Email)
	invitationLink := strings.TrimRight(s.publicURL, "/") + "/" + student.InvitationCode
	subject := RenderInvitationTemplate(settings.EmailSubject, settings, student, invitationLink)
	body := RenderInvitationTemplate(settings.EmailTemplate, settings, student, invitationLink)
	messageID, err := s.sendBrevo(ctx, recipientEmail, student.Name, subject, body, emailHTML(body, settings, student, invitationLink))
	if err != nil {
		return nil, err
	}
	updated, err := s.repo.MarkEmailSent(ctx, student.ID, messageID)
	if err != nil {
		return nil, err
	}
	if updated != nil {
		student = updated
	}
	s.recordLog(ctx, models.EmailLogInput{
		StudentID:  &student.ID,
		Email:      recipientEmail,
		MessageID:  messageID,
		Subject:    subject,
		Event:      "request",
		EventTime:  student.EmailSentAt,
		RawPayload: json.RawMessage(`{"source":"application_send"}`),
	})

	return map[string]interface{}{
		"student_id":    student.ID,
		"name":          student.Name,
		"email":         recipientEmail,
		"sent":          true,
		"skipped":       false,
		"already_sent":  false,
		"email_sent_at": student.EmailSentAt,
		"message":       "Email undangan berhasil dikirim",
		"brevo_id":      messageID,
		"invite_link":   invitationLink,
	}, nil
}

func (s *EmailService) sendBrevo(ctx context.Context, recipientEmail string, recipientName string, subject string, body string, htmlBody string) (string, error) {
	payload := brevoEmailPayload{
		Sender: brevoAddress{
			Email: s.senderEmail,
			Name:  s.senderName,
		},
		To: []brevoAddress{{
			Email: recipientEmail,
			Name:  recipientName,
		}},
		Subject:     subject,
		TextContent: body,
		HTMLContent: htmlBody,
		Tags:        []string{"graduation-invitation"},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.apiKey)

	res, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gagal menghubungi Brevo: %w", err)
	}
	defer res.Body.Close()

	responseBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Brevo menolak email (%s): %s", res.Status, strings.TrimSpace(string(responseBody)))
	}

	var response brevoEmailResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", nil
	}
	return response.MessageID, nil
}

func (s *EmailService) ResetSent(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	return s.repo.ResetEmailSent(ctx, id)
}

func (s *EmailService) WebhookSecretConfigured() bool {
	return s.webhookSecret != ""
}

func (s *EmailService) ValidateWebhookSecret(value string) bool {
	return s.webhookSecret == "" || strings.TrimSpace(value) == s.webhookSecret
}

func (s *EmailService) HandleBrevoEvent(ctx context.Context, raw json.RawMessage) (map[string]interface{}, error) {
	input := parseBrevoRawEvent(raw)
	if input.Event == "" || input.Email == "" {
		return nil, errors.New("payload webhook Brevo tidak valid")
	}
	settings, err := s.events.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !isGraduationEmailEvent(input, raw, settings) {
		return map[string]interface{}{
			"email":   input.Email,
			"event":   input.Event,
			"ignored": true,
			"message": "Event Brevo bukan email undangan graduation.",
		}, nil
	}

	var studentID *uuid.UUID
	student, err := s.findStudentForEmailEvent(ctx, input.Email, input.MessageID)
	if err != nil {
		return nil, err
	}
	if student != nil {
		studentID = &student.ID
		if isEmailSentEvidence(input.Event) && student.EmailSentAt == nil {
			if updated, err := s.repo.MarkEmailSent(ctx, student.ID, input.MessageID); err == nil {
				student = updated
			} else {
				return nil, err
			}
		}
	}

	input.StudentID = studentID
	if len(input.RawPayload) == 0 {
		input.RawPayload = raw
	}
	if err := s.recordLog(ctx, input); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"email":      input.Email,
		"event":      input.Event,
		"message_id": input.MessageID,
		"matched":    student != nil,
	}
	if student != nil {
		result["student_id"] = student.ID
		result["name"] = student.Name
	}
	return result, nil
}

func (s *EmailService) SyncBrevoHistory(ctx context.Context, days int) (map[string]interface{}, error) {
	if s.apiKey == "" {
		return nil, errors.New("konfigurasi Brevo API key belum tersedia")
	}
	if days <= 0 || days > 90 {
		days = 90
	}
	limit := 5000
	offset := 0
	total := 0
	matched := 0
	for {
		endpoint := "https://api.brevo.com/v3/smtp/statistics/events?days=" + strconv.Itoa(days) +
			"&limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("api-key", s.apiKey)
		res, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("gagal menghubungi Brevo: %w", err)
		}
		responseBody, _ := io.ReadAll(io.LimitReader(res.Body, 10*1024*1024))
		res.Body.Close()
		if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("Brevo menolak sinkronisasi (%s): %s", res.Status, strings.TrimSpace(string(responseBody)))
		}
		var parsed brevoEventsResponse
		if err := json.Unmarshal(responseBody, &parsed); err != nil {
			return nil, err
		}
		for _, event := range parsed.Events {
			result, err := s.HandleBrevoEvent(ctx, event)
			if err != nil {
				continue
			}
			if result["ignored"] == true {
				continue
			}
			total++
			if result["matched"] == true {
				matched++
			}
		}
		if len(parsed.Events) < limit {
			break
		}
		offset += limit
	}
	return map[string]interface{}{"synced": total, "matched": matched, "days": days}, nil
}

func (s *EmailService) findStudentForEmailEvent(ctx context.Context, email, messageID string) (*models.Student, error) {
	messageID = strings.TrimSpace(messageID)
	if messageID != "" {
		student, err := s.repo.FindByBrevoMessageID(ctx, messageID)
		if err != nil || student != nil {
			return student, err
		}
	}
	return s.repo.FindByEmail(ctx, email)
}

func (s *EmailService) recordLog(ctx context.Context, input models.EmailLogInput) error {
	if s.logs == nil {
		return nil
	}
	return s.logs.Create(ctx, input)
}

func parseBrevoRawEvent(raw json.RawMessage) models.EmailLogInput {
	var data map[string]interface{}
	_ = json.Unmarshal(raw, &data)
	eventTime := parseBrevoEventTime(data)
	return models.EmailLogInput{
		Email:      stringField(data, "email"),
		MessageID:  firstStringField(data, "message-id", "messageId", "message_id"),
		Subject:    stringField(data, "subject"),
		Event:      normalizeBrevoEvent(stringField(data, "event")),
		EventTime:  eventTime,
		Reason:     stringField(data, "reason"),
		Link:       stringField(data, "link"),
		RawPayload: raw,
	}
}

func isGraduationEmailEvent(input models.EmailLogInput, raw json.RawMessage, settings *models.EventSettings) bool {
	var data map[string]interface{}
	_ = json.Unmarshal(raw, &data)
	for _, tag := range stringSliceField(data, "tags") {
		if strings.EqualFold(strings.TrimSpace(tag), "graduation-invitation") {
			return true
		}
	}
	subject := strings.ToLower(strings.TrimSpace(input.Subject))
	if subject == "" {
		return false
	}
	if strings.Contains(subject, "undangan") {
		return true
	}
	if settings != nil {
		eventTitle := strings.ToLower(strings.TrimSpace(settings.EventTitle))
		if eventTitle != "" && strings.Contains(subject, eventTitle) {
			return true
		}
	}
	return false
}

func parseBrevoEventTime(data map[string]interface{}) *time.Time {
	for _, key := range []string{"ts_event", "ts", "ts_epoch"} {
		if value, ok := numericField(data, key); ok {
			if key == "ts_epoch" && value > 9999999999 {
				value = value / 1000
			}
			t := time.Unix(value, 0).UTC()
			return &t
		}
	}
	if value := firstStringField(data, "date", "createdAt"); value != "" {
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z"} {
			if parsed, err := time.Parse(layout, value); err == nil {
				return &parsed
			}
		}
	}
	return nil
}

func stringField(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func firstStringField(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := stringField(data, key); value != "" {
			return value
		}
	}
	return ""
}

func stringSliceField(data map[string]interface{}, key string) []string {
	value, ok := data[key]
	if !ok {
		return nil
	}
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func numericField(data map[string]interface{}, key string) (int64, bool) {
	value, ok := data[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func normalizeBrevoEvent(event string) string {
	event = strings.ToLower(strings.TrimSpace(event))
	if event == "request" {
		return "sent"
	}
	return event
}

func isEmailSentEvidence(event string) bool {
	switch normalizeBrevoEvent(event) {
	case "sent", "delivered", "opened", "unique_opened", "proxy_open", "unique_proxy_open", "click":
		return true
	default:
		return false
	}
}

func emailHTML(body string, settings *models.EventSettings, student *models.Student, invitationLink string) string {
	if settings == nil {
		defaults := DefaultEventSettings()
		settings = &defaults
	}
	if student == nil {
		student = &models.Student{}
	}
	student.PopulateSeatNumbers()

	primary := emailColor(settings.ThemePrimary, "#0f172a")
	secondary := emailColor(settings.ThemeSecondary, "#1e293b")
	accent := emailColor(settings.ThemeAccent, "#facc15")
	surface := emailColor(settings.ThemeSurface, "#ffffff")
	text := emailColor(settings.ThemeText, "#0f172a")
	muted := "#64748b"
	eventTitle := escapeText(settings.EventTitle)
	schoolName := escapeText(settings.SchoolName)
	logo := strings.TrimSpace(settings.SchoolLogoURL)
	logoHTML := ""
	if logo != "" {
		logoHTML = `<img src="` + escapeAttr(logo) + `" alt="Logo sekolah" width="56" height="56" style="display:block;width:56px;height:56px;border-radius:16px;object-fit:contain;background:#ffffff;padding:6px;margin:0 auto 18px auto;">`
	}

	return `<!doctype html>
<html>
<body style="margin:0;padding:0;background:#eef2f7;font-family:Inter,Segoe UI,Roboto,Arial,sans-serif;color:` + text + `;">
  <div style="display:none;max-height:0;overflow:hidden;color:transparent;opacity:0;">` + escapeText(settings.OpeningText) + `</div>
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#eef2f7;margin:0;padding:32px 14px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:680px;background:` + surface + `;border-radius:28px;overflow:hidden;box-shadow:0 24px 70px rgba(15,23,42,.16);">
          <tr>
            <td style="padding:38px 28px 34px 28px;text-align:center;background:linear-gradient(135deg,` + primary + ` 0%,` + secondary + ` 100%);color:#ffffff;">
              ` + logoHTML + `
              <div style="display:inline-block;padding:7px 12px;border:1px solid rgba(255,255,255,.28);border-radius:999px;color:` + accent + `;font-size:11px;font-weight:800;letter-spacing:2.4px;text-transform:uppercase;">Undangan Resmi</div>
              <h1 style="margin:18px 0 8px 0;font-size:30px;line-height:1.18;font-weight:850;color:#ffffff;">` + eventTitle + `</h1>
              <p style="margin:0;color:rgba(255,255,255,.78);font-size:15px;line-height:1.6;">` + schoolName + ` &bull; Tahun Kelulusan ` + escapeText(settings.GraduationYear) + `</p>
            </td>
          </tr>
          <tr>
            <td style="padding:30px 28px 10px 28px;">
              <p style="margin:0 0 8px 0;color:` + muted + `;font-size:12px;font-weight:800;letter-spacing:1.8px;text-transform:uppercase;">Kepada</p>
              <h2 style="margin:0;color:` + text + `;font-size:26px;line-height:1.2;font-weight:850;">` + escapeText(student.Name) + `</h2>
              <p style="margin:8px 0 0 0;color:` + muted + `;font-size:14px;line-height:1.6;">` + escapeText(student.ClassName) + ` &bull; ` + escapeText(student.Major) + `</p>
            </td>
          </tr>
          <tr>
            <td style="padding:18px 28px 4px 28px;">
              <div style="border:1px solid #e2e8f0;border-radius:20px;padding:22px;background:#f8fafc;">
                <p style="margin:0;color:` + text + `;font-size:16px;line-height:1.8;">` + escapeText(settings.OpeningText) + `</p>
              </div>
            </td>
          </tr>
          <tr>
            <td style="padding:18px 28px 0 28px;">
              <table role="presentation" width="100%" cellspacing="0" cellpadding="0">
                <tr>
                  <td width="50%" valign="top" style="padding:0 6px 12px 0;">
                    ` + emailInfoCard("Tanggal", settings.EventDate, text, muted) + `
                  </td>
                  <td width="50%" valign="top" style="padding:0 0 12px 6px;">
                    ` + emailInfoCard("Waktu", settings.EventTime, text, muted) + `
                  </td>
                </tr>
                <tr>
                  <td width="50%" valign="top" style="padding:0 6px 12px 0;">
                    ` + emailInfoCard("Tempat", settings.VenueName, text, muted) + `
                  </td>
                  <td width="50%" valign="top" style="padding:0 0 12px 6px;">
                    ` + emailInfoCard("Dress Code", settings.DressCode, text, muted) + `
                  </td>
                </tr>
              </table>
              <div style="border:1px solid #e2e8f0;border-radius:18px;padding:16px 18px;background:#ffffff;">
                <p style="margin:0 0 6px 0;color:` + muted + `;font-size:11px;font-weight:800;letter-spacing:1.4px;text-transform:uppercase;">Alamat</p>
                <p style="margin:0;color:` + text + `;font-size:14px;line-height:1.7;font-weight:650;">` + escapeText(settings.VenueAddress) + `</p>
              </div>
            </td>
          </tr>
          <tr>
            <td style="padding:18px 28px 0 28px;">
              <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="border-radius:22px;overflow:hidden;background:` + primary + `;">
                <tr>
                  <td align="center" width="50%" style="padding:22px 10px;border-right:1px solid rgba(255,255,255,.14);">
                    <p style="margin:0 0 6px 0;color:rgba(255,255,255,.65);font-size:11px;font-weight:800;letter-spacing:1.4px;text-transform:uppercase;">Nomor Siswa</p>
                    <p style="margin:0;color:` + accent + `;font-size:28px;line-height:1;font-weight:900;">` + escapeText(student.StudentSeatNumber) + `</p>
                  </td>
                  <td align="center" width="50%" style="padding:22px 10px;">
                    <p style="margin:0 0 6px 0;color:rgba(255,255,255,.65);font-size:11px;font-weight:800;letter-spacing:1.4px;text-transform:uppercase;">Nomor Pendamping</p>
                    <p style="margin:0;color:` + accent + `;font-size:28px;line-height:1;font-weight:900;">` + escapeText(student.CompanionSeatNumber) + `</p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:28px 28px 10px 28px;">
              <a href="` + escapeAttr(invitationLink) + `" style="display:inline-block;background:` + accent + `;color:#111827;text-decoration:none;border-radius:999px;padding:15px 26px;font-size:14px;font-weight:900;box-shadow:0 14px 30px rgba(15,23,42,.18);">Buka Undangan Digital</a>
              <p style="margin:14px 0 0 0;color:` + muted + `;font-size:12px;line-height:1.6;">Tautan berisi QR Code untuk registrasi kehadiran.</p>
            </td>
          </tr>
          <tr>
            <td style="padding:14px 28px 30px 28px;">
              <div style="border-left:4px solid ` + accent + `;padding:14px 16px;background:#fffbeb;border-radius:14px;color:#713f12;font-size:13px;line-height:1.7;">` + escapeText(settings.AdditionalNote) + `</div>
            </td>
          </tr>
          <tr>
            <td style="padding:22px 28px;background:#f8fafc;border-top:1px solid #e2e8f0;">
              <p style="margin:0 0 10px 0;color:` + muted + `;font-size:11px;font-weight:800;letter-spacing:1.4px;text-transform:uppercase;">Pesan Panitia</p>
              <div style="color:` + muted + `;font-size:13px;line-height:1.75;">` + emailTextBlock(body) + `</div>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
}

func emailInfoCard(label, value, text, muted string) string {
	return `<div style="border:1px solid #e2e8f0;border-radius:18px;padding:16px 18px;background:#ffffff;min-height:74px;">
  <p style="margin:0 0 6px 0;color:` + muted + `;font-size:11px;font-weight:800;letter-spacing:1.4px;text-transform:uppercase;">` + escapeText(label) + `</p>
  <p style="margin:0;color:` + text + `;font-size:14px;line-height:1.55;font-weight:700;">` + escapeText(value) + `</p>
</div>`
}

func emailTextBlock(value string) string {
	escaped := html.EscapeString(strings.TrimSpace(value))
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")
	return escaped
}

func escapeText(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func escapeAttr(value string) string {
	return html.EscapeString(strings.TrimSpace(value))
}

func emailColor(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if len(value) == 7 && strings.HasPrefix(value, "#") {
		return value
	}
	return fallback
}
