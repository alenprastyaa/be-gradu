package utils

import (
	"errors"
	"math/big"
	"net/url"
	"regexp"
	"strings"
)

var (
	digitsOnly       = regexp.MustCompile(`^\d+$`)
	nonDigitPattern  = regexp.MustCompile(`\D`)
	trailingZeroesRE = regexp.MustCompile(`\.0+$`)
)

func NormalizeWhatsapp(number string) (string, error) {
	cleaned := normalizeNumericString(number)
	cleaned = nonDigitPattern.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimLeft(cleaned, "0")
	if cleaned == "" {
		return "", errors.New("nomor WhatsApp tidak valid")
	}
	if strings.HasPrefix(cleaned, "62") {
		cleaned = "62" + strings.TrimPrefix(cleaned, "62")
	} else if strings.HasPrefix(cleaned, "8") {
		cleaned = "62" + cleaned
	} else {
		return "", errors.New("nomor WhatsApp tidak valid")
	}
	cleaned = normalizeNumericString(cleaned)
	if !strings.HasPrefix(cleaned, "62") || !digitsOnly.MatchString(cleaned) {
		return "", errors.New("nomor WhatsApp tidak valid")
	}
	// Lebih longgar dari aturan lama agar nomor valid dari Excel/input manual tidak mudah ditolak.
	if len(cleaned) < 10 || len(cleaned) > 16 {
		return "", errors.New("nomor WhatsApp tidak valid")
	}
	return cleaned, nil
}

func normalizeNumericString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if strings.ContainsAny(value, "eE") {
		if parsed, _, err := big.ParseFloat(value, 10, 256, big.ToNearestEven); err == nil {
			return parsed.Text('f', 0)
		}
	}
	if trailingZeroesRE.MatchString(value) {
		return trailingZeroesRE.ReplaceAllString(value, "")
	}
	return value
}

func WhatsappLink(number string, message string) string {
	return "https://wa.me/" + number + "?text=" + url.QueryEscape(message)
}
