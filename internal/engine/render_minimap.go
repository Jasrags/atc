package engine

import (
	"image/color"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	minimapWidth  = 140
	minimapHeight = 70
	minimapMargin = 10
)

var (
	minimapBg     = color.RGBA{0x0a, 0x0a, 0x0a, 0xcc}
	minimapBorder = color.RGBA{0x33, 0x44, 0x33, 0xff}
	minimapRwy    = color.RGBA{0x66, 0x66, 0x66, 0xff}
	minimapDot    = color.RGBA{0x00, 0xcc, 0x44, 0xff}
	minimapView   = color.RGBA{0x44, 0x88, 0x44, 0x88} // semi-transparent viewport rect
)

// drawMinimap renders a small overview map in the bottom-left corner when zoomed in.
// Shows the full map extent with aircraft as tiny dots and the current viewport outlined.
func (g *Game) drawMinimap(screen *ebiten.Image) {
	// Only show when zoomed in past 1.5x.
	if g.camera.Zoom < 1.5 {
		return
	}

	mx := float32(minimapMargin)
	my := float32(g.height) - float32(minimapHeight) - float32(inputHeight) - float32(radioHeight) - minimapMargin

	// Background.
	vector.DrawFilledRect(screen, mx, my, minimapWidth, minimapHeight, minimapBg, false)
	vector.StrokeRect(screen, mx, my, minimapWidth, minimapHeight, 1, minimapBorder, false)

	// Scale: map grid coordinates to minimap pixels.
	scaleX := float64(minimapWidth) / float64(g.gameMap.Width)
	scaleY := float64(minimapHeight) / float64(g.gameMap.Height)

	// Draw runways.
	for _, rw := range g.gameMap.Runways {
		rx := float64(mx) + float64(rw.X)*scaleX
		ry := float64(my) + float64(rw.Y)*scaleY
		vector.DrawFilledRect(screen, float32(rx)-2, float32(ry), 4, 1, minimapRwy, false)
	}

	// Draw aircraft as tiny dots.
	for _, ac := range g.aircraft {
		if ac.State == aircraft.Crashed || ac.State == aircraft.Landed {
			continue
		}
		ax := float64(mx) + ac.X*scaleX
		ay := float64(my) + ac.Y*scaleY
		vector.DrawFilledCircle(screen, float32(ax), float32(ay), 1.5, minimapDot, false)
	}

	// Draw current viewport rectangle.
	// Get the world coordinates of the screen corners.
	topLeftX, topLeftY := g.camera.ScreenToWorld(0, float64(hudHeight))
	botRightX, botRightY := g.camera.ScreenToWorld(float64(g.width-stripWidth-sidebarPadding*2),
		float64(g.height)-float64(inputHeight)-float64(radioHeight))

	// Clamp to map bounds.
	if topLeftX < 0 {
		topLeftX = 0
	}
	if topLeftY < 0 {
		topLeftY = 0
	}
	if botRightX > float64(g.gameMap.Width) {
		botRightX = float64(g.gameMap.Width)
	}
	if botRightY > float64(g.gameMap.Height) {
		botRightY = float64(g.gameMap.Height)
	}

	vx := float32(float64(mx) + topLeftX*scaleX)
	vy := float32(float64(my) + topLeftY*scaleY)
	vw := float32((botRightX - topLeftX) * scaleX)
	vh := float32((botRightY - topLeftY) * scaleY)

	vector.StrokeRect(screen, vx, vy, vw, vh, 1, minimapView, false)
}
