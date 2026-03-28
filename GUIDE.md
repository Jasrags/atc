# How to Play ATC

## Overview

You are the controller managing aircraft from approach through landing, ground taxi, and departure. Arrivals enter your airspace from the radar edges — guide them to the runway, land them, and taxi them to a gate. Departures spawn at gates — push them back, taxi to the runway, and clear them for takeoff. Each successful arrival (at gate) and departure (leaves airspace) scores a point. If two aircraft collide, the game is over.

## Real-World ATC Flow

This game simulates several positions in the real ATC system. Here's how it maps:

### Arrivals

```
Center (ARTCC)
  |  hands off ~30-50nm out, descending from cruise
  v
TRACON Approach            <-- YOU: vector, sequence, clear approach
  |  vectors, sequences, issues approach clearance
  v
Tower (Local Control)      <-- YOU: clear to land, watch runway
  |  clears to land, clears when runway vacated
  v
Ground Control             <-- YOU: taxi to gate
  |  taxis to gate
  v
Ramp (arrival complete)
```

### Departures

```
Clearance Delivery
  |  IFR clearance (route, altitude, squawk)
  v
Ground Control             <-- YOU: pushback, taxi to runway
  |  pushback approval, taxi routing to hold short line
  v
Tower (Local Control)      <-- YOU: takeoff clearance
  |  position and hold, takeoff clearance, initial heading/altitude
  |  watches for runway conflicts, wake turbulence spacing
  v
TRACON Departure           <-- YOU: climb out, vectors
  |  radar contact, climb instructions, vectors
  |  deconflicts with arrival traffic
  v
Center (ARTCC)             (departure leaves your airspace — scored)
```

In the game, you handle all of these roles simultaneously through the command input.

## Getting Started

Launch the game:

```bash
./atc
```

You'll see the main menu. Press **N** (or select "New Game" and press Enter) to open the game setup screen.

## Game Setup

Before each game, you configure five settings using a combined setup screen:

```
  GAME SETUP

  Map
  > San Diego TRACON
    Chicago O'Hare
    Tutorial

  Difficulty
    Easy
  > Normal
    Hard

  Game Mode
  > Arrivals Only

  Callsign Style
  > ICAO (AA123)
    Short (A12)

  Plane Trails
    On
  > Off
```

### Navigation

| Key | Action |
|-----|--------|
| Tab | Move to next section |
| Shift+Tab | Move to previous section |
| Up/Down (or k/j) | Select option within section |
| Enter | Start the game |
| Esc | Back to main menu |

### Settings Explained

**Map** controls the radar size, runway layout, and navigation fixes:

| Map | Size | Runways | Best for |
|-----|------|---------|----------|
| Tutorial | 90x40 | 9/27 | Learning the basics |
| San Diego TRACON | 120x50 | 9/27 | Single runway, realistic fixes |
| Chicago O'Hare | 120x50 | 10L/28R, 10R/28L | Parallel runway challenge |

**Difficulty** affects how fast aircraft spawn and how many you'll manage:

| | Easy | Normal | Hard |
|--|------|--------|------|
| Spawn rate | Slow (1.5x) | Standard | Fast (0.6x) |
| Max aircraft | 8 | 15 | 25 |
| Speed range | 1-3 | 2-4 | 2-5 |

**Callsign Style** changes the format of aircraft identifiers:
- **ICAO** (AA123) — realistic airline + flight number, 5 characters
- **Short** (A12) — single letter + 2 digits, faster to type

**Plane Trails** shows the last 5 positions of each aircraft as dots on the radar. Helpful for visualizing flight paths and turn arcs.

## The Radar Screen

Once the game starts, your screen is divided into three areas:

```
+-- HUD (top) ----+-------- Flight Strips (right) ---+
|  Score, aircraft |  AA123           APPR            |
|  count, time     |   090  ↓08  S3                   |
|                  |   → H270 A3                      |
+-- Radar ---------+  ─────────────────────────────── |
| ╭──────────────╮ |  UA456           TAXI            |
| │  ^ MAFAN     │ |   Gate G2 | TX A B              |
| │       * NKX  │ |  ─────────────────────────────── |
| │              │ |  DL789           GATE            |
| │  9 ======= 27│ |   Gate G1 | Rwy 27              |
| │    |--A--|   │ |                                   |
| │    # # #     │ |                                   |
| ╰──────────────╯ |                                   |
+------------------+-----------------------------------+
| RADIO                                                |
| 01:23 AA123: approach, with you, heading 270 at 5000 |
| 01:25 ATC → AA123: turn right heading 090, ALT 3     |
| 01:30 DL789: at gate G1, requesting pushback          |
+------------------------------------------------------+
  ATC> AA123 _
  Commands: [H] Heading  [A] Altitude  [S] Speed  [L] Land
```

### Radar Symbols

**Airborne aircraft:**

