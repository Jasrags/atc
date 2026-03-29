# ATC - Terminal Air Traffic Control Game

## Project Overview

A terminal-based air traffic control simulation built in Go using the charmbracelet TUI stack (bubbletea, bubbles, lipgloss). Players guide aircraft from radar screen edges to runways using classic ATC text commands.

## Roadmap

See **[ROADMAP.md](ROADMAP.md)** for the feature priority list. Always reference and update that document when planning or completing work. Features are ordered by gameplay impact:

1. Departures — arrival/departure sequencing puzzle
2. Separation rules — distance-based enforcement, not just binary collision
3. Expanded commands — direct to fix, hold, go around, expedite, turn left/right
4. Aircraft types — Light/Medium/Heavy with performance and wake turbulence differences
5. Pilot patience — urgency/pressure system
6. Scenarios/stages — structured challenges with objectives
7. Ground operations — taxiways, runway crossing, gate assignment

## Tech Stack

- **Language**: Go 1.26+
- **TUI Framework**: charmbracelet/bubbletea (Elm Architecture: Model/Update/View)
- **Components**: charmbracelet/bubbles (textinput)
- **Styling**: charmbracelet/lipgloss
- **Live Rebuild**: entr (watch and rebuild on .go file changes)

## Project Structure

```
main.go                          # Entry point — tea.NewProgram with alt-screen, --dev flag
internal/
  game/
    model.go                     # Top-level bubbletea Model (Init/Update/View, game loop)
    commands.go                  # tea.Cmd factories (tickCmd at 100ms / 10 FPS)
    dev.go                       # Developer mode: / command parser and handlers
  aircraft/
    aircraft.go                  # Aircraft struct, state machine, movement, trail tracking
    departure.go                 # NewDeparture constructor, TakeoffTick (OnRunway → Departing)
    spawner.go                   # Edge spawning with difficulty ramp, callsign generation, departure spawning
  command/
    parser.go                    # Parse "AA123 H270 A3 S2 L" into Command structs
    executor.go                  # Apply parsed commands to aircraft map (immutable)
  collision/
    detector.go                  # O(n^2) same-grid-position + same-altitude check
  config/
    config.go                    # GameConfig, Difficulty, GameMode, CallsignStyle enums
  gamemap/
    gamemap.go                   # Map, Runway, Fix, TaxiNode, TaxiEdge, Gate types + pathfinding
    registry.go                  # Map definitions with taxiway networks (San Diego, Chicago, Tutorial)
  heading/
    heading.go                   # Shared heading math (Delta, AbsDelta)
  runway/
    runway.go                    # Runway CanLand validation
  radio/
    radio.go                     # RadioMessage, Log (immutable append-only), constructors
    phraseology.go               # ATC/pilot phraseology formatting (heading, altitude, land, etc.)
    renderer.go                  # Radio log viewport rendering with styled timestamps and directions
  cmdtree/
    cmdtree.go                   # Command tree state machine — resolves input text + aircraft state to options
    renderer.go                  # Renders clickable bubblezone option buttons below ATC> prompt
  radar/
    renderer.go                  # ASCII radar grid, nav fixes, runway numbers, trail dots, flight strips
  ui/
    styles.go                    # All lipgloss style definitions (centralized, includes radio styles)
    hud.go                       # Score/status bar, game over overlay, pause overlay
    help.go                      # Help overlay content (command reference, keybindings)
    menu.go                      # Main menu renderer
    setup.go                     # Combined game setup screen renderer
```

## Build & Run

