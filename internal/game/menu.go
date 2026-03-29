package game

import (
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Menu ---

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.menuScreen == menuSetup {
		return m.handleSetupKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.menuSelected > 0 {
			m.menuSelected--
		}
	case key.Matches(msg, m.keys.Down):
		if m.menuSelected < len(mainMenuItems)-1 {
			m.menuSelected++
		}
	case key.Matches(msg, m.keys.Submit):
		return m.selectMainMenuItem()
	case key.Matches(msg, m.keys.Quit, m.keys.Back):
		return m, tea.Quit
	case key.Matches(msg, m.keys.NewGame):
		m.menuScreen = menuSetup
		m.setupFocus = 0
	case key.Matches(msg, m.keys.HelpKey, m.keys.Help):
		m.screen = screenHelp
	}
	return m, nil
}

func (m Model) selectMainMenuItem() (tea.Model, tea.Cmd) {
	switch mainMenuItems[m.menuSelected].Key {
	case "n":
		m.menuScreen = menuSetup
		m.setupFocus = 0
		return m, nil
	case "h":
		m.screen = screenHelp
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

// --- Setup ---

func (m Model) setupSectionMax(section int) int {
	switch section {
	case setupMap:
		return len(m.maps) - 1
	case setupRole:
		return len(config.RoleOptions()) - 1
	case setupDiff:
		return len(config.DifficultyOptions()) - 1
	case setupMode:
		return len(config.GameModeOptions()) - 1
	case setupCallsign:
		return len(config.CallsignStyleOptions()) - 1
	case setupTrails:
		return len(config.PlaneTrailsOptions()) - 1
	}
	return 0
}

func (m Model) handleSetupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Tab):
		m.setupFocus = (m.setupFocus + 1) % setupSections
	case key.Matches(msg, m.keys.STab):
		m.setupFocus = (m.setupFocus - 1 + setupSections) % setupSections
	case key.Matches(msg, m.keys.Up):
		if m.setupSelections[m.setupFocus] > 0 {
			m.setupSelections[m.setupFocus]--
		}
	case key.Matches(msg, m.keys.Down):
		max := m.setupSectionMax(m.setupFocus)
		if m.setupSelections[m.setupFocus] < max {
			m.setupSelections[m.setupFocus]++
		}
	case key.Matches(msg, m.keys.Submit):
		m.gameConfig = m.buildConfigFromSetup()
		m.gameMap = gamemap.ByID(m.gameConfig.MapID)
		return m.startGame()
	case key.Matches(msg, m.keys.Back):
		m.menuScreen = menuMain
		m.menuSelected = 0
	}
	return m, nil
}

func (m Model) buildConfigFromSetup() config.GameConfig {
	return config.GameConfig{
		MapID:         m.maps[m.setupSelections[setupMap]].ID,
		Role:          config.RoleFromIndex(m.setupSelections[setupRole]),
		Difficulty:    config.Difficulty(m.setupSelections[setupDiff]),
		GameMode:      config.GameMode(m.setupSelections[setupMode]),
		CallsignStyle: config.CallsignStyle(m.setupSelections[setupCallsign]),
		PlaneTrails:   m.setupSelections[setupTrails] == 0, // 0=On, 1=Off
	}
}

func (m Model) buildSetupSections() []ui.SetupSection {
	mapNames := make([]string, len(m.maps))
	for i, gm := range m.maps {
		mapNames[i] = gm.Name
	}

	return []ui.SetupSection{
		{Title: "Map", Options: mapNames, Selected: m.setupSelections[setupMap]},
		{Title: "Role", Options: config.RoleOptions(), Selected: m.setupSelections[setupRole]},
		{Title: "Difficulty", Options: config.DifficultyOptions(), Selected: m.setupSelections[setupDiff]},
		{Title: "Game Mode", Options: config.GameModeOptions(), Selected: m.setupSelections[setupMode]},
		{Title: "Callsign Style", Options: config.CallsignStyleOptions(), Selected: m.setupSelections[setupCallsign]},
		{Title: "Plane Trails", Options: config.PlaneTrailsOptions(), Selected: m.setupSelections[setupTrails]},
	}
}

// --- Help ---

func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help, m.keys.Back, m.keys.Quit):
		if m.started {
			m.screen = screenPlaying
		} else {
			m.screen = screenMenu
		}
	}
	return m, nil
}

// --- Game Over ---

func (m Model) handleGameOverKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Restart):
		return m.startGame()
	case key.Matches(msg, m.keys.Quit, m.keys.Back):
		m.screen = screenMenu
		m.menuSelected = 0
		return m, nil
	}
	return m, nil
}
