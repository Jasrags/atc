package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderHUD(t *testing.T) {
	result := RenderHUD(5, 3, 90*time.Second, "TRACON")

	if !strings.Contains(result, "5") {
		t.Error("expected score in HUD")
	}
	if !strings.Contains(result, "3") {
		t.Error("expected aircraft count in HUD")
	}
	if !strings.Contains(result, "01:30") {
		t.Error("expected formatted time in HUD")
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
