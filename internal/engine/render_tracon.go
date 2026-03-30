package engine

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

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
	g.drawTraconCompassRose(screen)
	g.drawTraconAircraft(screen)
	g.drawMinimap(screen)
}

// --- Grid ---

func (g *Game) drawTraconGrid(screen *ebiten.Image) {
	for y := 0; y < g.gameMap.Height; y += 10 {
		for x := 0; x < g.gameMap.Width; x += 10 {
			sx, sy := g.camera.WorldToScreen(float64(x), float64(y))
			vector.DrawFilledRect(screen, float32(sx), float32(sy), 1, 1, traconGrid, false)
		}
	}
}

// --- Range rings ---

func (g *Game) drawTraconRangeRings(screen *ebiten.Image) {
	rw := g.gameMap.PrimaryRunway()
	cx, cy := g.camera.WorldToScreen(float64(rw.X), float64(rw.Y))

	for _, radius := range []float64{10, 20, 30, 40} {
		r := float32(g.camera.ScaledSize(radius))
		drawCircle(screen, float32(cx), float32(cy), r, traconRangeRing)
	}
}

// --- Fixes ---

func (g *Game) drawTraconFixes(screen *ebiten.Image) {
	for _, fix := range g.gameMap.Fixes {
		sx, sy := g.camera.WorldToScreen(float64(fix.X), float64(fix.Y))

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

		fixFont := scaledFontSize(3, g.camera.Zoom, 7, 18)
		drawLabel(screen, sx+7, sy-3, fix.Name, fixFont, traconFixLabel)
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

	sx1, sy1 := g.camera.WorldToScreen(x1, y1)
	sx2, sy2 := g.camera.WorldToScreen(x2, y2)

	// Runway centerline.
	rwyWidth := float32(3 * g.camera.Zoom)
	if rwyWidth < 1 {
		rwyWidth = 1
	}
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), rwyWidth, traconRunway, false)

	// Extended approach course — dashed (extend 25 cells in world space).
	ext1X, ext1Y := g.camera.WorldToScreen(x2+dx*25, y2+dy*25)
	ext2X, ext2Y := g.camera.WorldToScreen(x1-dx*25, y1-dy*25)
	drawDashedLine(screen,
		float32(sx2), float32(sy2), float32(ext1X), float32(ext1Y),
		6, 4, 1, traconApproach)
	drawDashedLine(screen,
		float32(sx1), float32(sy1), float32(ext2X), float32(ext2Y),
		6, 4, 1, traconApproach)

	// Runway numbers (fixed screen-space offset from runway ends).
	numApproach := gamemap.RunwayNumber(rw.Heading)
	numOpposite := gamemap.RunwayNumber(rw.OppositeHeading())

	rwyFont := scaledFontSize(4, g.camera.Zoom, 9, 22)
	drawLabel(screen, sx2+dx*16, sy2-dy*16-4,
		fmt.Sprintf("%d", numApproach), rwyFont, traconRunwayNum)
	drawLabel(screen, sx1-dx*16-14, sy1+dy*16-4,
		fmt.Sprintf("%d", numOpposite), rwyFont, traconRunwayNum)
}

// --- Compass rose ---

var compassLabels = map[int]string{
	0: "N", 30: "03", 60: "06", 90: "E",
	120: "12", 150: "15", 180: "S", 210: "21",
	240: "24", 270: "W", 300: "30", 330: "33",
}