| Symbol | Meaning |
|--------|---------|
| `@` | Aircraft in flight (with callsign label) |
| `X` | Crashed aircraft |
| `.` | Trail dot (previous position) |

**Ground aircraft:**

| Symbol | Meaning |
|--------|---------|
| `v` | Taxiing |
| `#` | At gate |
| `<` | Pushing back |
| `!` | Holding short of runway |
| `>` | On runway (taking off) |

**Map features:**

| Symbol | Meaning |
|--------|---------|
| `=` | Runway |
| `9` / `27` | Runway heading numbers |
| `-` / `\|` | Taxiway |
| `#` | Gate position (blue) |
| `:` | Hold-short line (yellow) |
| `^` | Waypoint fix |
| `*` | VOR navigation aid |
| `o` | Airport |
| `+` | Intersection fix |

### Flight Strips

The right panel shows a strip for each active aircraft:

```
──────────────────────────────
AA123           APPR
 090  ↓08  S3
 → H270 A3
──────────────────────────────
```

- **Line 1**: Callsign (color-coded by state) and status
  - APPR = approaching (green)
  - LAND = cleared to land (orange)
  - DEPT = departing (green)
  - TAXI = taxiing on ground (yellow)
  - GATE = at gate (yellow)
  - PUSH = pushing back (yellow)
  - HOLD = holding short of runway (yellow)
  - TKOF = on runway, taking off (yellow)
  - DONE = landed (dim)
  - CRASH = crashed (red)
- **Line 2**: Airborne aircraft show heading, altitude (with ↑/↓ arrows), and speed. Ground aircraft show gate assignment, runway, and taxi route.
- **Line 3**: Pending commands (only shown for airborne aircraft when targets differ from current values)

**Click any flight strip** to auto-fill the callsign in the command input. Then you only need to type the command itself.

The strip panel scrolls automatically when there are many aircraft. Use the mouse wheel to scroll through strips.

### HUD

The top bar shows a stats table:

| Field | Meaning |
|-------|---------|
| SCORE | Arrivals at gate + departures leaving airspace |
| AIRCRAFT | Currently active aircraft count |
| TIME | Elapsed game time (MM:SS) |

### Radio Panel

Between the game area and the command input, the radio panel shows all communications:

- **Inbound** (pilot → you): Aircraft check-ins, pushback requests, position reports — shown in cyan
- **Outbound** (you → pilot): Your commands translated into ATC phraseology — shown in green
- **Emergency**: Collision alerts — shown in red

The radio panel scrolls. Recent messages appear at the bottom.

## Issuing Commands

Type commands in the input bar at the bottom and press Enter:

```
<CALLSIGN> <COMMAND> [<COMMAND>...]
```

### Airborne Commands

| Command | Description | Example |
|---------|-------------|---------|
| `H<0-359>` | Set heading (degrees) | `AA123 H270` |
| `A<1-40>` | Set altitude (thousands of feet) | `AA123 A3` |
| `S<1-5>` | Set speed | `AA123 S2` |
| `L` | Clear aircraft to land | `AA123 L` |
| `GA` | Go around (abort landing) | `AA123 GA` |

Airborne commands can be chained: `AA123 H270 A3 S2`

### Ground Commands

| Command | Description | Example |
|---------|-------------|---------|
| `PB [runway]` | Approve pushback (optional: expect runway) | `DL789 PB 27` |
| `TX <taxiway...>` | Taxi via named taxiways | `DL789 TX B A D` |
| `HS <runway>` | Hold short of runway | `DL789 HS 27` |
| `CR <runway>` | Clear to cross runway | `DL789 CR 27` |
| `T` | Clear for takeoff | `DL789 T` |
| `GATE <gate>` | Assign gate (after landing) | `AA123 GATE G2` |

### Command Tree (Interactive)

When you select an aircraft (click a strip or type a callsign followed by a space), a context-sensitive command menu appears below the input:

```
ATC> AA123 _
Commands: [H] Heading  [A] Altitude  [S] Speed  [L] Land
```

Click any option or press its key. The tree adapts to the aircraft's state — ground aircraft show `PB`, `TX`, `HS`, `T` instead of airborne commands. After choosing a command with a value (like heading), the tree shows value options:

```
ATC> AA123 H _
Values: [030] [060] [090] [120] [150] [180] [210] [240] [270] [300] [330] [360]
```

After picking a value, chain more commands or press Enter to send.

### Quick Entry with Mouse

**Click a flight strip** to auto-fill the callsign, then type or click commands:

```
Click strip for AA123 → input shows "AA123 "
Type: H270 A3
Press Enter
```

## Arrival Flow

### 1. Approach and Landing

To successfully land an aircraft, all four conditions must be met:

