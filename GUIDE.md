# How to Play ATC

## Overview

You are an air traffic controller guiding aircraft through approach, landing, ground taxi, and departure. Arrivals enter your airspace and must be vectored to the runway. Departures spawn at gates and need pushback, taxi, and takeoff clearance. Each successful operation scores a point. Collisions end the game.

## Quick Start

```bash
make run          # TRACON mode (approach radar)
make run-tower    # Tower mode (airport surface)
make run-combined # Combined mode (both)
make dev          # Developer mode with /commands
```

## Controller Roles

The game offers three perspectives on the same airspace:

### TRACON (Approach/Departure)

You manage the terminal airspace — vectoring arrivals toward the runway and guiding departures to climb out. The STARS-style radar shows aircraft as bright green dots with history trails and data blocks. Ground operations are automated.

### Tower (Local/Ground Control)

You manage the runway environment and airport surface. TRACON delivers arrivals pre-sequenced on final approach. You handle landing clearances, taxi routing, pushback, and takeoff. The ASDE-X surface radar shows runways as filled rectangles, taxiways as labeled paths, and aircraft as directional chevrons. An approach inset shows incoming traffic.

### Combined

Full control of both approach and surface — the most challenging mode.

## The Display

### TRACON Radar (STARS-style)

- **Aircraft**: Bright green dots with fading history trails showing their path
- **Data blocks**: Callsign + altitude (hundreds of feet) + speed, connected by leader lines
- **Runway**: White centerline with dashed extended approach course
- **Fixes**: Dim blue triangles (waypoints), circles (VORs/airports), crosses (intersections)
- **Range rings**: Concentric circles centered on the airport
- **Compass rose**: Heading marks every 10 degrees around the scope edge

### Tower Radar (ASDE-X-style)

- **Runway**: Filled gray rectangle with threshold markings and centerline dashes
- **Taxiways**: Gray paths with letter labels (A, B, C, D...)
- **Gates**: Blue rectangles with ID labels (G1, G2...)
- **Hold-short**: Yellow dashed bars across taxiways
- **Aircraft**: Directional chevrons — green for arrivals, cyan for departures
- **Approach inset**: Small box in the corner showing incoming traffic on final

### Flight Strip Sidebar

The right panel shows a strip for each aircraft:

- **Callsign** (color-coded): green = normal, yellow = landing/waiting, orange = impatient, red = angry/conflict, cyan = ground
- **State tag**: APPR, LAND, DEPT, TAXI, GATE, PUSH, HOLD, TKOF
- **Airborne info**: heading, altitude with climb/descend arrow, speed, pending targets
- **Ground info**: gate assignment, runway assignment
- **Tower mode**: strips split into ARRIVALS and DEPARTURES sections

Click any strip to auto-fill the callsign in the command input.

### HUD Bar

Top of screen showing: role, score, aircraft count, near misses, elapsed time, speed/freeze status.

### Radio Log

Above the command input, showing the last 5 messages:
- **Cyan**: Pilot messages (check-ins, requests)
- **Green**: Your ATC commands (translated to phraseology)
- **Gray**: System messages
- **Yellow**: Urgent (traffic alerts)
- **Red**: Emergency (collisions)

## Controls

### Camera

| Control | Action |
|---------|--------|
| Scroll wheel | Zoom in/out (centered on cursor, 0.5x-8x) |
| Click + drag | Pan the viewport |
| Arrow keys | Pan (when input is empty) |
| +/- | Zoom from keyboard (when input is empty) |
| Home | Reset to default view for current role |

### Time

| Control | Action |
|---------|--------|
| p | Freeze / unfreeze time (when input is empty) |
| [ | Slow down (1x minimum) |
| ] | Speed up (12x maximum) |

### Game

| Control | Action |
|---------|--------|
| Type + Enter | Submit ATC command |
| Click strip | Select aircraft (fills callsign in input) |
| Esc | Quit |
| R | Restart (on game over screen) |

### Minimap

Appears automatically in the bottom-left corner when zoomed past 1.5x. Shows the full map with aircraft as dots and your current viewport outlined.

## Issuing Commands

Type in the command prompt at the bottom and press Enter:

```
<CALLSIGN> <COMMAND> [<COMMAND>...]
```

Commands can be chained: `AA123 H270 A3 S2` issues heading, altitude, and speed in one go.

### Airborne Commands

| Command | Description | Example |
|---------|-------------|---------|
| `H<0-359>` | Set heading | `AA123 H270` |
| `A<1-40>` | Set altitude (thousands of feet) | `AA123 A3` |
| `S<1-5>` | Set speed | `AA123 S2` |
| `L [runway]` | Clear to land (optionally on specific runway) | `AA123 L` or `AA123 L 28R` |
| `D <fix>` | Direct to waypoint (auto-navigates) | `AA123 D MAFAN` |
| `HLD <fix>` | Hold at fix (right-turn racetrack) | `AA123 HLD BOKNE` |
| `TL <heading>` | Turn LEFT to heading | `AA123 TL 270` |
| `TR <heading>` | Turn RIGHT to heading | `AA123 TR 090` |
| `EX` | Expedite altitude change (2x rate) | `AA123 A3 EX` |
| `GA` | Go around (abort landing) | `AA123 GA` |

### Ground Commands

| Command | Description | Example |
|---------|-------------|---------|
| `PB [runway]` | Approve pushback (optional: expect runway) | `DL789 PB 27` |
| `TX <taxiway...>` | Taxi via named taxiways | `DL789 TX B A D` |
| `HS <runway>` | Hold short of runway | `DL789 HS 27` |
| `CR <runway>` | Clear to cross runway | `DL789 CR 27` |
| `T` | Clear for takeoff | `DL789 T` |
| `GATE <gate>` | Assign gate (after landing) | `AA123 GATE G2` |

