package game

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/collision"
	"github.com/Jasrags/atc/internal/command"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/radar"
	"github.com/Jasrags/atc/internal/runway"
	"github.com/Jasrags/atc/internal/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minWidth  = 60
	minHeight = 24
)

type screen int

const (
	screenMenu screen = iota
	screenPlaying
	screenHelp
	screenPaused
	screenGameOver
)

type menuScreen int

const (
	menuMain menuScreen = iota
	menuMapSelect
)

var mainMenuItems = []ui.MenuItem{
	{Label: "New Game", Key: "n"},
	{Label: "Help", Key: "h"},
	{Label: "Quit", Key: "q"},
}

// Model is the top-level bubbletea model for the ATC game.
type Model struct {
	screen       screen
	menuScreen   menuScreen
	menuSelected int
	maps         []gamemap.Map
	gameMap      gamemap.Map
	aircraft   map[string]aircraft.Aircraft
	runways    []runway.Runway
	spawner      *aircraft.Spawner
	input        textinput.Model
	score        int
	messages     []string
	width        int
	height       int
	elapsed      time.Duration
	startTime    time.Time
	pauseElapsed time.Duration
	started      bool
}

// NewModel creates a new model starting at the main menu.
func NewModel() Model {
	maps := gamemap.All()
	if len(maps) == 0 {
		panic("gamemap.All returned no maps")
	}
	gm := maps[0]
	return Model{
		screen:   screenMenu,
		maps:     maps,
		gameMap:  gm,
		aircraft: make(map[string]aircraft.Aircraft),
		runways:  buildRunways(gm),
	}
}

func buildRunways(gm gamemap.Map) []runway.Runway {
	runways := make([]runway.Runway, len(gm.Runways))
	for i, r := range gm.Runways {
		runways[i] = runway.New(r.X, r.Y, r.Heading, r.Length)
	}
	return runways
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickMsg:
		return m.handleTick(msg)
	}

	if m.screen == screenPlaying {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case screenMenu:
		return m.handleMenuKey(msg)
	case screenPlaying:
		return m.handlePlayingKey(msg)
	case screenHelp:
		return m.handleHelpKey(msg)
	case screenPaused:
		return m.handlePausedKey(msg)
	case screenGameOver:
		return m.handleGameOverKey(msg)
	}

	return m, nil
}

// --- Menu ---

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.menuScreen == menuMapSelect {
		return m.handleMapSelectKey(msg)
	}

	switch msg.String() {
	case "up", "k":
		if m.menuSelected > 0 {
			m.menuSelected--
		}
	case "down", "j":
		if m.menuSelected < len(mainMenuItems)-1 {
			m.menuSelected++
		}
	case "enter":
		return m.selectMainMenuItem()
	case "q", "esc":
		return m, tea.Quit
	case "n":
		m.menuScreen = menuMapSelect
		m.menuSelected = 0
	case "h", "?":
		m.screen = screenHelp
	}
	return m, nil
}

func (m Model) selectMainMenuItem() (tea.Model, tea.Cmd) {
	switch mainMenuItems[m.menuSelected].Key {
	case "n":
		m.menuScreen = menuMapSelect
		m.menuSelected = 0
		return m, nil
	case "h":
		m.screen = screenHelp
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleMapSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuSelected > 0 {
			m.menuSelected--
		}
	case "down", "j":
		if m.menuSelected < len(m.maps)-1 {
			m.menuSelected++
		}
	case "enter":
		m.gameMap = m.maps[m.menuSelected]
		return m.startGame()
	case "esc":
		m.menuScreen = menuMain
		m.menuSelected = 0
	}
	return m, nil
}

func (m Model) startGame() (Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "AA123 H270 A3 S2"
	ti.CharLimit = 40
	ti.Width = 40
	ti.Focus()

	m.screen = screenPlaying
	m.aircraft = make(map[string]aircraft.Aircraft)
	m.runways = buildRunways(m.gameMap)
	m.spawner = aircraft.NewSpawner(time.Now().UnixNano())
	m.input = ti
	m.score = 0
	m.messages = []string{ui.FormatInfo(fmt.Sprintf("Welcome to %s! Type ? for help.", m.gameMap.Name))}
	m.elapsed = 0
	m.pauseElapsed = 0
	m.started = true
	m.startTime = time.Now()

	return m, tea.Batch(tickCmd(), textinput.Blink)
}

// --- Playing ---

func (m Model) handlePlayingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
		m.menuSelected = 0
		return m, nil

	case "?":
		m.screen = screenHelp
		return m, nil

	case "p":
		m.screen = screenPaused
		m.pauseElapsed = m.elapsed
		m = m.addMessage(ui.FormatInfo("Game paused"))
		return m, nil
	}

	if msg.Type == tea.KeyEnter {
		input := strings.TrimSpace(m.input.Value())
		if input != "" {
			m = m.processCommand(input)
			m.input.Reset()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) processCommand(input string) Model {
	cmd, err := command.Parse(input)
	if err != nil {
		m = m.addMessage(ui.FormatError(err.Error()))
		return m
	}

	newPlanes, msg, err := command.Execute(cmd, m.aircraft)
	if err != nil {
		m = m.addMessage(ui.FormatError(err.Error()))
		return m
	}

	m.aircraft = newPlanes
	m = m.addMessage(ui.FormatSuccess(msg))
	return m
}

// --- Help ---

func (m Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "q":
		if m.started {
			m.screen = screenPlaying
		} else {
			m.screen = screenMenu
		}
	}
	return m, nil
}

