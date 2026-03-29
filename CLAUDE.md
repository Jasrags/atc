# ATC - Air Traffic Control Game

## Project Overview

A 2D air traffic control simulation built in Go using Ebitengine. Players guide aircraft using classic ATC text commands across TRACON (approach radar), Tower (surface), and Combined modes.

## Roadmap

See **[ROADMAP.md](ROADMAP.md)** for the feature priority list. Always reference and update that document when planning or completing work.

## Tech Stack

- **Language**: Go 1.26+
- **Engine**: Ebitengine v2.9+ (2D game engine, Update/Draw loop)
- **Text**: Ebitengine text/v2 with Go Mono font (embedded via golang.org/x/image/font/gofont)
- **Live Rebuild**: entr (watch and rebuild on .go file changes)

## Project Structure

```
main.go                          # Entry point — ebiten.RunGame, --dev and --role flags
internal/
  engine/
    game.go                      # Game struct, Update/Draw/Layout, tick pipeline, command processing
    camera.go                    # Camera zoom/pan, world↔screen coordinate transforms
    input.go                     # Text input handling (keyboard characters, cursor, Enter submit)
    colors.go                    # Color palettes (TRACON + Tower), font init, shared drawing helpers
    blink.go                     # Blink/pulse timing utilities for visual effects
    render_tracon.go             # STARS approach radar: range rings, fixes, runway, aircraft dots + trails
    render_tower.go              # ASDE-X surface radar: filled runways, taxiways, gates, chevrons
    render_hud.go                # HUD bar, radio log, command input prompt, game over overlay
    render_strips.go             # Flight strip sidebar with click-to-select
    render_minimap.go            # Minimap overlay when zoomed in
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
    config.go                    # GameConfig, Difficulty, GameMode, CallsignStyle, Role enums
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
  cmdtree/
    cmdtree.go                   # Command tree state machine — resolves input text + aircraft state to options
```

## Build & Run

```bash
make build       # Build binary
make run         # Build and run (TRACON mode)
make run-tower   # Build and run in Tower mode
make run-combined # Build and run in Combined mode
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

- **Ebitengine Game Loop**: All state lives in `engine.Game`. `Update()` runs at 10 TPS for physics. `Draw()` renders at 60 FPS with position interpolation. `Layout()` handles window resizing.
- **Immutability**: `Aircraft.Tick()` returns a new Aircraft, never mutates. `command.Execute()` returns a new aircraft map. All state transitions use `WithState()`, `WithHeading()` etc.
- **Camera System**: World coordinates (grid cells) transform to screen pixels via `Camera.WorldToScreen()`. Zoom/pan centered on cursor. Drawing elements scale with zoom (runways, blips), text stays fixed in screen space.
- **Dual-mode rendering**: TRACON uses STARS approach radar style (dots, trails, leader lines, range rings). Tower uses ASDE-X surface radar style (filled runways, chevrons, gate rectangles, hold-short bars).
- **Pure Draw**: `Draw()` is a pure render function with no state mutation. Hit testing, strip layout, and interpolation state are computed in `Update()`.
- **Heading interpolation**: Uses `((target - current + 540) % 360) - 180` for shortest-turn direction via `heading.Delta()`.
- **Float positions, integer grid**: Physics uses float64 for smooth movement. `math.Round()` converts to grid coordinates for collision checks.

## Key Constants

- Physics TPS: 10 (100ms per tick), render at 60 FPS with interpolation
- Cell size: 8 pixels per grid cell (before zoom)
- Turn rate: 1 deg/tick (10 deg/s, ~9s for a 90-degree turn)
- Speed scale: `speed * 0.04` cells/tick (speed 3 = 1.2 cells/s)
- Altitude change: 1 per 5 ticks (~0.5s per 1000ft)
- Speed change: 1 per 10 ticks (~1s per speed unit)
- Spawn interval: 5s initially, ramps down to 1.5s over 5 minutes
- Landing tolerance: 2 grid cells from runway, +/-10 degrees heading, altitude must be 1
- Camera zoom: 0.5x to 8x, scroll wheel centered on cursor

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

## Controls

| Key | Action |
|-----|--------|
| Type + Enter | Submit ATC command |
| Scroll wheel | Zoom in/out (centered on cursor) |
| Click + drag | Pan the viewport |
| Arrow keys | Pan (when input empty) |
| +/- | Zoom from keyboard (when input empty) |
| Home | Reset to default view for current role |
| p | Freeze / unfreeze time (when input empty) |
| [ / ] | Decrease / increase speed 1x-12x (when input empty) |
| Esc | Quit |
| R | Restart (on game over screen) |
| Click strip | Select aircraft (populates command input) |

## Developer Mode

Launch with `make dev` or `./atc -dev` to enable `/` commands in the ATC prompt.

| Command | Effect |
|---------|--------|
| `/god` | Toggle god mode (collisions don't end game) |
| `/pause` | Toggle automatic spawner on/off |
| `/clear` | Remove all aircraft |
| `/spawn` | Spawn one arrival |
| `/spawn dep` | Spawn one departure at a gate |

## Testing

- Table-driven tests throughout
- Always run with `-race` flag: `make test-race`
- Target 80%+ coverage on core logic packages
- Game logic packages (aircraft, command, collision, config, gamemap, heading, runway, radio) are engine-agnostic and fully testable

## Code Style

- Follow standard Go conventions: `gofmt`, `goimports`, `go vet`
- Accept interfaces, return structs
- Wrap errors with `fmt.Errorf("context: %w", err)`
- No mutation of existing structs — all state transitions return new values
- Named constants for all magic numbers (gridSpeedScale, turnRate, altTickRate, etc.)