func (g *Game) drawTraconCompassRose(screen *ebiten.Image) {
	// Draw around the outermost range ring.
	rw := g.gameMap.PrimaryRunway()
	cx, cy := g.camera.WorldToScreen(float64(rw.X), float64(rw.Y))
	radius := g.camera.ScaledSize(42) // just outside the 40-cell range ring

	for deg := 0; deg < 360; deg += 10 {
		rad := float64(deg) * math.Pi / 180
		sin := math.Sin(rad)
		cos := -math.Cos(rad)

		// Tick mark.
		inner := radius - 4
		outer := radius
		if deg%30 == 0 {
			outer = radius + 2 // longer tick for labeled headings
		}

		x1 := float64(cx) + sin*inner
		y1 := float64(cy) + cos*inner
		x2 := float64(cx) + sin*outer
		y2 := float64(cy) + cos*outer
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 0.8, traconRangeRing, false)

		// Label at 30-degree intervals.
		if label, ok := compassLabels[deg]; ok {
			compassFont := scaledFontSize(3, g.camera.Zoom, 7, 16)
			lx := float64(cx) + sin*(radius+8) - 4
			ly := float64(cy) + cos*(radius+8) - 4
			drawLabel(screen, lx, ly, label, compassFont, traconFixLabel)
		}
	}
}

// --- Live aircraft rendering ---

func (g *Game) drawTraconAircraft(screen *ebiten.Image) {
	for _, ac := range g.aircraft {
		// Skip ground aircraft that aren't departing.
		if ac.State.IsGround() && ac.State != aircraft.OnRunway {
			continue
		}

		ix, iy := g.interpolatedPosition(ac)
		sx, sy := g.camera.WorldToScreen(ix, iy)

		// History trail — fading dots from oldest (dimmest) to newest.
		trailR := float32(1.5 * g.camera.Zoom)
		if trailR < 1 {
			trailR = 1
		}
		for i, pos := range ac.Trail {
			tx, ty := g.camera.WorldToScreen(float64(pos[0]), float64(pos[1]))
			alpha := float32(0.15) + float32(i)*float32(0.15)
			if alpha > 0.8 {
				alpha = 0.8
			}
			c := traconTrail
			c.A = uint8(float32(255) * alpha)
			vector.DrawFilledCircle(screen, float32(tx), float32(ty), trailR, c, false)
		}

		// Target blip — scales with zoom.
		blipR := float32(3.5 * g.camera.Zoom)
		if blipR < 2 {
			blipR = 2
		}
		if blipR > 8 {
			blipR = 8
		}
		targetColor := traconTarget
		violating := isViolating(ac.Callsign, g.activeViolations)
		if violating {
			// Blink red at 2Hz (on 70% of the time).
			if blinkVisible(g.elapsed, 500*time.Millisecond, 0.7) {
				targetColor = traconConflict
			} else {
				targetColor = colorWithAlpha(traconConflict, 0.3)
			}
		} else if ac.State == aircraft.Landing {
			// Pulse yellow for cleared-to-land aircraft.
			alpha := pulseAlpha(g.elapsed, 2*time.Second, 0.5, 1.0)
			targetColor = colorWithAlpha(traconLanding, alpha)
		}
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), blipR, targetColor, false)

		// Leader line — fixed screen-space length.
		leaderLen := float64(25)
		lx := sx + leaderLen
		ly := sy - leaderLen
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(lx), float32(ly), 0.8, traconLeader, false)

		// Data block — color reflects patience/state.
		dbColor := traconDataBlock
		switch {
		case violating:
			dbColor = traconConflict
		case ac.PatienceLevel() >= 3:
			// Angry — blink red.
			if blinkVisible(g.elapsed, 400*time.Millisecond, 0.6) {
				dbColor = traconConflict
			}
		case ac.PatienceLevel() >= 2:
			dbColor = color.RGBA{0xff, 0x88, 0x00, 0xff} // orange
		case ac.PatienceLevel() >= 1:
			dbColor = color.RGBA{0xcc, 0xcc, 0x00, 0xff} // yellow
		}

		altArrow := " "
		if ac.Altitude < ac.TargetAltitude {
			altArrow = "\u2191" // ↑
		} else if ac.Altitude > ac.TargetAltitude {
			altArrow = "\u2193" // ↓
		}
		acFont := scaledFontSize(4, g.camera.Zoom, 9, 20)
		drawLabel(screen, lx+3, ly-acFont*1.1, ac.Callsign, acFont, dbColor)
		drawLabel(screen, lx+3, ly,
			fmt.Sprintf("%02d %02d%s", ac.Altitude, ac.Speed*10, altArrow), acFont*0.85, dbColor)
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
