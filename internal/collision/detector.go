package collision

import "github.com/Jasrags/atc/internal/aircraft"

// Collision represents two aircraft that have collided.
type Collision struct {
	Callsign1 string
	Callsign2 string
}

// Check finds all collision pairs among the given aircraft.
// Two aircraft collide if they occupy the same grid cell at the same altitude.
// Only checks Approaching and Landing aircraft.
func Check(planes map[string]aircraft.Aircraft) []Collision {
	active := make([]aircraft.Aircraft, 0, len(planes))
	for _, ac := range planes {
		if ac.State == aircraft.Approaching || ac.State == aircraft.Landing {
			active = append(active, ac)
		}
	}

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
