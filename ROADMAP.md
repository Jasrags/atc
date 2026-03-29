# ATC Roadmap

Feature priorities for closing the gap with Tower!3D Pro and IATC4. Each feature is tracked here with status, and broken into implementation tasks when work begins.

---

## Priority 0: Ebitengine Migration

**Status: Not Started**

Migrate from charmbracelet/bubbletea (terminal UI) to Ebitengine (2D game engine) for proper graphical rendering. The TUI cannot provide the visual fidelity needed for Tower mode surface operations, zoom/pan, or future features like weather overlays and radar scopes.

### Why Migrate

- ASCII radar is too coarse for Tower ground ops — callsigns overlap, airport surface is tiny
- No zoom/pan capability in terminal rendering
- Limited to monospace character grid — no smooth movement, no real coordinates
- Mouse interaction limited to bubblezone hacks
- Cannot render radar scope sweeps, weather, altitude tags properly

### Architecture Preservation

The game logic is cleanly separated from the presentation layer. **~70% of the codebase ports unchanged:**

| Package | Lines | Changes |
|---------|-------|---------|
| `aircraft/` | 634 | **None** — Tick(), GroundTick(), state machine, spawner |
| `command/` | 469 | **None** — parser, executor |
| `collision/` | 107 | **None** — separation/collision detection |
| `config/` | 190 | **None** — GameConfig, roles, difficulty |
| `gamemap/` | 487 | **None** — map definitions, taxiway graph, pathfinding |
| `heading/` | 17 | **None** — heading math |
| `runway/` | 60 | **None** — landing validation |
| `radio/` | 297 | **None** — RadioMessage, Log, phraseology |
| `game/` | 1550 | **Rewrite glue** — Replace bubbletea Model with Ebitengine Game. Tick pipeline, command processing, spawning logic preserved. |
| `radar/` | 429 | **Replace** — ASCII grid → Ebitengine draw calls |
| `ui/` | 371 | **Replace** — lipgloss styles → ebitenui widgets |
| `cmdtree/` | 361 | **Adapt** — State machine stays, rendering becomes ebitenui buttons |

### Tech Stack

