package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/engine"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	devMode := flag.Bool("dev", false, "enable developer mode (/ commands)")
	role := flag.String("role", "tracon", "controller role: tracon, tower, combined")
	flag.Parse()

	gm := gamemap.ByID("san")

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(10)

	r := config.RoleTRACON
	switch *role {
	case "tower":
		r = config.RoleTower
	case "combined":
		r = config.RoleCombined
	}

	mapName := strings.TrimSuffix(gm.Name, " TRACON")
	ebiten.SetWindowTitle(fmt.Sprintf("ATC — %s %s", mapName, r.String()))

	g := engine.NewGame(gm, r, *devMode)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatalf("ebitengine: %v", err)
	}
}
