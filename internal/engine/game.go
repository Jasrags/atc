package engine

import (
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
)

// Game implements ebiten.Game for the ATC radar display.
type Game struct {
	gameMap gamemap.Map
	role    config.Role
	width   int
	height  int
}

// NewGame creates a new Ebitengine game displaying the given map and role.
func NewGame(gm gamemap.Map, role config.Role) *Game {
	return &Game{
		gameMap: gm,
		role:    role,
		width:   defaultWidth,
		height:  defaultHeight,
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.role {
	case config.RoleTower:
		g.drawTower(screen)
	default:
		// TRACON and Combined both use the approach radar view.
		g.drawTRACON(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	return outsideWidth, outsideHeight
}
