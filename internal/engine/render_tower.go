package engine

import (
	"fmt"
	"math"

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
	g.drawTowerApproachInset(screen)
	g.drawTowerDemoAircraft(screen)
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

	// Runway endpoints in world coords.
	x1 := float64(rw.X) - halfLen*dx
	y1 := float64(rw.Y) - halfLen*dy
	x2 := float64(rw.X) + halfLen*dx
	y2 := float64(rw.Y) + halfLen*dy

	sx1, sy1 := worldToScreen(x1, y1)
	sx2, sy2 := worldToScreen(x2, y2)

	// Perpendicular direction for width.
	perpX := dy
	perpY := dx
	runwayWidth := float64(cellSize) * 1.8 // ~14px wide

	// Draw filled runway rectangle using two thick triangles.
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
	vertices := []ebiten.Vertex{
		{DstX: c1x, DstY: c1y, ColorR: float32(towerRunwayFill.R) / 255, ColorG: float32(towerRunwayFill.G) / 255, ColorB: float32(towerRunwayFill.B) / 255, ColorA: 1},
		{DstX: c2x, DstY: c2y, ColorR: float32(towerRunwayFill.R) / 255, ColorG: float32(towerRunwayFill.G) / 255, ColorB: float32(towerRunwayFill.B) / 255, ColorA: 1},
		{DstX: c3x, DstY: c3y, ColorR: float32(towerRunwayFill.R) / 255, ColorG: float32(towerRunwayFill.G) / 255, ColorB: float32(towerRunwayFill.B) / 255, ColorA: 1},
		{DstX: c4x, DstY: c4y, ColorR: float32(towerRunwayFill.R) / 255, ColorG: float32(towerRunwayFill.G) / 255, ColorB: float32(towerRunwayFill.B) / 255, ColorA: 1},
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

	// Threshold markings — short perpendicular bars at each end.
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

// --- Taxiways (gray paths with letter labels) ---

func (g *Game) drawTowerTaxiways(screen *ebiten.Image) {
	gm := g.gameMap

	// Track which taxiway names we've already labeled (to avoid duplicates).
	labeled := make(map[string]bool)

	for _, edge := range gm.TaxiEdges {
		from := gm.NodeByID(edge.From)
		to := gm.NodeByID(edge.To)
		if from == nil || to == nil {
			continue
		}
		sx1, sy1 := worldToScreen(float64(from.X), float64(from.Y))
		sx2, sy2 := worldToScreen(float64(to.X), float64(to.Y))

		// Taxiway path — wider than TRACON style.
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 3, towerTaxiway, false)

		// Intersection dots.
		vector.DrawFilledCircle(screen, float32(sx1), float32(sy1), 2, towerTaxiway, false)
		vector.DrawFilledCircle(screen, float32(sx2), float32(sy2), 2, towerTaxiway, false)

		// Taxiway label at midpoint (once per taxiway name).
		if !labeled[edge.Taxiway] {
			mx := (sx1 + sx2) / 2
			my := (sy1 + sy2) / 2
			drawLabel(screen, mx-3, my-12, edge.Taxiway, 8, towerTaxiLabel)
			labeled[edge.Taxiway] = true
		}
	}
}

// --- Gates (labeled rectangles) ---

func (g *Game) drawTowerGates(screen *ebiten.Image) {
	for _, gate := range g.gameMap.Gates {
		node := g.gameMap.NodeByID(gate.NodeID)
		if node == nil {
			continue
		}
		sx, sy := worldToScreen(float64(node.X), float64(node.Y))

		// Gate rectangle.
		gw, gh := float32(20), float32(12)
		vector.DrawFilledRect(screen, float32(sx)-gw/2, float32(sy)-gh/2, gw, gh, towerGateFill, false)
		vector.StrokeRect(screen, float32(sx)-gw/2, float32(sy)-gh/2, gw, gh, 1, towerGateEdge, false)

		// Gate label centered.
		drawLabel(screen, sx-6, sy-4, gate.ID, 7, towerGateLabel)
	}
}

// --- Hold-short markings (yellow dashed bars across taxiway) ---

func (g *Game) drawTowerHoldShorts(screen *ebiten.Image) {
	for _, node := range g.gameMap.TaxiNodes {
		if node.Type != gamemap.NodeHoldShort {
			continue
		}
		sx, sy := worldToScreen(float64(node.X), float64(node.Y))

		// Draw perpendicular dashed yellow bar.
		barHalf := float32(8)
		drawDashedLine(screen,
			float32(sx)-barHalf, float32(sy),
			float32(sx)+barHalf, float32(sy),
			3, 2, 1.5, towerHoldShort)

		// Small label.
		drawLabel(screen, sx+10, sy-4, "HS", 6, towerHoldShort)
	}
}

// --- Approach inset (shows arriving traffic) ---

func (g *Game) drawTowerApproachInset(screen *ebiten.Image) {
	// Inset in the upper-right area of the screen.
	iw, ih := float32(200), float32(100)
	ix := float32(g.width) - iw - 20
	iy := float32(20)

	// Border.
	vector.StrokeRect(screen, ix, iy, iw, ih, 1, towerInsetBorder, false)

	// Label.
	drawLabel(screen, float64(ix)+4, float64(iy)+2, "FINAL APPROACH", 7, towerInsetLabel)

	// Runway reference line inside inset.
	rwY := iy + ih - 15
	vector.StrokeLine(screen, ix+20, rwY, ix+iw-20, rwY, 2, towerRunwayEdge, false)

	// Demo approach aircraft inside inset.
	drawChevron(screen, ix+iw/2-30, iy+ih/2, 270, 4, traconTarget)
	drawLabel(screen, float64(ix+iw/2-30)+6, float64(iy+ih/2)-4, "UA159", 6, towerDataTag)

	drawChevron(screen, ix+iw/2+20, iy+ih/2-15, 268, 4, traconTarget)
	drawLabel(screen, float64(ix+iw/2+20)+6, float64(iy+ih/2-15)-4, "AA341", 6, towerDataTag)
}

// --- Demo aircraft on surface ---

func (g *Game) drawTowerDemoAircraft(screen *ebiten.Image) {
	demos := []struct {
		x, y     float64
		callsign string
		heading  int
		isDep    bool // departure = cyan, arrival = green
	}{
		{x: 57, y: 26, callsign: "BA221", heading: 270, isDep: true},  // on runway
		{x: 52, y: 28, callsign: "SW482", heading: 270, isDep: false}, // taxiing on A
		{x: 55, y: 30, callsign: "DL789", heading: 180, isDep: true},  // at gate G3 pushing back
	}

	for _, ac := range demos {
		sx, sy := worldToScreen(ac.x, ac.y)

		c := towerTarget
		if ac.isDep {
			c = towerDeparture
		}

		// Directional chevron.
		drawChevron(screen, float32(sx), float32(sy), ac.heading, 5, c)

		// Short leader line + callsign tag.
		lx := sx + 10
		ly := sy - 8
		vector.StrokeLine(screen, float32(sx), float32(sy), float32(lx), float32(ly), 0.6, towerLeader, false)
		drawLabel(screen, lx+2, ly-4, ac.callsign, 8, towerDataTag)
	}
}

// --- Helper ---

// whitePixel returns a 1x1 white pixel image for use with DrawTriangles.
// The actual color comes from the vertex color fields.
func whitePixel() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(colorWhite)
	return img
}
