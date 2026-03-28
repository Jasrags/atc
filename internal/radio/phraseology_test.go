package radio

import (
	"testing"
	"time"
)

func TestFormatHeadingChangeRight(t *testing.T) {
	got := FormatHeadingChange("AA123", 180, 270)
	want := "AA123, turn right heading 270"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHeadingChangeLeft(t *testing.T) {
	got := FormatHeadingChange("AA123", 270, 180)
	want := "AA123, turn left heading 180"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHeadingChangeWrapRight(t *testing.T) {
	// 350 → 010 should be right turn (20 degrees)
	got := FormatHeadingChange("UA1", 350, 10)
	want := "UA1, turn right heading 010"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatHeadingChangeWrapLeft(t *testing.T) {
	// 010 → 350 should be left turn (20 degrees)
	got := FormatHeadingChange("UA1", 10, 350)
	want := "UA1, turn left heading 350"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAltitudeClimb(t *testing.T) {
	got := FormatAltitudeChange("DL5", 3, 10)
	want := "DL5, climb and maintain 10000"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatAltitudeDescend(t *testing.T) {
	got := FormatAltitudeChange("DL5", 10, 3)
	want := "DL5, descend and maintain 3000"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatSpeedChange(t *testing.T) {
	got := FormatSpeedChange("SW22", 3)
	want := "SW22, adjust speed 3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatClearedToLand(t *testing.T) {
	got := FormatClearedToLand("AA123")
	want := "AA123, cleared to land"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatEnteringAirspace(t *testing.T) {
	got := FormatEnteringAirspace("AA123", 270, 5)
	want := "approach, AA123 with you, heading 270 at 5000"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatLanded(t *testing.T) {
	got := FormatLanded("AA123")
	want := "AA123 clear of the active"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatCollision(t *testing.T) {
	got := FormatCollision("AA1", "UA2")
	want := "COLLISION ALERT: AA1 and UA2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCommandPhraseology(t *testing.T) {
	msg := CommandPhraseology(30*time.Second, "AA123", []string{"HDG 270", "ALT 3"})
	if msg.Direction != Outbound {
		t.Error("command phraseology should be outbound")
	}
	if msg.To != "AA123" {
		t.Errorf("expected to=AA123, got %s", msg.To)
	}
	want := "AA123, HDG 270, ALT 3"
	if msg.Text != want {
		t.Errorf("got %q, want %q", msg.Text, want)
	}
}
