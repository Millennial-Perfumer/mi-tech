package helper

import (
	"fmt"
	"regexp"
	"time"
)

var (
	gstinRegex = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
)

// IsValidGSTIN validates the format of GSTIN in India
func IsValidGSTIN(gstin string) bool {
	return gstinRegex.MatchString(gstin)
}

// GetFinancialYear computes the financial year string (e.g. "26-27" for date 2026-06-15)
func GetFinancialYear(t time.Time) string {
	year := t.Year()
	// FY runs from April 1st to March 31st
	if t.Month() < time.April {
		year = year - 1
	}
	nextYearShort := (year + 1) % 100
	return fmt.Sprintf("%02d-%02d", year%100, nextYearShort)
}
