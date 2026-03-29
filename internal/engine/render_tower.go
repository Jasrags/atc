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

// drawTower renders the ASDE-X style airport surface radar display.
func (g *Game) drawTower(screen *ebiten.Image) {
	screen.Fill(towerBg)
	g.drawTowerTaxiways(screen)
	g.drawTowerGates(screen)
	g.drawTowerHoldShorts(screen)
	g.drawTowerRunways(screen)
	g.drawTowerAircraft(screen)
	g.drawTowerApproachInset(screen)
	g.drawMinimap(screen)
}

// --- Runways (filled rectangles with threshold markings) ---

func (g *Game) drawTowerRunways(screen *ebiten.Image) {
	for _, rw := range g.gameMap.Runways {
		g.drawTowerRunway(screen, rw)
	}
}

func (g *Game) drawTowerRunway(screen *ebiten.Image, rw gamemap.Runway) {
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

	// Perpendicular direction for width — scales with zoom.
	perpX := dy
	perpY := dx
	runwayWidth := g.camera.ScaledSize(1.8)

	// Four corners of the runway rectangle.
	hw := runwayWidth / 2
	c1x := float32(sx1 + perpX*hw)
	c1y := float32(sy1 + perpY*hw)
	c2x := float32(sx1 - perpX*hw)
	c2y := float32(sy1 - perpY*hw)
	c3x := float32(sx2 - perpX*hw)
	c3y := float32(sy2 - perpY*hw)
	c4x := float32(sx2 + perpX*hw)
	c4y := float32(sy2 + perpY*hw)

	// Fill using two triangles.
	r := float32(towerRunwayFill.R) / 255
	gc := float32(towerRunwayFill.G) / 255
	b := float32(towerRunwayFill.B) / 255
	vertices := []ebiten.Vertex{
		{DstX: c1x, DstY: c1y, ColorR: r, ColorG: gc, ColorB: b, ColorA: 1},
		{DstX: c2x, DstY: c2y, ColorR: r, ColorG: gc, ColorB: b, ColorA: 1},
		{DstX: c3x, DstY: c3y, ColorR: r, ColorG: gc, ColorB: b, ColorA: 1},
		{DstX: c4x, DstY: c4y, ColorR: r, ColorG: gc, ColorB: b, ColorA: 1},
	}
	indices := []uint16{0, 1, 2, 0, 2, 3}
	screen.DrawTriangles(vertices, indices, whitePixelImg, &ebiten.DrawTrianglesOptions{})

	// Runway outline.
	vector.StrokeLine(screen, c1x, c1y, c2x, c2y, 1, towerRunwayEdge, false)
	vector.StrokeLine(screen, c2x, c2y, c3x, c3y, 1, towerRunwayEdge, false)
	vector.StrokeLine(screen, c3x, c3y, c4x, c4y, 1, towerRunwayEdge, false)
	vector.StrokeLine(screen, c4x, c4y, c1x, c1y, 1, towerRunwayEdge, false)

	// Dashed centerline.
	drawDashedLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 4, 3, 0.8, towerRunwayEdge)

	// Threshold markings — multiple bars (FAA style).
	threshW := float32(hw * 0.8)
	for _, offset := range []float64{0, 0.4, -0.4} {
		tx := float32(sx1) + float32(dx*offset*cellSize*g.camera.Zoom)
		ty := float32(sy1) + float32(dy*offset*cellSize*g.camera.Zoom)
		vector.StrokeLine(screen, tx+float32(perpX)*threshW, ty+float32(perpY)*threshW,
			tx-float32(perpX)*threshW, ty-float32(perpY)*threshW, 1.5, towerThreshold, false)
	}
	// Threshold bars at the other end.
	for _, offset := range []float64{0, 0.4, -0.4} {
		tx := float32(sx2) + float32(dx*offset*cellSize*g.camera.Zoom)
		ty := float32(sy2) + float32(dy*offset*cellSize*g.camera.Zoom)
		vector.StrokeLine(screen, tx+float32(perpX)*threshW, ty+float32(perpY)*threshW,
			tx-float32(perpX)*threshW, ty-float32(perpY)*threshW, 1.5, towerThreshold, false)
	}

	// Runway numbers inside the surface, near each threshold.
	numApproach := gamemap.RunwayNumber(rw.Heading)
	numOpposite := gamemap.RunwayNumber(rw.OppositeHeading())

	// Inset from threshold toward center.
	inset := g.camera.ScaledSize(1.5)
	rwyFontSize := scaledFontSize(4, g.camera.Zoom, 10, 28)
	drawLabel(screen, sx2-dx*inset-8, sy2+dy*inset-6,
		fmt.Sprintf("%d", numApproach), rwyFontSize, towerRunwayNum)
	drawLabel(screen, sx1+dx*inset-8, sy1-dy*inset-6,
		fmt.Sprintf("%d", numOpposite), rwyFontSize, towerRunwayNum)

	// Draw connector stubs from runway entry nodes to the runway centerline.
	// Only draw connectors belonging to this runway.
	for _, node := range g.gameMap.TaxiNodes {
		if node.Type != gamemap.NodeRunwayEntry {
			continue
		}
		// Filter: only entry nodes for this runway (check if runway name contains the node's runway designator).
		if node.Runway != "" && !strings.Contains(rw.Name, node.Runway) {
			continue
		}
		nsx, nsy := g.camera.WorldToScreen(float64(node.X), float64(node.Y))
		// Project node position onto the runway centerline.
		// Nearest point on line (x1,y1)-(x2,y2) to point (node.X, node.Y).
		nx, ny := float64(node.X), float64(node.Y)
		lineLen := halfLen * 2
		if lineLen == 0 {
			continue
		}
		// Project node onto runway centerline via dot product.
		projT := ((nx-x1)*(x2-x1) + (ny-y1)*(y2-y1)) / (lineLen * lineLen)
		if projT < 0 {
			projT = 0
		}
		if projT > 1 {
			projT = 1
		}
		projX := x1 + projT*(x2-x1)
		projY := y1 + projT*(y2-y1)
		rsx, rsy := g.camera.WorldToScreen(projX, projY)

		tw := float32(g.camera.Zoom * 2)
		if tw < 1 {
			tw = 1
		}
		vector.StrokeLine(screen, float32(nsx), float32(nsy), float32(rsx), float32(rsy), tw, towerTaxiway, false)
	}
}

