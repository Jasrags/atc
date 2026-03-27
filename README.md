# ATC - Terminal Air Traffic Control

A terminal-based air traffic control game built with Go and the [Charm](https://charm.sh) TUI stack.

Guide aircraft safely to the runway using classic ATC text commands. Avoid collisions. Land as many planes as you can before the skies get too crowded.

## Gameplay

Aircraft appear at the edges of your radar screen flying toward the center. You issue commands to change their heading, altitude, and speed, then clear them for landing when they're aligned with the runway.

- **Collision** (same position + same altitude) = game over
- **Successful landing** = +1 score
- **Difficulty ramps** over time: more aircraft, faster spawns

## Maps

| Map | Size | Runways | Description |
|-----|------|---------|-------------|
| San Diego TRACON | 120x50 | 9/27 | Coastal approach with 16 nav fixes |
| Chicago O'Hare | 120x50 | 10L/28R, 10R/28L | Parallel runway operations |
| Tutorial | 90x40 | 9/27 | Small map for learning the basics |

## Installation

```bash
go install github.com/Jasrags/atc@latest
```

Or clone and build:

```bash
git clone https://github.com/Jasrags/atc.git
cd atc
make build
./atc
```

## Requirements

- Go 1.26+
- Terminal with at least 60x24 cells (larger recommended for full-size maps)

## Game Setup

When starting a new game, a setup screen lets you configure:

| Setting | Options | Effect |
|---------|---------|--------|
| **Map** | San Diego, Chicago, Tutorial | Radar size, runways, nav fixes |
| **Difficulty** | Easy / Normal / Hard | Spawn rate, max aircraft, speed range |
| **Game Mode** | Arrivals Only | Traffic type |
| **Callsign Style** | ICAO (AA123) / Short (A12) | Callsign format — short is faster to type |
| **Plane Trails** | On / Off | Show last 5 positions as dots behind aircraft |

Navigate with Tab (sections), Up/Down (options), Enter (start).

## Commands

Type commands in the input bar and press Enter:

```
<CALLSIGN> <CMD> [<CMD>...]
```

| Command | Description | Example |
|---------|-------------|---------|
| `H<0-359>` | Set heading | `AA123 H270` |
| `A<1-40>` | Set altitude (thousands of ft) | `AA123 A3` |
| `S<1-5>` | Set speed | `AA123 S2` |
| `L` | Clear to land | `AA123 L` |

Commands can be combined: `AA123 H270 A3 S2`

### Landing Requirements

To land an aircraft, it must be:
- Cleared with the `L` command
- Within 2 cells of the runway
- Heading within +/-10 degrees of the runway heading
- At altitude 1 (1000ft)

## Flight Strips

The right panel shows flight strips for each aircraft:

```
──────────────────────────────
AA123           APPR
 090  ↓08  S3
 → H270 A3
──────────────────────────────
```

- Line 1: Callsign (color-coded) + state (APPR/LAND/CRASH)
- Line 2: Current heading, altitude with climb/descend arrow, speed
- Line 3: Pending target commands (shown only when different from current)

## Keybindings

| Key | Action |
|-----|--------|
| `Enter` | Submit command |
| `P` | Pause / Resume |
| `?` | Toggle help |
| `Esc` | Back to menu |
| `R` | Restart (game over screen) |
| `Ctrl+C` | Quit |

## Development

```bash
make          # Format, vet, test, build
make run      # Build and run
make watch    # Rebuild on code changes (requires entr)
make test     # Run tests
make test-race # Run tests with race detector
make cover    # Coverage summary
make lint     # Run staticcheck
make help     # Show all make targets
```

### Watch Mode

Since this is a TUI app (needs direct terminal access), `make watch` rebuilds the binary on code changes but doesn't auto-run. Use two terminals:

```bash
# Terminal 1: watch and rebuild
brew install entr
make watch

# Terminal 2: run the game (re-run after rebuilds)
./atc
```

## Project Structure

```
main.go                    Entry point
internal/
  game/                    Bubbletea model, game loop, screen management
  aircraft/                Aircraft type, movement physics, spawner
  command/                 Command parser and executor
  collision/               Collision detection
  config/                  GameConfig, difficulty params, enums
  gamemap/                 Map definitions (runways, fixes, registry)
  heading/                 Shared heading math (Delta, AbsDelta)
  runway/                  Runway landing validation
  radar/                   ASCII radar grid renderer, flight strips
  ui/                      Lipgloss styles, HUD, help, menu, setup screen
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## License

MIT
