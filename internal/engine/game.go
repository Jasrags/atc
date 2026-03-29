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
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
	gameTPS       = 10 // match the TUI tick rate: 10 updates per second
)

// Game implements ebiten.Game for the ATC radar display.
type Game struct {
	gameMap    gamemap.Map
	gameConfig config.GameConfig
	width      int
	height     int

	// Aircraft state
	aircraft       map[string]aircraft.Aircraft
	spawner        *aircraft.Spawner
	runways        []runway.Runway
	spawnDeparture bool

	// Time
	tickCount int           // total ticks since game start
	elapsed   time.Duration // simulated elapsed time (tickCount * 100ms)

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

	return &Game{
		gameMap:          gm,
		gameConfig:       cfg,
		width:            defaultWidth,
		height:           defaultHeight,
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
	return outsideWidth, outsideHeight
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

	// In TRACON mode, auto-pushback departures.
	if g.gameConfig.Role == config.RoleTRACON {
		ac.State = aircraft.Pushback
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

		// Departing aircraft that leave airspace — remove.
		if next.State == aircraft.Departing && next.IsOffScreen(g.gameMap.Width, g.gameMap.Height) {
			continue
		}
		// Airborne aircraft that leave — remove.
		if next.State.IsAirborne() && next.IsOffScreen(g.gameMap.Width, g.gameMap.Height) {
			continue
		}

		// Landing check.
		if next.State == aircraft.Landing {
			for _, rw := range g.runways {
				if rw.CanLand(next.GridX(), next.GridY(), next.Heading, next.Altitude) {
					next.State = aircraft.Landed
					break
				}
			}
		}

		// Landed aircraft — remove after landing (simplified: no ground routing yet).
		if next.State == aircraft.Landed {
			continue
		}
		// AtGate — arrival complete, remove.
		if next.State == aircraft.AtGate {
			continue
		}

		newAircraft[k] = next
	}

	g.aircraft = newAircraft

	// Separation violations.
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

// sortedAircraft returns aircraft sorted by callsign for consistent rendering order.
func (g *Game) sortedAircraft() []aircraft.Aircraft {
	planes := make([]aircraft.Aircraft, 0, len(g.aircraft))
	for _, ac := range g.aircraft {
		planes = append(planes, ac)
	}
	return planes
}