// --- Taxiways ---

func (g *Game) drawTowerTaxiways(screen *ebiten.Image) {
	gm := g.gameMap
	labeled := make(map[string]bool)

	for _, edge := range gm.TaxiEdges {
		from := gm.NodeByID(edge.From)
		to := gm.NodeByID(edge.To)
		if from == nil || to == nil {
			continue
		}
		sx1, sy1 := g.camera.WorldToScreen(float64(from.X), float64(from.Y))
		sx2, sy2 := g.camera.WorldToScreen(float64(to.X), float64(to.Y))

		tw := float32(g.camera.Zoom * 2)
		if tw < 1 {
			tw = 1
		}
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), tw, towerTaxiway, false)
		vector.DrawFilledCircle(screen, float32(sx1), float32(sy1), tw*0.8, towerTaxiway, false)
		vector.DrawFilledCircle(screen, float32(sx2), float32(sy2), tw*0.8, towerTaxiway, false)

		if !labeled[edge.Taxiway] {
			mx := (sx1 + sx2) / 2
			my := (sy1 + sy2) / 2
			twFont := scaledFontSize(3, g.camera.Zoom, 8, 22)
			drawLabel(screen, mx-3, my-twFont*1.5, edge.Taxiway, twFont, towerTaxiLabel)
			labeled[edge.Taxiway] = true
		}
	}
}

// --- Gates ---

func (g *Game) drawTowerGates(screen *ebiten.Image) {
	for _, gate := range g.gameMap.Gates {
		node := g.gameMap.NodeByID(gate.NodeID)
		if node == nil {
			continue
		}
		sx, sy := g.camera.WorldToScreen(float64(node.X), float64(node.Y))

		gw := float32(g.camera.ScaledSize(2.5))
		gh := float32(g.camera.ScaledSize(1.5))
		if gw < 10 {
			gw = 10
		}
		if gh < 6 {
			gh = 6
		}
		vector.DrawFilledRect(screen, float32(sx)-gw/2, float32(sy)-gh/2, gw, gh, towerGateFill, false)
		vector.StrokeRect(screen, float32(sx)-gw/2, float32(sy)-gh/2, gw, gh, 1, towerGateEdge, false)

		gateFont := scaledFontSize(3, g.camera.Zoom, 7, 20)
		drawLabel(screen, sx-gateFont*0.8, sy-gateFont*0.5, gate.ID, gateFont, towerGateLabel)
	}
}

// --- Hold-short markings ---

func (g *Game) drawTowerHoldShorts(screen *ebiten.Image) {
	for _, node := range g.gameMap.TaxiNodes {
		if node.Type != gamemap.NodeHoldShort {
			continue
		}
		sx, sy := g.camera.WorldToScreen(float64(node.X), float64(node.Y))

		barHalf := float32(g.camera.ScaledSize(1.0))
		if barHalf < 4 {
			barHalf = 4
		}
		drawDashedLine(screen,
			float32(sx)-barHalf, float32(sy),
			float32(sx)+barHalf, float32(sy),
			3, 2, 1.5, towerHoldShort)

		hsFont := scaledFontSize(2.5, g.camera.Zoom, 6, 18)
		drawLabel(screen, sx+float64(barHalf)+4, sy-hsFont*0.5, "HS", hsFont, towerHoldShort)
	}
}

