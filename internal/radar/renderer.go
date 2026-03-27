package radar

import (
	"fmt"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/runway"
	"github.com/Jasrags/atc/internal/ui"
)

// Render builds the ASCII radar grid with aircraft and runway positioned on it.
func Render(width, height int, rw runway.Runway, planes []aircraft.Aircraft) string {
	grid := makeGrid(width, height)
	placeRunway(grid, width, height, rw)
	placeAircraft(grid, width, height, planes)
	return renderGrid(grid, width, height)
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

func placeRunway(grid [][]rune, width, height int, rw runway.Runway) {
	cells := rw.Cells()
	for i, c := range cells {
		x, y := c[0], c[1]
		if x >= 0 && x < width && y >= 0 && y < height {
			ch := '='
			if i == 0 {
				ch = '['
			} else if i == len(cells)-1 {
				ch = ']'
			}
			grid[y][x] = ch
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

		// Place aircraft symbol
		switch ac.State {
		case aircraft.Crashed:
			grid[gy][gx] = 'X'
		default:
			grid[gy][gx] = '@'
		}

		// Place callsign label (up to 3 chars after the symbol)
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

	// Top border
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
			case '=', '[', ']':
				sb.WriteString(ui.RunwayStyle.Render(string(ch)))
			default:
				sb.WriteRune(ch)
			}
		}
		sb.WriteString("|\n")
	}

	// Bottom border
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
