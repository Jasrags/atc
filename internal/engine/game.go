package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/collision"
	"github.com/Jasrags/atc/internal/command"
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/radio"
	"github.com/Jasrags/atc/internal/runway"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
	gameTPS       = 10

	minSpeedMul = 1
	maxSpeedMul = 12
)

type gameScreen int

const (
	screenPlaying gameScreen = iota
	screenGameOver
)

// Game implements ebiten.Game for the ATC radar display.
type Game struct {
	gameMap    gamemap.Map
	gameConfig config.GameConfig
	width      int
	height     int
	camera     Camera
	screen     gameScreen

	// Aircraft state
	aircraft       map[string]aircraft.Aircraft
	prevPositions  map[string][2]float64 // previous tick positions for interpolation
	spawner        *aircraft.Spawner
	runways        []runway.Runway
	spawnDeparture bool

	// Scoring
	score      int
	nearMisses int

	// Time
	tickCount       int
	elapsed         time.Duration
	timeFrozen      bool
	speedMultiplier int
	lastUpdateTime  time.Time // wall clock of last Update call, for draw interpolation

	// Radio
	radioLog radio.Log

	// Input
	input TextInput

	// Violations
	activeViolations map[string]bool

	// UI state
	stripHits []stripHit // bounding boxes of rendered strips for click detection

	// Developer mode
	devMode       bool
	godMode       bool
	spawnerPaused bool
}

// NewGame creates a new Ebitengine game with live aircraft.
func NewGame(gm gamemap.Map, role config.Role, devMode bool) *Game {
	cfg := config.DefaultConfig()
	cfg.Role = role
	cfg.PlaneTrails = true

	runways := make([]runway.Runway, len(gm.Runways))
	for i, r := range gm.Runways {
		runways[i] = runway.New(r.X, r.Y, r.Heading, r.Length)
	}

	var cam Camera
	switch role {
	case config.RoleTower:
		cam = FitSurface(gm, defaultWidth, defaultHeight)
	default:
		cam = FitMap(gm, defaultWidth, defaultHeight)
	}

	g := &Game{
		gameMap:          gm,
		gameConfig:       cfg,
		width:            defaultWidth,
		height:           defaultHeight,
		camera:           cam,
		screen:           screenPlaying,
		aircraft:         make(map[string]aircraft.Aircraft),
		prevPositions:    make(map[string][2]float64),
		spawner:          aircraft.NewSpawner(time.Now().UnixNano(), cfg),
		runways:          runways,
		speedMultiplier:  1,
		radioLog:         radio.NewLog(),
		input:            NewTextInput(),
		activeViolations: make(map[string]bool),
		devMode:          devMode,
	}

	g.addRadio(radio.SystemMessage(0,
		fmt.Sprintf("Welcome to %s. Type commands to direct aircraft.", gm.Name), radio.Normal))

	return g
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if g.screen == screenGameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.restart()
		}
		return nil
	}

	// Camera controls always active.
	g.updateCamera()

	// Time controls — only when input is empty (same as TUI).
	if g.input.IsEmpty() {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.timeFrozen = !g.timeFrozen
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
			g.speedMultiplier++
			if g.speedMultiplier > maxSpeedMul {
				g.speedMultiplier = maxSpeedMul
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
			g.speedMultiplier--
			if g.speedMultiplier < minSpeedMul {
				g.speedMultiplier = minSpeedMul
			}
		}
	}

	// Text input — always processes characters.
	submitted := g.input.Update()
	if submitted != "" {
		if strings.HasPrefix(submitted, "/") && g.devMode {
			g.processDevCommand(strings.TrimPrefix(submitted, "/"))
		} else {
			g.processCommand(submitted)
		}
	}

	// Precompute strip layout for click detection (before Draw, not inside it).
	g.stripHits = g.computeStripHits()

	// Physics — skip when frozen.
	if !g.timeFrozen {
		for i := 0; i < g.speedMultiplier; i++ {
			// Snapshot the penultimate tick for interpolation —
			// always capture the state just before the final tick.
			if i == g.speedMultiplier-1 {
				g.snapshotPositions()
			}

			g.tickCount++
			g.elapsed = time.Duration(g.tickCount) * (time.Second / gameTPS)

			if !g.spawnerPaused {
				g.tickSpawn()
			}
			g.tickPhysics()

			if g.screen == screenGameOver {
				break
			}
		}
	}

	g.lastUpdateTime = time.Now()
	return nil
}

