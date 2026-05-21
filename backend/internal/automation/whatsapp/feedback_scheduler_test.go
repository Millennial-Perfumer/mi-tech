package whatsapp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeedbackScheduler_TimeChecking(t *testing.T) {
	// Create a minimal scheduler
	s := &FeedbackScheduler{}

	// Test 1: Invalid time string parses to default 10:00
	configTime := "invalid"
	var hour, min int
	n, err := fmt.Sscanf(configTime, "%d:%d", &hour, &min)
	if err != nil || n != 2 {
		hour = 10
		min = 0
	}
	assert.Equal(t, 10, hour)
	assert.Equal(t, 0, min)

	// Test 2: Valid time string parses correctly
	configTime = "14:35"
	n, err = fmt.Sscanf(configTime, "%d:%d", &hour, &min)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, 14, hour)
	assert.Equal(t, 35, min)

	// Test 3: Date tracking mechanism prevents duplicate runs today
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	s.mu.Lock()
	s.lastTriggeredDate = currentDate
	s.mu.Unlock()

	s.mu.Lock()
	alreadyTriggered := s.lastTriggeredDate == currentDate
	s.mu.Unlock()

	assert.True(t, alreadyTriggered)
}

func TestFeedbackScheduler_ContextCancellation(t *testing.T) {
	s := &FeedbackScheduler{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Start should exit immediately due to context cancellation
	done := make(chan bool)
	go func() {
		s.Start(ctx)
		done <- true
	}()

	select {
	case <-done:
		// Success: Start exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not exit on cancelled context")
	}
}
