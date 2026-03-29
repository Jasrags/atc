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

func TestRoleFromIndex(t *testing.T) {
	tests := []struct {
		idx  int
		want Role
	}{
		{0, RoleTRACON},
		{1, RoleTower},
		{2, RoleCombined},
		{-1, RoleTRACON}, // out of bounds
		{99, RoleTRACON}, // out of bounds
	}
	for _, tt := range tests {
		got := RoleFromIndex(tt.idx)
		if got != tt.want {
			t.Errorf("RoleFromIndex(%d) = %v, want %v", tt.idx, got, tt.want)
		}
	}
}

func TestRoleOptions(t *testing.T) {
	opts := RoleOptions()
	if len(opts) != 3 {
		t.Fatalf("expected 3 role options, got %d", len(opts))
	}
	if opts[0] != "TRACON" || opts[1] != "Tower" || opts[2] != "Combined" {
		t.Errorf("unexpected role options: %v", opts)
	}
}

func TestIsCommandAllowed(t *testing.T) {
	tests := []struct {
		name string
		role Role
		cmd  string
		want bool
	}{
		// TRACON blocks ground commands
		{"tracon blocks PB", RoleTRACON, "PB", false},
		{"tracon blocks TX", RoleTRACON, "TX", false},
		{"tracon blocks T", RoleTRACON, "T", false},
		{"tracon allows H", RoleTRACON, "H", true},
		{"tracon allows D", RoleTRACON, "D", true},
		{"tracon allows TL", RoleTRACON, "TL", true},

		// Tower blocks TRACON-specific commands
		{"tower blocks D", RoleTower, "D", false},
		{"tower blocks TL", RoleTower, "TL", false},
		{"tower blocks TR", RoleTower, "TR", false},
		{"tower blocks EX", RoleTower, "EX", false},
		{"tower allows H", RoleTower, "H", true},
		{"tower allows A", RoleTower, "A", true},
		{"tower allows S", RoleTower, "S", true},
		{"tower allows L", RoleTower, "L", true},
		{"tower allows PB", RoleTower, "PB", true},
		{"tower allows TX", RoleTower, "TX", true},
		{"tower allows T", RoleTower, "T", true},
		{"tower allows GA", RoleTower, "GA", true},

		// Combined allows everything
		{"combined allows D", RoleCombined, "D", true},
		{"combined allows PB", RoleCombined, "PB", true},
		{"combined allows TL", RoleCombined, "TL", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.IsCommandAllowed(tt.cmd)
			if got != tt.want {
				t.Errorf("%s.IsCommandAllowed(%q) = %v, want %v", tt.role, tt.cmd, got, tt.want)
			}
		})
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
