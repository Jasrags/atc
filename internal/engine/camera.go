package engine

import (
	"math"

	"github.com/Jasrags/atc/internal/gamemap"
)

const (
	minZoom  = 0.5
	maxZoom  = 8.0
	zoomStep = 1.15 // multiplicative zoom per scroll tick
	panSpeed = 2.0  // pixels per key-press frame at zoom 1.0
)

// Camera controls the viewport transform from world (grid) coordinates to screen pixels.
type Camera struct {
	// Center of the viewport in world (grid) coordinates.
	CenterX float64
	CenterY float64

	// Zoom level: 1.0 = one grid cell is cellSize pixels.
	// Higher values zoom in (more pixels per cell).
	Zoom float64

	// Screen dimensions (updated each frame from Layout).
	screenW int
	screenH int

	// Drag state.
	dragging   bool
	dragStartX int
	dragStartY int
	dragCamX   float64
	dragCamY   float64
}

// NewCamera creates a camera centered on the given world position.
func NewCamera(centerX, centerY, zoom float64) Camera {
	return Camera{
		CenterX: centerX,
		CenterY: centerY,
		Zoom:    zoom,
	}
}

// FitMap returns a camera that fits the entire map in the screen.
func FitMap(gm gamemap.Map, screenW, screenH int) Camera {
	// Calculate zoom to fit the map with some margin.
	zoomX := float64(screenW) / (float64(gm.Width) * cellSize)
	zoomY := float64(screenH) / (float64(gm.Height) * cellSize)
	zoom := math.Min(zoomX, zoomY) * 0.9 // 90% to leave margin

	if zoom < minZoom {
		zoom = minZoom
	}

	return Camera{
		CenterX: float64(gm.Width) / 2,
		CenterY: float64(gm.Height) / 2,
		Zoom:    zoom,
		screenW: screenW,
		screenH: screenH,
	}
}

// FitSurface returns a camera centered on the airport surface bounding box.
func FitSurface(gm gamemap.Map, screenW, screenH int) Camera {
	minX, minY, maxX, maxY := surfaceBounds(gm)

	cx := float64(minX+maxX) / 2
	cy := float64(minY+maxY) / 2

	// Zoom to fit the surface area with generous margin.
	surfW := float64(maxX-minX+10) * cellSize // +10 cells padding
	surfH := float64(maxY-minY+10) * cellSize

	zoomX := float64(screenW) / surfW
	zoomY := float64(screenH) / surfH
	zoom := math.Min(zoomX, zoomY)

	if zoom < minZoom {
		zoom = minZoom
	}
	if zoom > maxZoom {
		zoom = maxZoom
	}

	return Camera{
		CenterX: cx,
		CenterY: cy,
		Zoom:    zoom,
		screenW: screenW,
		screenH: screenH,
	}
}

// WorldToScreen converts world grid coordinates to screen pixel coordinates.
func (c *Camera) WorldToScreen(worldX, worldY float64) (float64, float64) {
	// World position in base pixels (before zoom).
	px := worldX * cellSize
	py := worldY * cellSize

	// Camera center in base pixels.
	camPX := c.CenterX * cellSize
	camPY := c.CenterY * cellSize

	// Offset from camera center, scaled by zoom.
	sx := float64(c.screenW)/2 + (px-camPX)*c.Zoom
	sy := float64(c.screenH)/2 + (py-camPY)*c.Zoom

	return sx, sy
}

// ScreenToWorld converts screen pixel coordinates to world grid coordinates.
func (c *Camera) ScreenToWorld(screenX, screenY float64) (float64, float64) {
	camPX := c.CenterX * cellSize
	camPY := c.CenterY * cellSize

	px := camPX + (screenX-float64(c.screenW)/2)/c.Zoom
	py := camPY + (screenY-float64(c.screenH)/2)/c.Zoom

	return px / cellSize, py / cellSize
}

// ZoomAt zooms in/out centered on a screen position.
func (c *Camera) ZoomAt(screenX, screenY float64, factor float64) {
	// World position under the cursor before zoom.
	wx, wy := c.ScreenToWorld(screenX, screenY)

	// Apply zoom.
	c.Zoom *= factor
	if c.Zoom < minZoom {
		c.Zoom = minZoom
	}
	if c.Zoom > maxZoom {
		c.Zoom = maxZoom
	}

	// After zoom, the cursor should still point at the same world position.
	// Adjust camera center so that (wx, wy) maps back to (screenX, screenY).
	camPX := wx*cellSize - (screenX-float64(c.screenW)/2)/c.Zoom
	camPY := wy*cellSize - (screenY-float64(c.screenH)/2)/c.Zoom

	c.CenterX = camPX / cellSize
	c.CenterY = camPY / cellSize
}

// Pan moves the camera center by the given screen-space delta.
func (c *Camera) Pan(dx, dy float64) {
	c.CenterX -= dx / (c.Zoom * cellSize)
	c.CenterY -= dy / (c.Zoom * cellSize)
}

// StartDrag begins a drag operation from the given screen position.
func (c *Camera) StartDrag(screenX, screenY int) {
	c.dragging = true
	c.dragStartX = screenX
	c.dragStartY = screenY
	c.dragCamX = c.CenterX
	c.dragCamY = c.CenterY
}

// UpdateDrag updates the camera position during a drag operation.
func (c *Camera) UpdateDrag(screenX, screenY int) {
	if !c.dragging {
		return
	}
	dx := float64(screenX - c.dragStartX)
	dy := float64(screenY - c.dragStartY)
	c.CenterX = c.dragCamX - dx/(c.Zoom*cellSize)
	c.CenterY = c.dragCamY - dy/(c.Zoom*cellSize)
}

// EndDrag ends a drag operation.
func (c *Camera) EndDrag() {
	c.dragging = false
}

// ScaledSize returns a world-unit size scaled to screen pixels.
func (c *Camera) ScaledSize(worldUnits float64) float64 {
	return worldUnits * cellSize * c.Zoom
}

// surfaceBounds calculates the bounding box of all airport surface features.
func surfaceBounds(gm gamemap.Map) (minX, minY, maxX, maxY int) {
	minX, minY = gm.Width, gm.Height
	maxX, maxY = 0, 0

	for _, node := range gm.TaxiNodes {
		if node.X < minX {
			minX = node.X
		}
		if node.X > maxX {
			maxX = node.X
		}
		if node.Y < minY {
			minY = node.Y
		}
		if node.Y > maxY {
			maxY = node.Y
		}
	}

	for _, rw := range gm.Runways {
		x1 := rw.X - rw.Length/2
		x2 := rw.X + rw.Length/2
		if x1 < minX {
			minX = x1
		}
		if x2 > maxX {
			maxX = x2
		}
		if rw.Y < minY {
			minY = rw.Y
		}
		if rw.Y > maxY {
			maxY = rw.Y
		}
	}

	// Fallback if no surface features.
	if minX > maxX {
		minX = 0
		maxX = gm.Width
		minY = 0
		maxY = gm.Height
	}

	return
}
