# ATC - Terminal Air Traffic Control Game

## Project Overview

A terminal-based air traffic control simulation built in Go using the charmbracelet TUI stack (bubbletea, bubbles, lipgloss). Players guide aircraft from radar screen edges to a runway using classic ATC text commands.

## Tech Stack

- **Language**: Go 1.25+
- **TUI Framework**: charmbracelet/bubbletea (Elm Architecture: Model/Update/View)
- **Components**: charmbracelet/bubbles (textinput)
- **Styling**: charmbracelet/lipgloss
- **Live Rebuild**: entr (watch and rebuild on .go file changes)

## Project Structure

```
main.go                          # Entry point — tea.NewProgram with alt-screen
internal/
  game/
    model.go                     # Top-level bubbletea Model (Init/Update/View, game loop)
    commands.go                  # tea.Cmd factories (tickCmd at 100ms / 10 FPS)
  aircraft/
    aircraft.go                  # Aircraft struct, state machine, heading/altitude/speed interpolation, movement
    spawner.go                   # Edge spawning with difficulty ramp, callsign generation
  command/
    parser.go                    # Parse "AA123 H270 A3 S2 L" into Command structs
    executor.go                  # Apply parsed commands to aircraft map (immutable)
  collision/
    detector.go                  # O(n^2) same-grid-position + same-altitude check
  runway/
    runway.go                    # Runway definition, CanLand validation, cell rendering
  radar/
    renderer.go                  # ASCII radar grid builder, aircraft sidebar panel
  ui/
    styles.go                    # All lipgloss style definitions (centralized)
    hud.go                       # Score/status bar, messages, game over overlay, pause overlay
    help.go                      # Help overlay content (command reference, keybindings)
```

## Build & Run

```bash
make build       # Build binary
make run         # Build and run
make watch       # Live reload with air
make test        # Run tests
make test-race   # Run tests with -race
make cover       # Coverage summary
make cover-html  # HTML coverage report
make fmt         # gofmt + goimports
make vet         # go vet
make lint        # staticcheck
make clean       # Remove artifacts
make help        # Show all targets
```

## Architecture Patterns

- **Elm Architecture**: All state lives in `game.Model`. `Update()` handles messages and returns new state + commands. `View()` is a pure render from state to string. No side effects in Update or View.
- **Immutability**: `Aircraft.Tick()` returns a new Aircraft, never mutates. `command.Execute()` returns a new aircraft map. All state transitions produce new values.
- **Game loop**: `tea.Tick(100ms)` fires `tickMsg`. The tick handler advances aircraft, checks collisions, checks landings, spawns new aircraft, then schedules the next tick. The loop stops on game over or pause.
- **Heading interpolation**: Uses `((target - current + 540) % 360) - 180` for shortest-turn direction. Turn rate is 3 degrees per tick.
- **Float positions, integer grid**: Physics uses float64 for smooth movement. `math.Round()` converts to grid coordinates for display and collision checks.

## Key Constants

- Radar size: 60x30 cells
- Tick interval: 100ms (10 FPS)
- Turn rate: 3 degrees/tick
- Speed scale: `speed * 0.3` cells/tick
- Spawn interval: 5s initially, ramps down to 1.5s over 5 minutes
- Max aircraft: 5 initially, ramps up to 15
- Landing tolerance: 2 grid cells from runway, +/-10 degrees heading, altitude must be 1

## Command Grammar

```
<CALLSIGN> <CMD> [<CMD>...]
  H<0-359>    Set target heading
  A<1-40>     Set target altitude (thousands of feet)
  S<1-5>      Set target speed
  L           Clear to land
```

## Testing

- Table-driven tests throughout
- Always run with `-race` flag: `make test-race`
- Target 80%+ coverage on core logic packages
- Integration tests in `game/model_test.go` test full game loop via synthetic `tea.Msg` values

## Code Style

- Follow standard Go conventions: `gofmt`, `goimports`, `go vet`
- Accept interfaces, return structs
- Wrap errors with `fmt.Errorf("context: %w", err)`
- No mutation of existing structs — all state transitions return new values
- Package-level lipgloss styles in `ui/styles.go`, not scattered inline
