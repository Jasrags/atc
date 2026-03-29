package command

import (
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
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

	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].TargetHeading != 270 {
		t.Errorf("target heading = %d, want 270", newPlanes["AA123"].TargetHeading)
	}
	if len(changes) == 0 {
		t.Error("expected at least one change")
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

	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
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

	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
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

	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
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

	_, _, err := Execute(cmd, planes, config.RoleCombined)
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

	_, _, err := Execute(cmd, planes, config.RoleCombined)
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

	_, _, err := Execute(cmd, planes, config.RoleCombined)
	if err == nil {
		t.Error("expected error for already cleared aircraft")
	}
}

func TestExecuteGoAround(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Landing
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", GoAround: true}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Approaching {
		t.Errorf("state = %v, want Approaching", newPlanes["AA123"].State)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecuteGoAroundNotLanding(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", GoAround: true}
	_, _, err := Execute(cmd, planes, config.RoleCombined)
	if err == nil {
		t.Error("expected error for go around on non-landing aircraft")
	}
}

func TestExecutePushback(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.AtGate
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", PushbackApproved: true, ExpectRunway: "27"}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Pushback {
		t.Errorf("state = %v, want Pushback", newPlanes["AA123"].State)
	}
	if newPlanes["AA123"].AssignedRunway != "27" {
		t.Errorf("runway = %s, want 27", newPlanes["AA123"].AssignedRunway)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecutePushbackNotAtGate(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", PushbackApproved: true}
	_, _, err := Execute(cmd, planes, config.RoleCombined)
	if err == nil {
		t.Error("expected error for pushback on airborne aircraft")
	}
}

func TestExecuteTakeoff(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.HoldShort
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", Takeoff: true}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.OnRunway {
		t.Errorf("state = %v, want OnRunway", newPlanes["AA123"].State)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecuteTakeoffNotInPosition(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", Takeoff: true}
	_, _, err := Execute(cmd, planes, config.RoleCombined)
	if err == nil {
		t.Error("expected error for takeoff on airborne aircraft")
	}
}

func TestExecuteTaxiRoute(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Pushback
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", TaxiRoute: []string{"A", "B"}}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Taxiing {
		t.Errorf("state = %v, want Taxiing", newPlanes["AA123"].State)
	}
	if len(newPlanes["AA123"].TaxiRoute) != 2 {
		t.Errorf("taxi route len = %d, want 2", len(newPlanes["AA123"].TaxiRoute))
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecuteHoldShort(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Taxiing
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", HoldShort: "27"}
	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.HoldShort {
		t.Errorf("state = %v, want HoldShort", newPlanes["AA123"].State)
	}
}

func TestExecuteCrossRunway(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.HoldShort
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", CrossRunway: "27"}
	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Taxiing {
		t.Errorf("state = %v, want Taxiing", newPlanes["AA123"].State)
	}
}

func TestExecuteAssignGate(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Landed
	planes["AA123"] = ac

	cmd := Command{Callsign: "AA123", AssignGate: "G3"}
	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Taxiing {
		t.Errorf("state = %v, want Taxiing", newPlanes["AA123"].State)
	}
	if newPlanes["AA123"].AssignedGate != "G3" {
		t.Errorf("gate = %s, want G3", newPlanes["AA123"].AssignedGate)
	}
}

func TestExecuteAirborneOnGround(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.Taxiing
	planes["AA123"] = ac

	h := 270
	cmd := Command{Callsign: "AA123", Heading: &h}
	_, _, err := Execute(cmd, planes, config.RoleCombined)
	if err == nil {
		t.Error("expected error for heading command on ground aircraft")
	}
}

func TestExecuteMultiCommand(t *testing.T) {
	planes := makePlanes()
	h, a, s := 180, 5, 4
	cmd := Command{Callsign: "AA123", Heading: &h, Altitude: &a, Speed: &s}

	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
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
	if len(changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(changes))
	}
}

func TestTRACONRejectsGroundCommands(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.State = aircraft.AtGate
	planes["AA123"] = ac

	tests := []struct {
		name string
		cmd  Command
	}{
		{"pushback", Command{Callsign: "AA123", PushbackApproved: true}},
		{"taxi", Command{Callsign: "AA123", TaxiRoute: []string{"A"}}},
		{"hold short", Command{Callsign: "AA123", HoldShort: "27"}},
		{"cross runway", Command{Callsign: "AA123", CrossRunway: "27"}},
		{"gate", Command{Callsign: "AA123", AssignGate: "G1"}},
		{"takeoff", Command{Callsign: "AA123", Takeoff: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Execute(tt.cmd, planes, config.RoleTRACON)
			if err == nil {
				t.Errorf("expected TRACON to reject %s command", tt.name)
			}
		})
	}
}

func TestTRACONAllowsAirborneCommands(t *testing.T) {
	planes := makePlanes()
	h := 270
	cmd := Command{Callsign: "AA123", Heading: &h}

	_, _, err := Execute(cmd, planes, config.RoleTRACON)
	if err != nil {
		t.Errorf("TRACON should allow heading command: %v", err)
	}
}

func TestExecuteDirectFix(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", DirectFix: "MAFAN"}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].TargetFixName != "MAFAN" {
		t.Errorf("TargetFixName = %s, want MAFAN", newPlanes["AA123"].TargetFixName)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecuteTurnLeft(t *testing.T) {
	planes := makePlanes()
	h := 180
	cmd := Command{Callsign: "AA123", TurnLeft: &h}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].TargetHeading != 180 {
		t.Errorf("TargetHeading = %d, want 180", newPlanes["AA123"].TargetHeading)
	}
	if newPlanes["AA123"].ForceTurnDir != aircraft.ForceTurnLeft {
		t.Errorf("ForceTurnDir = %d, want ForceTurnLeft", newPlanes["AA123"].ForceTurnDir)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecuteTurnRight(t *testing.T) {
	planes := makePlanes()
	h := 90
	cmd := Command{Callsign: "AA123", TurnRight: &h}
	newPlanes, _, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].ForceTurnDir != aircraft.ForceTurnRight {
		t.Errorf("ForceTurnDir = %d, want ForceTurnRight", newPlanes["AA123"].ForceTurnDir)
	}
}

func TestExecuteExpedite(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", Expedite: true}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newPlanes["AA123"].ExpeditedAlt {
		t.Error("expected ExpeditedAlt=true")
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestTowerRejectsTRACONCommands(t *testing.T) {
	planes := makePlanes()
	h := 180

	tests := []struct {
		name string
		cmd  Command
	}{
		{"direct fix", Command{Callsign: "AA123", DirectFix: "MAFAN"}},
		{"turn left", Command{Callsign: "AA123", TurnLeft: &h}},
		{"turn right", Command{Callsign: "AA123", TurnRight: &h}},
		{"expedite", Command{Callsign: "AA123", Expedite: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Execute(tt.cmd, planes, config.RoleTower)
			if err == nil {
				t.Errorf("expected Tower to reject %s command", tt.name)
			}
		})
	}
}

func TestTowerAllowsBasicAndGroundCommands(t *testing.T) {
	planes := makePlanes()

	// Heading command should work for airborne aircraft in Tower mode
	h := 270
	_, _, err := Execute(Command{Callsign: "AA123", Heading: &h}, planes, config.RoleTower)
	if err != nil {
		t.Errorf("Tower should allow heading: %v", err)
	}

	// Altitude command should work
	a := 3
	_, _, err = Execute(Command{Callsign: "AA123", Altitude: &a}, planes, config.RoleTower)
	if err != nil {
		t.Errorf("Tower should allow altitude: %v", err)
	}

	// Land command should work
	_, _, err = Execute(Command{Callsign: "AA123", ClearToLand: true}, planes, config.RoleTower)
	if err != nil {
		t.Errorf("Tower should allow land: %v", err)
	}

	// Ground commands should work
	gatePlanes := makePlanes()
	ac := gatePlanes["AA123"]
	ac.State = aircraft.AtGate
	gatePlanes["AA123"] = ac
	_, _, err = Execute(Command{Callsign: "AA123", PushbackApproved: true}, gatePlanes, config.RoleTower)
	if err != nil {
		t.Errorf("Tower should allow pushback: %v", err)
	}
}

func TestExecuteLandWithRunway(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", ClearToLand: true, LandRunway: "28R"}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].State != aircraft.Landing {
		t.Errorf("state = %v, want Landing", newPlanes["AA123"].State)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
	// Should mention runway in changes
	found := false
	for _, c := range changes {
		if c == "CLEARED TO LAND RWY 28R" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected runway in changes, got %v", changes)
	}
}

func TestExecuteHoldAtFix(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", HoldFix: "MAFAN"}
	newPlanes, changes, err := Execute(cmd, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].HoldingFixName != "MAFAN" {
		t.Errorf("HoldingFixName = %s, want MAFAN", newPlanes["AA123"].HoldingFixName)
	}
	if newPlanes["AA123"].TargetFixName != "MAFAN" {
		t.Errorf("TargetFixName = %s, want MAFAN (should fly to fix first)", newPlanes["AA123"].TargetFixName)
	}
	if len(changes) == 0 {
		t.Error("expected changes")
	}
}

func TestExecuteHeadingCancelsHold(t *testing.T) {
	planes := makePlanes()
	ac := planes["AA123"]
	ac.HoldingFixName = "MAFAN"
	planes["AA123"] = ac

	h := 180
	newPlanes, _, err := Execute(Command{Callsign: "AA123", Heading: &h}, planes, config.RoleCombined)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPlanes["AA123"].HoldingFixName != "" {
		t.Errorf("HoldingFixName should be cleared by heading command, got %s", newPlanes["AA123"].HoldingFixName)
	}
}

func TestTowerRejectsHold(t *testing.T) {
	planes := makePlanes()
	cmd := Command{Callsign: "AA123", HoldFix: "MAFAN"}
	_, _, err := Execute(cmd, planes, config.RoleTower)
	if err == nil {
		t.Error("expected Tower to reject HLD command")
	}
}
