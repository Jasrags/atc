package radar

import (
	"fmt"
	"math"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/ui"
	zone "github.com/lrstanley/bubblezone"
)

// Render builds the ASCII radar grid with map features, aircraft, and runways.
func Render(gm gamemap.Map, planes []aircraft.Aircraft) string {
	grid := makeGrid(gm.Width, gm.Height)
	placeTaxiways(grid, gm.Width, gm.Height, gm)
	placeFixes(grid, gm.Width, gm.Height, gm.Fixes)
	for _, rw := range gm.Runways {
		placeRunway(grid, gm.Width, gm.Height, rw)
	}
	placeTrails(grid, gm.Width, gm.Height, planes)
	placeAircraft(grid, gm.Width, gm.Height, planes)
	return renderGrid(grid, gm.Width, gm.Height)
}

func makeGrid(width, height int) [][]rune {
	grid := make([][]rune, height)
	for y := range grid {
		row := make([]rune, width)
		for x := range row {
			row[x] = ' '
		}
		grid[y] = row
	}
	return grid
}

func placeTaxiways(grid [][]rune, width, height int, gm gamemap.Map) {
	// Draw taxiway edges as lines between nodes
	for _, edge := range gm.TaxiEdges {
		fromNode := gm.NodeByID(edge.From)
		toNode := gm.NodeByID(edge.To)
		if fromNode == nil || toNode == nil {
			continue
		}
		placeTaxiLine(grid, width, height, fromNode.X, fromNode.Y, toNode.X, toNode.Y)
	}

	// Draw node markers on top
	for _, node := range gm.TaxiNodes {
		if node.X < 0 || node.X >= width || node.Y < 0 || node.Y >= height {
			continue
		}
		switch node.Type {
		case gamemap.NodeGate:
			grid[node.Y][node.X] = '~' // gate marker (styled separately)
		case gamemap.NodeHoldShort:
			if grid[node.Y][node.X] == ' ' || grid[node.Y][node.X] == '-' || grid[node.Y][node.X] == '|' {
				grid[node.Y][node.X] = ':'
			}
		}
	}
}

// placeTaxiLine draws a taxiway segment between two points using - and | characters.
func placeTaxiLine(grid [][]rune, width, height, x1, y1, x2, y2 int) {
	// Horizontal segment
	if y1 == y2 {
		minX, maxX := x1, x2
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX; x <= maxX; x++ {
			if x >= 0 && x < width && y1 >= 0 && y1 < height && grid[y1][x] == ' ' {
				grid[y1][x] = '-'
			}
		}
		return
	}

	// Vertical segment
	if x1 == x2 {
		minY, maxY := y1, y2
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY; y <= maxY; y++ {
			if x1 >= 0 && x1 < width && y >= 0 && y < height && grid[y][x1] == ' ' {
				grid[y][x1] = '|'
			}
		}
		return
	}

	// Diagonal: draw L-shaped path (horizontal then vertical)
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		if x >= 0 && x < width && y1 >= 0 && y1 < height && grid[y1][x] == ' ' {
			grid[y1][x] = '-'
		}
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		if x2 >= 0 && x2 < width && y >= 0 && y < height && grid[y][x2] == ' ' {
			grid[y][x2] = '|'
		}
	}
}

func placeFixes(grid [][]rune, width, height int, fixes []gamemap.Fix) {
	for _, f := range fixes {
		if f.X < 0 || f.X >= width || f.Y < 0 || f.Y >= height {
			continue
		}

		// Place the symbol character (using ASCII fallbacks for grid runes)
		switch f.Type {
		case gamemap.FixWaypoint:
			grid[f.Y][f.X] = '^' // △
		case gamemap.FixAirport:
			grid[f.Y][f.X] = 'o' // ◎
		case gamemap.FixVOR:
			grid[f.Y][f.X] = '*' // ◉
		case gamemap.FixIntersection:
			grid[f.Y][f.X] = '+' // ✦
		}

		// Place label after symbol
		for i, ch := range f.Name {
			lx := f.X + 2 + i
			if lx < width {
				grid[f.Y][lx] = ch
			}
		}
	}
}

