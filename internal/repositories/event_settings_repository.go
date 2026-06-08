package repositories

import (
	"context"
	"errors"

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
	rows, err := r.db.Query(ctx, eventTemplateSelect()+` ORDER BY is_active DESC, updated_at DESC, template_name`)
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
	row := r.db.QueryRow(ctx, eventTemplateSelect()+` WHERE is_active = TRUE ORDER BY updated_at DESC LIMIT 1`)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, eventTemplateSelect()+` WHERE id=$1`, id)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) Create(ctx context.Context, settings models.EventSettings) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO event_templates (
			template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
			event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
			schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
			theme_primary, theme_secondary, theme_accent,
			theme_background, theme_surface, theme_text,
			event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42)
		RETURNING id, template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, settings.TemplateName, settings.IsActive, settings.EventTitle, settings.SchoolName, settings.GraduationYear, settings.RecipientGreeting, settings.OpeningText,
		settings.EventDate, settings.EventTime, settings.VenueName, settings.VenueAddress, settings.MapsURL, settings.DressCode, settings.AdditionalNote,
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
			school_name=$4,
			graduation_year=$5,
			recipient_greeting=$6,
			opening_text=$7,
			event_date=$8,
			event_time=$9,
			venue_name=$10,
			venue_address=$11,
			maps_url=$12,
			dress_code=$13,
			additional_note=$14,
			schedule_title=$15,
			schedule_headers=$16,
			schedule_rows=$17,
			school_logo_url=$18,
			school_logo_key=$19,
			whatsapp_template=$20,
			email_subject=$21,
			email_template=$22,
			audio_url=$23,
			audio_key=$24,
			audio_title=$25,
			audio_autoplay=$26,
			theme_primary=$27,
			theme_secondary=$28,
			theme_accent=$29,
			theme_background=$30,
			theme_surface=$31,
			theme_text=$32,
			event_datetime=$33,
			layout_variant=$34,
			show_countdown=$35,
			show_map=$36,
			show_qr=$37,
			show_note=$38,
			layout_sections=$39,
			seat_map_columns=$40,
			seat_map_color_mode=$41,
			seat_map_layout=$42,
			updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
		RETURNING id, template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, settings.ID, settings.TemplateName, settings.EventTitle, settings.SchoolName, settings.GraduationYear, settings.RecipientGreeting, settings.OpeningText,
		settings.EventDate, settings.EventTime, settings.VenueName, settings.VenueAddress, settings.MapsURL, settings.DressCode, settings.AdditionalNote,
		settings.ScheduleTitle, settings.ScheduleHeaders, settings.ScheduleRows, settings.SchoolLogoURL, settings.SchoolLogoKey, settings.WhatsappTemplate, settings.EmailSubject, settings.EmailTemplate, settings.AudioURL, settings.AudioKey, settings.AudioTitle, settings.AudioAutoplay,
		settings.ThemePrimary, settings.ThemeSecondary, settings.ThemeAccent,
		settings.ThemeBackground, settings.ThemeSurface, settings.ThemeText,
		settings.EventDatetime, settings.LayoutVariant, settings.ShowCountdown, settings.ShowMap, settings.ShowQR, settings.ShowNote, settings.LayoutSections, settings.SeatMapColumns, settings.SeatMapColorMode, settings.SeatMapLayout)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) Activate(ctx context.Context, id uuid.UUID) (*models.EventSettings, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE event_templates SET is_active=FALSE`); err != nil {
		return nil, err
	}
	row := tx.QueryRow(ctx, `
		UPDATE event_templates SET is_active=TRUE, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
		RETURNING id, template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, id)
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
	_, err := r.db.Exec(ctx, `DELETE FROM event_templates WHERE id=$1 AND is_active=FALSE`, id)
	return err
}

func (r *EventSettingsRepository) UpdateActiveSeatMap(ctx context.Context, columns int, colorMode, layout string) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE event_templates
		SET seat_map_columns=$1, seat_map_color_mode=$2, seat_map_layout=$3, updated_at=CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM event_templates
			WHERE is_active = TRUE
			ORDER BY updated_at DESC
			LIMIT 1
		)
		RETURNING id, template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
			schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
			theme_primary, theme_secondary, theme_accent,
			theme_background, theme_surface, theme_text,
			event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, columns, colorMode, layout)
	return scanEventSettings(row)
}

func (r *EventSettingsRepository) UpdateActiveSchoolLogo(ctx context.Context, url string, key string) (*models.EventSettings, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE event_templates
		SET school_logo_url=$1, school_logo_key=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM event_templates
			WHERE is_active = TRUE
			ORDER BY updated_at DESC
			LIMIT 1
		)
		RETURNING id, template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
		          event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
		          schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		          theme_primary, theme_secondary, theme_accent,
		          theme_background, theme_surface, theme_text,
		          event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at
	`, url, key)
	return scanEventSettings(row)
}

func eventTemplateSelect() string {
	return `SELECT id, template_name, is_active, event_title, school_name, graduation_year, recipient_greeting, opening_text,
		event_date, event_time, venue_name, venue_address, maps_url, dress_code, additional_note,
		schedule_title, schedule_headers, schedule_rows, school_logo_url, school_logo_key, whatsapp_template, email_subject, email_template, audio_url, audio_key, audio_title, audio_autoplay,
		theme_primary, theme_secondary, theme_accent,
		theme_background, theme_surface, theme_text,
		event_datetime, layout_variant, show_countdown, show_map, show_qr, show_note, layout_sections, seat_map_columns, seat_map_color_mode, seat_map_layout, updated_at FROM event_templates`
}

func scanEventSettings(row pgx.Row) (*models.EventSettings, error) {
	var settings models.EventSettings
	if err := row.Scan(
		&settings.ID,
		&settings.TemplateName,
		&settings.IsActive,
		&settings.EventTitle,
		&settings.SchoolName,
		&settings.GraduationYear,
		&settings.RecipientGreeting,
		&settings.OpeningText,
		&settings.EventDate,
		&settings.EventTime,
		&settings.VenueName,
		&settings.VenueAddress,
		&settings.MapsURL,
		&settings.DressCode,
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
