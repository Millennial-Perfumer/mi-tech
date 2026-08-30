package util

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
	if err == nil && phonenumbers.IsValidNumber(num) {
		return phonenumbers.Format(num, phonenumbers.E164)
	}

	// If it fails (e.g. missing + for some formats), try explicitly parsing as E164 if it has no prefix
	if !strings.HasPrefix(phone, "+") {
		// Try appending a + and parsing again, in case it was E164 without the plus
		numPlus, errPlus := phonenumbers.Parse("+"+phone, "ZZ") // ZZ is unknown region, forces +
		if errPlus == nil && phonenumbers.IsValidNumber(numPlus) {
			return phonenumbers.Format(numPlus, phonenumbers.E164)
		}
	}

	// If all parsing fails, return original cleaned of obvious noise just in case,
	// but preserve original behavior of not touching fully invalid strings
	// but the original code stripped characters.
	var cleaned strings.Builder
	for _, r := range phone {
		if (r >= '0' && r <= '9') || r == '+' {
			cleaned.WriteRune(r)
		}
	}
	cleanedStr := cleaned.String()

	if cleanedStr == "" {
		return phone
	}
	return cleanedStr
}
