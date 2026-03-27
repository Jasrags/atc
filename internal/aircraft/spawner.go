package aircraft

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Jasrags/atc/internal/config"
)

var airlinesICAO = []string{"AA", "UA", "DL", "SW", "BA", "LH", "AF", "JL", "QF", "EK"}

// Spawner generates new aircraft at the edges of the radar.
type Spawner struct {
	rng          *rand.Rand
	lastSpawn    time.Duration
	baseInterval time.Duration
	cfg          config.GameConfig
}

// NewSpawner creates a spawner with a seeded random source and game config.
func NewSpawner(seed int64, cfg config.GameConfig) *Spawner {
	return &Spawner{
		rng:          rand.New(rand.NewSource(seed)),
		baseInterval: 5 * time.Second,
		cfg:          cfg,
	}
}

// ShouldSpawn returns true if it's time to spawn a new aircraft.
func (s *Spawner) ShouldSpawn(elapsed time.Duration, currentCount int) bool {
	if currentCount >= s.maxAircraft(elapsed) {
		return false
	}

	interval := s.spawnInterval(elapsed)
	if elapsed-s.lastSpawn >= interval {
		s.lastSpawn = elapsed
		return true
	}
	return false
}

func (s *Spawner) spawnInterval(elapsed time.Duration) time.Duration {
	params := s.cfg.Difficulty.Params()
	minutes := elapsed.Minutes()
	// Start at 5s, decrease to 1.5s over 5 minutes
	interval := 5.0 - (minutes * 0.7)
	if interval < 1.5 {
		interval = 1.5
	}
	// Apply difficulty multiplier
	interval *= params.IntervalMultiplier
	return time.Duration(interval * float64(time.Second))
}

func (s *Spawner) maxAircraft(elapsed time.Duration) int {
	params := s.cfg.Difficulty.Params()
	minutes := elapsed.Minutes()
	max := 5 + int(minutes*2)
	if max > params.MaxAircraft {
		max = params.MaxAircraft
	}
	return max
}

// Spawn creates a new aircraft at a random edge of the radar area.
func (s *Spawner) Spawn(width, height int) Aircraft {
	callsign := s.generateCallsign()
	edge := s.rng.Intn(4) // 0=N, 1=E, 2=S, 3=W

	var x, y float64

	switch edge {
	case 0: // North
		x = float64(s.rng.Intn(width))
		y = 0
	case 1: // East
		x = float64(width - 1)
		y = float64(s.rng.Intn(height))
	case 2: // South
		x = float64(s.rng.Intn(width))
		y = float64(height - 1)
	case 3: // West
		x = 0
		y = float64(s.rng.Intn(height))
	}

	// Heading biased toward center +/- 20 degrees of randomness
	centerX := float64(width) / 2
	centerY := float64(height) / 2
	toCenter := math.Atan2(centerX-x, -(centerY-y)) * 180 / math.Pi
	if toCenter < 0 {
		toCenter += 360
	}
	heading := int(toCenter) + (s.rng.Intn(40) - 20)
	heading = ((heading % 360) + 360) % 360

	altitude := 5 + s.rng.Intn(16) // 5000-20000ft

	params := s.cfg.Difficulty.Params()
	speedRange := params.MaxSpeed - params.MinSpeed + 1
	speed := params.MinSpeed + s.rng.Intn(speedRange)

	ac := New(callsign, x, y, heading, altitude, speed)
	ac.TrailEnabled = s.cfg.PlaneTrails
	return ac
}

func (s *Spawner) generateCallsign() string {
	if s.cfg.CallsignStyle == config.CallsignShort {
		letter := 'A' + rune(s.rng.Intn(26))
		num := 10 + s.rng.Intn(90) // 10-99
		return fmt.Sprintf("%c%d", letter, num)
	}
	// ICAO style
	airline := airlinesICAO[s.rng.Intn(len(airlinesICAO))]
	num := 100 + s.rng.Intn(900)
	return fmt.Sprintf("%s%d", airline, num)
}
