package cmdtree

import (
	"fmt"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
)

// Phase represents the current stage of command input.
type Phase int

const (
	PhaseIdle     Phase = iota // No input yet
	PhaseCallsign              // Callsign entered, showing command options
	PhaseValue                 // Command chosen, showing value options
	PhaseChain                 // Value entered, showing chain options or send
)

// Option represents a single clickable/typable choice in the command tree.
type Option struct {
	Label    string // Display text (e.g., "H", "270", "Enter")
	Value    string // What gets appended to input (e.g., "H", "270", "")
	ZoneID   string // Unique ID for click detection
	IsSubmit bool   // True for the "Send" / Enter option
}

// Tree holds the current command tree state derived from input text and aircraft state.
type Tree struct {
	Phase   Phase
	Options []Option
}

// commandDef defines a command available in a given phase.
type commandDef struct {
	key   string
	label string
}

// airCommands are available for airborne aircraft (Approaching).
var airCommands = []commandDef{
	{"H", "Heading"},
	{"A", "Altitude"},
	{"S", "Speed"},
	{"L", "Land"},
	{"D", "Direct"},
	{"TL", "Turn Left"},
	{"TR", "Turn Right"},
	{"EX", "Expedite"},
}

// landingCommands are available for aircraft cleared to land.
var landingCommands = []commandDef{
	{"GA", "Go Around"},
}

// gateCommands are available for aircraft at gates.
var gateCommands = []commandDef{
	{"PB", "Pushback"},
}

// pushbackCommands are available after pushback.
var pushbackCommands = []commandDef{
	{"TX", "Taxi"},
}

// taxiCommands are available for taxiing aircraft.
var taxiCommands = []commandDef{
	{"TX", "Taxi"},
	{"HS", "Hold Short"},
	{"GATE", "Gate"},
}

// holdShortCommands are available for aircraft holding short.
var holdShortCommands = []commandDef{
	{"T", "Takeoff"},
	{"CR", "Cross Rwy"},
	{"TX", "Taxi"},
}

// Resolve examines the current input text and the selected aircraft's state
// to determine what command tree options to show.
func Resolve(inputText string, acState aircraft.State, role config.Role) Tree {
	tokens := strings.Fields(strings.TrimSpace(inputText))

	// No tokens or empty input — idle, no tree
	if len(tokens) == 0 {
		return Tree{Phase: PhaseIdle}
	}

	// One token = callsign entered, show command options based on state
	if len(tokens) == 1 {
		// Only show tree if input ends with space (user finished typing callsign)
		if !strings.HasSuffix(inputText, " ") {
			return Tree{Phase: PhaseIdle}
		}
		return Tree{
			Phase:   PhaseCallsign,
			Options: filterByRole(commandOptions(acState), role),
		}
	}

	// Two+ tokens — parse what commands have been entered
	lastToken := strings.ToUpper(tokens[len(tokens)-1])
	trailingSpace := strings.HasSuffix(inputText, " ")

	// If the last token is a command prefix without a value and has trailing space,
	// show value options for that command
	if !trailingSpace {
		// User is still typing — check if it's a bare command letter
		if len(lastToken) == 1 && isCommandPrefix(lastToken) {
			return Tree{
				Phase:   PhaseValue,
				Options: valueOptions(lastToken),
			}
		}
		// Otherwise user is mid-type, no tree
		return Tree{Phase: PhaseIdle}
	}

	// Trailing space after a completed token — determine what to show next
	if isCompleteCommand(lastToken) {
		// Last token was a no-value command (L, GA) or a command+value (H270)
		return Tree{
			Phase:   PhaseChain,
			Options: chainOptions(acState, tokens, role),
		}
	}

	// After a command prefix with value, show chain options
	if len(lastToken) >= 2 && isCommandPrefix(string(lastToken[0])) {
		return Tree{
			Phase:   PhaseChain,
			Options: chainOptions(acState, tokens, role),
		}
	}

	return Tree{Phase: PhaseIdle}
}

func isCommandPrefix(s string) bool {
	switch s {
	case "H", "A", "S":
		return true
	}
	return false
}