// snapshotPositions saves current aircraft positions for draw interpolation.
func (g *Game) snapshotPositions() {
	for callsign, ac := range g.aircraft {
		g.prevPositions[callsign] = [2]float64{ac.X, ac.Y}
	}
}

// interpolatedPosition returns a smoothly interpolated position between
// the previous tick position and the current position based on time elapsed
// since the last Update call.
func (g *Game) interpolatedPosition(ac aircraft.Aircraft) (float64, float64) {
	prev, hasPrev := g.prevPositions[ac.Callsign]
	if !hasPrev || g.timeFrozen {
		return ac.X, ac.Y
	}

	// Fraction of a tick interval elapsed since last Update.
	tickDuration := time.Second / gameTPS
	elapsed := time.Since(g.lastUpdateTime)
	t := float64(elapsed) / float64(tickDuration)
	if t > 1 {
		t = 1
	}
	if t < 0 {
		t = 0
	}

	// Lerp from previous to current.
	x := prev[0] + (ac.X-prev[0])*t
	y := prev[1] + (ac.Y-prev[1])*t
	return x, y
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw radar view.
	switch g.gameConfig.Role {
	case config.RoleTower:
		g.drawTower(screen)
	default:
		g.drawTRACON(screen)
	}

	// Overlay UI (always on top of radar).
	g.drawHUD(screen)
	g.drawStrips(screen)
	g.drawRadioLog(screen)
	g.drawInput(screen)

	if g.screen == screenGameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	g.camera.screenW = outsideWidth
	g.camera.screenH = outsideHeight
	return outsideWidth, outsideHeight
}

// --- Command processing ---

func (g *Game) processCommand(input string) {
	cmd, err := command.Parse(input)
	if err != nil {
		g.addRadio(radio.SystemMessage(g.elapsed, err.Error(), radio.Normal))
		return
	}

	newPlanes, changes, err := command.Execute(cmd, g.aircraft, g.gameConfig.Role)
	if err != nil {
		g.addRadio(radio.SystemMessage(g.elapsed, err.Error(), radio.Normal))
		return
	}

	// Reset patience timer.
	if ac, exists := newPlanes[cmd.Callsign]; exists {
		ac = ac.ResetPatience()
		newPlanes[cmd.Callsign] = ac
	}

	// Resolve direct-to-fix.
	if cmd.DirectFix != "" {
		ac := newPlanes[cmd.Callsign]
		found := false
		for _, fix := range g.gameMap.Fixes {
			if fix.Name == cmd.DirectFix {
				ac.TargetFixX = float64(fix.X)
				ac.TargetFixY = float64(fix.Y)
				newPlanes[cmd.Callsign] = ac
				found = true
				break
			}
		}
		if !found {
			g.addRadio(radio.SystemMessage(g.elapsed,
				fmt.Sprintf("unknown fix: %s", cmd.DirectFix), radio.Normal))
			ac.TargetFixName = ""
			newPlanes[cmd.Callsign] = ac
		}
	}

	// Resolve hold-at-fix — set holding fix position only (navigation via updateHolding).
	if cmd.HoldFix != "" {
		ac := newPlanes[cmd.Callsign]
		found := false
		for _, fix := range g.gameMap.Fixes {
			if fix.Name == cmd.HoldFix {
				ac.HoldingFixX = float64(fix.X)
				ac.HoldingFixY = float64(fix.Y)
				newPlanes[cmd.Callsign] = ac
				found = true
				break
			}
		}
		if !found {
			g.addRadio(radio.SystemMessage(g.elapsed,
				fmt.Sprintf("unknown fix: %s", cmd.HoldFix), radio.Normal))
			ac.HoldingFixName = ""
			ac.TargetFixName = ""
			newPlanes[cmd.Callsign] = ac
		}
	}

	// Resolve taxi route.
	if len(cmd.TaxiRoute) > 0 {
		ac := newPlanes[cmd.Callsign]
		startNodeID := g.findNearestNode(ac)
		if startNodeID == "" {
			g.addRadio(radio.SystemMessage(g.elapsed, "no taxi nodes available", radio.Normal))
			ac = ac.WithState(aircraft.Pushback)
			ac.TaxiRoute = nil
			newPlanes[cmd.Callsign] = ac
		} else {
			path, routeErr := g.gameMap.ResolveTaxiRoute(startNodeID, cmd.TaxiRoute)
			if routeErr != nil {
				g.addRadio(radio.SystemMessage(g.elapsed,
					fmt.Sprintf("invalid taxi route: %s", routeErr), radio.Normal))
				ac = ac.WithState(aircraft.Pushback)
				ac.TaxiRoute = nil
			} else {
				ac.TaxiPath = g.nodeIDsToPositions(path)
				ac.TaxiPathIndex = 0
			}
			newPlanes[cmd.Callsign] = ac
		}
	}

	// Resolve gate assignment.
	if cmd.AssignGate != "" {
		ac := newPlanes[cmd.Callsign]
		gate := g.gameMap.GateByID(cmd.AssignGate)
		if gate == nil {
			g.addRadio(radio.SystemMessage(g.elapsed,
				fmt.Sprintf("unknown gate: %s", cmd.AssignGate), radio.Normal))
		} else {
			gateNode := g.gameMap.NodeByID(gate.NodeID)
			if gateNode != nil {
				ac.TaxiPath = [][2]int{{ac.GridX(), ac.GridY()}, {gateNode.X, gateNode.Y}}
				ac.TaxiPathIndex = 0
				newPlanes[cmd.Callsign] = ac
			}
		}
	}

	g.aircraft = newPlanes
	g.addRadio(radio.CommandPhraseology(g.elapsed, cmd.Callsign, changes))
}