// --- Paused ---

func (m Model) handlePausedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "p":
		m.screen = screenPlaying
		m.startTime = time.Now().Add(-m.pauseElapsed)
		m = m.addMessage(ui.FormatInfo("Game resumed"))
		return m, tickCmd()
	case "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenMenu
		m.menuSelected = 0
		return m, nil
	}
	return m, nil
}

// --- Game Over ---

func (m Model) handleGameOverKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		return m.startGame()
	case "q", "esc":
		m.screen = screenMenu
		m.menuSelected = 0
		return m, nil
	}
	return m, nil
}

// --- Tick ---

func (m Model) handleTick(msg tickMsg) (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying {
		return m, nil
	}

	if !m.started {
		m.started = true
		m.startTime = time.Time(msg)
	}

	m.elapsed = time.Time(msg).Sub(m.startTime)

	// Advance all aircraft
	newAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		next := ac.Tick()
		if !next.IsOffScreen(m.gameMap.Width, m.gameMap.Height) {
			newAircraft[k] = next
		}
	}
	m.aircraft = newAircraft

	// Check collisions
	collisions := collision.Check(m.aircraft)
	if len(collisions) > 0 {
		for _, c := range collisions {
			ac1 := m.aircraft[c.Callsign1]
			ac1.State = aircraft.Crashed
			m.aircraft[c.Callsign1] = ac1

			ac2 := m.aircraft[c.Callsign2]
			ac2.State = aircraft.Crashed
			m.aircraft[c.Callsign2] = ac2

			m = m.addMessage(
				ui.FormatError(fmt.Sprintf("COLLISION: %s and %s!", c.Callsign1, c.Callsign2)))
		}
		m.screen = screenGameOver
		return m, nil
	}

	// Check landings against all runways
	for k, ac := range m.aircraft {
		if ac.State != aircraft.Landing {
			continue
		}
		for _, rw := range m.runways {
			if rw.CanLand(ac.GridX(), ac.GridY(), ac.Heading, ac.Altitude) {
				ac.State = aircraft.Landed
				m.aircraft[k] = ac
				m.score++
				m = m.addMessage(
					ui.FormatSuccess(fmt.Sprintf("%s landed safely! Score: %d", ac.Callsign, m.score)))
				break
			}
		}
	}

	// Remove landed aircraft
	for k, ac := range m.aircraft {
		if ac.State == aircraft.Landed {
			delete(m.aircraft, k)
		}
	}

	// Spawn new aircraft
	if m.spawner.ShouldSpawn(m.elapsed, len(m.aircraft)) {
		ac := m.spawner.Spawn(m.gameMap.Width, m.gameMap.Height)
		if _, exists := m.aircraft[ac.Callsign]; !exists {
			m.aircraft[ac.Callsign] = ac
			m = m.addMessage(ui.FormatInfo(ac.Callsign + " entering airspace"))
		}
	}

	return m, tickCmd()
}

// --- View ---

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	if m.width < minWidth || m.height < minHeight {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.MessageError.Render("Terminal too small. Need at least 60x24."))
	}

	switch m.screen {
	case screenMenu:
		var menuView string
		if m.menuScreen == menuMapSelect {
			mapItems := make([]ui.MenuItem, len(m.maps))
			for i, gm := range m.maps {
				mapItems[i] = ui.MenuItem{Label: gm.Name, Key: gm.ID}
			}
			menuView = ui.RenderMapSelect(mapItems, m.menuSelected)
		} else {
			menuView = ui.RenderMenu(mainMenuItems, m.menuSelected)
		}
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			menuView)

	case screenHelp:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.RenderHelp())

	case screenPaused:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.RenderPaused(m.score, m.elapsed))

	case screenGameOver:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.RenderGameOver(m.score))

	case screenPlaying:
		return m.renderPlaying()
	}

	return ""
}

func (m Model) renderPlaying() string {
	planes := m.sortedAircraft()

	hud := ui.RenderHUD(m.score, len(planes), m.elapsed, m.messages)
	radarView := radar.Render(m.gameMap, planes)
	sidebar := radar.RenderSidebar(planes)
	gameArea := lipgloss.JoinHorizontal(lipgloss.Top, radarView, " ", sidebar)

	prompt := ui.InputPrompt.Render("ATC> ")
	inputView := prompt + m.input.View()

	return lipgloss.JoinVertical(lipgloss.Left, hud, gameArea, "", inputView)
}

const maxStoredMessages = 50

func (m Model) addMessage(msg string) Model {
	m.messages = append(m.messages, msg)
	if len(m.messages) > maxStoredMessages {
		m.messages = m.messages[len(m.messages)-maxStoredMessages:]
	}
	return m
}

func (m Model) sortedAircraft() []aircraft.Aircraft {
	planes := make([]aircraft.Aircraft, 0, len(m.aircraft))
	for _, ac := range m.aircraft {
		planes = append(planes, ac)
	}
	sort.Slice(planes, func(i, j int) bool {
		return planes[i].Callsign < planes[j].Callsign
	})
	return planes
}
