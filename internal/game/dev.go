package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/radio"
)

// processDevCommand handles / prefixed developer commands.
// Returns the updated model with feedback in the radio log.
func (m Model) processDevCommand(input string) Model {
	if !m.devMode {
		m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
			"dev commands require --dev flag", radio.Normal))
		return m
	}

	args := strings.Fields(input)
	if len(args) == 0 {
		return m
	}

	cmd := strings.ToLower(args[0])
	switch cmd {
	case "help":
		return m.devHelp()
	case "spawn":
		return m.devSpawn(args[1:])
	case "clear":
		return m.devClear()
	case "god":
		return m.devGod()
	case "pause":
		return m.devPause()
	case "speed":
		return m.devSpeed(args[1:])
	default:
		m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
			fmt.Sprintf("unknown dev command: /%s (try /help)", cmd), radio.Normal))
		return m
	}
}

func (m Model) devHelp() Model {
	elapsed := m.stopwatch.Elapsed()
	m = m.addRadio(radio.SystemMessage(elapsed, "--- DEV COMMANDS ---", radio.Normal))
	m = m.addRadio(radio.SystemMessage(elapsed, "/spawn        — spawn arrival", radio.Normal))
	m = m.addRadio(radio.SystemMessage(elapsed, "/spawn dep    — spawn departure", radio.Normal))
	m = m.addRadio(radio.SystemMessage(elapsed, "/clear        — remove all aircraft", radio.Normal))
	m = m.addRadio(radio.SystemMessage(elapsed, "/god          — toggle god mode", radio.Normal))
	m = m.addRadio(radio.SystemMessage(elapsed, "/pause        — toggle spawner", radio.Normal))
	m = m.addRadio(radio.SystemMessage(elapsed, "/speed <1-5>  — set game speed", radio.Normal))
	return m
}

func (m Model) devSpawn(args []string) Model {
	elapsed := m.stopwatch.Elapsed()

	if len(args) > 0 && strings.ToLower(args[0]) == "dep" {
		m = m.trySpawnDeparture(elapsed)
		return m
	}

	ac := m.spawner.Spawn(m.gameMap.Width, m.gameMap.Height)
	if _, exists := m.aircraft[ac.Callsign]; exists {
		m = m.addRadio(radio.SystemMessage(elapsed,
			fmt.Sprintf("callsign collision: %s already exists", ac.Callsign), radio.Normal))
		return m
	}
	m.aircraft[ac.Callsign] = ac
	m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign,
		radio.FormatEnteringAirspace(ac.Callsign, ac.Heading, ac.Altitude)))
	return m
}

func (m Model) devClear() Model {
	count := len(m.aircraft)
	m.aircraft = make(map[string]aircraft.Aircraft)
	m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
		fmt.Sprintf("cleared %d aircraft", count), radio.Normal))
	return m
}

func (m Model) devGod() Model {
	m.godMode = !m.godMode
	status := "OFF"
	if m.godMode {
		status = "ON"
	}
	m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
		fmt.Sprintf("god mode: %s", status), radio.Normal))
	return m
}

func (m Model) devPause() Model {
	m.spawnerPaused = !m.spawnerPaused
	status := "resumed"
	if m.spawnerPaused {
		status = "paused"
	}
	m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
		fmt.Sprintf("spawner %s", status), radio.Normal))
	return m
}

func (m Model) devSpeed(args []string) Model {
	if len(args) == 0 {
		m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
			fmt.Sprintf("speed: %dx (usage: /speed <1-5>)", m.speedMultiplier), radio.Normal))
		return m
	}

	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 || n > 5 {
		m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
			"speed must be 1-5", radio.Normal))
		return m
	}

	m.speedMultiplier = n
	m = m.addRadio(radio.SystemMessage(m.stopwatch.Elapsed(),
		fmt.Sprintf("speed: %dx", n), radio.Normal))
	return m
}

// DevStatus returns a short status string for the HUD.
func (m Model) DevStatus() string {
	if !m.devMode {
		return ""
	}
	var parts []string
	parts = append(parts, "DEV")
	if m.godMode {
		parts = append(parts, "GOD")
	}
	if m.spawnerPaused {
		parts = append(parts, "NOSPAWN")
	}
	if m.speedMultiplier != 1 {
		parts = append(parts, fmt.Sprintf("%dx", m.speedMultiplier))
	}
	return "[" + strings.Join(parts, " ") + "]"
}
