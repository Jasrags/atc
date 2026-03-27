package runway

import "testing"

func TestCanLand(t *testing.T) {
	rw := New(40, 20, 270, 5)

	tests := []struct {
		name     string
		x, y     int
		heading  int
		altitude int
		want     bool
	}{
		{"exact alignment", 40, 20, 270, 1, true},
		{"within heading tolerance +10", 40, 20, 280, 1, true},
		{"within heading tolerance -10", 40, 20, 260, 1, true},
		{"within position tolerance", 41, 21, 270, 1, true},
		{"too far from runway", 45, 20, 270, 1, false},
		{"wrong heading", 40, 20, 90, 1, false},
		{"too high altitude", 40, 20, 270, 3, false},
		{"altitude zero invalid", 40, 20, 270, 0, false},
		{"heading wrap 0/360 tolerance", 40, 20, 265, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rw.CanLand(tt.x, tt.y, tt.heading, tt.altitude)
			if got != tt.want {
				t.Errorf("CanLand(%d,%d, h=%d, a=%d) = %v, want %v",
					tt.x, tt.y, tt.heading, tt.altitude, got, tt.want)
			}
		})
	}
}

func TestCanLandHeadingWrap(t *testing.T) {
	rw := New(40, 20, 5, 5)

	tests := []struct {
		name    string
		heading int
		want    bool
	}{
		{"heading 5 exact", 5, true},
		{"heading 355 within tolerance", 355, true},
		{"heading 15 within tolerance", 15, true},
		{"heading 350 outside tolerance", 350, false},
		{"heading 20 outside tolerance", 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rw.CanLand(40, 20, tt.heading, 1)
			if got != tt.want {
				t.Errorf("CanLand(h=%d) = %v, want %v", tt.heading, got, tt.want)
			}
		})
	}
}

func TestCells(t *testing.T) {
	rw := New(40, 20, 270, 5)
	cells := rw.Cells()

	if len(cells) != 5 {
		t.Fatalf("expected 5 cells, got %d", len(cells))
	}

	// For heading 270 (west), runway extends east-west
	// All cells should have the same Y
	for _, c := range cells {
		if c[1] != 20 {
			t.Errorf("expected y=20, got y=%d for cell %v", c[1], c)
		}
	}
}
