package aircraft

import (
	"math"

	"github.com/Jasrags/atc/internal/heading"
)

// State represents the current state of an aircraft.
type State int

const (
	Approaching State = iota
	Landing
	Landed
	Crashed
	Taxiing    // Moving on ground along taxiway
	AtGate     // Parked at gate
	Pushback   // Pushing back from gate
	HoldShort  // Holding short of runway
	OnRunway   // On the runway, cleared for takeoff
	Departing  // Airborne after takeoff, climbing out
)

func (s State) String() string {
	switch s {
	case Approaching:
		return "APPR"
	case Landing:
		return "LAND"
	case Landed:
		return "DONE"
	case Crashed:
		return "CRASH"
	case Taxiing:
		return "TAXI"
	case AtGate:
		return "GATE"
	case Pushback:
		return "PUSH"
	case HoldShort:
		return "HOLD"
	case OnRunway:
		return "TKOF"
	case Departing:
		return "DEPT"
	default:
		return "?"
	}
}

// IsGround reports whether the aircraft is on the ground surface.
func (s State) IsGround() bool {
	switch s {
	case Landed, Taxiing, AtGate, Pushback, HoldShort, OnRunway:
		return true
	}
	return false
}

// IsAirborne reports whether the aircraft is in the air.
func (s State) IsAirborne() bool {
	switch s {
	case Approaching, Landing, Departing:
		return true
	}
	return false
}

// Aircraft represents a single aircraft on the radar.
type Aircraft struct {
	Callsign       string
	X              float64
	Y              float64
	Heading        int // Current heading 0-359
	TargetHeading  int
	Altitude       int // In thousands of feet (1-40)
	TargetAltitude int
	Speed          int // 1-5
	TargetSpeed    int
	State          State
	TrailEnabled   bool     // whether to record position history
	Trail          [][2]int // previous grid positions for trail rendering
	tickCount      int      // internal frame counter for throttled updates

	// Ground operations
	AssignedGate   string   // gate ID (e.g., "G1")
	AssignedRunway string   // runway for departure (e.g., "27")
	TaxiRoute      []string // ordered taxiway names to follow
	TaxiPath       [][2]int // resolved node positions to traverse
	TaxiPathIndex  int      // current position in TaxiPath
}

const MaxTrailLength = 5

// New creates an aircraft with the given parameters.
func New(callsign string, x, y float64, heading, altitude, speed int) Aircraft {
	return Aircraft{
		Callsign:       callsign,
		X:              x,
		Y:              y,
		Heading:        heading,
		TargetHeading:  heading,
		Altitude:       altitude,
		TargetAltitude: altitude,
		Speed:          speed,
		TargetSpeed:    speed,
		State:          Approaching,
	}
}

// GridX returns the rounded X position for grid display.
func (a Aircraft) GridX() int {
	return int(math.Round(a.X))
}

// GridY returns the rounded Y position for grid display.
func (a Aircraft) GridY() int {
	return int(math.Round(a.Y))
}

// Tick advances the aircraft by one frame, returning a new Aircraft.
func (a Aircraft) Tick() Aircraft {
	if a.State == Landed || a.State == Crashed {
		return a
	}

	// Ground aircraft don't use heading-based movement — they follow taxiway paths
	// (handled separately by the ground movement system)
	if a.State.IsGround() {
		return a
	}

	next := a
	next.tickCount++

	// Record trail before moving
	if next.TrailEnabled {
		pos := [2]int{a.GridX(), a.GridY()}
		trail := make([][2]int, len(a.Trail), len(a.Trail)+1)
		copy(trail, a.Trail)
		trail = append(trail, pos)
		if len(trail) > MaxTrailLength {
			trail = trail[len(trail)-MaxTrailLength:]
		}
		next.Trail = trail
	}

	next = next.interpolateHeading()
	next = next.interpolateAltitude()
	next = next.interpolateSpeed()
	next = next.move()
	return next
}

const (
	turnRate       = 1    // degrees per tick (10 deg/s at 10 FPS, 9s for 90-degree turn)
	gridSpeedScale = 0.04 // grid cells per tick per speed unit (speed 3 = ~67s to cross 80 cells)
	altTickRate    = 5    // change altitude by 1 every N ticks (~2s per 1000ft)
	spdTickRate    = 10   // change speed by 1 every N ticks (~1s per speed unit)
)

func (a Aircraft) interpolateHeading() Aircraft {
	if a.Heading == a.TargetHeading {
		return a
	}
	next := a
	delta := heading.Delta(a.Heading, a.TargetHeading)

	absDelta := delta
	if absDelta < 0 {
		absDelta = -absDelta
	}
	if absDelta <= turnRate {
		next.Heading = a.TargetHeading
	} else if delta > 0 {
		next.Heading = (a.Heading + turnRate) % 360
	} else {
		next.Heading = (a.Heading - turnRate + 360) % 360
	}
	return next
}

func (a Aircraft) interpolateAltitude() Aircraft {
	if a.Altitude == a.TargetAltitude {
		return a
	}
	if a.tickCount%altTickRate != 0 {
		return a
	}
	next := a
	if a.Altitude < a.TargetAltitude {
		next.Altitude = a.Altitude + 1
	} else {
		next.Altitude = a.Altitude - 1
	}
	return next
}

func (a Aircraft) interpolateSpeed() Aircraft {
	if a.Speed == a.TargetSpeed {
		return a
	}
	if a.tickCount%spdTickRate != 0 {
		return a
	}
	next := a
	if a.Speed < a.TargetSpeed {
		next.Speed = a.Speed + 1
	} else {
		next.Speed = a.Speed - 1
	}
	return next
}

func (a Aircraft) move() Aircraft {
	next := a
	rad := float64(a.Heading) * math.Pi / 180.0
	speed := float64(a.Speed) * gridSpeedScale
	next.X += speed * math.Sin(rad)
	next.Y -= speed * math.Cos(rad)
	return next
}

const (
	groundTickRate = 3 // advance one node every N ticks (~0.3s per node)
)

// GroundTick advances a ground aircraft along its taxi path, returning a new Aircraft.
// Returns the aircraft unchanged if it has no path or has reached the end.
func (a Aircraft) GroundTick() Aircraft {
	if !a.State.IsGround() || a.State == Landed {
		return a
	}

	// No path to follow — stationary
	if len(a.TaxiPath) == 0 || a.TaxiPathIndex >= len(a.TaxiPath)-1 {
		return a
	}

	next := a
	next.tickCount++

	if next.tickCount%groundTickRate != 0 {
		return next
	}

	// Advance to next node in path
	next.TaxiPathIndex++
	pos := next.TaxiPath[next.TaxiPathIndex]
	next.X = float64(pos[0])
	next.Y = float64(pos[1])

	// If we reached the end of the path, clear it
	if next.TaxiPathIndex >= len(next.TaxiPath)-1 {
		next.TaxiPath = nil
		next.TaxiPathIndex = 0
		next.TaxiRoute = nil
	}

	return next
}

// IsOffScreen reports whether the aircraft has left the radar area.
func (a Aircraft) IsOffScreen(width, height int) bool {
	return a.X < -2 || a.X > float64(width+2) ||
		a.Y < -2 || a.Y > float64(height+2)
}
