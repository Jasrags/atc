package aircraft

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var airlines = []string{"AA", "UA", "DL", "SW", "BA", "LH", "AF", "JL", "QF", "EK"}

// Spawner generates new aircraft at the edges of the radar.
type Spawner struct {
	rng          *rand.Rand
	lastSpawn    time.Duration
	baseInterval time.Duration
}

// NewSpawner creates a spawner with a seeded random source.
func NewSpawner(seed int64) *Spawner {
	return &Spawner{
		rng:          rand.New(rand.NewSource(seed)),
		baseInterval: 5 * time.Second,
	}
}

// ShouldSpawn returns true if it's time to spawn a new aircraft.
// The spawn interval decreases as elapsed time increases.
func (s *Spawner) ShouldSpawn(elapsed time.Duration, currentCount int) bool {
	maxCount := s.maxAircraft(elapsed)
	if currentCount >= maxCount {
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
	minutes := elapsed.Minutes()
	// Start at 5s, decrease to 1.5s over 5 minutes
	interval := 5.0 - (minutes * 0.7)
	if interval < 1.5 {
		interval = 1.5
	}
	return time.Duration(interval * float64(time.Second))
}

func (s *Spawner) maxAircraft(elapsed time.Duration) int {
	minutes := elapsed.Minutes()
	// Start at 5, increase to 15 over 5 minutes
	max := 5 + int(minutes*2)
	if max > 15 {
		max = 15
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
	speed := 2 + s.rng.Intn(3)     // 2-4

	return New(callsign, x, y, heading, altitude, speed)
}

func (s *Spawner) generateCallsign() string {
	airline := airlines[s.rng.Intn(len(airlines))]
	num := 100 + s.rng.Intn(900)
	return fmt.Sprintf("%s%d", airline, num)
}
