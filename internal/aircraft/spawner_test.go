package aircraft

import (
	"testing"
	"time"
)

func TestSpawnPosition(t *testing.T) {
	spawner := NewSpawner(42)
	width, height := 60, 30

	for i := 0; i < 50; i++ {
		ac := spawner.Spawn(width, height)

		if ac.X < 0 || ac.X >= float64(width) {
			t.Errorf("aircraft spawned out of X bounds: %f", ac.X)
		}
		if ac.Y < 0 || ac.Y >= float64(height) {
			t.Errorf("aircraft spawned out of Y bounds: %f", ac.Y)
		}

		onEdge := ac.X == 0 || ac.X == float64(width-1) ||
			ac.Y == 0 || ac.Y == float64(height-1)
		if !onEdge {
			t.Errorf("aircraft not on edge: (%f, %f)", ac.X, ac.Y)
		}
	}
}

func TestSpawnParameters(t *testing.T) {
	spawner := NewSpawner(42)

	for i := 0; i < 50; i++ {
		ac := spawner.Spawn(60, 30)

		if ac.Heading < 0 || ac.Heading >= 360 {
			t.Errorf("invalid heading: %d", ac.Heading)
		}
		if ac.Altitude < 5 || ac.Altitude > 20 {
			t.Errorf("invalid altitude: %d", ac.Altitude)
		}
		if ac.Speed < 2 || ac.Speed > 4 {
			t.Errorf("invalid speed: %d", ac.Speed)
		}
		if ac.State != Approaching {
			t.Errorf("expected Approaching state, got %v", ac.State)
		}
		if len(ac.Callsign) < 4 || len(ac.Callsign) > 5 {
			t.Errorf("unexpected callsign length: %q", ac.Callsign)
		}
	}
}

func TestShouldSpawn(t *testing.T) {
	spawner := NewSpawner(42)

	// Should not spawn immediately (lastSpawn=0, elapsed=0)
	if spawner.ShouldSpawn(0, 0) {
		t.Error("should not spawn at elapsed=0")
	}

	// Should spawn after 5 seconds
	if !spawner.ShouldSpawn(5*time.Second, 0) {
		t.Error("should spawn after 5 seconds")
	}

	// Should not spawn if at max capacity
	if spawner.ShouldSpawn(10*time.Second, 15) {
		t.Error("should not spawn at max capacity")
	}
}

func TestCallsignUniqueness(t *testing.T) {
	spawner := NewSpawner(42)
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		ac := spawner.Spawn(60, 30)
		if seen[ac.Callsign] {
			// With 9000 possible callsigns and 100 spawns, duplicates are very unlikely
			// but not impossible. Only fail if we get many.
			continue
		}
		seen[ac.Callsign] = true
	}

	if len(seen) < 90 {
		t.Errorf("expected mostly unique callsigns, got %d unique out of 100", len(seen))
	}
}
