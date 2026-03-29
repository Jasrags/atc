package engine

import (
	"fmt"
	"math"

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

	sx1, sy1 := worldToScreen(x1, y1)
	sx2, sy2 := worldToScreen(x2, y2)

	// Perpendicular direction for width.
	perpX := dy
	perpY := dx
	runwayWidth := float64(cellSize) * 1.8

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
	screen.DrawTriangles(vertices, indices, whitePixel(), &ebiten.DrawTrianglesOptions{})

	// Runway outline.
	vector.StrokeLine(screen, c1x, c1y, c2x, c2y, 1, towerRunwayEdge, false)
	vector.StrokeLine(screen, c2x, c2y, c3x, c3y, 1, towerRunwayEdge, false)
	vector.StrokeLine(screen, c3x, c3y, c4x, c4y, 1, towerRunwayEdge, false)
	vector.StrokeLine(screen, c4x, c4y, c1x, c1y, 1, towerRunwayEdge, false)

	// Dashed centerline.
	drawDashedLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 4, 3, 0.8, towerRunwayEdge)

	// Threshold markings.
	threshW := float32(hw * 0.8)
	vector.StrokeLine(screen, float32(sx1)+float32(perpX)*threshW, float32(sy1)+float32(perpY)*threshW,
		float32(sx1)-float32(perpX)*threshW, float32(sy1)-float32(perpY)*threshW, 2, towerThreshold, false)
	vector.StrokeLine(screen, float32(sx2)+float32(perpX)*threshW, float32(sy2)+float32(perpY)*threshW,
		float32(sx2)-float32(perpX)*threshW, float32(sy2)-float32(perpY)*threshW, 2, towerThreshold, false)

	// Runway numbers at each threshold.
	numApproach := gamemap.RunwayNumber(rw.Heading)
	numOpposite := gamemap.RunwayNumber(rw.OppositeHeading())

	drawLabel(screen, sx2+dx*cellSize*2+2, sy2-dy*cellSize*2-5,
		fmt.Sprintf("%d", numApproach), 11, towerRunwayNum)
	drawLabel(screen, sx1-dx*cellSize*2-18, sy1+dy*cellSize*2-5,
		fmt.Sprintf("%d", numOpposite), 11, towerRunwayNum)
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
		sx1, sy1 := worldToScreen(float64(from.X), float64(from.Y))
		sx2, sy2 := worldToScreen(float64(to.X), float64(to.Y))

		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 3, towerTaxiway, false)
		vector.DrawFilledCircle(screen, float32(sx1), float32(sy1), 2, towerTaxiway, false)
		vector.DrawFilledCircle(screen, float32(sx2), float32(sy2), 2, towerTaxiway, false)

		if !labeled[edge.Taxiway] {
			mx := (sx1 + sx2) / 2
			my := (sy1 + sy2) / 2
			drawLabel(screen, mx-3, my-12, edge.Taxiway, 8, towerTaxiLabel)
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
		sx, sy := worldToScreen(float64(node.X), float64(node.Y))

		gw, gh := float32(20), float32(12)
		vector.DrawFilledRect(screen, float32(sx)-gw/2, float32(sy)-gh/2, gw, gh, towerGateFill, false)
		vector.StrokeRect(screen, float32(sx)-gw/2, float32(sy)-gh/2, gw, gh, 1, towerGateEdge, false)

		drawLabel(screen, sx-6, sy-4, gate.ID, 7, towerGateLabel)
	}
}

// --- Hold-short markings ---

func (g *Game) drawTowerHoldShorts(screen *ebiten.Image) {
	for _, node := range g.gameMap.TaxiNodes {
		if node.Type != gamemap.NodeHoldShort {
			continue
		}
		sx, sy := worldToScreen(float64(node.X), float64(node.Y))

		barHalf := float32(8)
		drawDashedLine(screen,
			float32(sx)-barHalf, float32(sy),
			float32(sx)+barHalf, float32(sy),
			3, 2, 1.5, towerHoldShort)

		drawLabel(screen, sx+10, sy-4, "HS", 6, towerHoldShort)
	}
}

// --- Live aircraft on surface ---

func (g *Game) drawTowerAircraft(screen *ebiten.Image) {
	for _, ac := range g.aircraft {
		// Airborne aircraft on approach are shown in the inset, not on the main surface.
		if ac.State == aircraft.Approaching || ac.State == aircraft.Crashed {
			continue
		}

		sx, sy := worldToScreen(ac.X, ac.Y)

		// Color: departures = cyan, arrivals/ground = green, conflict = red.
		c := towerTarget
		if ac.State == aircraft.Departing || ac.State == aircraft.OnRunway ||
			ac.State == aircraft.Pushback || (ac.State == aircraft.AtGate && ac.AssignedRunway != "") {
			c = towerDeparture
		}
		if isViolating(ac.Callsign, g.activeViolations) {
			c = towerConflict
		}

		// Directional chevron for moving/positioned aircraft.
		if ac.State == aircraft.Approaching || ac.State == aircraft.Landing ||
			ac.State == aircraft.Departing || ac.State == aircraft.OnRunway ||
			ac.State == aircraft.Taxiing {
			drawChevron(screen, float32(sx), float32(sy), ac.Heading, 5, c)
		} else {
			// Static aircraft (gate, pushback, hold short) — filled square.
			vector.DrawFilledRect(screen, float32(sx)-3, float32(sy)-3, 6, 6, c, false)
		}

		// Short leader line + callsign tag.
		lx := sx + 10
		ly := sy - 8
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(lx), float32(ly), 0.6, towerLeader, false)
		drawLabel(screen, lx+2, ly-4, ac.Callsign, 8, towerDataTag)
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
		dx := ac.X - float64(rw.X)
		dy := ac.Y - float64(rw.Y)
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

// --- Helper ---

func whitePixel() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(colorWhite)
	return img
}
