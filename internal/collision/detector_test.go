package collision

import (
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
)

func TestNoCollision(t *testing.T) {
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 20, 20, 0, 5, 1),
	}

	collisions := Check(planes)
	if len(collisions) != 0 {
		t.Errorf("expected 0 collisions, got %d", len(collisions))
	}
}

func TestNoCollisionDifferentAltitude(t *testing.T) {
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 10, 10, 0, 8, 1),
	}

	collisions := Check(planes)
	if len(collisions) != 0 {
		t.Errorf("expected 0 collisions (different altitude), got %d", len(collisions))
	}
}

func TestCollisionSamePositionAndAltitude(t *testing.T) {
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 10, 10, 0, 5, 1),
	}

	collisions := Check(planes)
	if len(collisions) != 1 {
		t.Fatalf("expected 1 collision, got %d", len(collisions))
	}
}

func TestIgnoreLandedAircraft(t *testing.T) {
	ac1 := aircraft.New("AA1", 10, 10, 0, 5, 1)
	ac1.State = aircraft.Landed
	ac2 := aircraft.New("AA2", 10, 10, 0, 5, 1)

	planes := map[string]aircraft.Aircraft{
		"AA1": ac1,
		"AA2": ac2,
	}

	collisions := Check(planes)
	if len(collisions) != 0 {
		t.Errorf("expected 0 collisions (one landed), got %d", len(collisions))
	}
}

func TestIgnoreCrashedAircraft(t *testing.T) {
	ac1 := aircraft.New("AA1", 10, 10, 0, 5, 1)
	ac1.State = aircraft.Crashed
	ac2 := aircraft.New("AA2", 10, 10, 0, 5, 1)

	planes := map[string]aircraft.Aircraft{
		"AA1": ac1,
		"AA2": ac2,
	}

	collisions := Check(planes)
	if len(collisions) != 0 {
		t.Errorf("expected 0 collisions (one crashed), got %d", len(collisions))
	}
}

func TestMultipleCollisions(t *testing.T) {
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 10, 10, 0, 5, 1),
		"AA3": aircraft.New("AA3", 20, 20, 0, 8, 1),
		"AA4": aircraft.New("AA4", 20, 20, 0, 8, 1),
	}

	collisions := Check(planes)
	if len(collisions) != 2 {
		t.Errorf("expected 2 collisions, got %d", len(collisions))
	}
}

func TestIgnoreDepartingCollision(t *testing.T) {
	ac1 := aircraft.New("AA1", 10, 10, 0, 5, 1)
	ac1.State = aircraft.Departing
	ac2 := aircraft.New("AA2", 10, 10, 0, 5, 1)

	planes := map[string]aircraft.Aircraft{
		"AA1": ac1,
		"AA2": ac2,
	}

	collisions := Check(planes)
	if len(collisions) != 1 {
		t.Errorf("expected 1 collision (departing is airborne), got %d", len(collisions))
	}
}

// --- Separation tests ---

func TestNoSeparationViolation(t *testing.T) {
	// 15 cells apart — well above 3-cell minimum
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 25, 10, 0, 5, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
	}
}

func TestSeparationViolationClose(t *testing.T) {
	// 2 cells apart laterally, same altitude — violation
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 12, 10, 0, 5, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Distance > float64(SeparationMinLateral) {
		t.Errorf("distance %f should be < %d", violations[0].Distance, SeparationMinLateral)
	}
}

func TestSeparationOKWithVerticalSeparation(t *testing.T) {
	// 2 cells apart laterally but 2 altitude units apart — OK
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 12, 10, 0, 7, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (adequate vertical sep), got %d", len(violations))
	}
}

func TestSeparationOKWithAlt1Diff(t *testing.T) {
	// 2 cells apart laterally, 1 altitude unit apart — exactly at minimum, no violation
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 12, 10, 0, 6, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (altDiff=1 meets minimum), got %d", len(violations))
	}
}

func TestSeparationExcludesCollisions(t *testing.T) {
	// Same grid cell, same altitude — this is a collision, not a separation violation
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 10, 10, 0, 5, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 0 {
		t.Errorf("exact collisions should not appear as violations, got %d", len(violations))
	}
}

func TestSeparationIgnoresGroundAircraft(t *testing.T) {
	ac1 := aircraft.New("AA1", 10, 10, 0, 5, 1)
	ac2 := aircraft.New("AA2", 12, 10, 0, 5, 1)
	ac2.State = aircraft.Taxiing

	planes := map[string]aircraft.Aircraft{
		"AA1": ac1,
		"AA2": ac2,
	}

	violations := CheckSeparation(planes)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (one is ground), got %d", len(violations))
	}
}

func TestSeparationDiagonal(t *testing.T) {
	// Distance = sqrt(2^2 + 1^2) = sqrt(5) ≈ 2.24 — within 3 cells
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 12, 11, 0, 5, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 1 {
		t.Fatalf("expected 1 diagonal violation, got %d", len(violations))
	}
}

func TestMultipleViolations(t *testing.T) {
	planes := map[string]aircraft.Aircraft{
		"AA1": aircraft.New("AA1", 10, 10, 0, 5, 1),
		"AA2": aircraft.New("AA2", 11, 10, 0, 5, 1),
		"AA3": aircraft.New("AA3", 50, 50, 0, 8, 1),
		"AA4": aircraft.New("AA4", 51, 50, 0, 8, 1),
	}

	violations := CheckSeparation(planes)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(violations))
	}
}
