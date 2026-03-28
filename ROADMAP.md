# ATC Roadmap

Feature priorities for closing the gap with Tower!3D Pro and IATC4. Each feature is tracked here with status, and broken into implementation tasks when work begins.

## Priority 1: Ground Operations MVP (Active)

**Status: In Progress**

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

### Phase 6: Departures & Runway Occupancy

Complete the arrival/departure loop with runway management.

- [ ] Departure spawner: aircraft appear at gates with pushback requests
- [ ] Departure sequence: pushback → taxi → hold short → takeoff → climb out → hand off at airspace edge
- [ ] Takeoff roll: aircraft accelerates along runway, lifts off, transitions to airborne state
- [ ] Climb-out: default heading + altitude assignment, player can override
- [ ] Runway occupancy tracking: only one operation (landing, takeoff, crossing) at a time per runway
- [ ] Hold short enforcement: aircraft at hold-short nodes wait until cleared
- [ ] Runway incursion = collision (game over) if two aircraft on the same runway simultaneously
- [ ] Scoring: +1 per departure handed off, +1 per arrival at gate (was: +1 per landing)
- [ ] Departure flight strips with distinct styling

### MVP Scope

Phases 1-6 together constitute the Ground Operations MVP. Phases 1-2 (radio + command tree) are interaction foundations that benefit the entire game. Phases 3-6 add the ground gameplay. Each phase is playable — you don't need all 6 to test and enjoy what's built.

### Known Follow-ups

- [ ] Wire phraseology formatters (`FormatHeadingChange`, `FormatAltitudeChange`, etc.) into `CommandPhraseology` so radio log shows real ATC phrasing instead of abbreviated codes
- [ ] `isCommand` in parser treats single-letter taxiway names `L`, `T`, `GA` as command tokens — will break `TX L T` routes if maps use those taxiway names. Guard or namespace taxiway names when defining maps.
- [ ] Add integration tests for mouse click → command tree → input manipulation path in `game/model.go`. Bubblezone zone detection requires global manager init and synthetic mouse events.

---

## Priority 2: Separation Rules (Future)

**Status: Not Started**

Replace binary collision with distance-based separation enforcement.

- Minimum lateral separation: 3 NM (grid cells as proxy)
- Minimum vertical separation: 1000ft (1 altitude unit) when laterally close
- Separation violation warnings (visual + message) before collision
- Point penalty for separation violations (-50 per violation tick)
- Near-miss tracking for stats
- Collision still ends game, but separation violations degrade score first

---

## Priority 3: Expanded Command Set (Future)

**Status: Not Started**

Richer ATC commands closer to real phraseology. Note: `T`, `PB`, `TX`, `HS`, `CR`, `GATE` are now part of Ground Operations (Priority 1). This section covers the remaining airborne commands.

- `D <fix>` — Direct to waypoint/fix (aircraft auto-navigates to the named fix)
- `HLD <fix>` — Hold at fix (circle a waypoint)
- `GA` — Go around (abort landing, climb, maintain heading)
- `EX` — Expedite (double altitude change rate)
- `TL <heading>` / `TR <heading>` — Turn left/right to heading (force turn direction)
- `L <runway>` — Clear to land on specific runway (for multi-runway airports)

---

## Priority 4: Aircraft Types (Future)

**Status: Not Started**

Different aircraft categories with gameplay-affecting differences.

- Categories: Light (L), Medium (M), Heavy (H)
- Light: faster turns, slower speed range, shorter landing distance
- Heavy: slower turns, faster speed range, longer runway occupancy after landing
- Wake turbulence: extra separation required behind Heavy aircraft
- Display aircraft type on flight strips (e.g., B738, A321, C172)
- Spawner assigns type based on map/difficulty

---

## Priority 5: Pilot Patience / Pressure System (Future)

**Status: Not Started**

Aircraft request instructions and expect timely responses.

- New aircraft announce entry with a request ("AA123 with you, requesting approach")
- Patience timer per aircraft — starts when they enter or request something
- If ignored too long: repeated requests, then penalty, then aircraft leaves (score loss)
- Visual indicator on flight strip (patience bar or color change)
- Pressure creates urgency to manage multiple aircraft actively
- Optional: global pressure bar (IATC4 style) — if aggregate pressure hits 100%, game over

---

## Priority 6: Scenarios / Stages (Future)

**Status: Not Started**

Structured challenges beyond infinite sandbox mode.

- Scenario definition format (YAML or Go structs): timed aircraft spawns, objectives, par score
- Objective types: land N aircraft, clear N departures, survive N minutes, handle rush hour wave
- Star rating (1-3 stars) based on score/time/separation violations
- Scenario select screen in menu
- Tutorial scenarios that teach one concept at a time
- Campaign progression: unlock harder scenarios by completing easier ones

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

---

## Notes

- Each priority can be broken into phases when implementation begins
- Features should be developed with tests (TDD where practical)
- Code review after each feature before committing
- Update GUIDE.md and CLAUDE.md as features land
