package gamemap

import "testing"

func TestRunwayNumber(t *testing.T) {
	tests := []struct {
		heading int
		want    int
	}{
		{270, 27},
		{90, 9},
		{280, 28},
		{100, 10},
		{360, 36},
		{0, 36},
		{5, 1},
		{355, 36},
	}
	for _, tt := range tests {
		got := RunwayNumber(tt.heading)
		if got != tt.want {
			t.Errorf("RunwayNumber(%d) = %d, want %d", tt.heading, got, tt.want)
		}
	}
}

func TestOppositeHeading(t *testing.T) {
	tests := []struct {
		heading int
		want    int
	}{
		{270, 90},
		{90, 270},
		{0, 180},
		{180, 0},
	}
	for _, tt := range tests {
		rw := Runway{Heading: tt.heading}
		got := rw.OppositeHeading()
		if got != tt.want {
			t.Errorf("OppositeHeading(%d) = %d, want %d", tt.heading, got, tt.want)
		}
	}
}

func TestPrimaryRunwayFallback(t *testing.T) {
	m := Map{Width: 80, Height: 35}
	rw := m.PrimaryRunway()
	if rw.Heading != 270 {
		t.Errorf("fallback heading = %d, want 270", rw.Heading)
	}
}

func TestAllMaps(t *testing.T) {
	maps := All()
	if len(maps) < 3 {
		t.Fatalf("expected at least 3 maps, got %d", len(maps))
	}

	ids := make(map[string]bool)
	for _, m := range maps {
		if m.ID == "" {
			t.Error("map has empty ID")
		}
		if ids[m.ID] {
			t.Errorf("duplicate map ID: %s", m.ID)
		}
		ids[m.ID] = true

		if m.Width <= 0 || m.Height <= 0 {
			t.Errorf("map %s has invalid dimensions: %dx%d", m.ID, m.Width, m.Height)
		}
		if len(m.Runways) == 0 {
			t.Errorf("map %s has no runways", m.ID)
		}

		// Verify all fixes are within bounds
		for _, f := range m.Fixes {
			if f.X < 0 || f.X >= m.Width || f.Y < 0 || f.Y >= m.Height {
				t.Errorf("map %s: fix %s at (%d,%d) out of bounds %dx%d",
					m.ID, f.Name, f.X, f.Y, m.Width, m.Height)
			}
		}

		// Verify all runways are within bounds
		for _, r := range m.Runways {
			if r.X < 0 || r.X >= m.Width || r.Y < 0 || r.Y >= m.Height {
				t.Errorf("map %s: runway %s at (%d,%d) out of bounds",
					m.ID, r.Name, r.X, r.Y)
			}
		}
	}
}

func TestByID(t *testing.T) {
	m := ByID("san")
	if m.ID != "san" {
		t.Errorf("expected san, got %s", m.ID)
	}

	m = ByID("nonexistent")
	if m.ID != "tutorial" {
		t.Errorf("expected tutorial fallback, got %s", m.ID)
	}
}

func TestFixSymbols(t *testing.T) {
	tests := []struct {
		ft   FixType
		want string
	}{
		{FixWaypoint, "△"},
		{FixAirport, "◎"},
		{FixVOR, "◉"},
		{FixIntersection, "✦"},
	}
	for _, tt := range tests {
		got := tt.ft.Symbol()
		if got != tt.want {
			t.Errorf("FixType(%d).Symbol() = %q, want %q", tt.ft, got, tt.want)
		}
	}
}
