package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MapID != "san" {
		t.Errorf("MapID = %q, want san", cfg.MapID)
	}
	if cfg.Difficulty != DifficultyNormal {
		t.Errorf("Difficulty = %v, want Normal", cfg.Difficulty)
	}
	if cfg.GameMode != ModeArrivalsOnly {
		t.Errorf("GameMode = %v, want ArrivalsOnly", cfg.GameMode)
	}
	if cfg.CallsignStyle != CallsignICAO {
		t.Errorf("CallsignStyle = %v, want ICAO", cfg.CallsignStyle)
	}
	if cfg.PlaneTrails {
		t.Error("PlaneTrails should default to false")
	}
}

func TestDifficultyParams(t *testing.T) {
	tests := []struct {
		diff           Difficulty
		wantMultiplier float64
		wantMaxAC      int
		wantMinSpd     int
		wantMaxSpd     int
	}{
		{DifficultyEasy, 1.5, 8, 1, 3},
		{DifficultyNormal, 1.0, 15, 2, 4},
		{DifficultyHard, 0.6, 25, 2, 5},
	}

	for _, tt := range tests {
		t.Run(tt.diff.String(), func(t *testing.T) {
			p := tt.diff.Params()
			if p.IntervalMultiplier != tt.wantMultiplier {
				t.Errorf("IntervalMultiplier = %f, want %f", p.IntervalMultiplier, tt.wantMultiplier)
			}
			if p.MaxAircraft != tt.wantMaxAC {
				t.Errorf("MaxAircraft = %d, want %d", p.MaxAircraft, tt.wantMaxAC)
			}
			if p.MinSpeed != tt.wantMinSpd {
				t.Errorf("MinSpeed = %d, want %d", p.MinSpeed, tt.wantMinSpd)
			}
			if p.MaxSpeed != tt.wantMaxSpd {
				t.Errorf("MaxSpeed = %d, want %d", p.MaxSpeed, tt.wantMaxSpd)
			}
		})
	}
}

func TestDifficultyString(t *testing.T) {
	tests := []struct {
		d    Difficulty
		want string
	}{
		{DifficultyEasy, "Easy"},
		{DifficultyNormal, "Normal"},
		{DifficultyHard, "Hard"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("Difficulty(%d).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestCallsignStyleString(t *testing.T) {
	tests := []struct {
		c    CallsignStyle
		want string
	}{
		{CallsignICAO, "ICAO (AA123)"},
		{CallsignShort, "Short (A12)"},
	}
	for _, tt := range tests {
		if got := tt.c.String(); got != tt.want {
			t.Errorf("CallsignStyle(%d).String() = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestGameModeString(t *testing.T) {
	if got := ModeArrivalsOnly.String(); got != "Arrivals Only" {
		t.Errorf("GameMode.String() = %q, want %q", got, "Arrivals Only")
	}
}

func TestOptionLists(t *testing.T) {
	if len(DifficultyOptions()) != 3 {
		t.Errorf("expected 3 difficulty options, got %d", len(DifficultyOptions()))
	}
	if len(GameModeOptions()) != 1 {
		t.Errorf("expected 1 game mode option, got %d", len(GameModeOptions()))
	}
	if len(CallsignStyleOptions()) != 2 {
		t.Errorf("expected 2 callsign style options, got %d", len(CallsignStyleOptions()))
	}
	if len(PlaneTrailsOptions()) != 2 {
		t.Errorf("expected 2 plane trails options, got %d", len(PlaneTrailsOptions()))
	}
}
