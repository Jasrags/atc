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
	default:
		return "?"
	}
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
}

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

	next := a
	next = next.interpolateHeading()
	next = next.interpolateAltitude()
	next = next.interpolateSpeed()
	next = next.move()
	return next
}

const (
	turnRate       = 3   // degrees per tick
	gridSpeedScale = 0.3 // grid cells per tick per speed unit
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

// IsOffScreen reports whether the aircraft has left the radar area.
func (a Aircraft) IsOffScreen(width, height int) bool {
	return a.X < -2 || a.X > float64(width+2) ||
		a.Y < -2 || a.Y > float64(height+2)
}
