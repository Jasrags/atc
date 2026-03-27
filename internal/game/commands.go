package game

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const tickInterval = 100 * time.Millisecond // 10 FPS

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
