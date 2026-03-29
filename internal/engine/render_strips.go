package engine

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	stripWidth     = 240
	stripHeight    = 60
	stripPadding   = 3
	stripMargin    = 10 // left margin inside sidebar
	sidebarPadding = 5
)

// Strip colors.
var (
	sidebarBg      = color.RGBA{0x0a, 0x0a, 0x0a, 0xee}
	sidebarBorder  = color.RGBA{0x22, 0x33, 0x22, 0xff}
	stripBg        = color.RGBA{0x11, 0x15, 0x11, 0xff}
	stripBorder    = color.RGBA{0x22, 0x33, 0x22, 0xff}
	stripCallsign  = color.RGBA{0x00, 0xdd, 0x44, 0xff}
	stripState     = color.RGBA{0x77, 0x88, 0x77, 0xff}
	stripData      = color.RGBA{0x00, 0xaa, 0x44, 0xff}
	stripTarget    = color.RGBA{0x66, 0x77, 0x66, 0xff}
	stripLanding   = color.RGBA{0xcc, 0xcc, 0x00, 0xff}
	stripGround    = color.RGBA{0x44, 0xaa, 0xcc, 0xff}
	stripDeparture = color.RGBA{0x44, 0xcc, 0xee, 0xff}
	stripConflict  = color.RGBA{0xff, 0x44, 0x44, 0xff}
	stripPatience1 = color.RGBA{0xcc, 0xcc, 0x00, 0xff} // waiting
	stripPatience2 = color.RGBA{0xff, 0x88, 0x00, 0xff} // impatient
	stripPatience3 = color.RGBA{0xff, 0x33, 0x33, 0xff} // angry
	sectionTitle   = color.RGBA{0x55, 0x66, 0x55, 0xff}
)

// stripHit stores the bounding box of a rendered strip for click detection.
type stripHit struct {
	callsign string
	x, y     float64
	w, h     float64
}

// stripLayout holds a positioned aircraft for rendering.
type stripLayout struct {
	ac      aircraft.Aircraft
	x, y    float64
	section string // "ARRIVALS", "DEPARTURES", "FLIGHT STRIPS", or "" for items
}

// layoutStrips computes the positioned strip list without drawing.
func (g *Game) layoutStrips() ([]stripLayout, float64, float64, float64, float64) {
	sx := float64(g.width - stripWidth - sidebarPadding*2)
	sy := float64(hudHeight)
	sw := float64(stripWidth + sidebarPadding*2)
	sh := float64(g.height) - float64(hudHeight) - float64(inputHeight) - float64(radioHeight)

	planes := g.sortedStrips()
	var layout []stripLayout
	y := sy + sidebarPadding

	if g.gameConfig.Role == config.RoleTower {
		var arrivals, departures []aircraft.Aircraft
		for _, ac := range planes {
			if isDeparture(ac) {
				departures = append(departures, ac)
			} else {
				arrivals = append(arrivals, ac)
			}
		}

		if len(arrivals) > 0 {
			layout = append(layout, stripLayout{section: "ARRIVALS", x: sx + stripMargin, y: y})
			y += 18
			for _, ac := range arrivals {
				if y+stripHeight > sy+sh {
					break
				}
				layout = append(layout, stripLayout{ac: ac, x: sx + sidebarPadding, y: y})
				y += stripHeight + stripPadding
			}
		}

		if len(departures) > 0 && y+18 < sy+sh {
			y += 4
			layout = append(layout, stripLayout{section: "DEPARTURES", x: sx + stripMargin, y: y})
			y += 18
			for _, ac := range departures {
				if y+stripHeight > sy+sh {
					break
				}
				layout = append(layout, stripLayout{ac: ac, x: sx + sidebarPadding, y: y})
				y += stripHeight + stripPadding
			}
		}
	} else {
		layout = append(layout, stripLayout{section: "FLIGHT STRIPS", x: sx + stripMargin, y: y})
		y += 18
		for _, ac := range planes {
			if g.gameConfig.Role == config.RoleTRACON && skipStrip(ac) {
				continue
			}
			if y+stripHeight > sy+sh {
				break
			}
			layout = append(layout, stripLayout{ac: ac, x: sx + sidebarPadding, y: y})
			y += stripHeight + stripPadding
		}
	}

	return layout, sx, sy, sw, sh
}

// computeStripHits builds click hit areas from the current layout. Called in Update().
func (g *Game) computeStripHits() []stripHit {
	layout, _, _, _, _ := g.layoutStrips()
	var hits []stripHit
	for _, sl := range layout {
		if sl.section != "" {
			continue // section headers are not clickable
		}
		hits = append(hits, stripHit{
			callsign: sl.ac.Callsign,
			x:        sl.x, y: sl.y,
			w: float64(stripWidth), h: float64(stripHeight),
		})
	}
	return hits
}

