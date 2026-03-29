package game

import "github.com/charmbracelet/bubbles/key"

// keyMap defines all keybindings for the game, implementing help.KeyMap.
type keyMap struct {
	// Playing
	Submit  key.Binding
	Freeze  key.Binding
	SpeedUp key.Binding
	SpeedDn key.Binding
	Help    key.Binding
	Back    key.Binding

	// Navigation
	Up   key.Binding
	Down key.Binding
	Tab  key.Binding
	STab key.Binding

	// Menu
	NewGame key.Binding
	HelpKey key.Binding
	Quit    key.Binding

	// Game over
	Restart key.Binding

	// Global
	ForceQuit key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit command"),
		),
		Freeze: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "freeze / resume time"),
		),
		SpeedUp: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "speed up"),
		),
		SpeedDn: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "slow down"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back / menu"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next section"),
		),
		STab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev section"),
		),
		NewGame: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new game"),
		),
		HelpKey: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
	}
}

// ShortHelp returns the short help bindings shown at the bottom of the screen.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Freeze, k.SpeedUp, k.SpeedDn, k.Help, k.Back, k.Quit}
}

// FullHelp returns the full help bindings grouped by section.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Freeze, k.SpeedUp, k.SpeedDn, k.Help},
		{k.Up, k.Down, k.Tab},
		{k.Back, k.Restart, k.Quit, k.ForceQuit},
	}
}
