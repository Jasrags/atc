package aircraft

import (
	"regexp"
	"testing"
	"time"

	"github.com/Jasrags/atc/internal/config"
)

func defaultSpawner(seed int64) *Spawner {
	return NewSpawner(seed, config.DefaultConfig())
}

func TestSpawnPosition(t *testing.T) {
	spawner := defaultSpawner(42)
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
	spawner := defaultSpawner(42)

	for i := 0; i < 50; i++ {
		ac := spawner.Spawn(60, 30)

		if ac.Heading < 0 || ac.Heading >= 360 {
			t.Errorf("invalid heading: %d", ac.Heading)
		}
		if ac.Altitude < 5 || ac.Altitude > 20 {
			t.Errorf("invalid altitude: %d", ac.Altitude)
		}
		// Normal difficulty: speed 2-4
		if ac.Speed < 2 || ac.Speed > 4 {
			t.Errorf("invalid speed for Normal difficulty: %d", ac.Speed)
		}
		if ac.State != Approaching {
			t.Errorf("expected Approaching state, got %v", ac.State)
		}
	}
}

func TestShouldSpawn(t *testing.T) {
	spawner := defaultSpawner(42)

	if spawner.ShouldSpawn(0, 0) {
		t.Error("should not spawn at elapsed=0")
	}

	if !spawner.ShouldSpawn(5*time.Second, 0) {
		t.Error("should spawn after 5 seconds")
	}

	// Normal difficulty max = 15
	if spawner.ShouldSpawn(10*time.Second, 15) {
		t.Error("should not spawn at max capacity")
	}
}

func TestCallsignICAO(t *testing.T) {
	spawner := defaultSpawner(42) // default = ICAO
	re := regexp.MustCompile(`^[A-Z]{2}\d{3}$`)

	for i := 0; i < 50; i++ {
		ac := spawner.Spawn(60, 30)
		if !re.MatchString(ac.Callsign) {
			t.Errorf("ICAO callsign %q doesn't match pattern AA123", ac.Callsign)
		}
	}
}

func TestCallsignShort(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.CallsignStyle = config.CallsignShort
	spawner := NewSpawner(42, cfg)
	re := regexp.MustCompile(`^[A-Z]\d{2}$`)

	for i := 0; i < 50; i++ {
		ac := spawner.Spawn(60, 30)
		if !re.MatchString(ac.Callsign) {
			t.Errorf("Short callsign %q doesn't match pattern A12", ac.Callsign)
		}
	}
}

func TestDifficultyAffectsSpeedRange(t *testing.T) {
	tests := []struct {
		diff   config.Difficulty
		minSpd int
		maxSpd int
	}{
		{config.DifficultyEasy, 1, 3},
		{config.DifficultyNormal, 2, 4},
		{config.DifficultyHard, 2, 5},
	}

	for _, tt := range tests {
		t.Run(tt.diff.String(), func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Difficulty = tt.diff
			spawner := NewSpawner(42, cfg)

			for i := 0; i < 100; i++ {
				ac := spawner.Spawn(60, 30)
				if ac.Speed < tt.minSpd || ac.Speed > tt.maxSpd {
					t.Errorf("speed %d outside range [%d, %d]", ac.Speed, tt.minSpd, tt.maxSpd)
				}
			}
		})
	}
}

func TestDifficultyMaxAircraft(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Difficulty = config.DifficultyEasy
	spawner := NewSpawner(42, cfg)

	// Easy: ramp starts at 5, caps at 8 over time
	// At 10s elapsed the ramp is 5 + int(0.166*2) = 5, so max is 5
	if spawner.ShouldSpawn(10*time.Second, 5) {
		t.Error("Easy: should not spawn when at ramp cap")
	}

	// At 5 minutes, ramp = 5 + int(5*2) = 15, capped at Easy max 8
	spawner2 := NewSpawner(42, cfg)
	if spawner2.ShouldSpawn(5*time.Minute, 8) {
		t.Error("Easy: should not spawn at difficulty max 8")
	}

	// Below cap should allow spawn
	spawner3 := NewSpawner(42, cfg)
	if !spawner3.ShouldSpawn(5*time.Minute, 7) {
		t.Error("Easy: should spawn when count 7 < max 8")
	}
}

func TestSpawnFinalApproach(t *testing.T) {
	spawner := defaultSpawner(42)

	tests := []struct {
		name       string
		rwX, rwY   int
		rwHeading  int
		mapW, mapH int
	}{
		{"heading 270 (SAN)", 85, 40, 270, 120, 50},
		{"heading 90 (opposite)", 35, 40, 90, 120, 50},
		{"heading 180 (south)", 60, 25, 180, 120, 50},
		{"small map", 45, 20, 270, 90, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := spawner.SpawnFinalApproach(tt.rwX, tt.rwY, tt.rwHeading, tt.mapW, tt.mapH)

			// Should be in Landing state, pre-cleared
			if ac.State != Landing {
				t.Errorf("state = %v, want Landing", ac.State)
			}

			// Heading should match runway
			if ac.Heading != tt.rwHeading {
				t.Errorf("heading = %d, want %d", ac.Heading, tt.rwHeading)
			}

			// Altitude should be 3 (approach altitude)
			if ac.Altitude != 3 {
				t.Errorf("altitude = %d, want 3", ac.Altitude)
			}

			// Target altitude should be 1 (descending)
			if ac.TargetAltitude != 1 {
				t.Errorf("target altitude = %d, want 1", ac.TargetAltitude)
			}

			// Speed should be 2 (approach speed)
			if ac.Speed != 2 {
				t.Errorf("speed = %d, want 2", ac.Speed)
			}

			// Should have assigned landing runway
			if ac.AssignedLandingRunway == "" {
				t.Error("expected assigned landing runway")
			}

			// No patience (already cleared)
			if ac.PatienceMax != 0 {
				t.Errorf("patience = %d, want 0", ac.PatienceMax)
			}

			// Should be within map bounds
			if ac.X < 1 || ac.X > float64(tt.mapW-2) {
				t.Errorf("X = %.1f, out of bounds [1, %d]", ac.X, tt.mapW-2)
			}
			if ac.Y < 1 || ac.Y > float64(tt.mapH-2) {
				t.Errorf("Y = %.1f, out of bounds [1, %d]", ac.Y, tt.mapH-2)
			}

			// Should be upstream of the runway (not at the runway position)
			if ac.X == float64(tt.rwX) && ac.Y == float64(tt.rwY) {
				t.Error("aircraft spawned at runway position, should be upstream")
			}
		})
	}
}

func TestPlaneTrailsPassedToAircraft(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PlaneTrails = true
	spawner := NewSpawner(42, cfg)

	ac := spawner.Spawn(60, 30)
	if !ac.TrailEnabled {
		t.Error("expected TrailEnabled to be true when PlaneTrails is on")
	}

	cfg.PlaneTrails = false
	spawner2 := NewSpawner(42, cfg)
	ac2 := spawner2.Spawn(60, 30)
	if ac2.TrailEnabled {
		t.Error("expected TrailEnabled to be false when PlaneTrails is off")
	}
}
