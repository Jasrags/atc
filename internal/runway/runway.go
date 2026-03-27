package runway

import "math"

// Runway represents an airport runway on the radar display.
type Runway struct {
	X       int // Center X position on the grid
	Y       int // Center Y position on the grid
	Heading int // Required approach heading (0-359)
	Length  int // Visual length in grid cells
}

// New creates a runway at the given position with the specified approach heading.
func New(x, y, heading, length int) Runway {
	return Runway{
		X:       x,
		Y:       y,
		Heading: heading,
		Length:  length,
	}
}

// CanLand checks whether an aircraft with the given parameters can land on this runway.
// Requirements: position within tolerance, heading within +/-10 degrees, altitude == 1.
func (r Runway) CanLand(x, y int, heading, altitude int) bool {
	if altitude != 1 {
		return false
	}

	dx := math.Abs(float64(x - r.X))
	dy := math.Abs(float64(y - r.Y))
	if dx > 2 || dy > 2 {
		return false
	}

	headingDiff := headingDelta(heading, r.Heading)
	return math.Abs(float64(headingDiff)) <= 10
}

// headingDelta returns the shortest signed difference between two headings.
// Positive means clockwise from a to b, negative means counter-clockwise.
func headingDelta(a, b int) int {
	diff := ((b - a + 540) % 360) - 180
	return diff
}

// Cells returns the grid cells occupied by the runway for rendering.
// The runway extends from the center in the direction opposite to the approach heading.
func (r Runway) Cells() [][2]int {
	cells := make([][2]int, r.Length)
	rad := float64(r.Heading) * math.Pi / 180.0

	// Runway body extends opposite to approach heading
	dx := math.Sin(rad)
	dy := -math.Cos(rad)

	for i := 0; i < r.Length; i++ {
		offset := float64(i) - float64(r.Length-1)/2.0
		cx := r.X + int(math.Round(offset*dx))
		cy := r.Y + int(math.Round(offset*dy))
		cells[i] = [2]int{cx, cy}
	}
	return cells
}
