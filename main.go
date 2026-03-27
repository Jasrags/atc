package main

import (
	"fmt"
	"os"

	"github.com/Jasrags/atc/internal/game"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := game.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
