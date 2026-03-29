package engine

import (
	"bytes"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gomono"
)

// cellSize is the pixel size of one game grid cell.
const cellSize = 8

// --- Font ---

var monoFaceSource *text.GoTextFaceSource

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(gomono.TTF))
	if err != nil {
		log.Fatalf("failed to load font: %v", err)
	}
	monoFaceSource = s
}

// colorWhite is used as the base texture for DrawTriangles (vertex colors override it).
var colorWhite = color.RGBA{0xff, 0xff, 0xff, 0xff}

// --- TRACON (STARS) palette ---

var (
	traconBg        = color.RGBA{0x0a, 0x0f, 0x0a, 0xff} // near-black green tint
	traconGrid      = color.RGBA{0x15, 0x22, 0x15, 0xff} // very dim grid dots
	traconRangeRing = color.RGBA{0x1a, 0x28, 0x1a, 0xff} // dim range rings
	traconRunway    = color.RGBA{0xaa, 0xbb, 0xdd, 0xff} // pale blue-white centerline
	traconRunwayNum = color.RGBA{0x88, 0x99, 0xbb, 0xff} // dimmer blue-white
	traconApproach  = color.RGBA{0x44, 0x55, 0x77, 0xff} // dim dashed approach course
	traconFix       = color.RGBA{0x33, 0x55, 0x88, 0xff} // dim blue fix symbols
	traconFixLabel  = color.RGBA{0x2a, 0x44, 0x6a, 0xff} // dimmer blue fix text
	traconTarget    = color.RGBA{0x00, 0xe8, 0x50, 0xff} // bright green aircraft blip
	traconDataBlock = color.RGBA{0x00, 0xcc, 0x44, 0xff} // green data block text
	traconLeader    = color.RGBA{0x00, 0x88, 0x33, 0xff} // dim green leader line
	traconTrail     = color.RGBA{0x00, 0x88, 0x33, 0xff} // history trail (base, faded per dot)
	traconLanding   = color.RGBA{0xcc, 0xcc, 0x00, 0xff} // yellow for cleared-to-land
	traconConflict  = color.RGBA{0xff, 0x33, 0x33, 0xff} // bright red conflict/alert
)

// --- Tower (ASDE-X) palette ---

var (
	towerBg          = color.RGBA{0x05, 0x05, 0x08, 0xff} // pure dark
	towerRunwayFill  = color.RGBA{0x55, 0x55, 0x55, 0xff} // medium gray runway surface
	towerRunwayEdge  = color.RGBA{0x88, 0x88, 0x88, 0xff} // lighter gray runway outline
	towerRunwayNum   = color.RGBA{0xcc, 0xcc, 0xcc, 0xff} // bright white runway numbers
	towerThreshold   = color.RGBA{0xcc, 0xcc, 0xcc, 0xff} // white threshold marks
	towerTaxiway     = color.RGBA{0x44, 0x44, 0x50, 0xff} // dark gray taxiway paths
	towerTaxiLabel   = color.RGBA{0x55, 0x55, 0x66, 0xff} // dim taxiway letter labels
	towerGateFill    = color.RGBA{0x22, 0x33, 0x44, 0xff} // dark blue-gray gate rectangle
	towerGateEdge    = color.RGBA{0x44, 0x66, 0x88, 0xff} // lighter blue gate outline
	towerGateLabel   = color.RGBA{0x55, 0x88, 0xaa, 0xff} // blue gate text
	towerHoldShort   = color.RGBA{0xcc, 0xaa, 0x22, 0xff} // yellow hold-short bars
	towerTarget      = color.RGBA{0x00, 0xe8, 0x50, 0xff} // bright green ground target
	towerDeparture   = color.RGBA{0x44, 0xcc, 0xee, 0xff} // cyan departure chevron
	towerDataTag     = color.RGBA{0x00, 0xcc, 0x44, 0xff} // green data tag text
	towerLeader      = color.RGBA{0x00, 0x77, 0x33, 0xff} // dim leader line
	towerConflict    = color.RGBA{0xff, 0x33, 0x33, 0xff} // bright red incursion
	towerInsetBorder = color.RGBA{0x33, 0x44, 0x55, 0xff} // dim border for approach inset
	towerInsetLabel  = color.RGBA{0x55, 0x66, 0x77, 0xff} // dim inset label text
)

// --- Shared drawing helpers ---

// drawLabel draws text at (x, y) with the given font size and color.
func drawLabel(screen *ebiten.Image, x, y float64, s string, size float64, c color.Color) {
	face := &text.GoTextFace{Source: monoFaceSource, Size: size}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(screen, s, face, op)
}

// drawCircle draws an unfilled circle with the given number of segments.
func drawCircle(screen *ebiten.Image, cx, cy, r float32, c color.Color) {
	segments := 24
	for i := 0; i < segments; i++ {
		a1 := float64(i) / float64(segments) * 2 * math.Pi
		a2 := float64(i+1) / float64(segments) * 2 * math.Pi
		x1 := cx + r*float32(math.Cos(a1))
		y1 := cy + r*float32(math.Sin(a1))
		x2 := cx + r*float32(math.Cos(a2))
		y2 := cy + r*float32(math.Sin(a2))
		vector.StrokeLine(screen, x1, y1, x2, y2, 1, c, false)
	}
}

// drawDashedLine draws a dashed line between two points.
func drawDashedLine(screen *ebiten.Image, x1, y1, x2, y2, dashLen, gapLen, width float32, c color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if length == 0 {
		return
	}
	nx := dx / length
	ny := dy / length

	pos := float32(0)
	for pos < length {
		end := pos + dashLen
		if end > length {
			end = length
		}
		sx := x1 + nx*pos
		sy := y1 + ny*pos
		ex := x1 + nx*end
		ey := y1 + ny*end
		vector.StrokeLine(screen, sx, sy, ex, ey, width, c, false)
		pos = end + gapLen
	}
}

// drawChevron draws a small directional chevron (triangle) pointing in the given heading.
func drawChevron(screen *ebiten.Image, cx, cy float32, heading int, size float32, c color.Color) {
	rad := float64(heading) * math.Pi / 180.0
	sin := float32(math.Sin(rad))
	cos := float32(-math.Cos(rad))

	// Tip (front of aircraft)
	tx := cx + sin*size
	ty := cy + cos*size

	// Two rear points (perpendicular to heading, behind center)
	perpX := cos
	perpY := -sin
	halfW := size * 0.6

	rx := cx - sin*size*0.5 + perpX*halfW
	ry := cy - cos*size*0.5 + perpY*halfW
	lx := cx - sin*size*0.5 - perpX*halfW
	ly := cy - cos*size*0.5 - perpY*halfW

	vector.StrokeLine(screen, tx, ty, rx, ry, 1.5, c, false)
	vector.StrokeLine(screen, rx, ry, lx, ly, 1.5, c, false)
	vector.StrokeLine(screen, lx, ly, tx, ty, 1.5, c, false)
}