func placeRunway(grid [][]rune, width, height int, rw gamemap.Runway) {
	cells := runwayCells(rw)

	for _, c := range cells {
		x, y := c[0], c[1]
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = '='
		}
	}

	// Place runway numbers at each end
	// First cell and last cell define the two ends
	if len(cells) >= 2 {
		first := cells[0]
		last := cells[len(cells)-1]

		// Direction vector from first -> last
		dx := sign(last[0] - first[0])
		dy := sign(last[1] - first[1])

		// Label before the first cell (approach end opposite to heading)
		numFirst := fmt.Sprintf("%d", gamemap.RunwayNumber(rw.OppositeHeading()))
		placeLabel(grid, width, first[0]-dx*(len(numFirst)+1), first[1]-dy*(len(numFirst)+1), numFirst)

		// Label after the last cell (approach end at heading)
		numLast := fmt.Sprintf("%d", gamemap.RunwayNumber(rw.Heading))
		placeLabel(grid, width, last[0]+dx*2, last[1]+dy*2, numLast)
	}
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func placeLabel(grid [][]rune, width int, startX, y int, label string) {
	if y < 0 || y >= len(grid) {
		return
	}
	for i, ch := range label {
		x := startX + i
		if x >= 0 && x < width {
			grid[y][x] = ch
		}
	}
}

func runwayCells(rw gamemap.Runway) [][2]int {
	cells := make([][2]int, rw.Length)
	rad := float64(rw.Heading) * math.Pi / 180.0

	dx := math.Sin(rad)
	dy := -math.Cos(rad)

	for i := 0; i < rw.Length; i++ {
		offset := float64(i) - float64(rw.Length-1)/2.0
		cx := rw.X + int(math.Round(offset*dx))
		cy := rw.Y + int(math.Round(offset*dy))
		cells[i] = [2]int{cx, cy}
	}
	return cells
}

func placeTrails(grid [][]rune, width, height int, planes []aircraft.Aircraft) {
	for _, ac := range planes {
		if ac.State == aircraft.Landed || len(ac.Trail) == 0 {
			continue
		}
		for _, pos := range ac.Trail {
			x, y := pos[0], pos[1]
			if x >= 0 && x < width && y >= 0 && y < height && grid[y][x] == ' ' {
				grid[y][x] = '.'
			}
		}
	}
}

func placeAircraft(grid [][]rune, width, height int, planes []aircraft.Aircraft) {
	for _, ac := range planes {
		if ac.State == aircraft.Landed {
			continue
		}
		gx, gy := ac.GridX(), ac.GridY()
		if gx < 0 || gx >= width || gy < 0 || gy >= height {
			continue
		}

		switch ac.State {
		case aircraft.Crashed:
			grid[gy][gx] = 'X'
		case aircraft.Taxiing:
			grid[gy][gx] = 'v'
		case aircraft.AtGate:
			grid[gy][gx] = '#'
		case aircraft.Pushback:
			grid[gy][gx] = '<'
		case aircraft.HoldShort:
			grid[gy][gx] = '!'
		case aircraft.OnRunway:
			grid[gy][gx] = '>'
		default:
			grid[gy][gx] = '@'
		}

		// Callsign label (up to 5 chars)
		label := ac.Callsign
		if len(label) > 5 {
			label = label[:5]
		}
		for i, ch := range label {
			lx := gx + 1 + i
			if lx < width {
				grid[gy][lx] = ch
			}
		}
	}
}

func renderGrid(grid [][]rune, width, height int) string {
	var sb strings.Builder

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ch := grid[y][x]
			switch ch {
			case '@':
				sb.WriteString(ui.AircraftNormal.Render(string(ch)))
			case 'X':
				sb.WriteString(ui.AircraftCrashed.Render(string(ch)))
			case 'v', '<', '>', '#':
				sb.WriteString(ui.AircraftGround.Render(string(ch)))
			case '!':
				sb.WriteString(ui.AircraftHold.Render(string(ch)))
			case '=':
				sb.WriteString(ui.RunwayStyle.Render(string(ch)))
			case '^', 'o', '*', '+':
				sb.WriteString(ui.FixStyle.Render(string(ch)))
			case '-', '|':
				sb.WriteString(ui.TaxiwayStyle.Render(string(ch)))
			case '~':
				sb.WriteString(ui.GateStyle.Render("#"))
			case ':':
				sb.WriteString(ui.HoldShortStyle.Render(":"))
			case '.':
				sb.WriteString(ui.Dim.Render("."))
			default:
				sb.WriteRune(ch)
			}
		}
		if y < height-1 {
			sb.WriteRune('\n')
		}
	}

	return ui.RadarBorder.Render(sb.String())
}

