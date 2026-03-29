package engine

import (
	"image/color"
	"math"
	"time"
)

// blinkVisible returns true if the element should be visible in the current blink cycle.
// period is the full on+off cycle duration, duty is the fraction spent "on" (0.0-1.0).
func blinkVisible(elapsed time.Duration, period time.Duration, duty float64) bool {
	phase := float64(elapsed%period) / float64(period)
	return phase < duty
}

// pulseAlpha returns a smoothly oscillating alpha value between minA and maxA.
// Used for pulsing glow effects.
func pulseAlpha(elapsed time.Duration, period time.Duration, minA, maxA float64) float64 {
	phase := float64(elapsed%period) / float64(period)
	// Sine wave oscillation.
	t := (math.Sin(phase*2*math.Pi) + 1) / 2 // 0.0 to 1.0
	return minA + t*(maxA-minA)
}

// colorWithAlpha returns a copy of the color with the alpha channel modified.
func colorWithAlpha(c color.RGBA, alpha float64) color.RGBA {
	if alpha > 1 {
		alpha = 1
	}
	if alpha < 0 {
		alpha = 0
	}
	return color.RGBA{c.R, c.G, c.B, uint8(alpha * 255)}
}
