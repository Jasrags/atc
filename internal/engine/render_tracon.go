package engine

import (
	"fmt"
	"math"

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
	g.drawTraconDemoAircraft(screen)
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
	// Center on the primary runway.
	rw := g.gameMap.PrimaryRunway()
	cx, cy := worldToScreen(float64(rw.X), float64(rw.Y))

	// Draw rings at 10, 20, 30, 40 cell intervals (~nm equivalent).
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
			// Small triangle.
			s := float32(4)
			vector.StrokeLine(screen, float32(sx), float32(sy)-s, float32(sx)-s*0.7, float32(sy)+s*0.5, 0.8, traconFix, false)
			vector.StrokeLine(screen, float32(sx)-s*0.7, float32(sy)+s*0.5, float32(sx)+s*0.7, float32(sy)+s*0.5, 0.8, traconFix, false)
			vector.StrokeLine(screen, float32(sx)+s*0.7, float32(sy)+s*0.5, float32(sx), float32(sy)-s, 0.8, traconFix, false)
		case gamemap.FixAirport:
			drawCircle(screen, float32(sx), float32(sy), 4, traconFix)
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), 1.5, traconFix, false)
		case gamemap.FixVOR:
			// Small hexagon-like shape.
			drawCircle(screen, float32(sx), float32(sy), 3, traconFix)
		case gamemap.FixIntersection:
			// Small plus.
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

	// Runway centerline — solid, bright.
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 3, traconRunway, false)

	// Extended approach course — dashed lines extending beyond runway.
	extLen := float64(25) * cellSize // extend 25 cells
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

// --- Demo aircraft (static, replaced by live aircraft in Phase 2) ---

func (g *Game) drawTraconDemoAircraft(screen *ebiten.Image) {
	// Simulate 3 aircraft with history trails and data blocks.
	demos := []struct {
		x, y       float64
		callsign   string
		alt, spd   int
		heading    int
		descending bool
		trail      [][2]float64 // history positions
	}{
		{
			x: 35, y: 12, callsign: "AA341", alt: 8, spd: 22, heading: 250, descending: true,
			trail: [][2]float64{{36.5, 10.5}, {36, 11}, {35.5, 11.5}},
		},
		{
			x: 25, y: 20, callsign: "SW482", alt: 5, spd: 19, heading: 180, descending: true,
			trail: [][2]float64{{25, 18}, {25, 18.7}, {25, 19.3}},
		},
		{
			x: 60, y: 24, callsign: "UA159", alt: 2, spd: 16, heading: 270, descending: true,
			trail: [][2]float64{{62.5, 24}, {62, 24}, {61.2, 24}},
		},
	}

	for _, ac := range demos {
		// History trail — fading dots.
		for i, pos := range ac.trail {
			sx, sy := worldToScreen(pos[0], pos[1])
			alpha := float32(0.2) + float32(i)*0.15 // older = dimmer
			r := float32(1.5)
			c := traconTrail
			c.A = uint8(float32(c.A) * alpha)
			vector.DrawFilledCircle(screen, float32(sx), float32(sy), r, c, false)
		}

		sx, sy := worldToScreen(ac.x, ac.y)

		// Target blip — bright filled circle.
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), 3.5, traconTarget, false)

		// Leader line — offset to upper-right (default octant).
		leaderLen := float64(25)
		lx := sx + leaderLen
		ly := sy - leaderLen
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(lx), float32(ly), 0.8, traconLeader, false)

		// Data block at end of leader line.
		arrow := " "
		if ac.descending {
			arrow = "\u2193" // ↓
		}
		drawLabel(screen, lx+3, ly-10, ac.callsign, 9, traconDataBlock)
		drawLabel(screen, lx+3, ly,
			fmt.Sprintf("%02d %02d%s", ac.alt, ac.spd, arrow), 8, traconDataBlock)
	}
}