// RenderFlightStrips builds the flight strip panel showing each aircraft's status.
// Each strip shows: callsign, heading, altitude (with climb/descend arrow), speed, and state.
// Format inspired by ATC-SIM progress strips.
func RenderFlightStrips(planes []aircraft.Aircraft) string {
	if len(planes) == 0 {
		return ui.SidebarBox.Render(ui.Dim.Render("  No aircraft in airspace"))
	}

	var sb strings.Builder
	sb.WriteString(ui.SidebarTitle.Render(" FLIGHT STRIPS ") + "\n")
	sb.WriteString(strings.Repeat("─", 28) + "\n")

	for _, ac := range planes {
		if ac.State == aircraft.Landed {
			continue
		}
		strip := renderStrip(ac)
		sb.WriteString(zone.Mark(ac.Callsign, strip))
		sb.WriteString(strings.Repeat("─", 28) + "\n")
	}

	return ui.SidebarBox.Render(sb.String())
}

func renderStrip(ac aircraft.Aircraft) string {
	var sb strings.Builder

	// Line 1: Callsign + State
	callsign := fmt.Sprintf("%-6s", ac.Callsign)
	state := ac.State.String()

	switch {
	case ac.State == aircraft.Landing:
		sb.WriteString(ui.AircraftLanding.Render(callsign))
	case ac.State == aircraft.Crashed:
		sb.WriteString(ui.AircraftCrashed.Render(callsign))
	case ac.State.IsGround():
		sb.WriteString(ui.AircraftGround.Render(callsign))
	default:
		sb.WriteString(ui.AircraftNormal.Render(callsign))
	}
	sb.WriteString(strings.Repeat(" ", max(16-len(callsign)-len(state), 0)))
	sb.WriteString(ui.Dim.Render(state))
	sb.WriteRune('\n')

	// Line 2: Context depends on air vs ground
	if ac.State.IsGround() {
		var info []string
		if ac.AssignedGate != "" {
			info = append(info, "Gate "+ac.AssignedGate)
		}
		if ac.AssignedRunway != "" {
			info = append(info, "Rwy "+ac.AssignedRunway)
		}
		if len(ac.TaxiRoute) > 0 {
			info = append(info, "TX "+strings.Join(ac.TaxiRoute, " "))
		}
		if len(info) > 0 {
			sb.WriteString(ui.Dim.Render(" " + strings.Join(info, " | ")))
		}
		sb.WriteRune('\n')
	} else {
		altArrow := " "
		if ac.Altitude < ac.TargetAltitude {
			altArrow = "↑"
		} else if ac.Altitude > ac.TargetAltitude {
			altArrow = "↓"
		}

		hdgStr := fmt.Sprintf("%03d", ac.Heading)
		altStr := fmt.Sprintf("%s%02d", altArrow, ac.Altitude)
		spdStr := fmt.Sprintf("S%d", ac.Speed)

		sb.WriteString(ui.HUDInfo.Render(fmt.Sprintf(" %s  %s  %s", hdgStr, altStr, spdStr)))
		sb.WriteRune('\n')

		// Line 3: Target info (if different from current)
		var targets []string
		if ac.TargetHeading != ac.Heading {
			targets = append(targets, fmt.Sprintf("H%03d", ac.TargetHeading))
		}
		if ac.TargetAltitude != ac.Altitude {
			targets = append(targets, fmt.Sprintf("A%d", ac.TargetAltitude))
		}
		if ac.TargetSpeed != ac.Speed {
			targets = append(targets, fmt.Sprintf("S%d", ac.TargetSpeed))
		}

		if len(targets) > 0 {
			sb.WriteString(ui.Dim.Render(" → " + strings.Join(targets, " ")))
		}
	}
	sb.WriteRune('\n')

	return sb.String()
}
