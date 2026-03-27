package command

import (
	"fmt"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
)

// Execute applies a parsed command to the aircraft map, returning a new map and confirmation message.
func Execute(cmd Command, planes map[string]aircraft.Aircraft) (map[string]aircraft.Aircraft, string, error) {
	ac, exists := planes[cmd.Callsign]
	if !exists {
		return planes, "", fmt.Errorf("unknown aircraft: %s", cmd.Callsign)
	}

	if ac.State == aircraft.Crashed {
		return planes, "", fmt.Errorf("%s has crashed", cmd.Callsign)
	}

	if ac.State == aircraft.Landed {
		return planes, "", fmt.Errorf("%s has already landed", cmd.Callsign)
	}

	var changes []string

	if cmd.Heading != nil {
		ac.TargetHeading = *cmd.Heading
		changes = append(changes, fmt.Sprintf("HDG %03d", *cmd.Heading))
	}

	if cmd.Altitude != nil {
		ac.TargetAltitude = *cmd.Altitude
		changes = append(changes, fmt.Sprintf("ALT %d", *cmd.Altitude))
	}

	if cmd.Speed != nil {
		ac.TargetSpeed = *cmd.Speed
		changes = append(changes, fmt.Sprintf("SPD %d", *cmd.Speed))
	}

	if cmd.ClearToLand {
		if ac.State == aircraft.Landing {
			return planes, "", fmt.Errorf("%s already cleared to land", cmd.Callsign)
		}
		ac.State = aircraft.Landing
		changes = append(changes, "CLEARED TO LAND")
	}

	// Build new map (immutable)
	newPlanes := make(map[string]aircraft.Aircraft, len(planes))
	for k, v := range planes {
		newPlanes[k] = v
	}
	newPlanes[cmd.Callsign] = ac

	msg := fmt.Sprintf("%s: %s", cmd.Callsign, strings.Join(changes, ", "))
	return newPlanes, msg, nil
}
