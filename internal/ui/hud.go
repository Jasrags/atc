package ui

import (
	"fmt"
	"strings"
	"time"
)

const maxMessages = 5

// RenderHUD builds the heads-up display with score, aircraft count, and elapsed time.
func RenderHUD(score int, aircraftCount int, elapsed time.Duration, messages []string) string {
	var sb strings.Builder

	title := HUDTitle.Render(" ATC ")
	scoreStr := HUDScore.Render(fmt.Sprintf("Score: %d", score))
	info := HUDInfo.Render(fmt.Sprintf("Aircraft: %d  Time: %s", aircraftCount, formatDuration(elapsed)))

	sb.WriteString(fmt.Sprintf("%s  %s  %s\n", title, scoreStr, info))

	// Recent messages
	start := 0
	if len(messages) > maxMessages {
		start = len(messages) - maxMessages
	}
	for _, msg := range messages[start:] {
		sb.WriteString(Dim.Render(msg) + "\n")
	}

	return sb.String()
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// FormatSuccess returns a styled success message.
func FormatSuccess(msg string) string {
	return MessageSuccess.Render("+ " + msg)
}

// FormatError returns a styled error message.
func FormatError(msg string) string {
	return MessageError.Render("! " + msg)
}

// FormatInfo returns a styled info message.
func FormatInfo(msg string) string {
	return MessageInfo.Render("> " + msg)
}

// RenderPaused builds the paused overlay.
func RenderPaused(score int, elapsed time.Duration) string {
	title := HUDTitle.Render(" PAUSED ")
	scoreStr := HUDScore.Render(fmt.Sprintf("Score: %d  Time: %s", score, formatDuration(elapsed)))
	hint := Dim.Render("P resume  |  Esc menu  |  Q quit")

	content := strings.Join([]string{title, "", scoreStr, "", hint}, "\n")
	return HelpBox.Render(content)
}

// RenderGameOver builds the game over overlay.
func RenderGameOver(score int, width, height int) string {
	title := GameOverTitle.Render(" GAME OVER ")
	scoreStr := GameOverScore.Render(fmt.Sprintf("Final Score: %d", score))
	restart := Dim.Render("R restart  |  Esc menu  |  Q quit")

	content := strings.Join([]string{title, "", scoreStr, "", restart}, "\n")

	box := HelpBox.Render(content)

	return box
}
