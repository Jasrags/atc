package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Radar display
	RadarBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))

	// Aircraft symbols
	AircraftNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	AircraftSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true)

	AircraftLanding = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	AircraftCrashed = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Runway
	RunwayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	// HUD
	HUDTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("63")).
			PaddingLeft(1).PaddingRight(1)

	HUDScore = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))

	HUDInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Messages
	MessageSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	MessageError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	MessageInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	// Command input
	InputPrompt = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true)

	// Help overlay
	HelpBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)

	HelpTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))

	HelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true)

	HelpDesc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Game over
	GameOverTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			PaddingLeft(2).PaddingRight(2)

	GameOverScore = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))

	// Sidebar
	SidebarBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	SidebarTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))

	Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)
