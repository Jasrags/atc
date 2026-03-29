package engine

import (
	"fmt"
	"image/color"
	"time"

	"github.com/Jasrags/atc/internal/radio"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HUD colors.
var (
	hudBg       = color.RGBA{0x0a, 0x0a, 0x0a, 0xdd}
	hudText     = color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
	hudLabel    = color.RGBA{0x77, 0x88, 0x77, 0xff}
	hudValue    = color.RGBA{0x00, 0xcc, 0x44, 0xff}
	hudWarning  = color.RGBA{0xff, 0xcc, 0x00, 0xff}
	inputBg     = color.RGBA{0x0a, 0x0f, 0x0a, 0xee}
	inputText   = color.RGBA{0x00, 0xff, 0x00, 0xff}
	inputPrompt = color.RGBA{0x00, 0xaa, 0x44, 0xff}
	inputCursor = color.RGBA{0x00, 0xff, 0x00, 0xff}
	radioPilot  = color.RGBA{0x00, 0xcc, 0xcc, 0xff} // cyan — pilot messages
	radioATC    = color.RGBA{0x00, 0xcc, 0x44, 0xff} // green — ATC messages
	radioSystem = color.RGBA{0x88, 0x88, 0x88, 0xff} // gray — system messages
	radioUrgent = color.RGBA{0xff, 0xcc, 0x00, 0xff} // yellow — urgent
	radioEmerg  = color.RGBA{0xff, 0x44, 0x44, 0xff} // red — emergency
	gameOverBg  = color.RGBA{0x22, 0x00, 0x00, 0xdd}
	gameOverTxt = color.RGBA{0xff, 0x44, 0x44, 0xff}
)

const (
	hudHeight   = 36
	inputHeight = 36
	radioHeight = 130
	radioLines  = 5
)

// drawHUD renders the top status bar.
func (g *Game) drawHUD(screen *ebiten.Image) {
	w := float32(g.width)

	// Background bar.
	vector.DrawFilledRect(screen, 0, 0, w, hudHeight, hudBg, false)

	y := float64(8)
	x := float64(12)

	// Role.
	drawLabel(screen, x, y, g.gameConfig.Role.String(), 16, hudValue)
	x += 100

	// Score.
	drawLabel(screen, x, y, "SCORE", 12, hudLabel)
	x += 55
	drawLabel(screen, x, y, fmt.Sprintf("%d", g.score), 16, hudValue)
	x += 45

	// Aircraft count.
	drawLabel(screen, x, y, "AIRCRAFT", 12, hudLabel)
	x += 80
	drawLabel(screen, x, y, fmt.Sprintf("%d", len(g.aircraft)), 16, hudValue)
	x += 35

	// Near misses.
	if g.nearMisses > 0 {
		drawLabel(screen, x, y, "NEAR MISS", 12, hudLabel)
		x += 90
		drawLabel(screen, x, y, fmt.Sprintf("%d", g.nearMisses), 16, hudWarning)
		x += 35
	}

	// Time.
	elapsed := g.elapsed
	m := int(elapsed.Minutes())
	s := int(elapsed.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", m, s)

	// Speed / freeze indicator.
	if g.timeFrozen {
		timeStr += " PAUSED"
	} else if g.speedMultiplier > 1 {
		timeStr += fmt.Sprintf(" %dx", g.speedMultiplier)
	}

	drawLabel(screen, float64(g.width)-160, y, timeStr, 16, hudText)
}

// drawRadioLog renders the radio message log above the input.
func (g *Game) drawRadioLog(screen *ebiten.Image) {
	baseY := float32(g.height) - inputHeight - radioHeight
	w := float32(g.width)

	// Semi-transparent background.
	vector.DrawFilledRect(screen, 0, baseY, w, float32(radioHeight), hudBg, false)

	// Separator line.
	vector.StrokeLine(screen, 0, baseY, w, baseY, 1, hudLabel, false)

	messages := g.radioLog.All()

	// Show last N messages.
	start := len(messages) - radioLines
	if start < 0 {
		start = 0
	}

	y := float64(baseY) + 6
	for _, msg := range messages[start:] {
		c := radioMsgColor(msg)
		timestamp := formatTimestamp(msg.Time)
		prefix := timestamp + " "

		drawLabel(screen, 12, y, prefix, 13, hudLabel)
		drawLabel(screen, 12+float64(len(prefix)*8), y, msg.Text, 13, c)
		y += 22
	}
}

// drawInput renders the ATC command prompt at the bottom of the screen.
func (g *Game) drawInput(screen *ebiten.Image) {
	baseY := float32(g.height) - inputHeight
	w := float32(g.width)

	// Background.
	vector.DrawFilledRect(screen, 0, baseY, w, float32(inputHeight), inputBg, false)

	// Separator line.
	vector.StrokeLine(screen, 0, baseY, w, baseY, 1, hudLabel, false)

	y := float64(baseY) + 8

	// Prompt.
	drawLabel(screen, 12, y, "ATC>", 16, inputPrompt)

	// Input text.
	txt := g.input.Text()
	drawLabel(screen, 65, y, txt, 16, inputText)

	// Blinking cursor.
	if g.tickCount%10 < 6 {
		cursorX := 65 + float64(g.input.Cursor())*9.6 // approximate monospace width at size 16
		vector.StrokeLine(screen, float32(cursorX), float32(y), float32(cursorX), float32(y)+16, 2, inputCursor, false)
	}
}

// drawGameOver renders the game over overlay.
func (g *Game) drawGameOver(screen *ebiten.Image) {
	w := float32(g.width)
	h := float32(g.height)

	// Dimmed overlay.
	vector.DrawFilledRect(screen, 0, 0, w, h, gameOverBg, false)

	cx := float64(g.width) / 2
	cy := float64(g.height) / 2

	drawLabel(screen, cx-80, cy-40, "GAME OVER", 28, gameOverTxt)
	drawLabel(screen, cx-60, cy, fmt.Sprintf("Score: %d", g.score), 20, hudText)
	drawLabel(screen, cx-100, cy+40, "R restart  |  Esc quit", 14, hudLabel)
}

func radioMsgColor(msg radio.Message) color.Color {
	switch msg.Priority {
	case radio.Emergency:
		return radioEmerg
	case radio.Urgent:
		return radioUrgent
	}
	switch msg.Direction {
	case radio.Inbound:
		return radioPilot
	case radio.Outbound:
		return radioATC
	}
	return radioSystem
}

func formatTimestamp(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
