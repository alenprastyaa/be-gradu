package models

import (
	"time"

	"github.com/google/uuid"
)

type School struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Address   string    `json:"address"`
	LogoURL   string    `json:"logo_url"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SchoolInput struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Address  string `json:"address"`
	LogoURL  string `json:"logo_url"`
	IsActive *bool  `json:"is_active"`
}
