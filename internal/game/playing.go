package game

import (
	"fmt"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/cmdtree"
	"github.com/Jasrags/atc/internal/command"
	"github.com/Jasrags/atc/internal/radio"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

// --- Playing ---

func (m Model) handlePlayingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Esc always works regardless of input state.
	if key.Matches(msg, m.keys.Back) {
		m.screen = screenMenu
		m.menuSelected = 0
		return m, m.stopwatch.Stop()
	}

	// Submit always works — process the command if input is non-empty.
	if key.Matches(msg, m.keys.Submit) {
		input := strings.TrimSpace(m.input.Value())
		if input != "" {
			if strings.HasPrefix(input, "/") {
				m = m.processDevCommand(strings.TrimPrefix(input, "/"))
			} else {
				m = m.processCommand(input)
			}
			m.input.Reset()
			m.cmdTree = cmdtree.Tree{}
		}
		return m, nil
	}

	// Single-char shortcuts only activate when the input is empty.
	// Otherwise they'd swallow letters the user is typing into the ATC prompt.
	if m.input.Value() == "" {
		switch {
		case key.Matches(msg, m.keys.Help):
			m.screen = screenHelp
			return m, nil
		case key.Matches(msg, m.keys.Freeze):
			return m.toggleFreeze()
		case key.Matches(msg, m.keys.SpeedUp):
			return m.changeSpeed(1), nil
		case key.Matches(msg, m.keys.SpeedDn):
			return m.changeSpeed(-1), nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m = m.resolveTree()
	return m, cmd
}

func (m Model) processCommand(input string) Model {
	elapsed := m.stopwatch.Elapsed()

	cmd, err := command.Parse(input)
	if err != nil {
		m = m.addRadio(radio.SystemMessage(elapsed, err.Error(), radio.Normal))
		return m
	}

	newPlanes, changes, err := command.Execute(cmd, m.aircraft, m.gameConfig.Role)
	if err != nil {
		m = m.addRadio(radio.SystemMessage(elapsed, err.Error(), radio.Normal))
		return m
	}

	// Reset patience timer — controller gave instructions
	if ac, exists := newPlanes[cmd.Callsign]; exists {
		ac = ac.ResetPatience()
		newPlanes[cmd.Callsign] = ac
	}

	// Resolve direct-to-fix: look up fix position from the map
	if cmd.DirectFix != "" {
		ac := newPlanes[cmd.Callsign]
		found := false
		for _, fix := range m.gameMap.Fixes {
			if fix.Name == cmd.DirectFix {
				ac.TargetFixX = float64(fix.X)
				ac.TargetFixY = float64(fix.Y)
				newPlanes[cmd.Callsign] = ac
				found = true
				break
			}
		}
		if !found {
			m = m.addRadio(radio.SystemMessage(elapsed,
				fmt.Sprintf("unknown fix: %s", cmd.DirectFix), radio.Normal))
			ac.TargetFixName = ""
			newPlanes[cmd.Callsign] = ac
		}
	}

	// Resolve taxi route into a path if TX command was issued
	if len(cmd.TaxiRoute) > 0 {
		ac := newPlanes[cmd.Callsign]
		startNodeID := m.findNearestNode(ac)
		if startNodeID == "" {
			m = m.addRadio(radio.SystemMessage(elapsed, "no taxi nodes available", radio.Normal))
		} else {
			path, routeErr := m.gameMap.ResolveTaxiRoute(startNodeID, cmd.TaxiRoute)
			if routeErr != nil {
				m = m.addRadio(radio.SystemMessage(elapsed,
					fmt.Sprintf("invalid taxi route: %s", routeErr), radio.Normal))
				ac.State = aircraft.Pushback // revert to previous ground state
				ac.TaxiRoute = nil
			} else {
				positions := m.nodeIDsToPositions(path)
				ac.TaxiPath = positions
				ac.TaxiPathIndex = 0
			}
			newPlanes[cmd.Callsign] = ac
		}
	}

	// Resolve gate assignment into a taxi path
	if cmd.AssignGate != "" {
		ac := newPlanes[cmd.Callsign]
		gate := m.gameMap.GateByID(cmd.AssignGate)
		if gate == nil {
			m = m.addRadio(radio.SystemMessage(elapsed,
				fmt.Sprintf("unknown gate: %s", cmd.AssignGate), radio.Normal))
		} else {
			gateNode := m.gameMap.NodeByID(gate.NodeID)
			if gateNode == nil {
				m = m.addRadio(radio.SystemMessage(elapsed, "gate position not found", radio.Normal))
			} else {
				ac.TaxiPath = [][2]int{{ac.GridX(), ac.GridY()}, {gateNode.X, gateNode.Y}}
				ac.TaxiPathIndex = 0
				newPlanes[cmd.Callsign] = ac
			}
		}
	}

	m.aircraft = newPlanes
	m = m.addRadio(radio.CommandPhraseology(elapsed, cmd.Callsign, changes))
	return m
}

// --- Time Control ---

const (
	minSpeed = 1
	maxSpeed = 12
)

// toggleFreeze pauses or resumes the physics simulation while keeping the
// playing screen visible. The player can still type and submit commands.
func (m Model) toggleFreeze() (tea.Model, tea.Cmd) {
	m.timeFrozen = !m.timeFrozen
	if m.timeFrozen {
		return m, m.stopwatch.Stop()
	}
	return m, tea.Batch(tickCmd(), m.stopwatch.Start())
}

// changeSpeed adjusts the speed multiplier by delta, clamping to [1, 12].
func (m Model) changeSpeed(delta int) Model {
	m.speedMultiplier += delta
	if m.speedMultiplier < minSpeed {
		m.speedMultiplier = minSpeed
	}
	if m.speedMultiplier > maxSpeed {
		m.speedMultiplier = maxSpeed
	}
	return m
}

// --- Mouse ---

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying {
		return m, nil
	}
	if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	// Check if click is on a command tree option
	for _, opt := range m.cmdTree.Options {
		if zone.Get(opt.ZoneID).InBounds(msg) {
			if opt.IsSubmit {
				input := strings.TrimSpace(m.input.Value())
				if input != "" {
					m = m.processCommand(input)
					m.input.Reset()
					m.cmdTree = cmdtree.Tree{}
				}
			} else {
				m = m.appendToInput(opt.Value)
			}
			return m, nil
		}
	}

	// Check if click is on a flight strip zone
	for callsign := range m.aircraft {
		if zone.Get(callsign).InBounds(msg) {
			m.input.SetValue(callsign + " ")
			m.input.CursorEnd()
			m = m.resolveTree()
			return m, nil
		}
	}

	return m, nil
}

// appendToInput adds a value to the input field and resolves the tree.
func (m Model) appendToInput(value string) Model {
	current := m.input.Value()
	// If the last token is a bare command prefix (H, A, S), append value directly
	// Otherwise add a space before the value
	trimmed := strings.TrimRight(current, " ")
	tokens := strings.Fields(trimmed)
	if len(tokens) > 0 {
		lastToken := strings.ToUpper(tokens[len(tokens)-1])
		if len(lastToken) == 1 && (lastToken == "H" || lastToken == "A" || lastToken == "S") {
			// Append directly to the command prefix (e.g., "H" + "270" = "H270")
			m.input.SetValue(trimmed + value + " ")
			m.input.CursorEnd()
			m = m.resolveTree()
			return m
		}
	}

	// Otherwise append with space
	if !strings.HasSuffix(current, " ") && current != "" {
		current += " "
	}
	m.input.SetValue(current + value + " ")
	m.input.CursorEnd()
	m = m.resolveTree()
	return m
}

// resolveTree updates the command tree based on the current input text.
func (m Model) resolveTree() Model {
	inputText := m.input.Value()
	tokens := strings.Fields(inputText)
	acState := aircraft.Approaching // default

	if len(tokens) > 0 {
		callsign := strings.ToUpper(tokens[0])
		if ac, exists := m.aircraft[callsign]; exists {
			acState = ac.State
		}
	}

	m.cmdTree = cmdtree.Resolve(inputText, acState, m.gameConfig.Role)
	return m
}
