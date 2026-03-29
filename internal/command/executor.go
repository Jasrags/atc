package command

import (
	"fmt"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
)

// Execute applies a parsed command to the aircraft map, returning a new map and list of changes applied.
// The role parameter filters which commands are allowed.
func Execute(cmd Command, planes map[string]aircraft.Aircraft, role config.Role) (map[string]aircraft.Aircraft, []string, error) {
	// Role-based command filtering
	if err := checkRolePermissions(cmd, role); err != nil {
		return planes, nil, err
	}

	ac, exists := planes[cmd.Callsign]
	if !exists {
		return planes, nil, fmt.Errorf("unknown aircraft: %s", cmd.Callsign)
	}

	if ac.State == aircraft.Crashed {
		return planes, nil, fmt.Errorf("%s has crashed", cmd.Callsign)
	}

	if ac.State == aircraft.Landed {
		// Only gate assignment is valid for landed aircraft
		if cmd.AssignGate == "" {
			return planes, nil, fmt.Errorf("%s has already landed", cmd.Callsign)
		}
	}

	var changes []string
	var err error

	// Airborne commands
	if cmd.Heading != nil {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.TargetHeading = *cmd.Heading
		ac.TargetFixName = ""  // cancel direct-to-fix
		ac.HoldingFixName = "" // cancel hold
		ac.ForceTurnDir = 0    // reset to shortest path
		changes = append(changes, fmt.Sprintf("HDG %03d", *cmd.Heading))
	}

	if cmd.Altitude != nil {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.TargetAltitude = *cmd.Altitude
		changes = append(changes, fmt.Sprintf("ALT %d", *cmd.Altitude))
	}

	if cmd.Speed != nil {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.TargetSpeed = *cmd.Speed
		changes = append(changes, fmt.Sprintf("SPD %d", *cmd.Speed))
	}

	if cmd.TurnLeft != nil {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.TargetHeading = *cmd.TurnLeft
		ac.ForceTurnDir = aircraft.ForceTurnLeft
		ac.TargetFixName = ""  // cancel direct-to-fix
		ac.HoldingFixName = "" // cancel hold
		changes = append(changes, fmt.Sprintf("TURN LEFT HDG %03d", *cmd.TurnLeft))
	}

	if cmd.TurnRight != nil {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.TargetHeading = *cmd.TurnRight
		ac.ForceTurnDir = aircraft.ForceTurnRight
		ac.TargetFixName = ""  // cancel direct-to-fix
		ac.HoldingFixName = "" // cancel hold
		changes = append(changes, fmt.Sprintf("TURN RIGHT HDG %03d", *cmd.TurnRight))
	}

	if cmd.DirectFix != "" {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		// Fix position is resolved in the game model (executor has no map access)
		ac.TargetFixName = cmd.DirectFix
		ac.ForceTurnDir = 0
		// Clear any active hold — direct overrides hold
		ac.HoldingFixName = ""
		changes = append(changes, fmt.Sprintf("DIRECT %s", cmd.DirectFix))
	}

	if cmd.HoldFix != "" {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		// Fix position is resolved in the game model (executor has no map access)
		ac.HoldingFixName = cmd.HoldFix
		ac.HoldingPhase = 0 // start inbound
		ac.HoldingTicks = 0
		// Also set as navigation target to fly to the fix first
		ac.TargetFixName = cmd.HoldFix
		ac.ForceTurnDir = 0
		changes = append(changes, fmt.Sprintf("HOLD AT %s", cmd.HoldFix))
	}

	if cmd.Expedite {
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.ExpeditedAlt = true
		changes = append(changes, "EXPEDITE")
	}

	if cmd.ClearToLand {
		if ac.State == aircraft.Landing {
			return planes, nil, fmt.Errorf("%s already cleared to land", cmd.Callsign)
		}
		if err = requireAirborne(ac); err != nil {
			return planes, nil, err
		}
		ac.State = aircraft.Landing
		ac.AssignedLandingRunway = cmd.LandRunway
		if cmd.LandRunway != "" {
			changes = append(changes, fmt.Sprintf("CLEARED TO LAND RWY %s", cmd.LandRunway))
		} else {
			changes = append(changes, "CLEARED TO LAND")
		}
	}

	if cmd.GoAround {
		if ac.State != aircraft.Landing {
			return planes, nil, fmt.Errorf("%s is not on approach", cmd.Callsign)
		}
		ac.State = aircraft.Approaching
		ac.TargetAltitude = 3 // climb to 3000
		changes = append(changes, "GO AROUND")
	}

	// Ground commands
	if cmd.Takeoff {
		if ac.State != aircraft.HoldShort && ac.State != aircraft.OnRunway {
			return planes, nil, fmt.Errorf("%s not in position for takeoff", cmd.Callsign)
		}
		ac = ac.ResetTickCount()
		ac.State = aircraft.OnRunway
		ac.TaxiPath = nil
		ac.TaxiRoute = nil
		changes = append(changes, "CLEARED FOR TAKEOFF")
	}

	if cmd.PushbackApproved {
		if ac.State != aircraft.AtGate {
			return planes, nil, fmt.Errorf("%s is not at a gate", cmd.Callsign)
		}
		ac.State = aircraft.Pushback
		if cmd.ExpectRunway != "" {
			ac.AssignedRunway = cmd.ExpectRunway
			changes = append(changes, fmt.Sprintf("PUSHBACK APPROVED, EXPECT RWY %s", cmd.ExpectRunway))
		} else {
			changes = append(changes, "PUSHBACK APPROVED")
		}
	}

	if len(cmd.TaxiRoute) > 0 {
		if !ac.State.IsGround() {
			return planes, nil, fmt.Errorf("%s is not on the ground", cmd.Callsign)
		}
		if ac.State == aircraft.AtGate {
			return planes, nil, fmt.Errorf("%s must pushback first", cmd.Callsign)
		}
		ac.TaxiRoute = cmd.TaxiRoute
		ac.State = aircraft.Taxiing
		changes = append(changes, fmt.Sprintf("TAXI VIA %s", strings.Join(cmd.TaxiRoute, " ")))
	}

	if cmd.HoldShort != "" {
		if !ac.State.IsGround() {
			return planes, nil, fmt.Errorf("%s is not on the ground", cmd.Callsign)
		}
		ac.State = aircraft.HoldShort
		changes = append(changes, fmt.Sprintf("HOLD SHORT RWY %s", cmd.HoldShort))
	}

	if cmd.CrossRunway != "" {
		if ac.State != aircraft.HoldShort {
			return planes, nil, fmt.Errorf("%s is not holding short", cmd.Callsign)
		}
		ac.State = aircraft.Taxiing
		changes = append(changes, fmt.Sprintf("CROSS RWY %s", cmd.CrossRunway))
	}

	if cmd.AssignGate != "" {
		if !ac.State.IsGround() {
			return planes, nil, fmt.Errorf("%s is not on the ground", cmd.Callsign)
		}
		ac.AssignedGate = cmd.AssignGate
		ac.State = aircraft.Taxiing
		changes = append(changes, fmt.Sprintf("TAXI TO GATE %s", cmd.AssignGate))
	}

	// Build new map (immutable)
	newPlanes := make(map[string]aircraft.Aircraft, len(planes))
	for k, v := range planes {
		newPlanes[k] = v
	}
	newPlanes[cmd.Callsign] = ac

	return newPlanes, changes, nil
}

