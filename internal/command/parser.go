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
}

// Parse converts a raw input string into a Command.
// Format: <CALLSIGN> <CMD1> [<CMD2>] [<CMD3>]
// Commands: H<0-359>, A<1-40>, S<1-5>, L
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

	for _, token := range tokens[1:] {
		token = strings.ToUpper(token)

		if token == "L" {
			cmd.ClearToLand = true
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
	}

	return cmd, nil
}
