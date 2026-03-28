package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const maxDisplayMessages = 5

// RenderHUD builds the heads-up display with score, aircraft count, and elapsed time.
func RenderHUD(score int, aircraftCount int, elapsed time.Duration, messages []string) string {
	var sb strings.Builder

	t := table.New().
		Headers("", "SCORE", "AIRCRAFT", "TIME").
		Row("ATC", fmt.Sprintf("%d", score), fmt.Sprintf("%d", aircraftCount), formatDuration(elapsed)).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("63"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
			}
			if col == 1 {
				return lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
			}
			return lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		})

	sb.WriteString(t.Render() + "\n")

	// Recent messages
	start := 0
	if len(messages) > maxDisplayMessages {
		start = len(messages) - maxDisplayMessages
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
func RenderGameOver(score int) string {
	title := GameOverTitle.Render(" GAME OVER ")
	scoreStr := GameOverScore.Render(fmt.Sprintf("Final Score: %d", score))
	restart := Dim.Render("R restart  |  Esc menu  |  Q quit")

	content := strings.Join([]string{title, "", scoreStr, "", restart}, "\n")
	return HelpBox.Render(content)
}