```bash
make build       # Build binary
make run         # Build and run
make dev         # Build and run with developer mode (/ commands)
make watch       # Rebuild on .go file changes (requires entr)
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
- **Immutability**: `Aircraft.Tick()` returns a new Aircraft, never mutates. `command.Execute()` returns a new aircraft map. All state transitions produce new values. Exception: `Spawner` is a pointer with mutable internal state (`lastSpawn`, `rng`), intentionally shared across model copies in the single-goroutine bubbletea loop.
- **Game loop**: `tea.Tick(100ms)` fires `tickMsg`. The tick handler advances aircraft, checks collisions, checks landings (against all runways), spawns new aircraft, then schedules the next tick. The loop stops on game over or pause.
- **Screen state machine**: `screen` enum (Menu, Playing, Help, Paused, GameOver) + `menuScreen` sub-state (Main, Setup) control all UI flow. Each screen has its own key handler.
- **Heading interpolation**: Uses `((target - current + 540) % 360) - 180` for shortest-turn direction via `heading.Delta()`.
- **Float positions, integer grid**: Physics uses float64 for smooth movement. `math.Round()` converts to grid coordinates for display and collision checks.
- **Throttled changes**: Altitude and speed use a tick counter modulo to pace transitions realistically.
- **Multi-runway landing**: The tick handler checks all `m.runways` for each landing aircraft, supporting maps with parallel runways (e.g., Chicago).

## Key Constants

- Tick interval: 100ms (10 FPS)
- Turn rate: 1 deg/tick (10 deg/s, ~9s for a 90-degree turn)
- Speed scale: `speed * 0.04` cells/tick (speed 3 = 1.2 cells/s, ~67s to cross 120-cell map)
- Altitude change: 1 per 5 ticks (~0.5s per 1000ft)
- Speed change: 1 per 10 ticks (~1s per speed unit)
- Spawn interval: 5s initially, ramps down to 1.5s over 5 minutes (scaled by difficulty multiplier)
- Aircraft count ramp: starts at 5, increases with time, capped by difficulty max (Easy: 8, Normal: 15, Hard: 25)
- Landing tolerance: 2 grid cells from runway, +/-10 degrees heading, altitude must be 1
- Trail length: last 5 grid positions (when enabled)

## Game Config

Settings are configured on the setup screen before each game and stored in `config.GameConfig`:

| Setting | Options | Effect |
|---------|---------|--------|
| Map | San Diego (120x50), Chicago (120x50), Tutorial (90x40) | Radar size, runways, nav fixes |
| Difficulty | Easy (1.5x interval, max 8, spd 1-3) / Normal (1.0x, 15, 2-4) / Hard (0.6x, 25, 2-5) | Spawn rate, aircraft cap, speed range |
| Game Mode | Arrivals Only | Traffic type |
| Callsign Style | ICAO (AA123) / Short (A12) | Spawner callsign format |
| Plane Trails | On / Off | Aircraft store last 5 positions, rendered as dots |

## Command Grammar

```
<CALLSIGN> <CMD> [<CMD>...]
  H<0-359>      Set target heading (shortest turn)
  A<1-40>       Set target altitude (thousands of feet)
  S<1-5>        Set target speed
  L [runway]    Clear to land (optionally on specific runway)
  D <fix>       Direct to named waypoint (auto-navigate)
  TL <heading>  Turn left to heading (forced direction)
  TR <heading>  Turn right to heading (forced direction)
  EX            Expedite altitude change (2x rate)
  GA            Go around (abort landing)
  T             Cleared for takeoff
  PB [runway]   Pushback approved (optional: expect runway)
  TX <tw...>    Taxi via named taxiways
  HS <runway>   Hold short of runway
  CR <runway>   Cleared to cross runway
  GATE <gate>   Taxi to gate
```

## Developer Mode

Launch with `make dev` or `./atc -dev` to enable `/` commands in the ATC prompt for testing without gameplay pressure.

| Command | Effect |
|---------|--------|
| `/help` | List all dev commands |
| `/spawn` | Spawn one arrival at random edge |
| `/spawn dep` | Spawn one departure at random gate |
| `/clear` | Remove all aircraft |
| `/god` | Toggle god mode (collisions don't end game) |
| `/pause` | Toggle automatic spawner on/off |
| `/speed <1-5>` | Set game speed multiplier (physics runs N times per frame) |

HUD shows `[DEV]` with active flags: `[DEV GOD NOSPAWN 3x]`. Single-char shortcuts (`p` for pause, `?` for help) only activate when the input field is empty — they never interfere with typing commands.

## Testing

- Table-driven tests throughout
- Always run with `-race` flag: `make test-race`
- Target 80%+ coverage on core logic packages
- Integration tests in `game/model_test.go` test full game loop via synthetic `tea.Msg` values
- `processCommand` integration tests cover valid commands, invalid syntax, and unknown callsigns

## Code Style

- Follow standard Go conventions: `gofmt`, `goimports`, `go vet`
- Accept interfaces, return structs
- Wrap errors with `fmt.Errorf("context: %w", err)`
- No mutation of existing structs — all state transitions return new values
- Package-level lipgloss styles in `ui/styles.go`, not scattered inline
- Named constants for all magic numbers (gridSpeedScale, turnRate, altTickRate, etc.)
