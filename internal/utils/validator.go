package utils

import (
	"errors"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func CleanText(value string) string {
	return strings.TrimSpace(value)
}

func NormalizeEmail(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil, nil
	}
	if !emailPattern.MatchString(cleaned) {
		return nil, errors.New("format email tidak valid")
	}
	return &cleaned, nil
}

func ValidateRequired(value string, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " wajib diisi")
	}
	return nil
}
