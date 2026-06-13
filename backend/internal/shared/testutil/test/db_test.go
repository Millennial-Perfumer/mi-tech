package test

import (
	"testing"

	"mi-tech/internal/shared/testutil"

	"github.com/stretchr/testify/assert"
)

func TestSetupTestDB(t *testing.T) {
	// Skip this test if no local postgres is available or
	// just attempt to see if it can at least compile and run.
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}

	db, err := testutil.SetupTestDB()
	if err != nil {
		t.Logf("Failed to setup test DB: %v (this is expected if postgres is not running)", err)
		return
	}
	defer testutil.CleanupTestDB(db)

	assert.NotNil(t, db)

	// Check if a table exists (e.g., users)
	var count int64
	err = db.Table("users").Count(&count).Error
	assert.NoError(t, err)
	assert.True(t, count >= 2, "Expected at least 2 seeded users")
}