// --- Live aircraft on surface ---

func (g *Game) drawTowerAircraft(screen *ebiten.Image) {
	for _, ac := range g.aircraft {
		// Airborne aircraft on approach are shown in the inset, not on the main surface.
		if ac.State == aircraft.Approaching || ac.State == aircraft.Crashed {
			continue
		}

		ix, iy := g.interpolatedPosition(ac)
		sx, sy := g.camera.WorldToScreen(ix, iy)

		// Color: departures = cyan, arrivals/ground = green, conflict = red blink.
		c := towerTarget
		if ac.State == aircraft.Departing || ac.State == aircraft.OnRunway ||
			ac.State == aircraft.Pushback || (ac.State == aircraft.AtGate && ac.AssignedRunway != "") {
			c = towerDeparture
		}
		if ac.State == aircraft.Landing {
			alpha := pulseAlpha(g.elapsed, 2*time.Second, 0.5, 1.0)
			c = colorWithAlpha(color.RGBA{0xcc, 0xcc, 0x00, 0xff}, alpha)
		}
		if isViolating(ac.Callsign, g.activeViolations) {
			if blinkVisible(g.elapsed, 500*time.Millisecond, 0.7) {
				c = towerConflict
			} else {
				c = colorWithAlpha(towerConflict, 0.3)
			}
		}

		// Chevron/symbol size scales with zoom.
		chevSize := float32(g.camera.Zoom * 4)
		if chevSize < 3 {
			chevSize = 3
		}
		if chevSize > 12 {
			chevSize = 12
		}

		// Directional chevron for moving/positioned aircraft.
		if ac.State == aircraft.Approaching || ac.State == aircraft.Landing ||
			ac.State == aircraft.Departing || ac.State == aircraft.OnRunway ||
			ac.State == aircraft.Taxiing {
			drawChevron(screen, float32(sx), float32(sy), ac.Heading, chevSize, c)
		} else {
			// Static aircraft (gate, pushback, hold short) — filled square.
			hs := chevSize * 0.6
			vector.DrawFilledRect(screen, float32(sx)-hs, float32(sy)-hs, hs*2, hs*2, c, false)
		}

		// Short leader line + callsign tag — font scales with zoom.
		acFont := scaledFontSize(3.5, g.camera.Zoom, 8, 22)
		lx := sx + acFont*1.2
		ly := sy - acFont
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(lx), float32(ly), 0.6, towerLeader, false)
		drawLabel(screen, lx+2, ly-acFont*0.4, ac.Callsign, acFont, towerDataTag)
	}
}

// --- Approach inset (shows arriving traffic on final) ---

func (g *Game) drawTowerApproachInset(screen *ebiten.Image) {
	iw, ih := float32(200), float32(100)
	ix := float32(g.width) - iw - 20
	iy := float32(20)

	// Border and label.
	vector.StrokeRect(screen, ix, iy, iw, ih, 1, towerInsetBorder, false)
	drawLabel(screen, float64(ix)+4, float64(iy)+2, "FINAL APPROACH", 7, towerInsetLabel)

	// Runway reference line inside inset.
	rwY := iy + ih - 15
	vector.StrokeLine(screen, ix+20, rwY, ix+iw-20, rwY, 2, towerRunwayEdge, false)

	// Draw approaching/landing aircraft inside the inset.
	rw := g.gameMap.PrimaryRunway()
	for _, ac := range g.aircraft {
		if ac.State != aircraft.Approaching && ac.State != aircraft.Landing {
			continue
		}

		// Map aircraft position relative to runway into the inset box.
		// The approach course is along the runway heading.
		// Distance from runway center, projected onto approach axis.
		acIX, acIY := g.interpolatedPosition(ac)
		dx := acIX - float64(rw.X)
		dy := acIY - float64(rw.Y)
		dist := math.Sqrt(dx*dx + dy*dy)

		// Map distance (0-60 cells) to inset X position.
		maxDist := float64(60)
		t := dist / maxDist
		if t > 1 {
			t = 1
		}

		// Aircraft further from runway = further left in inset.
		acX := ix + iw - 25 - float32(t)*float32(iw-50)
		acY := iy + ih/2 + float32(dy)*0.3 // slight vertical offset

		// Clamp inside inset.
		if acX < ix+5 {
			acX = ix + 5
		}
		if acY < iy+15 {
			acY = iy + 15
		}
		if acY > iy+ih-20 {
			acY = iy + ih - 20
		}

		drawChevron(screen, acX, acY, ac.Heading, 3, traconTarget)
		drawLabel(screen, float64(acX)+5, float64(acY)-3, ac.Callsign, 6, towerDataTag)
	}
}

// whitePixelImg is a 1x1 white texture used for DrawTriangles — created once.
var whitePixelImg = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(colorWhite)
	return img
}()
