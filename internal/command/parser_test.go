package command

import "testing"

func intPtr(v int) *int { return &v }

func TestParseValid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCall    string
		wantHeading *int
		wantAlt     *int
		wantSpeed   *int
		wantLand    bool
	}{
		{"single heading", "AA123 H270", "AA123", intPtr(270), nil, nil, false},
		{"single altitude", "UA456 A3", "UA456", nil, intPtr(3), nil, false},
		{"single speed", "DL789 S4", "DL789", nil, nil, intPtr(4), false},
		{"clear to land", "AA123 L", "AA123", nil, nil, nil, true},
		{"multi command", "AA123 H270 A3 S2", "AA123", intPtr(270), intPtr(3), intPtr(2), false},
		{"heading and land", "BA100 H90 L", "BA100", intPtr(90), nil, nil, true},
		{"lowercase input", "aa123 h270 a3", "AA123", intPtr(270), intPtr(3), nil, false},
		{"extra whitespace", "  AA123   H270  ", "AA123", intPtr(270), nil, nil, false},
		{"heading zero", "AA123 H0", "AA123", intPtr(0), nil, nil, false},
		{"heading 359", "AA123 H359", "AA123", intPtr(359), nil, nil, false},
		{"altitude 1", "AA123 A1", "AA123", nil, intPtr(1), nil, false},
		{"altitude 40", "AA123 A40", "AA123", nil, intPtr(40), nil, false},
		{"speed 1", "AA123 S1", "AA123", nil, nil, intPtr(1), false},
		{"speed 5", "AA123 S5", "AA123", nil, nil, intPtr(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd.Callsign != tt.wantCall {
				t.Errorf("callsign = %q, want %q", cmd.Callsign, tt.wantCall)
			}
			assertIntPtr(t, "heading", cmd.Heading, tt.wantHeading)
			assertIntPtr(t, "altitude", cmd.Altitude, tt.wantAlt)
			assertIntPtr(t, "speed", cmd.Speed, tt.wantSpeed)
			if cmd.ClearToLand != tt.wantLand {
				t.Errorf("clearToLand = %v, want %v", cmd.ClearToLand, tt.wantLand)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"callsign only", "AA123"},
		{"bad heading H999", "AA123 H999"},
		{"bad heading H-1", "AA123 H-1"},
		{"bad heading Habc", "AA123 Habc"},
		{"bad altitude A0", "AA123 A0"},
		{"bad altitude A99", "AA123 A99"},
		{"bad speed S0", "AA123 S0"},
		{"bad speed S6", "AA123 S6"},
		{"unknown token", "AA123 X5"},
		{"single letter unknown", "AA123 Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func assertIntPtr(t *testing.T, name string, got, want *int) {
	t.Helper()
	if got == nil && want == nil {
		return
	}
	if got == nil && want != nil {
		t.Errorf("%s = nil, want %d", name, *want)
		return
	}
	if got != nil && want == nil {
		t.Errorf("%s = %d, want nil", name, *got)
		return
	}
	if *got != *want {
		t.Errorf("%s = %d, want %d", name, *got, *want)
	}
}
