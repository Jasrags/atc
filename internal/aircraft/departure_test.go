package aircraft

import "testing"

func TestNewDeparture(t *testing.T) {
	ac := NewDeparture("UA100", 50, 30, "G2")
	if ac.State != AtGate {
		t.Errorf("state = %v, want AtGate", ac.State)
	}
	if ac.Callsign != "UA100" {
		t.Errorf("callsign = %s, want UA100", ac.Callsign)
	}
	if ac.AssignedGate != "G2" {
		t.Errorf("gate = %s, want G2", ac.AssignedGate)
	}
	if ac.Altitude != 0 {
		t.Errorf("altitude = %d, want 0", ac.Altitude)
	}
	if ac.GridX() != 50 || ac.GridY() != 30 {
		t.Errorf("position = (%d,%d), want (50,30)", ac.GridX(), ac.GridY())
	}
}

func TestTakeoffTickNotOnRunway(t *testing.T) {
	ac := New("T1", 50, 30, 270, 5, 3)
	ac.State = Approaching
	next := ac.TakeoffTick()
	if next.State != Approaching {
		t.Error("TakeoffTick should not change non-OnRunway aircraft")
	}
}

func TestTakeoffTickRoll(t *testing.T) {
	ac := New("T1", 50, 30, 270, 0, 0)
	ac.State = OnRunway
	ac.AssignedRunway = "27"

	// After a few ticks, should still be on runway (rolling)
	for i := 0; i < takeoffRollTicks-1; i++ {
		ac = ac.TakeoffTick()
	}
	if ac.State != OnRunway {
		t.Errorf("should still be on runway during roll, got %v", ac.State)
	}
}

func TestTakeoffTickLiftoff(t *testing.T) {
	ac := New("T1", 50, 30, 270, 0, 0)
	ac.State = OnRunway
	ac.AssignedRunway = "27"

	// After enough ticks, should transition to Departing
	for i := 0; i < takeoffRollTicks+1; i++ {
		ac = ac.TakeoffTick()
	}
	if ac.State != Departing {
		t.Errorf("state = %v, want Departing", ac.State)
	}
	if ac.Altitude != 1 {
		t.Errorf("altitude = %d, want 1 (just airborne)", ac.Altitude)
	}
	if ac.TargetAltitude != departAltitude {
		t.Errorf("target altitude = %d, want %d", ac.TargetAltitude, departAltitude)
	}
	if ac.Speed != departSpeed {
		t.Errorf("speed = %d, want %d", ac.Speed, departSpeed)
	}
}

func TestDepartingAircraftMoves(t *testing.T) {
	ac := New("T1", 50, 30, 270, 1, 3)
	ac.State = Departing

	// Departing is airborne — Tick() should move it
	next := ac.Tick()
	if next.X == ac.X && next.Y == ac.Y {
		t.Error("departing aircraft should move via Tick()")
	}
}
