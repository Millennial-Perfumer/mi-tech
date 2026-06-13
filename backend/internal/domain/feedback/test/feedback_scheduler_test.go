package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	feedbackService "mi-tech/internal/domain/feedback/service"

	"github.com/stretchr/testify/assert"
)

func TestFeedbackScheduler_TimeChecking(t *testing.T) {
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
}

func TestFeedbackScheduler_ContextCancellation(t *testing.T) {
	s := &feedbackService.FeedbackScheduler{}

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

func TestFeedbackScheduler_ShouldTrigger(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	assert.NoError(t, err)

	// Case 1: Already triggered today -> should NOT trigger
	now := time.Date(2026, 5, 22, 11, 0, 0, 0, loc)
	triggered, nextDate := feedbackService.ShouldTrigger(now, "2026-05-22", 10, 0, loc)
	assert.False(t, triggered)
	assert.Equal(t, "2026-05-22", nextDate)

	// Case 2: Not triggered today, current time is before scheduled time -> should NOT trigger
	now = time.Date(2026, 5, 22, 9, 59, 59, 0, loc)
	triggered, nextDate = feedbackService.ShouldTrigger(now, "2026-05-21", 10, 0, loc)
	assert.False(t, triggered)
	assert.Equal(t, "2026-05-21", nextDate)

	// Case 3: Not triggered today, current time is exactly scheduled time -> should trigger
	now = time.Date(2026, 5, 22, 10, 0, 0, 0, loc)
	triggered, nextDate = feedbackService.ShouldTrigger(now, "2026-05-21", 10, 0, loc)
	assert.True(t, triggered)
	assert.Equal(t, "2026-05-22", nextDate)

	// Case 4: Not triggered today, current time is after scheduled time -> should trigger (resiliency for missed window)
	now = time.Date(2026, 5, 22, 10, 5, 0, 0, loc)
	triggered, nextDate = feedbackService.ShouldTrigger(now, "2026-05-21", 10, 0, loc)
	assert.True(t, triggered)
	assert.Equal(t, "2026-05-22", nextDate)
}
