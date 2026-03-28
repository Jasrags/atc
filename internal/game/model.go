package game

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/cmdtree"
	"github.com/Jasrags/atc/internal/collision"
	"github.com/Jasrags/atc/internal/command"
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
	setupDiff      = 1
	setupMode      = 2
	setupCallsign  = 3
	setupTrails    = 4
	setupSections  = 5
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
	stripViewport   viewport.Model
	radioViewport   viewport.Model
	cmdTree         cmdtree.Tree
}

// NewModel creates a new model starting at the main menu.
func NewModel() Model {
	maps := gamemap.All()
	if len(maps) == 0 {
		panic("gamemap.All returned no maps")
	}
	gm := maps[0]
	cfg := config.DefaultConfig()
	h := help.New()
	h.ShowAll = true
	return Model{
		screen:     screenMenu,
		maps:       maps,
		gameMap:    gm,
		gameConfig: cfg,
		aircraft:   make(map[string]aircraft.Aircraft),
		runways:    buildRunways(gm),
		radioLog:   radio.NewLog(),
		keys:       newKeyMap(),
		help:       h,
		setupSelections: [setupSections]int{
			0, // map: first
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
		{Title: "Difficulty", Options: config.DifficultyOptions(), Selected: m.setupSelections[setupDiff]},
		{Title: "Game Mode", Options: config.GameModeOptions(), Selected: m.setupSelections[setupMode]},
		{Title: "Callsign Style", Options: config.CallsignStyleOptions(), Selected: m.setupSelections[setupCallsign]},
		{Title: "Plane Trails", Options: config.PlaneTrailsOptions(), Selected: m.setupSelections[setupTrails]},
	}
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
	m.started = true

	return m, tea.Batch(tickCmd(), textinput.Blink, m.stopwatch.Start())
}

// --- Playing ---

func (m Model) handlePlayingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.screen = screenMenu
		m.menuSelected = 0
		return m, m.stopwatch.Stop()

	case key.Matches(msg, m.keys.Help):
		m.screen = screenHelp
		return m, nil

	case key.Matches(msg, m.keys.Pause):
		m.screen = screenPaused
		return m, m.stopwatch.Stop()
	}

	if key.Matches(msg, m.keys.Submit) {
		input := strings.TrimSpace(m.input.Value())
		if input != "" {
			m = m.processCommand(input)
			m.input.Reset()
			m.cmdTree = cmdtree.Tree{}
		}
		return m, nil
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

	newPlanes, changes, err := command.Execute(cmd, m.aircraft)
	if err != nil {
		m = m.addRadio(radio.SystemMessage(elapsed, err.Error(), radio.Normal))
		return m
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

// --- Paused ---

func (m Model) handlePausedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Pause):
		m.screen = screenPlaying
		return m, tea.Batch(tickCmd(), m.stopwatch.Start())
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Back):
		m.screen = screenMenu
		m.menuSelected = 0
		return m, nil
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

// --- Tick ---

func (m Model) handleTick(msg tickMsg) (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying {
		return m, nil
	}

	elapsed := m.stopwatch.Elapsed()

	newAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		var next aircraft.Aircraft
		if ac.State.IsGround() && ac.State != aircraft.Landed {
			next = ac.GroundTick()
		} else {
			next = ac.Tick()
		}
		// Only airborne aircraft can leave the radar area
		if next.State.IsAirborne() && next.IsOffScreen(m.gameMap.Width, m.gameMap.Height) {
			continue
		}
		newAircraft[k] = next
	}
	m.aircraft = newAircraft

	collisions := collision.Check(m.aircraft)
	if len(collisions) > 0 {
		for _, c := range collisions {
			ac1 := m.aircraft[c.Callsign1]
			ac1.State = aircraft.Crashed
			m.aircraft[c.Callsign1] = ac1

			ac2 := m.aircraft[c.Callsign2]
			ac2.State = aircraft.Crashed
			m.aircraft[c.Callsign2] = ac2

			m = m.addRadio(radio.SystemMessage(elapsed,
				radio.FormatCollision(c.Callsign1, c.Callsign2), radio.Emergency))
		}
		m.screen = screenGameOver
		return m, nil
	}

	// Single pass: check landings, taxi completion, and remove completed arrivals
	activeAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		// Check landing
		if ac.State == aircraft.Landing {
			for _, rw := range m.runways {
				if rw.CanLand(ac.GridX(), ac.GridY(), ac.Heading, ac.Altitude) {
					ac.State = aircraft.Landed
					m.score++
					m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign, radio.FormatLanded(ac.Callsign)))
					break
				}
			}
		}

		// Check taxi completion
		if ac.State == aircraft.Taxiing && len(ac.TaxiPath) == 0 && ac.AssignedGate != "" {
			ac.State = aircraft.AtGate
			m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s at gate %s", ac.Callsign, ac.AssignedGate)))
		}

		// Remove completed arrivals (at gate)
		if ac.State == aircraft.AtGate {
			continue
		}
		activeAircraft[k] = ac
	}
	m.aircraft = activeAircraft

	if m.spawner.ShouldSpawn(elapsed, len(m.aircraft)) {
		ac := m.spawner.Spawn(m.gameMap.Width, m.gameMap.Height)
		if _, exists := m.aircraft[ac.Callsign]; !exists {
			m.aircraft[ac.Callsign] = ac
			m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign,
				radio.FormatEnteringAirspace(ac.Callsign, ac.Heading, ac.Altitude)))
		}
	}

	// Update flight strip viewport content
	m.stripViewport.SetContent(radar.RenderFlightStrips(m.sortedAircraft()))

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

	hud := ui.RenderHUD(m.score, len(planes), m.stopwatch.Elapsed())
	radarView := radar.Render(m.gameMap, planes)
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

	m.cmdTree = cmdtree.Resolve(inputText, acState)
	return m
}

// findNearestNode returns the ID of the closest taxi node to the aircraft's position.
func (m Model) findNearestNode(ac aircraft.Aircraft) string {
	gx, gy := ac.GridX(), ac.GridY()
	bestDist := math.MaxFloat64
	bestID := ""
	for _, node := range m.gameMap.TaxiNodes {
		dx := float64(node.X - gx)
		dy := float64(node.Y - gy)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			bestID = node.ID
		}
	}
	return bestID
}

// nodeIDsToPositions converts a list of node IDs into grid positions.
func (m Model) nodeIDsToPositions(nodeIDs []string) [][2]int {
	positions := make([][2]int, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		node := m.gameMap.NodeByID(id)
		if node != nil {
			positions = append(positions, [2]int{node.X, node.Y})
		}
	}
	return positions
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
