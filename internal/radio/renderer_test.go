package radio

import (
	"strings"
	"testing"
	"time"
)

func TestRenderLogEmpty(t *testing.T) {
	result := RenderLog(nil)
	if !strings.Contains(result, "No radio traffic") {
		t.Error("expected empty state message")
	}
}

func TestRenderLogInbound(t *testing.T) {
	msgs := []Message{
		PilotMessage(30*time.Second, "AA123", "requesting approach"),
	}
	result := RenderLog(msgs)
	if !strings.Contains(result, "AA123") {
		t.Error("expected callsign in output")
	}
	if !strings.Contains(result, "00:30") {
		t.Error("expected timestamp in output")
	}
	if !strings.Contains(result, "requesting approach") {
		t.Error("expected message text in output")
	}
}

func TestRenderLogOutbound(t *testing.T) {
	msgs := []Message{
		ATCMessage(45*time.Second, "UA456", "turn right heading 090"),
	}
	result := RenderLog(msgs)
	if !strings.Contains(result, "ATC") {
		t.Error("expected ATC prefix in output")
	}
	if !strings.Contains(result, "UA456") {
		t.Error("expected callsign in output")
	}
}

func TestRenderLogUrgent(t *testing.T) {
	msg := PilotMessage(60*time.Second, "DL789", "say again")
	msg.Priority = Urgent
	result := RenderLog([]Message{msg})
	if result == "" {
		t.Error("expected non-empty output for urgent message")
	}
}

func TestRenderLogEmergency(t *testing.T) {
	msg := SystemMessage(90*time.Second, "COLLISION ALERT", Emergency)
	result := RenderLog([]Message{msg})
	if !strings.Contains(result, "COLLISION ALERT") {
		t.Error("expected emergency text in output")
	}
}

func TestRenderLogMultiple(t *testing.T) {
	msgs := []Message{
		PilotMessage(10*time.Second, "AA1", "with you"),
		ATCMessage(12*time.Second, "AA1", "radar contact"),
		PilotMessage(20*time.Second, "UA2", "requesting vectors"),
	}
	result := RenderLog(msgs)

	if !strings.Contains(result, "AA1") {
		t.Error("expected AA1 in output")
	}
	if !strings.Contains(result, "UA2") {
		t.Error("expected UA2 in output")
	}
}

func TestFormatTimestamp(t *testing.T) {
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
		got := formatTimestamp(tt.d)
		if got != tt.want {
			t.Errorf("formatTimestamp(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