// --- Dev commands (simplified subset) ---

func (g *Game) processDevCommand(input string) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	switch strings.ToLower(args[0]) {
	case "god":
		g.godMode = !g.godMode
		status := "OFF"
		if g.godMode {
			status = "ON"
		}
		g.addRadio(radio.SystemMessage(g.elapsed, fmt.Sprintf("god mode: %s", status), radio.Normal))
	case "pause":
		g.spawnerPaused = !g.spawnerPaused
		status := "resumed"
		if g.spawnerPaused {
			status = "paused"
		}
		g.addRadio(radio.SystemMessage(g.elapsed, fmt.Sprintf("spawner %s", status), radio.Normal))
	case "clear":
		count := len(g.aircraft)
		g.aircraft = make(map[string]aircraft.Aircraft)
		g.addRadio(radio.SystemMessage(g.elapsed, fmt.Sprintf("cleared %d aircraft", count), radio.Normal))
	case "spawn":
		if len(args) > 1 && strings.ToLower(args[1]) == "dep" {
			g.trySpawnDeparture()
		} else {
			g.spawnArrival()
		}
	default:
		g.addRadio(radio.SystemMessage(g.elapsed,
			fmt.Sprintf("unknown: /%s", args[0]), radio.Normal))
	}
}

// --- Helpers ---

func (g *Game) addRadio(msg radio.Message) {
	g.radioLog = g.radioLog.Add(msg)
}

// restart overwrites all game state in-place to start a new game.
// Safe because Ebitengine calls Update single-threaded — no concurrent access.
func (g *Game) restart() {
	role := g.gameConfig.Role
	dev := g.devMode
	*g = *NewGame(g.gameMap, role, dev)
}

