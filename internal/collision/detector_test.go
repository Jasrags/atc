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
