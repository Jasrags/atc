package engine

import (
	"fmt"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/collision"
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/runway"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
	gameTPS       = 10
)

// Game implements ebiten.Game for the ATC radar display.
type Game struct {
	gameMap    gamemap.Map
	gameConfig config.GameConfig
	width      int
	height     int
	camera     Camera

	// Aircraft state
	aircraft       map[string]aircraft.Aircraft
	spawner        *aircraft.Spawner
	runways        []runway.Runway
	spawnDeparture bool

	// Time
	tickCount int
	elapsed   time.Duration

	// Violations
	activeViolations map[string]bool
}

// NewGame creates a new Ebitengine game with live aircraft.
func NewGame(gm gamemap.Map, role config.Role) *Game {
	cfg := config.DefaultConfig()
	cfg.Role = role
	cfg.PlaneTrails = true

	runways := make([]runway.Runway, len(gm.Runways))
	for i, r := range gm.Runways {
		runways[i] = runway.New(r.X, r.Y, r.Heading, r.Length)
	}

	// Initial camera depends on role.
	var cam Camera
	switch role {
	case config.RoleTower:
		cam = FitSurface(gm, defaultWidth, defaultHeight)
	default:
		cam = FitMap(gm, defaultWidth, defaultHeight)
	}

	return &Game{
		gameMap:          gm,
		gameConfig:       cfg,
		width:            defaultWidth,
		height:           defaultHeight,
		camera:           cam,
		aircraft:         make(map[string]aircraft.Aircraft),
		spawner:          aircraft.NewSpawner(time.Now().UnixNano(), cfg),
		runways:          runways,
		activeViolations: make(map[string]bool),
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	g.updateCamera()

	g.tickCount++
	g.elapsed = time.Duration(g.tickCount) * (time.Second / gameTPS)

	g.tickSpawn()
	g.tickPhysics()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameConfig.Role {
	case config.RoleTower:
		g.drawTower(screen)
	default:
		g.drawTRACON(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	g.camera.screenW = outsideWidth
	g.camera.screenH = outsideHeight
	return outsideWidth, outsideHeight
}

// --- Camera input ---

func (g *Game) updateCamera() {
	// Scroll wheel zoom — centered on cursor position.
	_, wy := ebiten.Wheel()
	if wy != 0 {
		cx, cy := ebiten.CursorPosition()
		if wy > 0 {
			g.camera.ZoomAt(float64(cx), float64(cy), zoomStep)
		} else {
			g.camera.ZoomAt(float64(cx), float64(cy), 1.0/zoomStep)
		}
	}

	// Click-and-drag pan. Mutually exclusive: start OR drag OR end per frame.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		g.camera.StartDrag(cx, cy)
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		g.camera.UpdateDrag(cx, cy)
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.camera.EndDrag()
	}

	// Keyboard pan (arrow keys).
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

	// +/- zoom from keyboard.
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		g.camera.ZoomAt(float64(g.width)/2, float64(g.height)/2, zoomStep)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		g.camera.ZoomAt(float64(g.width)/2, float64(g.height)/2, 1.0/zoomStep)
	}

	// Home resets to fit-all.
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
	return true
}

func (g *Game) spawnTowerArrival() bool {
	rw := g.gameMap.PrimaryRunway()
	ac := g.spawner.SpawnFinalApproach(rw.X, rw.Y, rw.Heading, g.gameMap.Width, g.gameMap.Height)
	if _, exists := g.aircraft[ac.Callsign]; exists {
		return false
	}
	g.aircraft[ac.Callsign] = ac
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

		if next.State == aircraft.Departing && next.IsOffScreen(g.gameMap.Width, g.gameMap.Height) {
			continue
		}
		if next.State.IsAirborne() && next.IsOffScreen(g.gameMap.Width, g.gameMap.Height) {
			continue
		}

		if next.State == aircraft.Landing {
			for _, rw := range g.runways {
				if rw.CanLand(next.GridX(), next.GridY(), next.Heading, next.Altitude) {
					next = next.WithState(aircraft.Landed)
					break
				}
			}
		}

		if next.State == aircraft.Landed {
			continue
		}
		if next.State == aircraft.AtGate {
			continue
		}

		newAircraft[k] = next
	}

	g.aircraft = newAircraft

	violations := collision.CheckSeparation(g.aircraft)
	currentPairs := make(map[string]bool)
	for _, v := range violations {
		pairKey := v.Callsign1 + ":" + v.Callsign2
		currentPairs[pairKey] = true
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
