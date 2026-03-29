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
	ac.PatienceMax = PatienceDefault
	return ac
}

// SpawnDeparture creates a new departure aircraft at a random gate.
// Returns the aircraft and true if a gate is available, or a zero aircraft and false if not.
func (s *Spawner) SpawnDeparture(gates []struct {
	ID   string
	X, Y int
}) (Aircraft, bool) {
	if len(gates) == 0 {
		return Aircraft{}, false
	}
	gate := gates[s.rng.Intn(len(gates))]
	callsign := s.generateCallsign()
	ac := NewDeparture(callsign, gate.X, gate.Y, gate.ID)
	ac.TrailEnabled = s.cfg.PlaneTrails
	return ac, true
}

// SpawnFinalApproach creates an arrival aircraft pre-sequenced on final approach.
// The aircraft is placed ~approachDist cells back from the runway, aligned with
// the runway heading, at low altitude, already cleared to land.
func (s *Spawner) SpawnFinalApproach(rwX, rwY, rwHeading, mapWidth, mapHeight int) Aircraft {
	callsign := s.generateCallsign()

	// Place aircraft upstream of the runway along the approach course.
	// "Upstream" = opposite of runway heading.
	approachDist := 50.0
	maxDist := float64(mapWidth) / 2
	if approachDist > maxDist {
		approachDist = maxDist
	}

	// Approach heading is the opposite of runway heading — the aircraft flies
	// toward the runway, so its heading equals the runway heading, but its
	// spawn position is offset in the opposite direction.
	rad := float64(rwHeading) * math.Pi / 180.0
	spawnX := float64(rwX) - approachDist*math.Sin(rad)
	spawnY := float64(rwY) + approachDist*math.Cos(rad)

	// Clamp to map bounds with a small margin
	spawnX = math.Max(1, math.Min(spawnX, float64(mapWidth-2)))
	spawnY = math.Max(1, math.Min(spawnY, float64(mapHeight-2)))

	ac := Aircraft{
		Callsign:              callsign,
		X:                     spawnX,
		Y:                     spawnY,
		Heading:               rwHeading,
		TargetHeading:         rwHeading,
		Altitude:              3,
		TargetAltitude:        1,
		Speed:                 2,
		TargetSpeed:           2,
		State:                 Landing,
		AssignedLandingRunway: fmt.Sprintf("%d", runwayNumber(rwHeading)),
		TrailEnabled:          s.cfg.PlaneTrails,
		PatienceMax:           0, // no patience — already cleared to land
	}
	return ac
}

// runwayNumber converts a heading to a runway number (heading/10, with 0 → 36).
// Mirrors gamemap.RunwayNumber to avoid a cross-package dependency.
func runwayNumber(heading int) int {
	n := (heading + 5) / 10
	if n == 0 {
		n = 36
	}
	return n
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