### Command Availability by Role

| Command | TRACON | Tower | Combined |
|---------|--------|-------|----------|
| H, A, S, L, GA | Yes | Yes | Yes |
| D, HLD, TL, TR, EX | Yes | No | Yes |
| PB, TX, HS, CR, T, GATE | No | Yes | Yes |

## Arrival Flow

### 1. Vector to Runway

Aircraft enter at the radar edges. Guide them toward the runway:

1. Set heading toward the runway: `AA123 H270`
2. Begin descent: `AA123 A5`, then later `AA123 A1`
3. Adjust speed if needed: `AA123 S2`

Or use direct-to-fix: `AA123 D TORIE` to navigate to a waypoint near the runway, then vector from there.

### 2. Clear to Land

When the aircraft is:
- Within 2 grid cells of the runway
- Within +/-10 degrees of the runway heading
- At altitude 1 (1000ft)

Issue: `AA123 L`

The aircraft lands automatically when all conditions are met.

### 3. Taxi to Gate (Combined/Tower mode)

After landing, assign a gate: `AA123 GATE G2`

The aircraft taxis to the gate. Arrival complete, +1 score.

### Holding

If you need an aircraft to wait (e.g., traffic ahead), put it in a hold:

```
AA123 HLD BOKNE
```

The aircraft flies to the fix and enters a standard right-turn racetrack pattern. It continues holding until you issue a new command (heading, direct, or land clears the hold).

## Departure Flow

Departures spawn at gates and request pushback on the radio.

### Full Sequence

1. **Pushback**: `DL789 PB 27` (approve, expect runway 27)
2. **Taxi**: `DL789 TX B A D` (via taxiways B, A, D)
3. **Hold short**: `DL789 HS 27`
4. **Takeoff**: `DL789 T`
5. Aircraft rolls, lifts off, climbs to 5000ft on runway heading
6. Scores +1 when leaving airspace (or at altitude 3 in Tower mode)

## Separation

Aircraft must maintain minimum separation:
- **Lateral**: 3 grid cells
- **Vertical**: 1 altitude unit (1000ft) when laterally close

Violations trigger:
- TRAFFIC ALERT radio message
- Violating aircraft blink red on radar
- Score penalty (-50 per violation event)
- Near-miss counter in HUD

**Collision** (same position + same altitude) = game over.

## Pilot Patience

Aircraft expect timely instructions. If ignored:
- **30 seconds**: "still waiting for vectors" (strip turns yellow)
- **40 seconds**: "requesting ANY instructions!" (strip turns orange)
- **50 seconds**: Aircraft leaves your airspace, -1 score (strip blinks red)

Any command resets the patience timer.

## Scoring

| Event | Points |
|-------|--------|
| Aircraft lands | +1 |
| Departure leaves airspace | +1 |
| Tower: departure handed off at altitude 3 | +1 |
| Separation violation | -50 |
| Aircraft leaves due to impatience | -1 |

Score cannot go below 0.

## Difficulty Progression

Within each game, difficulty increases over time:
- **Aircraft cap**: Starts at 5, increases by 2 per minute, capped by difficulty setting (Easy: 8, Normal: 15, Hard: 25)
- **Spawn interval**: Starts at 5 seconds, decreases to 1.5 seconds over 5 minutes, scaled by difficulty

## Maps

### San Diego (SAN)

Single runway 9/27. Coastal approach from the west. 4 gates. Good for learning.

Fixes: MAFAN, BOKNE, TORIE, SARGS, SWATT, NKX, MZB, SAN, NZY, PGY, NRS, TIJ, UN, LOWMA, LYNDI, LUCKI

### Chicago O'Hare (ORD)

Parallel runways 10L/28R and 10R/28L. 4 gates between runways. Multi-runway challenge — use `L 28R` or `L 28L` to direct aircraft to specific runways.

Fixes: PLANO, DUPAGE, CMSKY, MOBLE, ORD, MIDWY, PEKNY, BRAVE, GLENW, JOT

### Tutorial

Small map (90x40) with one runway 9/27 and 3 gates. Fewer fixes, less clutter. Best for first-time players.

Fixes: NORTH, EAST, SOUTH, WEST, CTR

## Developer Mode

Launch with `make dev` or `./atc -dev` to enable `/` commands:

| Command | Effect |
|---------|--------|
| `/god` | Toggle god mode (collisions don't end game) |
| `/pause` | Toggle automatic spawner |
| `/clear` | Remove all aircraft |
| `/spawn` | Spawn one arrival |
| `/spawn dep` | Spawn one departure at a gate |

## Tips

1. **Start with Tutorial on Easy** — small map, slow spawns, learn the basics
2. **Use Short callsigns** (set via `--role` flag) — "A12" is faster to type than "AA123"
3. **Click flight strips** — never type a callsign when you can click it
4. **Turn and descend early** — turns take ~9 seconds for 90 degrees, altitude changes are gradual
5. **Use direct-to-fix** — `D TORIE` is faster than manual heading vectors for long approaches
6. **Separate with altitude** — keep aircraft at different altitudes until final approach
7. **Use holds** — `HLD BOKNE` buys time when the runway is busy
8. **Freeze time** — press `p` to pause and assess the situation before it gets overwhelming
9. **Speed up during lulls** — press `]` to skip the quiet moments, `[` to slow back down
10. **Zoom in** — scroll wheel to zoom into the airport surface for ground operations
11. **Watch the radio** — pilot requests tell you who needs attention next
12. **Process departures during lulls** — pushback and taxi when arrivals are stable
