package util

// StrPtr returns a pointer to the given string, or nil if empty.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Float64Ptr returns a pointer to the given float64.
func Float64Ptr(f float64) *float64 {
	return &f
}

// DerefStr returns the value of a *string pointer, or "" if nil.
func DerefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefFloat64 returns the value of a *float64 pointer, or 0 if nil.
func DerefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
