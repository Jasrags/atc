package collision

import (
	"math"

	"github.com/Jasrags/atc/internal/aircraft"
)

const (
	// SeparationMinLateral is the minimum lateral distance in grid cells.
	SeparationMinLateral = 3
	// SeparationMinVertical is the minimum vertical separation in altitude units (1 = 1000ft).
	SeparationMinVertical = 1
	// ViolationPenalty is the score penalty per violation per tick.
	ViolationPenalty = 50
)

// Collision represents two aircraft that have collided.
type Collision struct {
	Callsign1 string
	Callsign2 string
}

// Violation represents two aircraft that are too close but not yet colliding.
type Violation struct {
	Callsign1 string
	Callsign2 string
	Distance  float64 // lateral distance in grid cells
	AltDiff   int     // vertical difference in altitude units
}

// Check finds all collision pairs among the given aircraft.
// Two aircraft collide if they occupy the same grid cell at the same altitude.
// Only checks airborne aircraft.
func Check(planes map[string]aircraft.Aircraft) []Collision {
	active := airborne(planes)

	var collisions []Collision
	for i := 0; i < len(active); i++ {
		for j := i + 1; j < len(active); j++ {
			a, b := active[i], active[j]
			if a.GridX() == b.GridX() &&
				a.GridY() == b.GridY() &&
				a.Altitude == b.Altitude {
				collisions = append(collisions, Collision{
					Callsign1: a.Callsign,
					Callsign2: b.Callsign,
				})
			}
		}
	}
	return collisions
}

// CheckSeparation finds all pairs of airborne aircraft that violate minimum separation.
// A violation occurs when lateral distance < SeparationMinLateral AND vertical separation < SeparationMinVertical.
// Pairs that are already colliding (same grid cell) are excluded — those are handled by Check.
func CheckSeparation(planes map[string]aircraft.Aircraft) []Violation {
	active := airborne(planes)

	var violations []Violation
	for i := 0; i < len(active); i++ {
		for j := i + 1; j < len(active); j++ {
			a, b := active[i], active[j]

			dx := float64(a.GridX() - b.GridX())
			dy := float64(a.GridY() - b.GridY())
			dist := math.Sqrt(dx*dx + dy*dy)

			altDiff := a.Altitude - b.Altitude
			if altDiff < 0 {
				altDiff = -altDiff
			}

			// Skip exact collisions — handled by Check
			if a.GridX() == b.GridX() && a.GridY() == b.GridY() && a.Altitude == b.Altitude {
				continue
			}

			if dist < float64(SeparationMinLateral) && altDiff < SeparationMinVertical {
				violations = append(violations, Violation{
					Callsign1: a.Callsign,
					Callsign2: b.Callsign,
					Distance:  dist,
					AltDiff:   altDiff,
				})
			}
		}
	}
	return violations
}

// airborne returns only aircraft in airborne states.
func airborne(planes map[string]aircraft.Aircraft) []aircraft.Aircraft {
	active := make([]aircraft.Aircraft, 0, len(planes))
	for _, ac := range planes {
		if ac.State.IsAirborne() {
			active = append(active, ac)
		}
	}
	return active
}
