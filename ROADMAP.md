# ATC Roadmap

Feature priorities for closing the gap with Tower!3D Pro and IATC4. Each feature is tracked here with status, and broken into implementation tasks when work begins.

## Priority 1: Departures

**Status: Not Started**

Add departure traffic to create the core arrival/departure sequencing puzzle.

- Aircraft spawn at gates/runway threshold with departure clearance requests
- New commands: T (cleared for takeoff), pushback/taxi (simplified)
- Departure flight strips (distinct from arrival strips)
- Aircraft climb out after takeoff, hand off when leaving airspace
- Runway occupancy — cannot land while a departure is on the runway
- Score for successful departures (aircraft exits airspace at assigned altitude/heading)

---

## Priority 2: Separation Rules

**Status: Not Started**

Replace binary collision with distance-based separation enforcement.

- Minimum lateral separation: 3 NM (grid cells as proxy)
- Minimum vertical separation: 1000ft (1 altitude unit) when laterally close
- Separation violation warnings (visual + message) before collision
- Point penalty for separation violations (-50 per violation tick)
- Near-miss tracking for stats
- Collision still ends game, but separation violations degrade score first

---

## Priority 3: Expanded Command Set

**Status: Not Started**

Richer ATC commands closer to real phraseology.

- `D <fix>` — Direct to waypoint/fix (aircraft auto-navigates to the named fix)
- `HLD <fix>` — Hold at fix (circle a waypoint)
- `GA` — Go around (abort landing, climb, maintain heading)
- `EX` — Expedite (double altitude change rate)
- `TL <heading>` / `TR <heading>` — Turn left/right to heading (force turn direction)
- `L <runway>` — Clear to land on specific runway (for multi-runway airports)
- `T` — Cleared for takeoff (departures)

---

## Priority 4: Aircraft Types

**Status: Not Started**

Different aircraft categories with gameplay-affecting differences.

- Categories: Light (L), Medium (M), Heavy (H)
- Light: faster turns, slower speed range, shorter landing distance
- Heavy: slower turns, faster speed range, longer runway occupancy after landing
- Wake turbulence: extra separation required behind Heavy aircraft
- Display aircraft type on flight strips (e.g., B738, A321, C172)
- Spawner assigns type based on map/difficulty

---

## Priority 5: Pilot Patience / Pressure System

**Status: Not Started**

Aircraft request instructions and expect timely responses.

- New aircraft announce entry with a request ("AA123 with you, requesting approach")
- Patience timer per aircraft — starts when they enter or request something
- If ignored too long: repeated requests, then penalty, then aircraft leaves (score loss)
- Visual indicator on flight strip (patience bar or color change)
- Pressure creates urgency to manage multiple aircraft actively
- Optional: global pressure bar (IATC4 style) — if aggregate pressure hits 100%, game over

---

## Priority 6: Scenarios / Stages

**Status: Not Started**

Structured challenges beyond infinite sandbox mode.

- Scenario definition format (YAML or Go structs): timed aircraft spawns, objectives, par score
- Objective types: land N aircraft, clear N departures, survive N minutes, handle rush hour wave
- Star rating (1-3 stars) based on score/time/separation violations
- Scenario select screen in menu
- Tutorial scenarios that teach one concept at a time
- Campaign progression: unlock harder scenarios by completing easier ones

---

## Priority 7: Ground Operations

**Status: Not Started**

Post-landing and pre-departure ground movement.

- Taxiway network defined per map (nodes + edges)
- After landing: aircraft must taxi to gate via assigned taxiways
- Before departure: aircraft taxis from gate to runway
- Runway crossing clearances
- Hold short commands
- Ground conflicts (two aircraft on same taxiway segment)
- This is the largest feature — consider as a v2.0 milestone

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
