package repositories

import (
	"context"
	"errors"

	"graduation-invitation/internal/authcontext"
	"graduation-invitation/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventSettingsRepository struct {
	db *pgxpool.Pool
}

func NewEventSettingsRepository(db *pgxpool.Pool) *EventSettingsRepository {
	return &EventSettingsRepository{db: db}
}

func (r *EventSettingsRepository) List(ctx context.Context) ([]models.EventSettings, error) {
	rows, err := r.db.Query(ctx, eventTemplateSelect()+` WHERE 1=1`+tenantSQL(ctx, 1)+` ORDER BY is_active DESC, updated_at DESC, template_name`, tenantArgs(ctx)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.EventSettings
	for rows.Next() {
		template, err := scanEventSettings(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, *template)
	}
	return templates, rows.Err()
}

func (r *EventSettingsRepository) GetActive(ctx context.Context) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, eventTemplateSelect()+` WHERE is_active = TRUE`+tenantSQL(ctx, 1)+` ORDER BY updated_at DESC LIMIT 1`, tenantArgs(ctx)...)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, eventTemplateSelect()+` WHERE id=$1`+tenantSQL(ctx, 2), tenantArgs(ctx, id)...)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) Create(ctx context.Context, settings models.EventSettings) (*models.EventSettings, error) {
	schoolID, ok := authcontext.SchoolID(ctx)
	if !ok {
		return nil, errors.New("school_id tidak tersedia")
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO event_templates (
			school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
			event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
			schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
			theme_primary, theme_secondary, theme_accent,
			theme_background, theme_surface, theme_text,
			event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45)
		RETURNING id, school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, *schoolID, settings.TemplateName, settings.IsActive, settings.EventTitle, settings.EventTitleSecond, settings.SchoolName, settings.GraduationYear, settings.RecipientGreeting, settings.OpeningText,
		settings.EventDate, settings.EventTime, settings.VenueName, settings.VenueAddress, settings.MapsURL, settings.DressCodeStudent, settings.DressCodeParent, settings.AdditionalNote,
		settings.ScheduleTitle, settings.ScheduleHeaders, settings.ScheduleRows, settings.SchoolLogoURL, settings.SchoolLogoKey, settings.WhatsappTemplate, settings.EmailSubject, settings.EmailTemplate, settings.AudioURL, settings.AudioKey, settings.AudioTitle, settings.AudioAutoplay,
		settings.ThemePrimary, settings.ThemeSecondary, settings.ThemeAccent,
		settings.ThemeBackground, settings.ThemeSurface, settings.ThemeText,
		settings.EventDatetime, settings.LayoutVariant, settings.ShowCountdown, settings.ShowMap, settings.ShowQR, settings.ShowNote, settings.LayoutSections, settings.SeatMapColumns, settings.SeatMapColorMode, settings.SeatMapLayout)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) Update(ctx context.Context, settings models.EventSettings) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE event_templates SET
			template_name=$2,
			event_title=$3,
			event_title_second=$4,
			school_name=$5,
			graduation_year=$6,
			recipient_greeting=$7,
			opening_text=$8,
			event_date=$9,
			event_time=$10,
			venue_name=$11,
			venue_address=$12,
			maps_url=$13,
			dress_code_student=$14,
			dress_code_parent=$15,
			additional_note=$16,
			schedule_title=$17,
			schedule_headers=$18,
			schedule_rows=$19,
			school_logo_url=$20,
			school_logo_key=$21,
			whatsapp_template=$22,
			email_subject=$23,
			email_template=$24,
			audio_url=$25,
			audio_key=$26,
			audio_title=$27,
			audio_autoplay=$28,
			theme_primary=$29,
			theme_secondary=$30,
			theme_accent=$31,
			theme_background=$32,
			theme_surface=$33,
			theme_text=$34,
			event_datetime=$35,
			layout_variant=$36,
			show_countdown=$37,
			show_map=$38,
			show_qr=$39,
			show_note=$40,
			layout_sections=$41,
			seat_map_columns=$42,
			seat_map_color_mode=$43,
			seat_map_layout=$44,
			updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 `+tenantSQL(ctx, 45)+`
		RETURNING id, school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, tenantArgs(ctx, settings.ID, settings.TemplateName, settings.EventTitle, settings.EventTitleSecond, settings.SchoolName, settings.GraduationYear, settings.RecipientGreeting, settings.OpeningText,
		settings.EventDate, settings.EventTime, settings.VenueName, settings.VenueAddress, settings.MapsURL, settings.DressCodeStudent, settings.DressCodeParent, settings.AdditionalNote,
		settings.ScheduleTitle, settings.ScheduleHeaders, settings.ScheduleRows, settings.SchoolLogoURL, settings.SchoolLogoKey, settings.WhatsappTemplate, settings.EmailSubject, settings.EmailTemplate, settings.AudioURL, settings.AudioKey, settings.AudioTitle, settings.AudioAutoplay,
		settings.ThemePrimary, settings.ThemeSecondary, settings.ThemeAccent,
		settings.ThemeBackground, settings.ThemeSurface, settings.ThemeText,
		settings.EventDatetime, settings.LayoutVariant, settings.ShowCountdown, settings.ShowMap, settings.ShowQR, settings.ShowNote, settings.LayoutSections, settings.SeatMapColumns, settings.SeatMapColorMode, settings.SeatMapLayout)...)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) Activate(ctx context.Context, id uuid.UUID) (*models.EventSettings, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE event_templates SET is_active=FALSE WHERE 1=1`+tenantSQL(ctx, 1), tenantArgs(ctx)...); err != nil {
		return nil, err
	}
	row := tx.QueryRow(ctx, `
		UPDATE event_templates SET is_active=TRUE, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 `+tenantSQL(ctx, 2)+`
		RETURNING id, school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, tenantArgs(ctx, id)...)
	settings, err := scanEventSettings(row)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *EventSettingsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM event_templates WHERE id=$1 AND is_active=FALSE`+tenantSQL(ctx, 2), tenantArgs(ctx, id)...)
	return err
}

func (r *EventSettingsRepository) UpdateActiveSeatMap(ctx context.Context, columns int, colorMode, layout string) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE event_templates
		SET seat_map_columns=$1, seat_map_color_mode=$2, seat_map_layout=$3, updated_at=CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM event_templates
			WHERE is_active = TRUE
			`+tenantSQL(ctx, 4)+`
			ORDER BY updated_at DESC
			LIMIT 1
		)
		RETURNING id, school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
			schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
			theme_primary, theme_secondary, theme_accent,
			theme_background, theme_surface, theme_text,
			event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, tenantArgs(ctx, columns, colorMode, layout)...)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) UpdateActiveSchoolLogo(ctx context.Context, url string, key string) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE event_templates
		SET school_logo_url=$1, school_logo_key=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM event_templates
			WHERE is_active = TRUE
			`+tenantSQL(ctx, 3)+`
			ORDER BY updated_at DESC
			LIMIT 1
		)
		RETURNING id, school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, tenantArgs(ctx, url, key)...)
	return scanEventSettings(row)
}

func eventTemplateSelect() string {
	return `SELECT id, school_id, template_name, is_active, event_title, event_title_second, school_name, graduation_year, recipient_greeting, opening_text,
		event_date, event_time, venue_name, venue_address, maps_url, dress_code_student, dress_code_parent, additional_note,
		schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		theme_primary, theme_secondary, theme_accent,
		theme_background, theme_surface, theme_text,
		event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at FROM event_templates`
}

func scanEventSettings(row pgx.Row) (*models.EventSettings, error) {
	var settings models.EventSettings
	if err := row.Scan(
		&settings.ID,
		&settings.SchoolID,
		&settings.TemplateName,
		&settings.IsActive,
		&settings.EventTitle,
		&settings.EventTitleSecond,
		&settings.SchoolName,
		&settings.GraduationYear,
		&settings.RecipientGreeting,
		&settings.OpeningText,
		&settings.EventDate,
		&settings.EventTime,
		&settings.VenueName,
		&settings.VenueAddress,
		&settings.MapsURL,
		&settings.DressCodeStudent,
		&settings.DressCodeParent,
		&settings.AdditionalNote,
		&settings.ScheduleTitle,
		&settings.ScheduleHeaders,
		&settings.ScheduleRows,
		&settings.SchoolLogoURL,
		&settings.SchoolLogoKey,
		&settings.WhatsappTemplate,
		&settings.EmailSubject,
		&settings.EmailTemplate,
		&settings.AudioURL,
		&settings.AudioKey,
		&settings.AudioTitle,
		&settings.AudioAutoplay,
		&settings.ThemePrimary,
		&settings.ThemeSecondary,
		&settings.ThemeAccent,
		&settings.ThemeBackground,
		&settings.ThemeSurface,
		&settings.ThemeText,
		&settings.EventDatetime,
		&settings.LayoutVariant,
		&settings.ShowCountdown,
		&settings.ShowMap,
		&settings.ShowQR,
		&settings.ShowNote,
		&settings.LayoutSections,
		&settings.SeatMapColumns,
		&settings.SeatMapColorMode,
		&settings.SeatMapLayout,
		&settings.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}
