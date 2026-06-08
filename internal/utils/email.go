package utils

import "net/url"

func MailtoLink(email string, subject string, body string) string {
	return "mailto:" + email + "?subject=" + url.QueryEscape(subject) + "&body=" + url.QueryEscape(body)
}
