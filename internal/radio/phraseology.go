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
// Each change code (e.g., "HDG 270", "ALT 3") is translated into proper ATC phraseology.
func CommandPhraseology(elapsed time.Duration, callsign string, changes []string) Message {
	phrases := make([]string, 0, len(changes))
	for _, c := range changes {
		phrases = append(phrases, changeToPhraseology(c))
	}
	text := fmt.Sprintf("%s, %s", callsign, strings.Join(phrases, ", "))
	return ATCMessage(elapsed, callsign, text)
}

// changeToPhraseology maps a single executor change code to ATC phraseology.
func changeToPhraseology(change string) string {
	switch {
	case strings.HasPrefix(change, "TURN LEFT HDG "):
		return "turn left heading " + strings.TrimPrefix(change, "TURN LEFT HDG ")
	case strings.HasPrefix(change, "TURN RIGHT HDG "):
		return "turn right heading " + strings.TrimPrefix(change, "TURN RIGHT HDG ")
	case strings.HasPrefix(change, "HDG "):
		return "fly heading " + strings.TrimPrefix(change, "HDG ")
	case strings.HasPrefix(change, "ALT "):
		return "maintain " + strings.TrimPrefix(change, "ALT ") + ",000"
	case strings.HasPrefix(change, "SPD "):
		return "adjust speed " + strings.TrimPrefix(change, "SPD ")
	case strings.HasPrefix(change, "DIRECT "):
		return "proceed direct " + strings.TrimPrefix(change, "DIRECT ")
	case change == "EXPEDITE":
		return "expedite altitude change"
	case strings.HasPrefix(change, "CLEARED TO LAND RWY "):
		return "cleared to land runway " + strings.TrimPrefix(change, "CLEARED TO LAND RWY ")
	case change == "CLEARED TO LAND":
		return "cleared to land"
	case change == "GO AROUND":
		return "go around, climb and maintain 3000"
	case change == "CLEARED FOR TAKEOFF":
		return "cleared for takeoff"
	case strings.HasPrefix(change, "PUSHBACK APPROVED, EXPECT RWY "):
		return "pushback approved, expect runway " + strings.TrimPrefix(change, "PUSHBACK APPROVED, EXPECT RWY ")
	case change == "PUSHBACK APPROVED":
		return "pushback approved"
	case strings.HasPrefix(change, "TAXI VIA "):
		return "taxi via " + strings.TrimPrefix(change, "TAXI VIA ")
	case strings.HasPrefix(change, "HOLD SHORT RWY "):
		return "hold short runway " + strings.TrimPrefix(change, "HOLD SHORT RWY ")
	case strings.HasPrefix(change, "CROSS RWY "):
		return "cross runway " + strings.TrimPrefix(change, "CROSS RWY ")
	case strings.HasPrefix(change, "TAXI TO GATE "):
		return "taxi to gate " + strings.TrimPrefix(change, "TAXI TO GATE ")
	default:
		return strings.ToLower(change)
	}
}
