package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMachineKeyRowEntityDecodesScopes(t *testing.T) {
	row := machineKeyRow{
		ID:         42,
		Name:       "codex",
		KeyHash:    "hash",
		ScopesJSON: `["orders:read","marketing:publish"]`,
	}

	key, err := row.entity()
	require.NoError(t, err)
	require.Equal(t, int64(42), key.ID)
	require.Equal(t, []string{"orders:read", "marketing:publish"}, key.Scopes)
}

func TestMachineKeyRowEntityHandlesEmptyScopes(t *testing.T) {
	key, err := (machineKeyRow{ScopesJSON: "[]"}).entity()
	require.NoError(t, err)
	require.Empty(t, key.Scopes)
}

func TestMachineKeyRowEntityRejectsMalformedScopes(t *testing.T) {
	_, err := (machineKeyRow{ScopesJSON: `{"orders:read"}`}).entity()
	require.Error(t, err)
}
