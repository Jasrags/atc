package aircraft

import (
	"math"
	"testing"

	"github.com/Jasrags/atc/internal/heading"
)

func TestTickStraightMovement(t *testing.T) {
	tests := []struct {
		name    string
		heading int
		wantDX  float64 // expected sign of X change
		wantDY  float64 // expected sign of Y change
	}{
		{"north (0)", 0, 0, -1},
		{"east (90)", 90, 1, 0},
		{"south (180)", 180, 0, 1},
		{"west (270)", 270, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := New("T1", 30, 30, tt.heading, 5, 3)
			next := ac.Tick()

			dx := next.X - ac.X
			dy := next.Y - ac.Y

			if tt.wantDX != 0 && math.Signbit(dx) != math.Signbit(tt.wantDX) {
				t.Errorf("heading %d: dx=%f, expected sign %f", tt.heading, dx, tt.wantDX)
			}
			if tt.wantDY != 0 && math.Signbit(dy) != math.Signbit(tt.wantDY) {
				t.Errorf("heading %d: dy=%f, expected sign %f", tt.heading, dy, tt.wantDY)
			}
			if tt.wantDX == 0 && math.Abs(dx) > 0.01 {
				t.Errorf("heading %d: expected dx~0, got %f", tt.heading, dx)
			}
			if tt.wantDY == 0 && math.Abs(dy) > 0.01 {
				t.Errorf("heading %d: expected dy~0, got %f", tt.heading, dy)
			}
		})
	}
}

func TestTickHeadingInterpolation(t *testing.T) {
	tests := []struct {
		name          string
		heading       int
		targetHeading int
		wantDirection string // "cw" or "ccw"
	}{
		{"turn right 90->180", 90, 180, "cw"},
		{"turn left 180->90", 180, 90, "ccw"},
		{"shortest right 350->10", 350, 10, "cw"},
		{"shortest left 10->350", 10, 350, "ccw"},
		{"turn right 0->90", 0, 90, "cw"},
		{"turn left 90->0", 90, 0, "ccw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := New("T1", 30, 30, tt.heading, 5, 1)
			ac.TargetHeading = tt.targetHeading
			next := ac.Tick()

			delta := heading.Delta(tt.heading, next.Heading)
			if tt.wantDirection == "cw" && delta <= 0 {
				t.Errorf("expected clockwise turn, heading went from %d to %d (delta=%d)",
					tt.heading, next.Heading, delta)
			}
			if tt.wantDirection == "ccw" && delta >= 0 {
				t.Errorf("expected counter-clockwise turn, heading went from %d to %d (delta=%d)",
					tt.heading, next.Heading, delta)
			}
		})
	}
}

func TestTickHeadingConverges(t *testing.T) {
	ac := New("T1", 30, 30, 90, 5, 1)
	ac.TargetHeading = 180

	// After enough ticks, heading should reach target
	for i := 0; i < 100; i++ {
		ac = ac.Tick()
		if ac.Heading == 180 {
			return
		}
	}
	t.Errorf("heading never reached target 180, stuck at %d", ac.Heading)
}

func TestTickAltitudeClimbDescend(t *testing.T) {
	t.Run("climb", func(t *testing.T) {
		ac := New("T1", 30, 30, 0, 5, 1)
		ac.TargetAltitude = 10
		next := ac.Tick()
		if next.Altitude <= ac.Altitude {
			t.Errorf("expected altitude increase, got %d -> %d", ac.Altitude, next.Altitude)
		}
	})

	t.Run("descend", func(t *testing.T) {
		ac := New("T1", 30, 30, 0, 10, 1)
		ac.TargetAltitude = 5
		next := ac.Tick()
		if next.Altitude >= ac.Altitude {
			t.Errorf("expected altitude decrease, got %d -> %d", ac.Altitude, next.Altitude)
		}
	})
}

func TestTickSpeedChange(t *testing.T) {
	ac := New("T1", 30, 30, 0, 5, 2)
	ac.TargetSpeed = 5
	next := ac.Tick()
	if next.Speed <= ac.Speed {
		t.Errorf("expected speed increase, got %d -> %d", ac.Speed, next.Speed)
	}
}

func TestTickLandedDoesNotMove(t *testing.T) {
	ac := New("T1", 30, 30, 0, 1, 1)
	ac.State = Landed
	next := ac.Tick()
	if next.X != ac.X || next.Y != ac.Y {
		t.Error("landed aircraft should not move")
	}
}

func TestIsOffScreen(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
		want bool
	}{
		{"on screen", 30, 15, false},
		{"left edge", -3, 15, true},
		{"right edge", 63, 15, true},
		{"top edge", 30, -3, true},
		{"bottom edge", 30, 33, true},
		{"just inside", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := New("T1", tt.x, tt.y, 0, 5, 1)
			got := ac.IsOffScreen(60, 30)
			if got != tt.want {
				t.Errorf("IsOffScreen(%f,%f) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