1. **Clearance**: You must issue the `L` command (e.g., `AA123 L`)
2. **Position**: Aircraft must be within 2 grid cells of the runway
3. **Heading**: Within +/-10 degrees of the runway heading
4. **Altitude**: Must be at altitude 1 (1000ft)

A typical approach sequence:

1. Aircraft enters at the edge — note its heading, altitude, and position
2. Issue heading command to aim toward the runway: `AA123 H270`
3. Begin descent: `AA123 A5` then later `AA123 A1`
4. When close and aligned, clear to land: `AA123 L`
5. The aircraft lands automatically when all conditions are met

### 2. Taxi to Gate

After landing, the aircraft stays on the map in DONE state. Assign it a gate:

```
AA123 GATE G2
```

The aircraft taxis directly to the assigned gate. When it arrives, it's removed from the map (arrival complete, +1 score).

**Tips:**
- Start turning early — turns take ~9 seconds for 90 degrees
- Altitude changes are gradual — descend well before the runway
- Speed affects how fast aircraft cross the map; slower gives more time to align
- Watch the flight strip's arrow (↑/↓) to confirm altitude is changing
- The pending targets line (→ H270 A3) confirms your commands were received

## Departure Flow

Departures spawn at gates and announce themselves on the radio:

```
01:30 DL789: at gate G1, requesting pushback
```

### Full Departure Sequence

1. **Pushback**: `DL789 PB 27` (approve pushback, expect runway 27)
2. **Taxi to runway**: `DL789 TX B A D` (taxi via taxiways B, A, D to the runway end)
3. **Hold short**: `DL789 HS 27` (hold short of runway 27)
4. **Takeoff**: `DL789 T` (cleared for takeoff)
5. Aircraft rolls on runway (~1.5 seconds), lifts off, climbs to 5000ft heading in the runway direction
6. Departing aircraft scores +1 when it leaves your airspace

### Runway Headings

Each map shows runway heading numbers at each end (e.g., `9` and `27` for runway 9/27). The heading number is the runway direction divided by 10. To land on runway 27, your aircraft needs a heading near 270.

Chicago O'Hare has parallel runways (10L/28R and 10R/28L). Aircraft can land on either one.

## Game States

### Pause

Press **P** to pause. The game freezes — no aircraft movement or spawning. Press **P** again to resume. You can also press **Esc** to return to the main menu (abandons the current game) or **Q** to quit entirely.

### Game Over

A collision occurs when two aircraft occupy the same grid cell at the same altitude. Both aircraft are marked as crashed and the game ends.

From the game over screen:
- **R** — restart with the same settings
- **Esc** — return to main menu
- **Q** — quit

### Help

Press **?** at any time during gameplay to see the in-game help screen, which shows command syntax, landing requirements, and all keybindings. Press **?** or **Esc** to return.

## Keybindings Reference

### During Gameplay

| Key | Action |
|-----|--------|
| Enter | Submit command |
| P | Pause / Resume |
| ? | Toggle help |
| Esc | Back to menu |
| Ctrl+C | Force quit |
| Mouse click (strip) | Auto-fill callsign |
| Mouse wheel | Scroll flight strips |

### Main Menu

| Key | Action |
|-----|--------|
| Up/Down (or k/j) | Navigate |
| Enter | Select |
| N | New game |
| H or ? | Help |
| Q or Esc | Quit |

### Game Over

| Key | Action |
|-----|--------|
| R | Restart |
| Esc or Q | Return to menu |

## Difficulty Progression

Within a single game session, difficulty increases over time regardless of your chosen difficulty level:

- **Aircraft count ramp**: Starts at 5 max, increases by 2 per minute, capped at the difficulty ceiling (8/15/25)
- **Spawn interval ramp**: Starts at 5 seconds between spawns, decreases to 1.5 seconds over 5 minutes, multiplied by the difficulty factor

This means even on Easy, the game gets progressively harder. The difficulty setting controls the ceiling and rate of escalation.

## Tips for New Players

1. **Start with Tutorial on Easy** — small map, slow spawns, plenty of time to learn
2. **One aircraft at a time** — focus on getting one aircraft landed before managing multiples
3. **Use Short callsigns** — "A12" is much faster to type than "AA123"
4. **Click flight strips** — never type a callsign manually if you can click it
5. **Use the command tree** — click options instead of memorizing commands
6. **Turn early** — aircraft take ~9 seconds to turn 90 degrees at the game's turn rate
7. **Descend early** — altitude changes are gradual; start descending well before the runway
8. **Watch the pending line** — the "→ H270 A3" line on flight strips confirms your command was received
9. **Use separation** — keep aircraft at different altitudes until they need to descend for landing
10. **Enable plane trails** — visual path history helps you understand turn radius and approach angles
11. **Assign gates quickly** — landed aircraft sit on the runway until you give them a gate
12. **Process departures during lulls** — pushback and taxi commands when arrivals are stable
13. **Watch the radio** — pilot requests tell you who needs attention next
