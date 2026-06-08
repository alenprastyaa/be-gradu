package models

import (
	"time"

	"github.com/google/uuid"
)

type EventSettings struct {
	ID                uuid.UUID  `json:"id"`
	SchoolID          *uuid.UUID `json:"school_id,omitempty"`
	TemplateName      string     `json:"template_name"`
	IsActive          bool       `json:"is_active"`
	EventTitle        string     `json:"event_title"`
	SchoolName        string     `json:"school_name"`
	GraduationYear    string     `json:"graduation_year"`
	RecipientGreeting string     `json:"recipient_greeting"`
	OpeningText       string     `json:"opening_text"`
	EventDate         string     `json:"event_date"`
	EventTime         string     `json:"event_time"`
	VenueName         string     `json:"venue_name"`
	VenueAddress      string     `json:"venue_address"`
	MapsURL           string     `json:"maps_url"`
	DressCode         string     `json:"dress_code"`
	AdditionalNote    string     `json:"additional_note"`
	ScheduleTitle     string     `json:"schedule_title"`
	ScheduleHeaders   string     `json:"schedule_headers"`
	ScheduleRows      string     `json:"schedule_rows"`
	SchoolLogoURL     string     `json:"school_logo_url"`
	SchoolLogoKey     string     `json:"school_logo_key"`
	WhatsappTemplate  string     `json:"whatsapp_template"`
	EmailSubject      string     `json:"email_subject"`
	EmailTemplate     string     `json:"email_template"`
	AudioURL          string     `json:"audio_url"`
	AudioKey          string     `json:"audio_key"`
	AudioTitle        string     `json:"audio_title"`
	AudioAutoplay     bool       `json:"audio_autoplay"`
	ThemePrimary      string     `json:"theme_primary"`
	ThemeSecondary    string     `json:"theme_secondary"`
	ThemeAccent       string     `json:"theme_accent"`
	ThemeBackground   string     `json:"theme_background"`
	ThemeSurface      string     `json:"theme_surface"`
	ThemeText         string     `json:"theme_text"`
	EventDatetime     string     `json:"event_datetime"`
	LayoutVariant     string     `json:"layout_variant"`
	ShowCountdown     bool       `json:"show_countdown"`
	ShowMap           bool       `json:"show_map"`
	ShowQR            bool       `json:"show_qr"`
	ShowNote          bool       `json:"show_note"`
	LayoutSections    string     `json:"layout_sections"`
	SeatMapColumns    int        `json:"seat_map_columns"`
	SeatMapColorMode  string     `json:"seat_map_color_mode"`
	SeatMapLayout     string     `json:"seat_map_layout"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
