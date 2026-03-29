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

	// Single-char shortcuts (p, ?) only activate when the input is empty.
	// Otherwise they'd swallow letters the user is typing into the ATC prompt.
	if m.input.Value() == "" {
		switch {
		case key.Matches(msg, m.keys.Help):
			m.screen = screenHelp
			return m, nil
		case key.Matches(msg, m.keys.Pause):
			m.screen = screenPaused
			return m, m.stopwatch.Stop()
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

// tickEffect accumulates side effects from pure per-aircraft transformations.
// Each transformation returns a new Aircraft plus a tickEffect; the caller
// merges all effects into the Model after the pipeline completes.
type tickEffect struct {
	scoreDelta int
	messages   []radio.Message
	remove     bool // aircraft should be removed from the active map
}

// merge combines another tickEffect into this one.
func (e tickEffect) merge(other tickEffect) tickEffect {
	e.scoreDelta += other.scoreDelta
	e.messages = append(e.messages, other.messages...)
	e.remove = e.remove || other.remove
	return e
}

// tickLanding checks whether a Landing aircraft has reached a runway and
// transitions it to Landed. Returns the (possibly updated) aircraft and effect.
func (m Model) tickLanding(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if ac.State != aircraft.Landing {
		return ac, tickEffect{}
	}
	for i, rw := range m.runways {
		if ac.AssignedLandingRunway != "" && i < len(m.gameMap.Runways) {
			if !strings.Contains(m.gameMap.Runways[i].Name, ac.AssignedLandingRunway) {
				continue
			}
		}
		if rw.CanLand(ac.GridX(), ac.GridY(), ac.Heading, ac.Altitude) {
			next := ac
			next.State = aircraft.Landed
			return next, tickEffect{
				scoreDelta: 1,
				messages:   []radio.Message{radio.PilotMessage(elapsed, ac.Callsign, radio.FormatLanded(ac.Callsign))},
			}
		}
	}
	return ac, tickEffect{}
}

// tickAutoGroundArrival handles TRACON auto-ground: landed aircraft auto-taxi
// to the nearest available gate.
func (m Model) tickAutoGroundArrival(ac aircraft.Aircraft) (aircraft.Aircraft, tickEffect) {
	if ac.State != aircraft.Landed || m.gameConfig.Role != config.RoleTRACON {
		return ac, tickEffect{}
	}
	gate := m.findAvailableGate()
	if gate == "" {
		return ac, tickEffect{}
	}
	gateObj := m.gameMap.GateByID(gate)
	if gateObj == nil {
		return ac, tickEffect{}
	}
	node := m.gameMap.NodeByID(gateObj.NodeID)
	if node == nil {
		return ac, tickEffect{}
	}
	next := ac
	next.AssignedGate = gate
	next.State = aircraft.Taxiing
	next.TaxiPath = [][2]int{{ac.GridX(), ac.GridY()}, {node.X, node.Y}}
	next.TaxiPathIndex = 0
	return next, tickEffect{}
}

// tickTaxiComplete checks whether a taxiing aircraft with a gate assignment
// has finished its path. Returns an AtGate aircraft with a radio message.
func (m Model) tickTaxiComplete(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if ac.State != aircraft.Taxiing || len(ac.TaxiPath) != 0 || ac.AssignedGate == "" {
		return ac, tickEffect{}
	}
	next := ac
	next.State = aircraft.AtGate
	return next, tickEffect{
		messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
			fmt.Sprintf("%s at gate %s", ac.Callsign, ac.AssignedGate))},
	}
}

// tickAutoGroundDeparture handles TRACON auto-departure: automates pushback →
// taxi → hold short → on-runway transitions.
func (m Model) tickAutoGroundDeparture(ac aircraft.Aircraft) (aircraft.Aircraft, tickEffect) {
	if m.gameConfig.Role != config.RoleTRACON {
		return ac, tickEffect{}
	}
	switch ac.State {
	case aircraft.Pushback:
		if len(ac.TaxiPath) == 0 {
			hsNode := m.findNearestHoldShort(ac)
			if hsNode != nil {
				next := ac
				next.State = aircraft.Taxiing
				next.TaxiPath = [][2]int{{ac.GridX(), ac.GridY()}, {hsNode.X, hsNode.Y}}
				next.TaxiPathIndex = 0
				return next, tickEffect{}
			}
		}
	case aircraft.Taxiing:
		if len(ac.TaxiPath) == 0 && ac.AssignedGate == "" {
			next := ac
			next.State = aircraft.HoldShort
			return next, tickEffect{}
		}
	case aircraft.HoldShort:
		next := ac.ResetTickCount()
		next.State = aircraft.OnRunway
		next.TaxiPath = nil
		next.TaxiRoute = nil
		return next, tickEffect{}
	}
	return ac, tickEffect{}
}