// checkRolePermissions validates that the command is allowed for the current role.
func checkRolePermissions(cmd Command, role config.Role) error {
	if role == config.RoleTRACON {
		if cmd.Takeoff {
			return fmt.Errorf("takeoff commands not available in TRACON mode")
		}
		if cmd.PushbackApproved {
			return fmt.Errorf("pushback commands not available in TRACON mode")
		}
		if len(cmd.TaxiRoute) > 0 {
			return fmt.Errorf("taxi commands not available in TRACON mode")
		}
		if cmd.HoldShort != "" {
			return fmt.Errorf("hold short commands not available in TRACON mode")
		}
		if cmd.CrossRunway != "" {
			return fmt.Errorf("cross runway commands not available in TRACON mode")
		}
		if cmd.AssignGate != "" {
			return fmt.Errorf("gate commands not available in TRACON mode")
		}
	}
	if role == config.RoleTower {
		if cmd.DirectFix != "" {
			return fmt.Errorf("direct-to-fix not available in Tower mode")
		}
		if cmd.HoldFix != "" {
			return fmt.Errorf("hold at fix not available in Tower mode")
		}
		if cmd.TurnLeft != nil {
			return fmt.Errorf("turn left not available in Tower mode")
		}
		if cmd.TurnRight != nil {
			return fmt.Errorf("turn right not available in Tower mode")
		}
		if cmd.Expedite {
			return fmt.Errorf("expedite not available in Tower mode")
		}
	}
	return nil
}

func requireAirborne(ac aircraft.Aircraft) error {
	if !ac.State.IsAirborne() {
		return fmt.Errorf("%s is not airborne", ac.Callsign)
	}
	return nil
}
