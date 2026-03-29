package command

import (
	"fmt"
	"strconv"
	"strings"
)

// Command represents a parsed player command.
type Command struct {
	Callsign    string
	Heading     *int
	Altitude    *int
	Speed       *int
	ClearToLand bool

	// Ground operations
	Takeoff          bool     // T — cleared for takeoff
	PushbackApproved bool     // PB — pushback approved
	ExpectRunway     string   // runway after PB (optional, e.g., "PB 27")
	TaxiRoute        []string // TX A B C1 — taxi via named taxiways
	HoldShort        string   // HS 27 — hold short of runway
	CrossRunway      string   // CR 27 — cleared to cross runway
	AssignGate       string   // GATE G1 — taxi to gate
	GoAround         bool     // GA — abort landing

	// Expanded commands
	DirectFix    string // D <fix> — direct to named waypoint
	TurnLeft     *int   // TL <heading> — turn left to heading
	TurnRight    *int   // TR <heading> — turn right to heading
	Expedite     bool   // EX — double altitude change rate
	LandRunway   string // L <runway> — clear to land on specific runway
}

// Parse converts a raw input string into a Command.
// Format: <CALLSIGN> <CMD1> [<CMD2>] [<CMD3>]
// Commands: H<0-359>, A<1-40>, S<1-5>, L, T, PB [runway], TX <taxiway...>,
//
//	HS <runway>, CR <runway>, GATE <gate>, GA
func Parse(input string) (Command, error) {
	tokens := strings.Fields(strings.TrimSpace(input))
	if len(tokens) == 0 {
		return Command{}, fmt.Errorf("empty command")
	}

	cmd := Command{
		Callsign: strings.ToUpper(tokens[0]),
	}

	if len(tokens) < 2 {
		return Command{}, fmt.Errorf("missing command after callsign %s", cmd.Callsign)
	}

	i := 1
	for i < len(tokens) {
		token := strings.ToUpper(tokens[i])

		switch token {
		case "L":
			cmd.ClearToLand = true
			i++
			// Optional runway argument (e.g., "L 28R")
			if i < len(tokens) && !isCommand(tokens[i]) {
				cmd.LandRunway = strings.ToUpper(tokens[i])
				i++
			}
			continue
		case "T":
			cmd.Takeoff = true
			i++
			continue
		case "GA":
			cmd.GoAround = true
			i++
			continue
		case "EX":
			cmd.Expedite = true
			i++
			continue
		case "D":
			i++
			if i >= len(tokens) {
				return Command{}, fmt.Errorf("D requires fix name (e.g., D MAFAN)")
			}
			cmd.DirectFix = strings.ToUpper(tokens[i])
			i++
			continue
		case "TL":
			i++
			if i >= len(tokens) {
				return Command{}, fmt.Errorf("TL requires heading (e.g., TL 270)")
			}
			v, err := strconv.Atoi(tokens[i])
			if err != nil {
				return Command{}, fmt.Errorf("invalid heading: TL %s", tokens[i])
			}
			if v < 0 || v > 359 {
				return Command{}, fmt.Errorf("heading must be 0-359, got %d", v)
			}
			cmd.TurnLeft = &v
			i++
			continue
		case "TR":
			i++
			if i >= len(tokens) {
				return Command{}, fmt.Errorf("TR requires heading (e.g., TR 270)")
			}
			v, err := strconv.Atoi(tokens[i])
			if err != nil {
				return Command{}, fmt.Errorf("invalid heading: TR %s", tokens[i])
			}
			if v < 0 || v > 359 {
				return Command{}, fmt.Errorf("heading must be 0-359, got %d", v)
			}
			cmd.TurnRight = &v
			i++
			continue
		case "PB":
			cmd.PushbackApproved = true
			i++
			// Optional runway argument
			if i < len(tokens) && !isCommand(tokens[i]) {
				cmd.ExpectRunway = strings.ToUpper(tokens[i])
				i++
			}
			continue
		case "HS":
			i++
			if i >= len(tokens) {
				return Command{}, fmt.Errorf("HS requires runway (e.g., HS 27)")
			}
			cmd.HoldShort = strings.ToUpper(tokens[i])
			i++
			continue
		case "CR":
			i++
			if i >= len(tokens) {
				return Command{}, fmt.Errorf("CR requires runway (e.g., CR 27)")
			}
			cmd.CrossRunway = strings.ToUpper(tokens[i])
			i++
			continue
		case "GATE":
			i++
			if i >= len(tokens) {
				return Command{}, fmt.Errorf("GATE requires gate ID (e.g., GATE G1)")
			}
			cmd.AssignGate = strings.ToUpper(tokens[i])
			i++
			continue
		case "TX":
			i++
			// Consume all remaining tokens as taxiway route
			for i < len(tokens) {
				cmd.TaxiRoute = append(cmd.TaxiRoute, strings.ToUpper(tokens[i]))
				i++
			}
			if len(cmd.TaxiRoute) == 0 {
				return Command{}, fmt.Errorf("TX requires taxiway names (e.g., TX A B C1)")
			}
			continue
		}

		if len(token) < 2 {
			return Command{}, fmt.Errorf("unknown command: %s", token)
		}

		prefix := token[0]
		valueStr := token[1:]

		switch prefix {
		case 'H':
			v, err := strconv.Atoi(valueStr)
			if err != nil {
				return Command{}, fmt.Errorf("invalid heading: %s", token)
			}
			if v < 0 || v > 359 {
				return Command{}, fmt.Errorf("heading must be 0-359, got %d", v)
			}
			cmd.Heading = &v

		case 'A':
			v, err := strconv.Atoi(valueStr)
			if err != nil {
				return Command{}, fmt.Errorf("invalid altitude: %s", token)
			}
			if v < 1 || v > 40 {
				return Command{}, fmt.Errorf("altitude must be 1-40, got %d", v)
			}
			cmd.Altitude = &v

		case 'S':
			v, err := strconv.Atoi(valueStr)
			if err != nil {
				return Command{}, fmt.Errorf("invalid speed: %s", token)
			}
			if v < 1 || v > 5 {
				return Command{}, fmt.Errorf("speed must be 1-5, got %d", v)
			}
			cmd.Speed = &v

		default:
			return Command{}, fmt.Errorf("unknown command: %s", token)
		}
		i++
	}

	return cmd, nil
}

// isCommand checks if a token is a known command keyword.
func isCommand(token string) bool {
	upper := strings.ToUpper(token)
	switch upper {
	case "L", "T", "GA", "PB", "HS", "CR", "GATE", "TX", "D", "TL", "TR", "EX":
		return true
	}
	if len(upper) >= 2 {
		switch upper[0] {
		case 'H', 'A', 'S':
			return true
		}
	}
	return false
}
