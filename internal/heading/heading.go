package heading

// Delta returns the shortest signed difference from heading a to heading b.
// Positive means clockwise, negative means counter-clockwise.
// At exactly 180 degrees, the result is -180 (counter-clockwise).
func Delta(a, b int) int {
	return ((b - a + 540) % 360) - 180
}

// AbsDelta returns the absolute shortest angular distance between two headings.
func AbsDelta(a, b int) int {
	d := Delta(a, b)
	if d < 0 {
		return -d
	}
	return d
}