// drawStrips renders the flight strip sidebar. Pure render — no state mutation.
func (g *Game) drawStrips(screen *ebiten.Image) {
	layout, sx, sy, sw, sh := g.layoutStrips()

	// Sidebar background.
	vector.DrawFilledRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), sidebarBg, false)
	vector.StrokeLine(screen, float32(sx), float32(sy), float32(sx), float32(sy+sh), 1, sidebarBorder, false)

	for _, sl := range layout {
		if sl.section != "" {
			drawLabel(screen, sl.x, sl.y, sl.section, 12, sectionTitle)
		} else {
			g.drawStrip(screen, sl.ac, sl.x, sl.y)
		}
	}
}

// drawStrip renders a single flight strip.
func (g *Game) drawStrip(screen *ebiten.Image, ac aircraft.Aircraft, x, y float64) {
	w := float64(stripWidth)
	h := float64(stripHeight)

	// Background.
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), stripBg, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 0.5, stripBorder, false)

	// Callsign (color-coded by state/patience).
	csColor := callsignColor(ac, g.activeViolations)
	drawLabel(screen, x+6, y+4, ac.Callsign, 14, csColor)

	// State tag (right-aligned).
	stateStr := ac.State.String()
	drawLabel(screen, x+w-50, y+4, stateStr, 11, stripState)

	// Line 2: depends on airborne vs ground.
	if ac.State.IsAirborne() {
		altArrow := " "
		if ac.Altitude < ac.TargetAltitude {
			altArrow = "\u2191"
		} else if ac.Altitude > ac.TargetAltitude {
			altArrow = "\u2193"
		}
		info := fmt.Sprintf("%03d  %s%02d  S%d", ac.Heading, altArrow, ac.Altitude, ac.Speed)
		drawLabel(screen, x+6, y+22, info, 11, stripData)

		// Line 3: targets (if different from current).
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
		if ac.HoldingFixName != "" {
			targets = append(targets, "HLD "+ac.HoldingFixName)
		} else if ac.TargetFixName != "" {
			targets = append(targets, "D "+ac.TargetFixName)
		}
		if len(targets) > 0 {
			drawLabel(screen, x+6, y+40, strings.Join(targets, " "), 10, stripTarget)
		}
	} else {
		// Ground info.
		var info string
		if ac.AssignedGate != "" {
			info += "Gate " + ac.AssignedGate
		}
		if ac.AssignedRunway != "" {
			if info != "" {
				info += " | "
			}
			info += "Rwy " + ac.AssignedRunway
		}
		if info != "" {
			drawLabel(screen, x+6, y+22, info, 11, stripData)
		}
	}

}

func callsignColor(ac aircraft.Aircraft, violations map[string]bool) color.Color {
	if isViolating(ac.Callsign, violations) {
		return stripConflict
	}
	switch {
	case ac.PatienceLevel() >= 3:
		return stripPatience3
	case ac.PatienceLevel() >= 2:
		return stripPatience2
	case ac.PatienceLevel() >= 1:
		return stripPatience1
	case ac.State == aircraft.Landing:
		return stripLanding
	case ac.State == aircraft.Departing || ac.State == aircraft.OnRunway:
		return stripDeparture
	case ac.State.IsGround():
		return stripGround
	default:
		return stripCallsign
	}
}

// isDeparture classifies an aircraft as a departure for Tower mode strip splitting.
// Departures are aircraft that were spawned at gates (heading outbound) vs arrivals
// that entered from the airspace (heading inbound to land).
func isDeparture(ac aircraft.Aircraft) bool {
	switch ac.State {
	case aircraft.Departing, aircraft.OnRunway, aircraft.Pushback, aircraft.HoldShort:
		return true
	case aircraft.AtGate:
		// AtGate aircraft are departures (spawned at gate, waiting for pushback).
		// Arrivals that reach the gate are removed before they appear in strips.
		return true
	case aircraft.Taxiing:
		// Taxiing with no gate assignment = departure heading to runway.
		// Taxiing with gate assignment = arrival heading to gate.
		return ac.AssignedGate == ""
	}
	return false
}

// skipStrip returns true if the aircraft should be hidden from strips in TRACON mode.
func skipStrip(ac aircraft.Aircraft) bool {
	switch ac.State {
	case aircraft.Taxiing, aircraft.AtGate, aircraft.Pushback, aircraft.HoldShort:
		return true
	case aircraft.Landed:
		return true
	}
	return false
}

// sortedStrips returns aircraft sorted by callsign.
func (g *Game) sortedStrips() []aircraft.Aircraft {
	planes := make([]aircraft.Aircraft, 0, len(g.aircraft))
	for _, ac := range g.aircraft {
		if ac.State == aircraft.Crashed || ac.State == aircraft.Landed {
			continue
		}
		planes = append(planes, ac)
	}
	sort.Slice(planes, func(i, j int) bool {
		return planes[i].Callsign < planes[j].Callsign
	})
	return planes
}
