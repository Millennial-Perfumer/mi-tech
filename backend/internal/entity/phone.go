package entity

import (
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// NormalizePhone canonicalizes phone numbers to E.164-like format starting with +91.
// Note: This logic defaults to the +91 (India) prefix for 10-digit formats, matching
// the business region. It uses a robust phone parser.
func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ""
	}

	// Try parsing with phonenumbers, defaulting to India (IN) region
	num, err := phonenumbers.Parse(phone, "IN")
	if err == nil {
		if phonenumbers.IsValidNumber(num) {
			return phonenumbers.Format(num, phonenumbers.E164)
		}
	}

	// For invalid or dummy numbers that couldn't be parsed, simply clean
	// out non-numeric characters and prepend default +91 prefix.
	var cleaned strings.Builder
	for _, r := range phone {
		if (r >= '0' && r <= '9') || r == '+' {
			cleaned.WriteRune(r)
		}
	}
	phone = cleaned.String()

	if !strings.HasPrefix(phone, "+") {
		if strings.HasPrefix(phone, "91") && len(phone) > 10 {
			phone = "+" + phone
		} else if len(phone) > 0 {
			phone = "+91" + phone
		}
	}

	return phone
}