// tickPatience advances the patience timer for airborne aircraft and generates
// nag messages or removal when patience expires.
func (m Model) tickPatience(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if !ac.State.IsAirborne() || ac.PatienceMax == 0 {
		return ac, tickEffect{}
	}
	next := ac
	next.PatienceTicks++

	nagThreshold := next.PatienceMax + next.PatienceNagCount*aircraft.PatienceNagEvery
	if next.PatienceTicks < nagThreshold {
		return next, tickEffect{}
	}

	next.PatienceNagCount++
	switch {
	case next.PatienceNagCount >= aircraft.PatienceLeaveAt:
		return next, tickEffect{
			scoreDelta: -1,
			messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s leaving your airspace, good day", ac.Callsign))},
			remove: true,
		}
	case next.PatienceNagCount >= aircraft.PatiencePenaltyAt:
		return next, tickEffect{
			messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s requesting ANY instructions!", ac.Callsign))},
		}
	default:
		return next, tickEffect{
			messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s still waiting for vectors", ac.Callsign))},
		}
	}
}

func (m Model) handleTick(msg tickMsg) (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying {
		return m, nil
	}

	elapsed := m.stopwatch.Elapsed()

	// Speed multiplier: run physics N times per render frame.
	ticks := m.speedMultiplier
	if ticks < 1 {
		ticks = 1
	}

	for tick := 0; tick < ticks; tick++ {
		m = m.tickPhysics(elapsed)
		if m.screen == screenGameOver {
			return m, nil
		}
	}

	// Phase 5: spawning (once per frame regardless of speed)
	if !m.spawnerPaused && m.spawner.ShouldSpawn(elapsed, len(m.aircraft)) {
		countBefore := len(m.aircraft)
		if m.spawnDeparture && len(m.gameMap.Gates) > 0 {
			m = m.trySpawnDeparture(elapsed)
		} else {
			ac := m.spawner.Spawn(m.gameMap.Width, m.gameMap.Height)
			if _, exists := m.aircraft[ac.Callsign]; !exists {
				m.aircraft[ac.Callsign] = ac
				m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign,
					radio.FormatEnteringAirspace(ac.Callsign, ac.Heading, ac.Altitude)))
			}
		}
		if len(m.aircraft) > countBefore {
			m.spawnDeparture = !m.spawnDeparture
		}
	}

	// Update flight strip viewport content
	m.stripViewport.SetContent(radar.RenderFlightStrips(m.sortedAircraft(), m.gameConfig.Role))

	return m, tickCmd()
}

// tickPhysics runs one cycle of aircraft movement, collision, separation, and
// state transitions. Extracted so the speed multiplier can call it N times.
func (m Model) tickPhysics(elapsed time.Duration) Model {
	// Phase 1: advance physics (each Tick/GroundTick/TakeoffTick is already pure)
	newAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		var next aircraft.Aircraft
		switch {
		case ac.State == aircraft.OnRunway:
			next = ac.TakeoffTick()
			if next.State == aircraft.Departing && ac.State == aircraft.OnRunway {
				hdg := m.runwayHeading(ac.AssignedRunway)
				if hdg == 0 {
					hdg = m.gameMap.PrimaryRunway().Heading
				}
				next = next.WithHeading(hdg)
			}
		case ac.State.IsGround() && ac.State != aircraft.Landed:
			next = ac.GroundTick()
		default:
			next = ac.Tick()
		}
		// Departing aircraft that leave airspace = successful departure
		if next.State == aircraft.Departing && next.IsOffScreen(m.gameMap.Width, m.gameMap.Height) {
			m.score++
			m = m.addRadio(radio.PilotMessage(elapsed, next.Callsign,
				fmt.Sprintf("%s leaving airspace, good day", next.Callsign)))
			continue
		}
		// Other airborne aircraft that leave = just removed
		if next.State.IsAirborne() && next.IsOffScreen(m.gameMap.Width, m.gameMap.Height) {
			continue
		}
		newAircraft[k] = next
	}
	m.aircraft = newAircraft

	// Phase 2: collision detection (immutable — build new map with crashed state)
	collisions := collision.Check(m.aircraft)
	if len(collisions) > 0 {
		for _, c := range collisions {
			m = m.addRadio(radio.SystemMessage(elapsed,
				radio.FormatCollision(c.Callsign1, c.Callsign2), radio.Emergency))
		}
		if !m.godMode {
			crashed := make(map[string]aircraft.Aircraft, len(m.aircraft))
			for k, ac := range m.aircraft {
				crashed[k] = ac
			}
			for _, c := range collisions {
				crashed[c.Callsign1] = crashed[c.Callsign1].WithState(aircraft.Crashed)
				crashed[c.Callsign2] = crashed[c.Callsign2].WithState(aircraft.Crashed)
			}
			m.aircraft = crashed
			m.screen = screenGameOver
			return m
		}
	}

	// Phase 3: separation violations
	violations := collision.CheckSeparation(m.aircraft)
	currentPairs := make(map[string]bool)
	for _, v := range violations {
		pairKey := v.Callsign1 + ":" + v.Callsign2
		currentPairs[pairKey] = true
		m.score -= collision.ViolationPenalty
		if !m.activeViolations[pairKey] {
			m.nearMisses++
			m = m.addRadio(radio.SystemMessage(elapsed,
				fmt.Sprintf("TRAFFIC ALERT: %s and %s — loss of separation (%.1f cells)",
					v.Callsign1, v.Callsign2, v.Distance), radio.Urgent))
		}
	}
	if m.score < 0 {
		m.score = 0
	}
	m.activeViolations = currentPairs

	// Phase 4: per-aircraft state pipeline (pure transformations)
	activeAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		var fx tickEffect
		var combined tickEffect

		ac, fx = m.tickLanding(ac, elapsed)
		combined = combined.merge(fx)

		ac, fx = m.tickAutoGroundArrival(ac)
		combined = combined.merge(fx)

		ac, fx = m.tickTaxiComplete(ac, elapsed)
		combined = combined.merge(fx)

		ac, fx = m.tickAutoGroundDeparture(ac)
		combined = combined.merge(fx)

		ac, fx = m.tickPatience(ac, elapsed)
		combined = combined.merge(fx)

		// Apply accumulated effects
		m.score += combined.scoreDelta
		for _, msg := range combined.messages {
			m = m.addRadio(msg)
		}

		// Remove aircraft that left or completed arrival
		if combined.remove || ac.State == aircraft.AtGate {
			continue
		}
		activeAircraft[k] = ac
	}
	if m.score < 0 {
		m.score = 0
	}
	m.aircraft = activeAircraft

	return m
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

