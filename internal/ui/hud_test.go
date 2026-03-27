package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderHUD(t *testing.T) {
	result := RenderHUD(5, 3, 90*time.Second, []string{"test message"})

	if !strings.Contains(result, "5") {
		t.Error("expected score in HUD")
	}
	if !strings.Contains(result, "3") {
		t.Error("expected aircraft count in HUD")
	}
	if !strings.Contains(result, "01:30") {
		t.Error("expected formatted time in HUD")
	}
	if !strings.Contains(result, "test message") {
		t.Error("expected message in HUD")
	}
}

func TestRenderHUDMessageLimit(t *testing.T) {
	msgs := make([]string, 10)
	for i := range msgs {
		msgs[i] = "msg"
	}

	result := RenderHUD(0, 0, 0, msgs)

	// Should only show last 5 messages
	count := strings.Count(result, "msg")
	if count != 5 {
		t.Errorf("expected 5 messages, got %d", count)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00"},
		{30 * time.Second, "00:30"},
		{90 * time.Second, "01:30"},
		{5*time.Minute + 15*time.Second, "05:15"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