func isCompleteCommand(token string) bool {
	switch token {
	case "L", "GA", "T", "PB":
		return true
	}
	// Command+value like H270, A3, S2
	if len(token) >= 2 && isCommandPrefix(string(token[0])) {
		return true
	}
	// Multi-word commands with arguments are handled as complete when followed by space
	return false
}

func commandOptions(state aircraft.State) []Option {
	var defs []commandDef
	switch state {
	case aircraft.Approaching, aircraft.Departing:
		defs = airCommands
	case aircraft.Landing:
		defs = landingCommands
	case aircraft.AtGate:
		defs = gateCommands
	case aircraft.Pushback:
		defs = pushbackCommands
	case aircraft.Taxiing:
		defs = taxiCommands
	case aircraft.HoldShort:
		defs = holdShortCommands
	default:
		defs = airCommands
	}

	options := make([]Option, len(defs))
	for i, d := range defs {
		options[i] = Option{
			Label:  d.key + " " + d.label,
			Value:  d.key,
			ZoneID: "cmd_" + d.key,
		}
	}
	return options
}

func valueOptions(prefix string) []Option {
	switch prefix {
	case "H":
		return headingOptions()
	case "A":
		return altitudeOptions()
	case "S":
		return speedOptions()
	}
	return nil
}

func headingOptions() []Option {
	headings := []int{30, 60, 90, 120, 150, 180, 210, 240, 270, 300, 330, 360}
	options := make([]Option, len(headings))
	for i, h := range headings {
		val := h
		if val == 360 {
			val = 0
		}
		label := fmt.Sprintf("%03d", h)
		value := fmt.Sprintf("%03d", val)
		options[i] = Option{
			Label:  label,
			Value:  value,
			ZoneID: "hdg_" + label,
		}
	}
	return options
}

func altitudeOptions() []Option {
	alts := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	options := make([]Option, len(alts))
	for i, a := range alts {
		label := fmt.Sprintf("%d", a)
		options[i] = Option{
			Label:  label,
			Value:  label,
			ZoneID: "alt_" + label,
		}
	}
	return options
}

func speedOptions() []Option {
	speeds := []int{1, 2, 3, 4, 5}
	options := make([]Option, len(speeds))
	for i, s := range speeds {
		label := fmt.Sprintf("%d", s)
		options[i] = Option{
			Label:  label,
			Value:  label,
			ZoneID: "spd_" + label,
		}
	}
	return options
}

func chainOptions(state aircraft.State, tokens []string, role config.Role) []Option {
	// Build set of commands already used
	used := make(map[string]bool)
	for _, t := range tokens[1:] {
		upper := strings.ToUpper(t)
		if len(upper) > 0 {
			used[string(upper[0])] = true
		}
		// Also track full commands like L, GA
		used[upper] = true
	}

	var options []Option

	// Offer commands not yet used, based on aircraft state
	var defs []commandDef
	switch state {
	case aircraft.Approaching, aircraft.Departing:
		defs = airCommands
	case aircraft.Landing:
		defs = landingCommands
	case aircraft.AtGate:
		defs = gateCommands
	case aircraft.Pushback:
		defs = pushbackCommands
	case aircraft.Taxiing:
		defs = taxiCommands
	case aircraft.HoldShort:
		defs = holdShortCommands
	default:
		defs = airCommands
	}

	for _, d := range defs {
		if !used[d.key] {
			options = append(options, Option{
				Label:  d.key + " " + d.label,
				Value:  d.key,
				ZoneID: "cmd_" + d.key,
			})
		}
	}

	// Always offer Send
	options = append(options, Option{
		Label:    "Enter ↵",
		Value:    "",
		ZoneID:   "cmd_send",
		IsSubmit: true,
	})

	return filterByRole(options, role)
}

// filterByRole removes options that the current role doesn't allow.
func filterByRole(options []Option, role config.Role) []Option {
	if role == config.RoleCombined {
		return options
	}
	// TRACON: filter out ground commands. Tower: filter out D, TL, TR, EX.
	var filtered []Option
	for _, opt := range options {
		if role.IsCommandAllowed(opt.Value) || opt.IsSubmit {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}
