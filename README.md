# ATC - Terminal Air Traffic Control

A terminal-based air traffic control game built with Go and the [Charm](https://charm.sh) TUI stack.

Guide aircraft safely to the runway using classic ATC text commands. Avoid collisions. Land as many planes as you can before the skies get too crowded.

**[Read the full player guide](GUIDE.md)** for detailed gameplay instructions, strategies, and tips.

## Quick Start

```bash
go install github.com/Jasrags/atc@latest
atc
```

Or clone and build:

```bash
git clone https://github.com/Jasrags/atc.git
cd atc
make run
```

## Requirements

- Go 1.26+
- Terminal with at least 60x24 cells (larger recommended for full-size maps)

## Maps

| Map | Size | Runways | Description |
|-----|------|---------|-------------|
| San Diego TRACON | 120x50 | 9/27 | Coastal approach with 16 nav fixes |
| Chicago O'Hare | 120x50 | 10L/28R, 10R/28L | Parallel runway operations |
| Tutorial | 90x40 | 9/27 | Small map for learning the basics |

## Commands

```
<CALLSIGN> <CMD> [<CMD>...]
```

| Command | Description | Example |
|---------|-------------|---------|
| `H<0-359>` | Set heading | `AA123 H270` |
| `A<1-40>` | Set altitude (thousands of ft) | `AA123 A3` |
| `S<1-5>` | Set speed | `AA123 S2` |
| `L` | Clear to land | `AA123 L` |

Combine commands: `AA123 H270 A3 S2`

Click a flight strip to auto-fill the callsign — then just type the command.

## Keybindings

| Key | Action |
|-----|--------|
| Enter | Submit command |
| P | Pause / Resume |
| ? | Help |
| Esc | Back to menu |
| R | Restart (game over) |
| Ctrl+C | Quit |
| Click strip | Auto-fill callsign |

## Game Setup

New Game opens a setup screen with: **Map**, **Difficulty** (Easy/Normal/Hard), **Game Mode**, **Callsign Style** (ICAO/Short), and **Plane Trails** (On/Off). See the [player guide](GUIDE.md#game-setup) for details.

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

## Project Structure

```
main.go                    Entry point
internal/
  game/                    Bubbletea model, game loop, keybindings
  aircraft/                Aircraft type, movement physics, spawner
  command/                 Command parser and executor
  collision/               Collision detection
  config/                  GameConfig, difficulty params, enums
  gamemap/                 Map definitions (runways, fixes, registry)
  heading/                 Shared heading math
  runway/                  Runway landing validation
  radar/                   ASCII radar grid renderer, flight strips
  ui/                      Lipgloss styles, HUD, help, menu, setup screen
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components (textinput, viewport, help, key, stopwatch)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling and table rendering
- [Bubble Zone](https://github.com/lrstanley/bubblezone) - Mouse click zones for aircraft selection

## License

MIT
