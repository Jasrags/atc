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
		// Altitude changes are throttled; tick enough times
		for i := 0; i < altTickRate+1; i++ {
			ac = ac.Tick()
		}
		if ac.Altitude <= 5 {
			t.Errorf("expected altitude increase after %d ticks, got %d", altTickRate+1, ac.Altitude)
		}
	})

	t.Run("descend", func(t *testing.T) {
		ac := New("T1", 30, 30, 0, 10, 1)
		ac.TargetAltitude = 5
		for i := 0; i < altTickRate+1; i++ {
			ac = ac.Tick()
		}
		if ac.Altitude >= 10 {
			t.Errorf("expected altitude decrease after %d ticks, got %d", altTickRate+1, ac.Altitude)
		}
	})
}

func TestTickSpeedChange(t *testing.T) {
	ac := New("T1", 30, 30, 0, 5, 2)
	ac.TargetSpeed = 5
	// Speed changes are throttled; tick enough times
	for i := 0; i < spdTickRate+1; i++ {
		ac = ac.Tick()
	}
	if ac.Speed <= 2 {
		t.Errorf("expected speed increase after %d ticks, got %d", spdTickRate+1, ac.Speed)
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

func TestTrailDisabledByDefault(t *testing.T) {
	ac := New("T1", 30, 30, 90, 5, 3)
	for i := 0; i < 10; i++ {
		ac = ac.Tick()
	}
	if len(ac.Trail) != 0 {
		t.Errorf("expected no trail when disabled, got %d entries", len(ac.Trail))
	}
}

func TestTrailEnabledRecordsPositions(t *testing.T) {
	ac := New("T1", 30, 30, 90, 5, 3)
	ac.TrailEnabled = true

	ac = ac.Tick()
	if len(ac.Trail) != 1 {
		t.Fatalf("expected 1 trail entry after first tick, got %d", len(ac.Trail))
	}
	if ac.Trail[0] != [2]int{30, 30} {
		t.Errorf("expected trail at (30,30), got %v", ac.Trail[0])
	}
}

func TestTrailCappedAtMaxLength(t *testing.T) {
	ac := New("T1", 30, 30, 90, 5, 3)
	ac.TrailEnabled = true

	for i := 0; i < MaxTrailLength+10; i++ {
		ac = ac.Tick()
	}
	if len(ac.Trail) != MaxTrailLength {
		t.Errorf("expected trail length %d, got %d", MaxTrailLength, len(ac.Trail))
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

func TestGroundTickAdvancesAlongPath(t *testing.T) {
	ac := New("T1", 10, 20, 0, 0, 0)
	ac.State = Taxiing
	ac.TaxiPath = [][2]int{{10, 20}, {15, 20}, {20, 20}}
	ac.TaxiPathIndex = 0

	for i := 0; i < groundTickRate+1; i++ {
		ac = ac.GroundTick()
	}
	if ac.GridX() != 15 || ac.GridY() != 20 {
		t.Errorf("expected position (15,20), got (%d,%d)", ac.GridX(), ac.GridY())
	}
}

func TestGroundTickClearsPathAtEnd(t *testing.T) {
	ac := New("T1", 10, 20, 0, 0, 0)
	ac.State = Taxiing
	ac.TaxiPath = [][2]int{{10, 20}, {15, 20}}
	ac.TaxiPathIndex = 0

	for i := 0; i < groundTickRate+1; i++ {
		ac = ac.GroundTick()
	}
	if len(ac.TaxiPath) != 0 {
		t.Error("taxi path should be cleared when destination reached")
	}
}

func TestGroundTickNoPathDoesNotMove(t *testing.T) {
	ac := New("T1", 10, 20, 0, 0, 0)
	ac.State = Taxiing

	next := ac.GroundTick()
	if next.X != ac.X || next.Y != ac.Y {
		t.Error("aircraft with no path should not move")
	}
}

func TestGroundTickAirborneDoesNothing(t *testing.T) {
	ac := New("T1", 10, 20, 90, 5, 3)
	ac.TaxiPath = [][2]int{{10, 20}, {15, 20}}

	next := ac.GroundTick()
	if next.X != ac.X || next.Y != ac.Y {
		t.Error("airborne aircraft should not use ground tick")
	}
}

func TestIsGroundStates(t *testing.T) {
	groundStates := []State{Landed, Taxiing, AtGate, Pushback, HoldShort, OnRunway}
	for _, s := range groundStates {
		if !s.IsGround() {
			t.Errorf("%v should be ground", s)
		}
	}
	airStates := []State{Approaching, Landing, Departing}
	for _, s := range airStates {
		if s.IsGround() {
			t.Errorf("%v should not be ground", s)
		}
	}
}

func TestPatienceLevelCalm(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.PatienceMax = PatienceDefault
	ac.PatienceTicks = 0
	if ac.PatienceLevel() != 0 {
		t.Errorf("expected calm (0), got %d", ac.PatienceLevel())
	}
}

func TestPatienceLevelWaiting(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.PatienceMax = 300
	ac.PatienceTicks = 200 // > 50%, < 75%
	if ac.PatienceLevel() != 1 {
		t.Errorf("expected waiting (1), got %d", ac.PatienceLevel())
	}
}

func TestPatienceLevelImpatient(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.PatienceMax = 300
	ac.PatienceTicks = 250 // > 75%
	ac.PatienceNagCount = 1
	if ac.PatienceLevel() != 2 {
		t.Errorf("expected impatient (2), got %d", ac.PatienceLevel())
	}
}

func TestPatienceLevelAngry(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.PatienceMax = 300
	ac.PatienceTicks = 700
	ac.PatienceNagCount = PatienceLeaveAt - 1 // angry one nag before leaving
	if ac.PatienceLevel() != 3 {
		t.Errorf("expected angry (3), got %d", ac.PatienceLevel())
	}
}

func TestPatienceLevelNoPatienceSystem(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.PatienceMax = 0 // no patience
	ac.PatienceTicks = 9999
	if ac.PatienceLevel() != 0 {
		t.Errorf("expected calm (0) when patience disabled, got %d", ac.PatienceLevel())
	}
}

func TestResetPatience(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.PatienceTicks = 500
	ac.PatienceNagCount = 3

	reset := ac.ResetPatience()
	if reset.PatienceTicks != 0 {
		t.Errorf("expected PatienceTicks=0, got %d", reset.PatienceTicks)
	}
	if reset.PatienceNagCount != 0 {
		t.Errorf("expected PatienceNagCount=0, got %d", reset.PatienceNagCount)
	}
	// Original unchanged
	if ac.PatienceTicks != 500 {
		t.Error("original should be unchanged (immutability)")
	}
}

func TestIsAirborneStates(t *testing.T) {
	airStates := []State{Approaching, Landing, Departing}
	for _, s := range airStates {
		if !s.IsAirborne() {
			t.Errorf("%v should be airborne", s)
		}
	}
	nonAirStates := []State{Landed, Crashed, Taxiing, AtGate, Pushback, HoldShort, OnRunway}
	for _, s := range nonAirStates {
		if s.IsAirborne() {
			t.Errorf("%v should not be airborne", s)
		}
	}
}

func TestFixNavigationUpdatesHeading(t *testing.T) {
	// Aircraft at (10,10) heading north, fix is at (50,10) — should turn east (heading ~90)
	ac := New("T1", 10, 10, 0, 5, 3)
	ac.TargetFixName = "EAST"
	ac.TargetFixX = 50
	ac.TargetFixY = 10

	// After one tick, target heading should be toward the fix (approximately 90)
	next := ac.Tick()
	if next.TargetHeading < 80 || next.TargetHeading > 100 {
		t.Errorf("expected target heading near 90, got %d", next.TargetHeading)
	}
}

func TestFixNavigationClearsOnArrival(t *testing.T) {
	// Aircraft very close to fix — should clear the fix
	ac := New("T1", 49, 10, 90, 5, 3)
	ac.TargetFixName = "NEAR"
	ac.TargetFixX = 50
	ac.TargetFixY = 10

	next := ac.Tick()
	if next.TargetFixName != "" {
		t.Errorf("expected fix cleared on arrival, got %s", next.TargetFixName)
	}
}

func TestForcedTurnLeft(t *testing.T) {
	// Heading 350, target 010 — shortest is right (+20), but we force left
	ac := New("T1", 50, 50, 350, 5, 3)
	ac.TargetHeading = 10
	ac.ForceTurnDir = ForceTurnLeft

	next := ac.Tick()
	if next.Heading != 349 {
		t.Errorf("expected left turn from 350 to 349, got %d", next.Heading)
	}
}

func TestForcedTurnRight(t *testing.T) {
	// Heading 10, target 350 — shortest is left (-20), but we force right
	ac := New("T1", 50, 50, 10, 5, 3)
	ac.TargetHeading = 350
	ac.ForceTurnDir = ForceTurnRight

	next := ac.Tick()
	if next.Heading != 11 {
		t.Errorf("expected right turn from 10 to 11, got %d", next.Heading)
	}
}

func TestExpeditedAltitude(t *testing.T) {
	ac := New("T1", 50, 50, 90, 5, 3)
	ac.TargetAltitude = 10
	ac.ExpeditedAlt = true

	// Run enough ticks for one altitude change at expedited rate
	normalRate := altTickRate
	expeditedRate := normalRate / 2
	for i := 0; i < expeditedRate+1; i++ {
		ac = ac.Tick()
	}
	if ac.Altitude <= 5 {
		t.Errorf("expected altitude increase with expedite, got %d", ac.Altitude)
	}
}

