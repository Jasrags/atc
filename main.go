package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/engine"
	"github.com/Jasrags/atc/internal/game"
	"github.com/Jasrags/atc/internal/gamemap"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hajimehoshi/ebiten/v2"
	zone "github.com/lrstanley/bubblezone"
)

func main() {
	devMode := flag.Bool("dev", false, "enable developer mode (/ commands)")
	guiMode := flag.Bool("gui", false, "launch graphical Ebitengine window instead of TUI")
	guiRole := flag.String("role", "tracon", "GUI role: tracon, tower, combined")
	flag.Parse()

	if *guiMode {
		runGUI(*guiRole)
		return
	}

	runTUI(*devMode)
}

func runGUI(role string) {
	gm := gamemap.ByID("san")

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(10) // match TUI tick rate: 10 physics updates per second

	r := config.RoleTRACON
	switch role {
	case "tower":
		r = config.RoleTower
	case "combined":
		r = config.RoleCombined
	}
	// Strip any existing facility suffix from the map name for the window title.
	mapName := strings.TrimSuffix(gm.Name, " TRACON")
	ebiten.SetWindowTitle(fmt.Sprintf("ATC — %s %s", mapName, r.String()))

	g := engine.NewGame(gm, r)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatalf("ebitengine: %v", err)
	}
}

func runTUI(devMode bool) {
	zone.NewGlobal()

	m := game.NewModel(devMode)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