// trySpawnDeparture attempts to spawn a departure aircraft at an unoccupied gate.
func (m Model) trySpawnDeparture(elapsed time.Duration) Model {
	// Build list of available gates (not occupied by another aircraft)
	occupiedGates := make(map[string]bool)
	for _, ac := range m.aircraft {
		if ac.AssignedGate != "" && ac.State.IsGround() {
			occupiedGates[ac.AssignedGate] = true
		}
	}

	var available []struct{ ID string; X, Y int }
	for _, g := range m.gameMap.Gates {
		if !occupiedGates[g.ID] {
			node := m.gameMap.NodeByID(g.NodeID)
			if node != nil {
				available = append(available, struct{ ID string; X, Y int }{g.ID, node.X, node.Y})
			}
		}
	}

	if len(available) == 0 {
		return m
	}

	ac, ok := m.spawner.SpawnDeparture(available)
	if !ok {
		return m
	}
	if _, exists := m.aircraft[ac.Callsign]; exists {
		return m
	}

	// In TRACON mode, auto-pushback departures immediately
	if m.gameConfig.Role == config.RoleTRACON {
		ac.State = aircraft.Pushback
	}

	m.aircraft[ac.Callsign] = ac
	m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign,
		fmt.Sprintf("%s at gate %s, requesting pushback", ac.Callsign, ac.AssignedGate)))
	return m
}

// findNearestHoldShort returns the nearest hold-short node to the aircraft, or nil if none.
func (m Model) findNearestHoldShort(ac aircraft.Aircraft) *gamemap.TaxiNode {
	gx, gy := ac.GridX(), ac.GridY()
	bestDist := math.MaxFloat64
	var bestNode *gamemap.TaxiNode
	for i := range m.gameMap.TaxiNodes {
		n := &m.gameMap.TaxiNodes[i]
		if n.Type != gamemap.NodeHoldShort {
			continue
		}
		dx := float64(n.X - gx)
		dy := float64(n.Y - gy)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			bestNode = n
		}
	}
	return bestNode
}

// findAvailableGate returns the ID of an unoccupied gate, or "" if none available.
func (m Model) findAvailableGate() string {
	occupied := make(map[string]bool)
	for _, ac := range m.aircraft {
		if ac.AssignedGate != "" {
			occupied[ac.AssignedGate] = true
		}
	}
	for _, g := range m.gameMap.Gates {
		if !occupied[g.ID] {
			return g.ID
		}
	}
	return ""
}

// runwayHeading returns the heading of the named runway, or 0 if not found.
func (m Model) runwayHeading(name string) int {
	for _, rw := range m.gameMap.Runways {
		// Match by runway number (e.g., "27" matches heading 270)
		num := gamemap.RunwayNumber(rw.Heading)
		if fmt.Sprintf("%d", num) == name {
			return rw.Heading
		}
		// Also check opposite end
		oppNum := gamemap.RunwayNumber(rw.OppositeHeading())
		if fmt.Sprintf("%d", oppNum) == name {
			return rw.OppositeHeading()
		}
	}
	return 0
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
