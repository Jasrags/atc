package command

import (
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
)

func makePlanes() map[string]aircraft.Aircraft {
	return map[string]aircraft.Aircraft{
		"AA123": aircraft.New("AA123", 30, 15, 90, 10, 3),
		"UA456": aircraft.New("UA456", 10, 10, 180, 5, 2),
	}
}

func TestExecuteHeading(t *testing.T) {
	planes := makePlanes()
	h := 270
	cmd := Command{Callsign: "AA123", Heading: &h}

	newPlanes, msg, err := Execute(cmd, planes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].TargetHeading != 270 {
		t.Errorf("target heading = %d, want 270", newPlanes["AA123"].TargetHeading)
	}
	if msg == "" {
		t.Error("expected confirmation message")
	}
	// Original map unchanged
	if planes["AA123"].TargetHeading == 270 {
		t.Error("original map should not be mutated")
	}
}

func TestExecuteAltitude(t *testing.T) {
	planes := makePlanes()
	a := 3
	cmd := Command{Callsign: "AA123", Altitude: &a}

	newPlanes, _, err := Execute(cmd, planes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].TargetAltitude != 3 {
		t.Errorf("target altitude = %d, want 3", newPlanes["AA123"].TargetAltitude)
	}
}

func TestExecuteSpeed(t *testing.T) {
	planes := makePlanes()
	s := 5
	cmd := Command{Callsign: "AA123", Speed: &s}

	newPlanes, _, err := Execute(cmd, planes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].TargetSpeed != 5 {
		t.Errorf("target speed = %d, want 5", newPlanes["AA123"].TargetSpeed)
	}
}

func TestExecuteClearToLand(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", ClearToLand: true}

	newPlanes, _, err := Execute(cmd, planes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Landing {
		t.Errorf("state = %v, want Landing", newPlanes["AA123"].State)
	}
}

func TestExecuteUnknownCallsign(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "XX999", ClearToLand: true}

	_, _, err := Execute(cmd, planes)
	if err == nil {
		t.Error("expected error for unknown callsign")
	}
}

func TestExecuteCrashedAircraft(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Crashed
	planes["AA123"] = ac

	h := 90
	cmd := Command{Callsign: "AA123", Heading: &h}

	_, _, err := Execute(cmd, planes)
	if err == nil {
		t.Error("expected error for crashed aircraft")
	}
}

func TestExecuteAlreadyCleared(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Landing
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", ClearToLand: true}

	_, _, err := Execute(cmd, planes)
	if err == nil {
		t.Error("expected error for already cleared aircraft")
	}
}

func TestExecuteMultiCommand(t *testing.T) {
	planes := makePlanes()
	h, a, s := 180, 5, 4
	cmd := Command{Callsign: "AA123", Heading: &h, Altitude: &a, Speed: &s}

	newPlanes, msg, err := Execute(cmd, planes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ac := newPlanes["AA123"]
	if ac.TargetHeading != 180 {
		t.Errorf("heading = %d, want 180", ac.TargetHeading)
	}
	if ac.TargetAltitude != 5 {
		t.Errorf("altitude = %d, want 5", ac.TargetAltitude)
	}
	if ac.TargetSpeed != 4 {
		t.Errorf("speed = %d, want 4", ac.TargetSpeed)
	}
	if msg == "" {
		t.Error("expected confirmation message")
	}
}
