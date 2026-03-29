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
	hudHeight   = 28
	inputHeight = 28
	radioHeight = 100
	radioLines  = 5
)

// drawHUD renders the top status bar.
func (g *Game) drawHUD(screen *ebiten.Image) {
	w := float32(g.width)

	// Background bar.
	vector.DrawFilledRect(screen, 0, 0, w, hudHeight, hudBg, false)

	y := float64(6)
	x := float64(10)

	// Role.
	drawLabel(screen, x, y, g.gameConfig.Role.String(), 11, hudValue)
	x += 80

	// Score.
	drawLabel(screen, x, y, "SCORE", 9, hudLabel)
	x += 45
	drawLabel(screen, x, y, fmt.Sprintf("%d", g.score), 11, hudValue)
	x += 40

	// Aircraft count.
	drawLabel(screen, x, y, "AIRCRAFT", 9, hudLabel)
	x += 65
	drawLabel(screen, x, y, fmt.Sprintf("%d", len(g.aircraft)), 11, hudValue)
	x += 30

	// Near misses.
	if g.nearMisses > 0 {
		drawLabel(screen, x, y, "NEAR MISS", 9, hudLabel)
		x += 75
		drawLabel(screen, x, y, fmt.Sprintf("%d", g.nearMisses), 11, hudWarning)
		x += 30
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

	drawLabel(screen, float64(g.width)-120, y, timeStr, 11, hudText)
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

	y := float64(baseY) + 4
	for _, msg := range messages[start:] {
		c := radioMsgColor(msg)
		timestamp := formatTimestamp(msg.Time)
		prefix := timestamp + " "

		drawLabel(screen, 10, y, prefix, 9, hudLabel)
		drawLabel(screen, 10+float64(len(prefix)*6), y, msg.Text, 9, c)
		y += 16
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

	y := float64(baseY) + 7

	// Prompt.
	drawLabel(screen, 10, y, "ATC>", 11, inputPrompt)

	// Input text.
	txt := g.input.Text()
	drawLabel(screen, 55, y, txt, 11, inputText)

	// Blinking cursor.
	if g.tickCount%10 < 6 { // blink every ~0.6s
		cursorX := 55 + float64(g.input.Cursor())*7.2 // approximate monospace width
		vector.StrokeLine(screen, float32(cursorX), float32(y), float32(cursorX), float32(y)+12, 1.5, inputCursor, false)
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

	drawLabel(screen, cx-60, cy-30, "GAME OVER", 18, gameOverTxt)
	drawLabel(screen, cx-50, cy, fmt.Sprintf("Score: %d", g.score), 14, hudText)
	drawLabel(screen, cx-70, cy+30, "R restart  |  Esc quit", 10, hudLabel)
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