func (g *Game) findNearestNode(ac aircraft.Aircraft) string {
	gx, gy := ac.GridX(), ac.GridY()
	bestDist := float64(1 << 60)
	bestID := ""
	for _, node := range g.gameMap.TaxiNodes {
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

func (g *Game) nodeIDsToPositions(nodeIDs []string) [][2]int {
	positions := make([][2]int, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		node := g.gameMap.NodeByID(id)
		if node != nil {
			positions = append(positions, [2]int{node.X, node.Y})
		}
	}
	return positions
}

// handleStripClick checks if a click is on a flight strip and populates the input.
// Returns true if a strip was clicked (preventing camera drag).
func (g *Game) handleStripClick(screenX, screenY int) bool {
	fx, fy := float64(screenX), float64(screenY)
	for _, hit := range g.stripHits {
		if fx >= hit.x && fx <= hit.x+hit.w && fy >= hit.y && fy <= hit.y+hit.h {
			g.input.SetText(hit.callsign + " ")
			return true
		}
	}
	return false
}

// --- Camera input ---

func (g *Game) updateCamera() {
	_, wy := ebiten.Wheel()
	if wy != 0 {
		cx, cy := ebiten.CursorPosition()
		if wy > 0 {
			g.camera.ZoomAt(float64(cx), float64(cy), zoomStep)
		} else {
			g.camera.ZoomAt(float64(cx), float64(cy), 1.0/zoomStep)
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Check strip click first — strips are in screen-fixed UI, not camera space.
		if !g.handleStripClick(cx, cy) {
			g.camera.StartDrag(cx, cy)
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		g.camera.UpdateDrag(cx, cy)
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.camera.EndDrag()
	}

	// Keyboard zoom (only when input empty to avoid consuming +/- while typing).
	if g.input.IsEmpty() {
		if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
			g.camera.ZoomAt(float64(g.width)/2, float64(g.height)/2, zoomStep)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
			g.camera.ZoomAt(float64(g.width)/2, float64(g.height)/2, 1.0/zoomStep)
		}
	}

	// Arrow key pan — only when input is empty to avoid conflicting with cursor movement.
	if g.input.IsEmpty() {
		panDelta := panSpeed / g.camera.Zoom
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
			g.camera.CenterX -= panDelta
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
			g.camera.CenterX += panDelta
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			g.camera.CenterY -= panDelta
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			g.camera.CenterY += panDelta
		}
	}

	// Home resets view.
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		switch g.gameConfig.Role {
		case config.RoleTower:
			g.camera = FitSurface(g.gameMap, g.width, g.height)
		default:
			g.camera = FitMap(g.gameMap, g.width, g.height)
		}
	}
}

// --- Tick pipeline ---

func (g *Game) tickSpawn() {
	if !g.spawner.ShouldSpawn(g.elapsed, len(g.aircraft)) {
		return
	}

	spawned := false
	if g.spawnDeparture && len(g.gameMap.Gates) > 0 {
		spawned = g.trySpawnDeparture()
	} else {
		spawned = g.spawnArrival()
	}
	if spawned {
		g.spawnDeparture = !g.spawnDeparture
	}
}

func (g *Game) spawnArrival() bool {
	if g.gameConfig.Role == config.RoleTower {
		return g.spawnTowerArrival()
	}
	ac := g.spawner.Spawn(g.gameMap.Width, g.gameMap.Height)
	if _, exists := g.aircraft[ac.Callsign]; exists {
		return false
	}
	g.aircraft[ac.Callsign] = ac
	g.addRadio(radio.PilotMessage(g.elapsed, ac.Callsign,
		radio.FormatEnteringAirspace(ac.Callsign, ac.Heading, ac.Altitude)))
	return true
}

func (g *Game) spawnTowerArrival() bool {
	rw := g.gameMap.PrimaryRunway()
	ac := g.spawner.SpawnFinalApproach(rw.X, rw.Y, rw.Heading, g.gameMap.Width, g.gameMap.Height)
	if _, exists := g.aircraft[ac.Callsign]; exists {
		return false
	}
	g.aircraft[ac.Callsign] = ac
	rwNum := fmt.Sprintf("%d", gamemap.RunwayNumber(rw.Heading))
	g.addRadio(radio.PilotMessage(g.elapsed, ac.Callsign,
		fmt.Sprintf("tower, %s on final runway %s", ac.Callsign, rwNum)))
	return true
}

func (g *Game) trySpawnDeparture() bool {
	occupiedGates := make(map[string]bool)
	for _, ac := range g.aircraft {
		if ac.AssignedGate != "" && ac.State.IsGround() {
			occupiedGates[ac.AssignedGate] = true
		}
	}

	var available []struct {
		ID   string
		X, Y int
	}
	for _, gate := range g.gameMap.Gates {
		if !occupiedGates[gate.ID] {
			node := g.gameMap.NodeByID(gate.NodeID)
			if node != nil {
				available = append(available, struct {
					ID   string
					X, Y int
				}{gate.ID, node.X, node.Y})
			}
		}
	}
	if len(available) == 0 {
		return false
	}

	ac, ok := g.spawner.SpawnDeparture(available)
	if !ok {
		return false
	}
	if _, exists := g.aircraft[ac.Callsign]; exists {
		return false
	}

	if g.gameConfig.Role == config.RoleTRACON {
		ac = ac.WithState(aircraft.Pushback)
	}

	g.aircraft[ac.Callsign] = ac
	// Ground comms only audible when player controls the surface (Tower/Combined).
	if g.gameConfig.Role != config.RoleTRACON {
		g.addRadio(radio.PilotMessage(g.elapsed, ac.Callsign,
			fmt.Sprintf("%s at gate %s, requesting pushback", ac.Callsign, ac.AssignedGate)))
	}
	return true
}

func (g *Game) tickPhysics() {
	newAircraft := make(map[string]aircraft.Aircraft, len(g.aircraft))

	for k, ac := range g.aircraft {
		var next aircraft.Aircraft
		switch {
		case ac.State == aircraft.OnRunway:
			next = ac.TakeoffTick()
			if next.State == aircraft.Departing && ac.State == aircraft.OnRunway {
				hdg := g.runwayHeading(ac.AssignedRunway)
				if hdg == 0 {
					hdg = g.gameMap.PrimaryRunway().Heading
				}
				next = next.WithHeading(hdg)
			}
		case ac.State.IsGround() && ac.State != aircraft.Landed:
			next = ac.GroundTick()
		default:
			next = ac.Tick()
		}

		// Departing aircraft that leave airspace.
		if next.State == aircraft.Departing && next.IsOffScreen(g.gameMap.Width, g.gameMap.Height) {
			g.score++
			g.addRadio(radio.PilotMessage(g.elapsed, next.Callsign,
				fmt.Sprintf("%s leaving airspace, good day", next.Callsign)))
			continue
		}
		// Other airborne that leave.
		if next.State.IsAirborne() && next.IsOffScreen(g.gameMap.Width, g.gameMap.Height) {
			continue
		}

		// Tower auto-handoff: departures at altitude 3+.
		if g.gameConfig.Role == config.RoleTower && next.State == aircraft.Departing && next.Altitude >= 3 {
			g.score++
			g.addRadio(radio.PilotMessage(g.elapsed, next.Callsign,
				fmt.Sprintf("%s contact departure, good day", next.Callsign)))
			continue
		}

		// Landing check.
		if next.State == aircraft.Landing {
			for _, rw := range g.runways {
				if rw.CanLand(next.GridX(), next.GridY(), next.Heading, next.Altitude) {
					next = next.WithState(aircraft.Landed)
					g.score++
					g.addRadio(radio.PilotMessage(g.elapsed, next.Callsign,
						radio.FormatLanded(next.Callsign)))
					break
				}
			}
		}

		// Landed — remove (simplified, no ground routing in GUI yet).
		if next.State == aircraft.Landed {
			continue
		}
		// AtGate — arrival complete.
		if next.State == aircraft.AtGate {
			continue
		}

		newAircraft[k] = next
	}

	g.aircraft = newAircraft

	// Collision detection.
	collisions := collision.Check(g.aircraft)
	if len(collisions) > 0 {
		for _, c := range collisions {
			g.addRadio(radio.SystemMessage(g.elapsed,
				radio.FormatCollision(c.Callsign1, c.Callsign2), radio.Emergency))
		}
		if !g.godMode {
			g.screen = screenGameOver
			return
		}
	}

	// Separation violations.
	violations := collision.CheckSeparation(g.aircraft)
	currentPairs := make(map[string]bool)
	for _, v := range violations {
		pairKey := v.Callsign1 + ":" + v.Callsign2
		currentPairs[pairKey] = true
		if !g.activeViolations[pairKey] {
			g.nearMisses++
			g.score -= collision.ViolationPenalty
			g.addRadio(radio.SystemMessage(g.elapsed,
				fmt.Sprintf("TRAFFIC ALERT: %s and %s — loss of separation (%.1f cells)",
					v.Callsign1, v.Callsign2, v.Distance), radio.Urgent))
		}
	}
	if g.score < 0 {
		g.score = 0
	}
	g.activeViolations = currentPairs
}

func (g *Game) runwayHeading(name string) int {
	for _, rw := range g.gameMap.Runways {
		num := gamemap.RunwayNumber(rw.Heading)
		if fmt.Sprintf("%d", num) == name {
			return rw.Heading
		}
		oppNum := gamemap.RunwayNumber(rw.OppositeHeading())
		if fmt.Sprintf("%d", oppNum) == name {
			return rw.OppositeHeading()
		}
	}
	return 0
}

func (g *Game) sortedAircraft() []aircraft.Aircraft {
	planes := make([]aircraft.Aircraft, 0, len(g.aircraft))
	for _, ac := range g.aircraft {
		planes = append(planes, ac)
	}
	return planes
}
