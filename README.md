# ATC - Terminal Air Traffic Control

A terminal-based air traffic control game built with Go and the [Charm](https://charm.sh) TUI stack.

Guide aircraft safely to the runway using classic ATC text commands. Avoid collisions. Land as many planes as you can before the skies get too crowded.

## Gameplay

Aircraft appear at the edges of your radar screen flying toward the center. You issue commands to change their heading, altitude, and speed, then clear them for landing when they're aligned with the runway.

- **Collision** (same position + same altitude) = game over
- **Successful landing** = +1 score
- **Difficulty ramps** over time: more aircraft, faster spawns

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

- Go 1.25+
- Terminal with at least 60x24 cells

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

## Keybindings

| Key | Action |
|-----|--------|
| `Enter` | Submit command |
| `P` | Pause / Resume |
| `?` | Toggle help |
| `R` | Restart (game over screen) |
| `Esc` / `Ctrl+C` | Quit |

## Development

```bash
make          # Format, vet, test, build
make run      # Build and run
make watch    # Live reload (requires air)
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
  game/                    Bubbletea model, game loop
  aircraft/                Aircraft type, movement physics, spawner
  command/                 Command parser and executor
  collision/               Collision detection
  runway/                  Runway definition and landing validation
  radar/                   ASCII radar grid renderer
  ui/                      Lipgloss styles, HUD, help overlay
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

## License

MIT
