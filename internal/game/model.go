package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/cmdtree"
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/radar"
	"github.com/Jasrags/atc/internal/radio"
	"github.com/Jasrags/atc/internal/runway"
	"github.com/Jasrags/atc/internal/ui"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

const (
	minWidth            = 60
	minHeight           = 24
	radioViewportHeight = 6
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
	menuSetup
)

// Setup section indices
const (
	setupMap       = 0
	setupRole      = 1
	setupDiff      = 2
	setupMode      = 3
	setupCallsign  = 4
	setupTrails    = 5
	setupSections  = 6
)

var mainMenuItems = []ui.MenuItem{
	{Label: "New Game", Key: "n"},
	{Label: "Help", Key: "h"},
	{Label: "Quit", Key: "q"},
}

// Model is the top-level bubbletea model for the ATC game.
type Model struct {
	screen          screen
	menuScreen      menuScreen
	menuSelected    int
	maps            []gamemap.Map
	gameMap         gamemap.Map
	gameConfig      config.GameConfig
	setupFocus      int
	setupSelections [setupSections]int
	aircraft        map[string]aircraft.Aircraft
	runways         []runway.Runway
	spawner         *aircraft.Spawner
	input           textinput.Model
	score           int
	radioLog        radio.Log
	width           int
	height          int
	stopwatch       stopwatch.Model
	started         bool
	keys            keyMap
	help            help.Model
	stripViewport      viewport.Model
	radioViewport      viewport.Model
	cmdTree            cmdtree.Tree
	spawnDeparture     bool           // alternates between arrival and departure spawns
	nearMisses         int            // total separation violations this game
	activeViolations   map[string]bool // tracks pairs currently in violation (to warn once)

	// Developer mode
	devMode         bool // --dev flag enables / commands
	godMode         bool // collisions logged but don't end game
	spawnerPaused   bool // automatic spawning disabled
	speedMultiplier int  // physics ticks per render frame (1 = normal)
}

// NewModel creates a new model starting at the main menu.
func NewModel(devMode bool) Model {
	maps := gamemap.All()
	if len(maps) == 0 {
		panic("gamemap.All returned no maps")
	}
	gm := maps[0]
	cfg := config.DefaultConfig()
	h := help.New()
	h.ShowAll = true
	return Model{
		screen:          screenMenu,
		maps:            maps,
		gameMap:          gm,
		gameConfig:       cfg,
		aircraft:         make(map[string]aircraft.Aircraft),
		runways:          buildRunways(gm),
		radioLog:         radio.NewLog(),
		keys:             newKeyMap(),
		help:             h,
		devMode:          devMode,
		speedMultiplier:  1,
		setupSelections: [setupSections]int{
			0, // map: first
			0, // role: TRACON
			1, // difficulty: Normal
			0, // mode: Arrivals Only
			0, // callsign: ICAO
			1, // trails: Off
		},
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
		m.stripViewport.Height = max(m.height-8, 20)
		m.radioViewport.Width = max(m.width-4, 60)
		m.radioViewport.Height = radioViewportHeight
		m.help.Width = max(m.width-4, 0)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tickMsg:
		return m.handleTick(msg)
	}

	// Forward stopwatch messages
	var swCmd tea.Cmd
	m.stopwatch, swCmd = m.stopwatch.Update(msg)

	if m.screen == screenPlaying {
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		var vpCmd tea.Cmd
		m.stripViewport, vpCmd = m.stripViewport.Update(msg)
		return m, tea.Batch(swCmd, inputCmd, vpCmd)
	}

	return m, swCmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.ForceQuit) {
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

// --- Start Game ---

func (m Model) startGame() (Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "AA123 H270 A3 S2"
	ti.CharLimit = 40
	ti.Width = 40
	ti.Focus()

	m.screen = screenPlaying
	m.aircraft = make(map[string]aircraft.Aircraft)
	m.runways = buildRunways(m.gameMap)
	m.spawner = aircraft.NewSpawner(time.Now().UnixNano(), m.gameConfig)
	m.input = ti
	m.score = 0
	m.radioLog = radio.NewLog()
	m.radioLog = m.radioLog.Add(radio.SystemMessage(0, fmt.Sprintf("Welcome to %s. Type ? for help.", m.gameMap.Name), radio.Normal))
	m.stopwatch = stopwatch.NewWithInterval(tickInterval)
	m.stripViewport = viewport.New(32, max(m.height-8, 20))
	m.radioViewport = viewport.New(max(m.width-4, 60), radioViewportHeight)
	m.nearMisses = 0
	m.activeViolations = make(map[string]bool)
	m.started = true

	return m, tea.Batch(tickCmd(), textinput.Blink, m.stopwatch.Start())
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
		if m.menuScreen == menuSetup {
			menuView = ui.RenderSetup(m.buildSetupSections(), m.setupFocus)
		} else {
			menuView = ui.RenderMenu(mainMenuItems, m.menuSelected)
		}
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			menuView)

	case screenHelp:
		helpView := ui.RenderCommandHelp() + "\n\n" + m.help.View(m.keys)
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.HelpBox.Render(helpView))

	case screenPaused:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.RenderPaused(m.score, m.stopwatch.Elapsed()))

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

	hud := ui.RenderHUD(m.score, len(planes), m.stopwatch.Elapsed(), m.gameConfig.Role.String(), m.nearMisses, m.DevStatus())
	// Build set of callsigns currently in violation for radar highlighting
	violating := make(map[string]bool)
	for pair := range m.activeViolations {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			violating[parts[0]] = true
			violating[parts[1]] = true
		}
	}
	radarView := radar.Render(m.gameMap, planes, violating)
	sidebar := m.stripViewport.View()
	gameArea := lipgloss.JoinHorizontal(lipgloss.Top, radarView, " ", sidebar)

	radioPanel := ui.RadioBorder.Render(
		ui.RadioTitle.Render(" RADIO ") + "\n" + m.radioViewport.View())

	prompt := ui.InputPrompt.Render("ATC> ")
	inputView := prompt + m.input.View()

	treeView := cmdtree.Render(m.cmdTree)
	if treeView != "" {
		inputView = inputView + "\n" + treeView
	}

	return zone.Scan(lipgloss.JoinVertical(lipgloss.Left, hud, gameArea, radioPanel, inputView))
}

func (m Model) addRadio(msg radio.Message) Model {
	m.radioLog = m.radioLog.Add(msg)
	m.radioViewport.SetContent(radio.RenderLog(m.radioLog.All()))
	m.radioViewport.GotoBottom()
	return m
}
