package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Jasrags/atc/internal/game"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

func main() {
	devMode := flag.Bool("dev", false, "enable developer mode (/ commands)")
	flag.Parse()

	zone.NewGlobal()

	m := game.NewModel(*devMode)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
