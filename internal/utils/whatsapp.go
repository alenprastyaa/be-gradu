package utils

import (
	"errors"
	"math/big"
	"net/url"
	"regexp"
	"strings"
)

var digitsOnly = regexp.MustCompile(`^\d{10,15}$`)

func NormalizeWhatsapp(number string) (string, error) {
	cleaned := strings.NewReplacer(" ", "", "+", "", "-", "", "(", "", ")", "", "\t", "", "\n", "", "\r", "").Replace(strings.TrimSpace(number))
	cleaned = normalizeNumericString(cleaned)
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "62" + strings.TrimPrefix(cleaned, "0")
	}
	if strings.HasPrefix(cleaned, "8") {
		cleaned = "62" + cleaned
	}
	if !strings.HasPrefix(cleaned, "62") || !digitsOnly.MatchString(cleaned) {
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
	if strings.HasSuffix(value, ".0") {
		return strings.TrimSuffix(value, ".0")
	}
	return value
}

func WhatsappLink(number string, message string) string {
	return "https://wa.me/" + number + "?text=" + url.QueryEscape(message)
}
