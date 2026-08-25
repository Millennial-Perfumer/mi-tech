package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDateRangeAcceptsMCPDateOnlyArguments(t *testing.T) {
	start, end := parseDateRange("2026-08-24", "2026-08-24")

	require.Equal(t, 2026, start.Year())
	require.Equal(t, time.August, start.Month())
	require.Equal(t, 24, start.Day())
	require.Equal(t, 0, start.Hour())
	require.Equal(t, 0, start.Minute())
	require.Equal(t, 0, start.Second())
	require.Equal(t, dashboardLocation, start.Location())

	require.Equal(t, 2026, end.Year())
	require.Equal(t, time.August, end.Month())
	require.Equal(t, 24, end.Day())
	require.Equal(t, 23, end.Hour())
	require.Equal(t, 59, end.Minute())
	require.Equal(t, 59, end.Second())
	require.Equal(t, 999999999, end.Nanosecond())
}

func TestParseDateRangePreservesRFC3339Arguments(t *testing.T) {
	start, end := parseDateRange("2026-08-24T00:00:00.000Z", "2026-08-24T23:59:59.999Z")

	require.Equal(t, time.Date(2026, time.August, 24, 0, 0, 0, 0, time.UTC), start)
	require.Equal(t, time.Date(2026, time.August, 24, 23, 59, 59, 999000000, time.UTC), end)
}