- **Engine**: [Ebitengine](https://ebitengine.org) v2.9+ (pure Go, Update/Draw loop matches existing Update/View)
- **UI Widgets**: [ebitenui](https://github.com/ebitenui/ebitenui) (text input, scrollable containers, buttons, layout)
- **Text**: Ebitengine `text/v2` with glyph caching for callsign labels + [etxt](https://github.com/tinne26/etxt) if needed
- **Fonts**: Monospace TTF for radar labels (JetBrains Mono, IBM Plex Mono, or similar)

### Phase 1: Engine Scaffold & Window

**Goal**: Ebitengine window opens, renders a static radar background with runway. Proves the engine works and establishes the project structure.

- [ ] Add Ebitengine dependency (`go get github.com/hajimehoshi/ebiten/v2`)
- [ ] New `internal/engine/` package with `Game` struct implementing `ebiten.Game` interface
- [ ] `Update()` stub — handles quit key (Esc/Ctrl+C)
- [ ] `Draw()` renders: dark background, radar grid dots, runway as a thick line, map border
- [ ] `Layout()` returns window size (resizable)
- [ ] New `main.go` entry point (or `--gui` flag alongside existing TUI)
- [ ] Render one static aircraft symbol (`@`) with callsign text label at a fixed position
- [ ] Load a monospace TTF font, configure `text/v2` face

### Phase 2: Aircraft Rendering & Movement

**Goal**: Aircraft move on screen in real-time. Core visual loop working.

- [ ] Wire existing `aircraft.Tick()` / `GroundTick()` into `Game.Update()` at 10 TPS
- [ ] Draw aircraft symbols as colored glyphs: `@` (airborne), `v` (taxi), `#` (gate), `!` (hold), `>` (runway)
- [ ] Draw callsign label offset from aircraft position
- [ ] Draw altitude/speed data tag near each aircraft (e.g., "AA123 ↑05 S3")
- [ ] Draw heading indicator line (short line extending from aircraft in heading direction)
- [ ] Draw trail dots for aircraft with trails enabled
- [ ] Draw nav fixes (waypoint/VOR/airport symbols + labels)
- [ ] Draw taxiway lines, gate markers, hold-short markers
- [ ] Separation violation highlighting (blinking red on violating aircraft)
- [ ] Color-code by state: green (approach), yellow (landing), cyan (ground), red (crashed)

### Phase 3: Camera & Viewport

**Goal**: Zoom and pan. This solves the Tower mode surface problem that triggered the migration.

- [ ] Camera struct: position (centerX, centerY), zoom level (1.0 = full map fits window)
- [ ] Mouse scroll wheel zooms in/out (0.5x to 8x range)
- [ ] Click-and-drag pans the camera
- [ ] Keyboard shortcuts: `+`/`-` zoom, arrow keys pan, `Home` reset to fit-all
- [ ] Tower mode: camera auto-centers on airport surface bounding box at ~3x zoom
- [ ] TRACON/Combined mode: camera fits full radar (zoom 1.0)
- [ ] World-to-screen and screen-to-world coordinate transforms
- [ ] Aircraft click detection uses world coordinates (zoom-independent)

### Phase 4: Command Input & Game Loop

**Goal**: Game is playable. Player can type commands and interact with aircraft.

- [ ] ebitenui TextInput widget for ATC command prompt (bottom of screen)
- [ ] Wire Enter key → existing `command.Parse()` + `command.Execute()` pipeline
- [ ] Wire command results to radio log
- [ ] Existing tick pipeline: spawning, physics, collision, separation, patience — all wired
- [ ] Time freeze (`p` key) and speed control (`[`/`]`) working
- [ ] Game state machine: Menu → Playing → GameOver (screen enum, same as current)
- [ ] Developer mode (`/` commands) working

### Phase 5: UI Panels (ebitenui)

**Goal**: Full HUD with all information panels. Feature parity with TUI version.

- [ ] Layout: radar viewport (main), flight strips (right sidebar), radio log (bottom), HUD (top bar), command input (bottom)
- [ ] HUD bar: score, aircraft count, elapsed time, role, speed/freeze status, near misses
- [ ] Flight strip sidebar: scrollable list of aircraft strips (callsign, state, heading/alt/speed, targets)
- [ ] Flight strip click → populates command input with callsign
- [ ] Radio log: scrollable message list with colored timestamps and direction indicators
- [ ] Command tree: clickable command options below input (H/A/S/L etc.) — reuse `cmdtree.Resolve()` state machine
- [ ] Setup screen: map/role/difficulty/callsign/trails selection (ebitenui widgets)
- [ ] Main menu: New Game, Help, Quit
- [ ] Help overlay: command reference
- [ ] Game over overlay: score + restart option

### Phase 6: Visual Polish

**Goal**: The game looks and feels like a real radar scope, not an ASCII art project.

Already done (from Phase 1-5):
- [x] Radar scope aesthetic: dark green/black background, bright green blips
- [x] Runway rendered as thick rectangle with threshold markings and runway numbers
- [x] Taxiways rendered as thinner lines with taxiway letter labels at intersections
- [x] Gates rendered as small rectangles with gate ID labels
- [x] Aircraft data blocks: leader line from symbol to data tag (callsign + alt + speed)
- [x] Hold-short bars rendered as dashed lines across taxiway

Remaining:
- [ ] Smooth aircraft position interpolation (sub-tick rendering for fluid movement)
- [ ] Blinking effects: separation warnings, patience urgency, cleared-to-land
- [ ] Compass rose or heading scale around radar border
- [ ] Minimap (when zoomed in) showing full radar with current viewport rectangle
- [ ] FAA airport diagram style Tower view:
  - [ ] Runway-taxiway visual connectivity (taxiway C/D meet runway edge, not floating)
  - [ ] Proportional taxiway widths (wider near intersections)
  - [ ] Terminal building outlines (dark gray polygons adjacent to gates)
  - [ ] Runway markings: centerline dashes, threshold bars, numbers rendered inside runway surface
  - [ ] Runway-entry connector paths drawn from taxiway nodes to runway rectangle edge

### Phase 7: Remove TUI & Cleanup

**Goal**: Single rendering path. Remove bubbletea dependency.

- [ ] Remove bubbletea, bubbles, lipgloss, bubblezone dependencies
- [ ] Remove `internal/ui/` (lipgloss styles)
- [ ] Remove `internal/radar/renderer.go` (ASCII grid)
- [ ] Remove `internal/cmdtree/renderer.go` (ASCII button rendering)
- [ ] Update `main.go` to launch Ebitengine directly (no `--gui` flag)
- [ ] Update CLAUDE.md project structure documentation
- [ ] Update Makefile targets
- [ ] Update all tests that depend on bubbletea message types

### Migration Strategy

- **Parallel development**: Keep TUI working while building the Ebitengine version. Use a `--gui` flag or separate binary during migration.
- **Phase 1-4 = MVP**: Playable game with graphical rendering. Ship this before adding polish.
- **Phase 5 = Feature parity**: All TUI features work in the graphical version.
- **Phase 6 = The payoff**: Visual quality that was never possible in a terminal.
- **Phase 7 = Cleanup**: Remove old code only after the new version is fully validated.

### Risk Mitigation

- **ebitenui learning curve**: Phase 5 is the highest-effort phase. Build a standalone ebitenui prototype of the radio log + command input first to learn the widget system before integrating.
- **Text rendering performance**: Many dynamic callsign labels updating each frame. Use `text/v2` glyph caching and only redraw changed labels. Profile early in Phase 2.
- **Click detection with zoom**: Phase 3's coordinate transforms must be correct before Phase 4's aircraft click detection. Test with unit tests on world↔screen transforms.
- **Test coverage**: Game logic tests (`aircraft/`, `command/`, `collision/`, etc.) continue passing throughout migration since those packages don't change. Only `game/` and rendering tests need rewriting.

---

## Priority 1: Controller Role System

**Status: TRACON Implemented, Tower Phases 1-4 Complete**

The player selects a role at game start that defines their scope of responsibility, available commands, visible UI panels, and what the game automates. Each role plays like a different game. See `atc-flight-strips.md` for the full design reference.

### Facility Model

Real-world ATC divides into distinct facilities. The game models these as selectable roles:

| Mode | Real Facility | Phase of Flight | Status |
|------|---------------|-----------------|--------|
| **TRACON** | TRACON — Approach + Departure | Terminal airspace, vectors, sequencing | ✓ Implemented |
| **Tower** | ATCT — Local Control + Ground | Runway ops, taxi, takeoff, landing | Not Started |
| **Combined** | All of the above | Full TRACON + Tower | ✓ Implemented (current default behavior) |
| **Center** | ARTCC — En-Route | Cruise, high altitude, flow management | Future |
| **Clearance/Ground** | ATCT — Clearance + Ground | Pre-departure, taxi routing | Future |

### TRACON (Approach/Departure) ✓

- [x] Role enum + setup screen selection (TRACON / Combined)
- [x] Command filtering: ground commands rejected
- [x] Command tree filtered by role
- [x] Auto-ground arrivals: landed → auto gate → auto taxi
- [x] Auto-ground departures: auto pushback → auto taxi → auto hold short → auto takeoff
- [x] Role shown in HUD

### Tower (Local + Ground) — Not Started

Tower is focused on the physical runway environment. TRACON delivers arrivals on final. Player manages the last mile + ground.

| Aspect | Details |
|--------|---------|
| **Scope** | Final approach → landing → gate (arrivals). Gate → pushback → taxi → takeoff → initial climb (departures). |
| **Commands** | L (clear to land), GA, T, PB, TX, HS, CR, GATE + limited H/A for go-arounds and initial departure climb |
| **Automated** | TRACON sequences arrivals onto final — they arrive at ~5nm, aligned, descending. Departures handed off to TRACON after initial climb (altitude ≥ 3). |
| **UI** | Zoomed airport surface view (taxiways, gates, runways prominent). Smaller approach radar inset. Tower-specific flight strips (see below). |
| **Scoring** | +1 per arrival at gate. +1 per departure handed off to TRACON. |
| **Game over** | Runway incursion, collision, or missed approach overflow. |

Implementation:
- [x] Tower auto-approach: arrivals spawn pre-sequenced on ~5nm final, low altitude, descending
- [x] Tower auto-departure handoff: departing aircraft auto-handoff at altitude 3+
- [x] Tower command filtering: D, TL, TR, EX blocked; H/A/S/L/GA + ground commands allowed
- [x] Add Tower to selectable role options
- [ ] Tower zoomed view: airport surface fills radar, approach corridor as inset *(deferred to Ebitengine migration Phase 3)*
- [ ] Runway occupancy enforcement: only one operation per runway at a time
- [ ] Runway incursion = game over
- [ ] Split flight strip panels: arrivals / departures

### Future Roles

**Center (ARTCC):** En-route cruise phase with large geographic sectors. Big-picture flow management, sector loading, long-range sequencing. Issues wheels-up time windows to TRACON/Tower.

**Clearance/Ground:** Pre-departure focus. IFR clearance issuance (route, altitude, squawk), pushback approval, taxi routing puzzle. Coordinates runway crossings with Tower.

---

## Priority 2: Ground Operations MVP (Complete)

**Status: Complete**

Full ground operations bringing departures, taxi, and surface management together as one cohesive feature. Subsumes the original "Departures" and "Ground Operations" items. Built in phases — each phase delivers playable value.

### Phase 1: Radio Comms Window ✓

Replace the HUD message bar with a proper radio communications log.

- [x] `RadioMessage` struct: timestamp, from, to, text, direction (inbound/outbound), priority
- [x] `radioLog []RadioMessage` replaces `messages []string` on the game model
- [x] Scrollable radio viewport (`viewport.Model`) between game area and input
- [x] Inbound messages (pilot → ATC) styled cyan, outbound (ATC → pilot) styled green
- [x] Urgent messages highlighted yellow, emergency messages red
- [x] Existing game events (entering airspace, landed, collision) converted to radio phraseology
- [x] Commands generate outbound radio messages with ATC phraseology ("AA123, turn right heading 090, descend and maintain 3000")

### Phase 2: Command Tree at Prompt ✓

Click-or-keyboard command menu below the ATC> input that builds commands interactively.

- [x] `CommandTree` state machine: Idle → CallsignSelected → CommandChosen → ValueInput → Execute
- [x] Clicking a flight strip (or typing callsign) opens the tree with state-sensitive options
- [x] State-sensitive root commands (APPR: H/A/S/L, LAND: GA — ground states to be added in Phase 3)
- [x] Clicking a command option appends to the input field — visible and editable
- [x] Sub-menus for values: compass rose grid for headings, numbered list for altitude, speed list
- [x] Chainable commands: after picking H270, tree offers A/S/L/Enter instead of closing
- [x] Raw typing still works and tree stays in sync with input text
- [x] Bubblezone marks each option for click detection

### Phase 3: Ground Aircraft States & Commands ✓

Extend the aircraft state machine and command set for surface operations.

- [x] New aircraft states: `Taxiing`, `AtGate`, `Pushback`, `HoldShort`, `OnRunway`, `Departing`
- [x] State definitions and basic transitions (full arrival/departure sequences wired in Phase 6)
- [x] New commands in parser/executor:
  - `T` — Cleared for takeoff
  - `PB` — Pushback approved (optional: `PB <runway>` to set expect runway)
  - `TX <taxiway...>` — Taxi via route (e.g., `TX A B C1`)
  - `HS <runway>` — Hold short of runway
  - `CR <runway>` — Cleared to cross runway
  - `GATE <gate>` — Taxi to gate (post-landing assignment)
  - `GA` — Go around (abort landing)
- [x] Command tree updated with ground command options per state
- [x] Flight strips show ground-specific info (taxi route, assigned gate, expect runway)
- [x] Ground aircraft symbols on radar: `v` (taxi), `#` (gate), `<` (pushback), `!` (hold), `>` (runway)
- [x] State validation: airborne commands rejected for ground aircraft and vice versa

### Phase 4: Taxiway Network & Map Data ✓

Define the airport surface layout as a graph for pathfinding and rendering.

- [x] `TaxiNode` struct: ID, X, Y, type (Intersection, HoldShort, Gate, RunwayEntry), optional runway
- [x] `TaxiEdge` struct: from/to node IDs, taxiway name (A, B, C, D...)
- [x] `Gate` struct: ID, node ID
- [x] Add `TaxiNodes`, `TaxiEdges`, `Gates` to `gamemap.Map`
- [x] Define taxiway layouts for all three maps (Tutorial, San Diego, Chicago)
- [x] `ResolveTaxiRoute`: given start node + taxiway names, walks the graph to produce ordered node path
- [x] `NodeByID`, `GateByID`, `Neighbors` lookup helpers
- [x] Validation tests: all nodes in bounds, all edges reference valid nodes, all gates reference valid nodes

### Phase 5: Ground Rendering & Taxi Movement ✓

Show ground traffic on the radar and move aircraft along taxiway paths.

- [x] Render taxiways as `-`/`|` chars (dimmed) on the radar grid
- [x] Render gates as `#` markers (styled blue)
- [x] Render hold-short points as `:` markers (styled yellow)
- [x] Ground aircraft symbols: `v` (taxiing), `#` (at gate), `<` (pushback), `!` (holding short), `>` (on runway)
- [x] Ground movement: `GroundTick()` advances aircraft along `TaxiPath` node positions (1 node per 3 ticks)
- [x] Aircraft snap to taxiway node positions — no heading-based free movement on ground
- [x] TX command resolves taxiway names into node path via `ResolveTaxiRoute`
- [x] GATE command creates direct path to gate node position
- [x] Landed aircraft stay in map for gate assignment (no longer removed on landing)
- [x] Taxiing aircraft auto-transition to AtGate when path completes with gate assignment
- [x] Aircraft removed from map when they reach AtGate state (arrival complete)

### Phase 6: Departures & Runway Occupancy ✓

Complete the arrival/departure loop with runway management.

- [x] Departure spawner: aircraft appear at gates with pushback requests (alternates with arrivals)
- [x] `NewDeparture()` creates aircraft at gate with `AtGate` state
- [x] `TakeoffTick()` handles `OnRunway` → `Departing` transition after 15-tick takeoff roll (~1.5s)
- [x] Departing aircraft climb to 5000ft at speed 3, heading set from assigned runway
- [x] Departing aircraft scored +1 when leaving airspace
- [x] Available gates tracked — departures only spawn at unoccupied gates
- [x] `isRunwayOccupied()` helper for future runway conflict detection
- [x] `runwayHeading()` resolves runway number to heading for departure direction
- [x] Departure flight strips show ground state info (gate, runway, taxi route)

### MVP Scope

Phases 1-6 together constitute the Ground Operations MVP. Phases 1-2 (radio + command tree) are interaction foundations that benefit the entire game. Phases 3-6 add the ground gameplay. Each phase is playable — you don't need all 6 to test and enjoy what's built.

### Known Follow-ups

- [x] **Tick loop immutability refactor** — Refactored `handleTick` into a pipeline of pure transformation functions (`tickLanding`, `tickAutoGroundArrival`, `tickTaxiComplete`, `tickAutoGroundDeparture`, `tickPatience`) that each return a new `Aircraft` plus a `tickEffect` struct for side effects. Collision block also uses immutable `WithState`/`WithHeading` methods. No in-place mutation remains.
- [x] **Extract `model.go` into smaller files** — Split from 1151 → 315 lines. Extracted: `tick.go` (tick pipeline + physics), `playing.go` (command processing + input), `menu.go` (menu/setup/help/pause/gameover), `helpers.go` (spawning + pathfinding). All files under 400 lines.
- [x] Wire phraseology formatters into `CommandPhraseology` — radio log now shows real ATC phrasing ("fly heading 270, maintain 3,000") instead of abbreviated codes ("HDG 270, ALT 3"). All 17 change codes mapped via `changeToPhraseology`.
- [x] Fix `TX` parser consuming command-like taxiway names — `TX L T` and `TX A GA B` now parse correctly. TX greedily consumes all remaining tokens as taxiway names.
- [ ] Add integration tests for mouse click → command tree → input manipulation path in `game/model.go`. Bubblezone zone detection requires global manager init and synthetic mouse events.

---

## Priority 3: Separation & Wake Turbulence

**Status: Separation Implemented, Wake Turbulence Future**

Replace binary collision with distance-based separation enforcement and add wake turbulence spacing.

**Separation rules: ✓**
- [x] Minimum lateral separation: 3 grid cells
- [x] Minimum vertical separation: 1 altitude unit (1000ft) when laterally close
- [x] Separation violation warnings: radio TRAFFIC ALERT with distance on first violation
- [x] Score penalty: -50 per violation per tick (score cannot go below 0)
- [x] Near-miss counter in HUD
- [x] Radar visual: violating aircraft shown as `?` with blinking red style
- [x] Only airborne aircraft checked (ground aircraft excluded)
- [x] Collision still ends game (same position + same altitude)
- [x] 100% test coverage on separation detection

**Wake turbulence (requires Aircraft Types — future):**
- Heavy/Super aircraft generate wake turbulence requiring extended spacing for following aircraft
- Minimum spacing intervals based on lead/follow weight category
- Tower mode enforces runway departure spacing (2 min behind heavy, 3 min behind super)
- Flight strips always display wake category
- Violations flagged on radar + radio warning

---

## Priority 4: Expanded Command Set

**Status: Implemented (except HLD)**

Richer ATC commands closer to real phraseology.

- [x] `D <fix>` — Direct to waypoint/fix (aircraft auto-navigates, recalculates heading each tick, clears on arrival within 2 cells)
- [ ] `HLD <fix>` — Hold at fix (circle a waypoint) — future
- [x] `GA` — Go around (abort landing, climb, maintain heading) — implemented in Ground Ops
- [x] `EX` — Expedite (double altitude change rate)
- [x] `TL <heading>` / `TR <heading>` — Turn left/right to heading (force turn direction, overrides shortest-path)
- [x] `L <runway>` — Clear to land on specific runway (for multi-runway airports like Chicago)

---

## Priority 5: Aircraft Types & Performance (Future)

**Status: Not Started**

Different aircraft categories with gameplay-affecting differences.

**Wake turbulence categories:** Light (L), Medium (M), Heavy (H), Super (J)

**Performance differences:**
- Light: faster turns, slower speed range (1-3), shorter landing distance
- Medium: standard performance (B737, A320)
- Heavy: slower turns, faster speed range (3-5), longer runway occupancy after landing
- Super: A380/AN-225 — extreme wake spacing required

**Flight strip display:**
- Aircraft type code (B738, A321, C172, A388)
- Equipment suffix (/L, /G, /W — IFR capability)
- Wake turbulence category prominently displayed
- Type affects strip color or icon for quick scan

**Spawner:**
- Assigns type based on map/difficulty
- Heavy traffic increases at higher difficulties
- Type distribution varies by airport (SAN = mostly medium, ORD = more heavies)

---

## Priority 6: Pilot Patience / Pressure System

**Status: Implemented**

Aircraft request instructions and expect timely responses.

- [x] Patience timer per aircraft — starts at spawn, 30s default before first nag
- [x] Escalation: "still waiting for vectors" → "requesting ANY instructions!" → leaves airspace (-1 score)
- [x] Nag every 10 seconds after first nag, 4 nags total before leaving
- [x] Any command resets the patience timer (PatienceTicks + NagCount)
- [x] Flight strip callsign color changes: green → yellow → orange → blinking red
- [x] Radio messages at each escalation level
- [ ] Optional: global pressure bar (IATC4 style) — future

---

## Priority 7: Scenarios / Stages (Future)

**Status: Not Started**

Structured challenges beyond infinite sandbox mode.

- Scenario definition format (YAML or Go structs): timed aircraft spawns, objectives, par score
- Objective types: land N aircraft, clear N departures, survive N minutes, handle rush hour wave
- Star rating (1-3 stars) based on score/time/separation violations
- Scenario select screen in menu
- Tutorial scenarios that teach one concept at a time
- Campaign progression: unlock harder scenarios by completing easier ones

---

## Priority 8: Role-Specific Flight Strips (Future)

**Status: Not Started**

Flight strips should render differently per controller role — each position only displays what is actionable at that phase. The strip is a state machine: status flags drive available actions.

**Tower strips** — lean and fast to scan:
- Callsign, aircraft type, wake category, sequence number
- Assigned runway, weather (wind/altimeter from ATIS)
- Status flags: TAXI / HOLD SHORT / POSITION / AIRBORNE / LANDED
- Arrivals additionally show expected exit taxiway
- Does NOT show: full route, squawk, cruise altitude, next sector

**TRACON strips** — data-dense, supplement the radar:
- Callsign, type + equipment suffix, squawk code
- Origin/destination, SID/STAR procedure
- Assigned altitude (AFL), filed altitude (RFL), heading, speed
- Next sector/facility for handoff
- Status flags: RADAR CONTACT / HANDOFF / RELEASED

**Strip field comparison by mode:**

| Field | Tower | TRACON | Combined |
|-------|-------|--------|----------|
| Callsign | yes | yes | yes |
| Aircraft type | wake category | performance | both |
| Squawk | no | primary ID | yes |
| Full route | no | yes | yes |
| SID/STAR | no | yes | yes |
| Assigned altitude | initial only | full | full |
| Heading | initial turn | yes | yes |
| Speed | no | yes | yes |
| Runway | primary focus | arrivals | yes |
| Taxi route | yes | no | yes |
| Sequence number | yes | arrivals | yes |
| Next sector | no | yes | no |

---

## Priority 9: Handoff Mechanics (Future)

**Status: Not Started**

Handoffs are the transitions between controller modes. In real ATC, a controller initiates a handoff, the receiver accepts, then the frequency change happens. A late or botched handoff creates downstream pressure.

**Handoff modes (configurable per difficulty):**
- Auto-handoff at boundary (Easy) — aircraft silently transitions
- Manual initiation required (Normal) — player clicks to initiate, AI accepts
- Handoff refusal if receiving sector is overloaded (Hard) — must reroute or hold

**Departure releases:**
- TRACON issues departure releases to Tower — Tower cannot send a departure without one during busy periods
- Creates coordination tension between roles
- In Combined mode, player manages both sides

**Go-around cascade:**
- Tower issues go-around → aircraft re-enters TRACON arrival sequence
- Displaces other sequenced traffic — best stress test for TRACON mode
- Radio comms show the cascade in real time

---

## Priority 10: ATIS & Weather (Future)

**Status: Not Started**

Automatic Terminal Information Service provides current active runway, weather, altimeter setting. All strips reference the active ATIS.

**ATIS system:**
- Active runway assignment based on wind direction
- Wind speed/direction, altimeter setting, visibility
- ATIS letter code (Alpha, Bravo, ...) changes when conditions update
- All flight strips show current ATIS reference

**Weather events:**
- Wind shift → runway change (high-pressure event, cascades across all modes)
- Reduced visibility → increased spacing requirements
- Crosswind limits → some runways close
- Thunderstorm cells → routing around weather

**Gameplay impact:**
- Tower strips show wind/altimeter prominently
- Runway change forces re-sequencing of all traffic
- Weather degrades over time within a session, increasing difficulty naturally

---

## Design Reference

See [`docs/atc-flight-strips.md`](docs/atc-flight-strips.md) for the full design document covering:
- Facility overview and flight lifecycle
- Controller mode details (Clearance, Ground, Tower, TRACON, Center)
- Flight strip field specifications per mode
- Gameplay notes: strips as state machines, handoff mechanics, departure releases, wake turbulence, ATIS, go-arounds
- References: FAA Order 7110.65, AIM Chapter 4, FAA AC 90-23G

---

## Completed Features

- [x] Core game loop (radar, commands, collision, landing, scoring)
- [x] Multiple maps (San Diego, Chicago, Tutorial)
- [x] Game setup screen (map, difficulty, callsign style, trails)
- [x] Flight strips with state/target display
- [x] Click-to-select aircraft (bubblezone)
- [x] Scrollable flight strip viewport
- [x] Auto-generated help from keybindings
- [x] Stopwatch-based elapsed time with native pause/resume
- [x] Lipgloss table HUD and styled radar borders
- [x] Configurable difficulty (Easy/Normal/Hard)
- [x] Plane trails
- [x] ICAO / Short callsign styles
- [x] Radio comms window with phraseology
- [x] Command tree (interactive click/keyboard command menu)
- [x] Ground aircraft states (Taxiing, AtGate, Pushback, HoldShort, OnRunway, Departing)
- [x] Ground commands (T, PB, TX, HS, CR, GATE, GA)
- [x] Taxiway network with pathfinding (all 3 maps)
- [x] Taxiway/gate/hold-short rendering on radar
- [x] Ground taxi movement along node paths
- [x] Departures (gate spawn, pushback, taxi, takeoff roll, climb out)
- [x] Arrival-to-gate flow (land → GATE → taxi → arrive)
- [x] Controller role system (TRACON with automation, Combined)
- [x] Separation rules (3-cell lateral, 1-alt vertical, -50/tick penalty, near-miss tracking)
- [x] Expanded commands: D (direct to fix), TL/TR (forced turn), EX (expedite), L <runway>
- [x] Pilot patience system (30s timer, nag escalation, aircraft leaves on timeout, visual indicators)
- [x] Tower mode: role selection, final approach spawning, auto-departure handoff, command filtering
- [x] Time freeze (p key) and player-facing speed control ([ ] keys, 1x-12x)

---

## Notes

- Each priority can be broken into phases when implementation begins
- Features should be developed with tests (TDD where practical)
- Code review after each feature before committing
- Update GUIDE.md and CLAUDE.md as features land
