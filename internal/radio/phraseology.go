package radio

import (
	"fmt"
	"strings"
	"time"
)

// Phraseology translates game commands and events into radio-style text.

// FormatHeadingChange returns ATC phraseology for a heading instruction.
func FormatHeadingChange(callsign string, currentHeading, targetHeading int) string {
	direction := "turn right"
	delta := ((targetHeading - currentHeading) + 360) % 360
	if delta > 180 {
		direction = "turn left"
	}
	return fmt.Sprintf("%s, %s heading %03d", callsign, direction, targetHeading)
}

// FormatAltitudeChange returns ATC phraseology for an altitude instruction.
func FormatAltitudeChange(callsign string, currentAlt, targetAlt int) string {
	if targetAlt > currentAlt {
		return fmt.Sprintf("%s, climb and maintain %d", callsign, targetAlt*1000)
	}
	return fmt.Sprintf("%s, descend and maintain %d", callsign, targetAlt*1000)
}

// FormatSpeedChange returns ATC phraseology for a speed instruction.
func FormatSpeedChange(callsign string, targetSpeed int) string {
	return fmt.Sprintf("%s, adjust speed %d", callsign, targetSpeed)
}

// FormatClearedToLand returns ATC phraseology for a landing clearance.
func FormatClearedToLand(callsign string) string {
	return fmt.Sprintf("%s, cleared to land", callsign)
}

// FormatEnteringAirspace returns a pilot check-in message.
func FormatEnteringAirspace(callsign string, heading, altitude int) string {
	return fmt.Sprintf("approach, %s with you, heading %03d at %d", callsign, heading, altitude*1000)
}

// FormatLanded returns a pilot message after landing.
func FormatLanded(callsign string) string {
	return fmt.Sprintf("%s clear of the active", callsign)
}

// FormatCollision returns an emergency collision message.
func FormatCollision(callsign1, callsign2 string) string {
	return fmt.Sprintf("COLLISION ALERT: %s and %s", callsign1, callsign2)
}

// CommandPhraseology converts parsed command changes into a single outbound ATC radio message.
// changes is a list of change descriptions (e.g., "HDG 270", "ALT 3", "CLEARED TO LAND").
func CommandPhraseology(elapsed time.Duration, callsign string, changes []string) Message {
	text := fmt.Sprintf("%s, %s", callsign, strings.Join(changes, ", "))
	return ATCMessage(elapsed, callsign, text)
}
