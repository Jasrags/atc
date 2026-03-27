package gamemap

// FixType distinguishes different kinds of navigation points on the radar.
type FixType int

const (
	FixWaypoint    FixType = iota // △ named waypoint/fix
	FixAirport                    // ◎ airport
	FixVOR                        // ◉ VOR navigation aid
	FixIntersection               // ✦ intersection fix
)

func (f FixType) Symbol() string {
	switch f {
	case FixWaypoint:
		return "△"
	case FixAirport:
		return "◎"
	case FixVOR:
		return "◉"
	case FixIntersection:
		return "✦"
	default:
		return "·"
	}
}

// Fix represents a named navigation point on the radar.
type Fix struct {
	Name string
	X    int
	Y    int
	Type FixType
}

// Runway represents an airport runway.
type Runway struct {
	Name      string // e.g. "9/27"
	X         int    // Center X position
	Y         int    // Center Y position
	Heading   int    // Primary approach heading (0-359)
	Length    int    // Visual length in grid cells
}

// OppositeHeading returns the reciprocal runway heading.
func (r Runway) OppositeHeading() int {
	return (r.Heading + 180) % 360
}

// RunwayNumber returns the runway number for a given heading (heading / 10).
func RunwayNumber(heading int) int {
	n := (heading + 5) / 10
	if n == 0 {
		n = 36
	}
	return n
}

// Map defines a complete game map with all its features.
type Map struct {
	ID          string
	Name        string
	Description string
	Width       int
	Height      int
	Runways     []Runway
	Fixes       []Fix
}

// PrimaryRunway returns the first runway, which is the main landing target.
func (m Map) PrimaryRunway() Runway {
	if len(m.Runways) == 0 {
		return Runway{X: m.Width / 2, Y: m.Height - 5, Heading: 270, Length: 5}
	}
	return m.Runways[0]
}
