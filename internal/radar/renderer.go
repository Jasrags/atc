package radar

import (
	"fmt"
	"math"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/ui"
)

// Render builds the ASCII radar grid with map features, aircraft, and runways.
func Render(gm gamemap.Map, planes []aircraft.Aircraft) string {
	grid := makeGrid(gm.Width, gm.Height)
	placeFixes(grid, gm.Width, gm.Height, gm.Fixes)
	for _, rw := range gm.Runways {
		placeRunway(grid, gm.Width, gm.Height, rw)
	}
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

	sb.WriteString("+" + strings.Repeat("-", width) + "+\n")

	for y := 0; y < height; y++ {
		sb.WriteRune('|')
		for x := 0; x < width; x++ {
			ch := grid[y][x]
			switch ch {
			case '@':
				sb.WriteString(ui.AircraftNormal.Render(string(ch)))
			case 'X':
				sb.WriteString(ui.AircraftCrashed.Render(string(ch)))
			case '=':
				sb.WriteString(ui.RunwayStyle.Render(string(ch)))
			case '^', 'o', '*', '+':
				sb.WriteString(ui.FixStyle.Render(string(ch)))
			default:
				sb.WriteRune(ch)
			}
		}
		sb.WriteString("|\n")
	}

	sb.WriteString("+" + strings.Repeat("-", width) + "+")

	return sb.String()
}

// RenderSidebar builds the aircraft info panel.
func RenderSidebar(planes []aircraft.Aircraft) string {
	if len(planes) == 0 {
		return ui.SidebarBox.Render(ui.Dim.Render("No aircraft"))
	}

	var sb strings.Builder
	sb.WriteString(ui.SidebarTitle.Render("  AIRCRAFT  ") + "\n")
	sb.WriteString(fmt.Sprintf(" %-6s %3s %3s %1s %4s\n", "CALL", "HDG", "ALT", "S", "ST"))
	sb.WriteString(strings.Repeat("-", 24) + "\n")

	for _, ac := range planes {
		if ac.State == aircraft.Landed {
			continue
		}
		line := fmt.Sprintf(" %-6s %03d  %02d %d %4s",
			ac.Callsign, ac.Heading, ac.Altitude, ac.Speed, ac.State)

		switch ac.State {
		case aircraft.Landing:
			sb.WriteString(ui.AircraftLanding.Render(line))
		case aircraft.Crashed:
			sb.WriteString(ui.AircraftCrashed.Render(line))
		default:
			sb.WriteString(line)
		}
		sb.WriteRune('\n')
	}

	return ui.SidebarBox.Render(sb.String())
}
