package utils

import (
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ExtractEmail extracts email address from a string
func ExtractEmail(text string) string {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	match := re.FindString(text)
	return match
}

// DecodeBase64 decodes base64 encoded string
func DecodeBase64(encoded string) string {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return encoded
	}
	return string(decoded)
}

// ExtractAmount extracts numeric amount from text
func ExtractAmount(text string) float64 {
	// Remove common currency symbols and text
	cleaned := strings.ReplaceAll(text, "$", "")
	cleaned = strings.ReplaceAll(cleaned, "€", "")
	cleaned = strings.ReplaceAll(cleaned, "£", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")

	// Find first number pattern
	re := regexp.MustCompile(`[\d.]+`)
	matches := re.FindAllString(cleaned, 1)

	if len(matches) == 0 {
		return 0
	}

	amount, err := strconv.ParseFloat(matches[0], 64)
	if err != nil {
		return 0
	}

	return amount
}

// ParseDate parses various date formats
func ParseDate(dateStr string) time.Time {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"Jan 02 2006",
		"January 02, 2006",
		time.RFC822,
		time.RFC2822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}

// ContainsAny checks if text contains any of the provided strings
func ContainsAny(text string, items []string) bool {
	text = strings.ToLower(text)
	for _, item := range items {
		if strings.Contains(text, strings.ToLower(item)) {
			return true
		}
	}

	return false
}

// ExtractDomain extracts domain from email address
func ExtractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}

	return ""
}

// SanitizeString removes special characters and extra spaces
func SanitizeString(str string) string {
	str = strings.TrimSpace(str)
	str = regexp.MustCompile(`\s+`).ReplaceAllString(str, " ")

	return str
}
