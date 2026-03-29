package engine

import (
	"fmt"
	"math"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawTRACON renders the STARS-style approach radar display.
func (g *Game) drawTRACON(screen *ebiten.Image) {
	screen.Fill(traconBg)
	g.drawTraconGrid(screen)
	g.drawTraconRangeRings(screen)
	g.drawTraconFixes(screen)
	g.drawTraconRunways(screen)
	g.drawTraconAircraft(screen)
}

// --- Grid ---

func (g *Game) drawTraconGrid(screen *ebiten.Image) {
	for y := 0; y < g.gameMap.Height; y += 10 {
		for x := 0; x < g.gameMap.Width; x += 10 {
			sx, sy := worldToScreen(float64(x), float64(y))
			vector.DrawFilledRect(screen, float32(sx), float32(sy), 1, 1, traconGrid, false)
		}
	}
}

// --- Range rings ---

func (g *Game) drawTraconRangeRings(screen *ebiten.Image) {
	rw := g.gameMap.PrimaryRunway()
	cx, cy := worldToScreen(float64(rw.X), float64(rw.Y))

	for _, radius := range []float64{10, 20, 30, 40} {
		r := float32(radius * cellSize)
		drawCircle(screen, float32(cx), float32(cy), r, traconRangeRing)
	}
}

// --- Fixes ---

func (g *Game) drawTraconFixes(screen *ebiten.Image) {
	for _, fix := range g.gameMap.Fixes {
		sx, sy := worldToScreen(float64(fix.X), float64(fix.Y))

		switch fix.Type {
		case gamemap.FixWaypoint:
			s := float32(4)
			vector.StrokeLine(screen, float32(sx), float32(sy)-s, float32(sx)-s*0.7, float32(sy)+s*0.5, 0.8, traconFix, false)
			vector.StrokeLine(screen, float32(sx)-s*0.7, float32(sy)+s*0.5, float32(sx)+s*0.7, float32(sy)+s*0.5, 0.8, traconFix, false)
			vector.StrokeLine(screen, float32(sx)+s*0.7, float32(sy)+s*0.5, float32(sx), float32(sy)-s, 0.8, traconFix, false)
		case gamemap.FixAirport:
			drawCircle(screen, float32(sx), float32(sy), 4, traconFix)
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), 1.5, traconFix, false)
		case gamemap.FixVOR:
			drawCircle(screen, float32(sx), float32(sy), 3, traconFix)
		case gamemap.FixIntersection:
			vector.StrokeLine(screen, float32(sx)-3, float32(sy), float32(sx)+3, float32(sy), 0.8, traconFix, false)
			vector.StrokeLine(screen, float32(sx), float32(sy)-3, float32(sx), float32(sy)+3, 0.8, traconFix, false)
		}

		drawLabel(screen, sx+7, sy-3, fix.Name, 7, traconFixLabel)
	}
}

// --- Runways ---

func (g *Game) drawTraconRunways(screen *ebiten.Image) {
	for _, rw := range g.gameMap.Runways {
		g.drawTraconRunway(screen, rw)
	}
}

func (g *Game) drawTraconRunway(screen *ebiten.Image, rw gamemap.Runway) {
	rad := float64(rw.Heading) * math.Pi / 180.0
	dx := math.Sin(rad)
	dy := -math.Cos(rad)

	halfLen := float64(rw.Length-1) / 2.0
	x1 := float64(rw.X) - halfLen*dx
	y1 := float64(rw.Y) - halfLen*dy
	x2 := float64(rw.X) + halfLen*dx
	y2 := float64(rw.Y) + halfLen*dy

	sx1, sy1 := worldToScreen(x1, y1)
	sx2, sy2 := worldToScreen(x2, y2)

	// Runway centerline.
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 3, traconRunway, false)

	// Extended approach course — dashed.
	extLen := float64(25) * cellSize
	drawDashedLine(screen,
		float32(sx2), float32(sy2),
		float32(sx2+dx*extLen), float32(sy2-dy*extLen),
		6, 4, 1, traconApproach)
	drawDashedLine(screen,
		float32(sx1), float32(sy1),
		float32(sx1-dx*extLen), float32(sy1+dy*extLen),
		6, 4, 1, traconApproach)

	// Runway numbers.
	numApproach := gamemap.RunwayNumber(rw.Heading)
	numOpposite := gamemap.RunwayNumber(rw.OppositeHeading())

	drawLabel(screen, sx2+dx*cellSize*2, sy2-dy*cellSize*2-4,
		fmt.Sprintf("%d", numApproach), 9, traconRunwayNum)
	drawLabel(screen, sx1-dx*cellSize*2-14, sy1+dy*cellSize*2-4,
		fmt.Sprintf("%d", numOpposite), 9, traconRunwayNum)
}

// --- Live aircraft rendering ---

func (g *Game) drawTraconAircraft(screen *ebiten.Image) {
	for _, ac := range g.aircraft {
		// Skip ground aircraft that aren't departing.
		if ac.State.IsGround() && ac.State != aircraft.OnRunway {
			continue
		}

		sx, sy := worldToScreen(ac.X, ac.Y)

		// History trail — fading dots from oldest (dimmest) to newest.
		for i, pos := range ac.Trail {
			tx, ty := worldToScreen(float64(pos[0]), float64(pos[1]))
			// Fade: older dots are dimmer. Index 0 = oldest.
			alpha := float32(0.15) + float32(i)*float32(0.15)
			if alpha > 0.8 {
				alpha = 0.8
			}
			c := traconTrail
			c.A = uint8(float32(255) * alpha)
			vector.DrawFilledCircle(screen, float32(tx), float32(ty), 1.5, c, false)
		}

		// Target blip.
		targetColor := traconTarget
		if g.activeViolations[ac.Callsign+":"] || isViolating(ac.Callsign, g.activeViolations) {
			targetColor = traconConflict
		}
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 3.5, targetColor, false)

		// Leader line — offset to upper-right by default.
		leaderLen := float64(25)
		lx := sx + leaderLen
		ly := sy - leaderLen
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(lx), float32(ly), 0.8, traconLeader, false)

		// Data block.
		altArrow := " "
		if ac.Altitude < ac.TargetAltitude {
			altArrow = "\u2191" // ↑
		} else if ac.Altitude > ac.TargetAltitude {
			altArrow = "\u2193" // ↓
		}
		drawLabel(screen, lx+3, ly-10, ac.Callsign, 9, traconDataBlock)
		drawLabel(screen, lx+3, ly,
			fmt.Sprintf("%02d %02d%s", ac.Altitude, ac.Speed*10, altArrow), 8, traconDataBlock)
	}
}

// isViolating checks if a callsign appears in any violation pair key.
func isViolating(callsign string, violations map[string]bool) bool {
	for pair := range violations {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 && (parts[0] == callsign || parts[1] == callsign) {
			return true
		}
	}
	return false
}
