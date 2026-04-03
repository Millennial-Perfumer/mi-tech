package entity

import (
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`[^0-9+]`)

// NormalizePhone canonicalizes phone numbers to E.164-like format starting with +91.
// Note: This logic defaults to the +91 (India) prefix for 10-digit formats, matching
// the business region. It removes non-numeric characters and handles varied formatting.
func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}
	// Remove all non-numeric characters except +
	phone = phoneRegex.ReplaceAllString(phone, "")

	if !strings.HasPrefix(phone, "+") {
		// If it's 10 digits, add +91
		if len(phone) == 10 {
			phone = "+91" + phone
		} else if len(phone) == 12 && strings.HasPrefix(phone, "91") {
			// If it's 91XXXXXXXXXX, add +
			phone = "+" + phone
		} else if !strings.HasPrefix(phone, "91") && len(phone) > 0 {
			// Default to adding +91 for anything else that looks like a number
			phone = "+91" + phone
		}
	}
	return phone
}
