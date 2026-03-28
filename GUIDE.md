# How to Play ATC

## Overview

You are an approach controller. Aircraft enter your airspace from the edges of the radar screen, and your job is to guide them safely to the runway using text commands. Each successful landing scores a point. If two aircraft collide (same position and altitude), the game is over.

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
|  Score, aircraft |  ─────────────────────────────── |
|  count, time,    |  AA123           APPR            |
|  messages        |   090  ↓08  S3                   |
|                  |   → H270 A3                      |
+-- Radar ---------+  ─────────────────────────────── |
| ╭──────────────╮ |  UA456           LAND            |
| │              │ |   265   02  S2                   |
| │  ^ MAFAN     │ |                                  |
| │       * NKX  │ |  ─────────────────────────────── |
| │              │ |                                   |
| │  9 ======= 27│ |                                   |
| │              │ |                                   |
| ╰──────────────╯ |                                   |
+------------------+-----------------------------------+
  ATC> _                          (command input)
```

### Radar Symbols

| Symbol | Meaning |
|--------|---------|
| `@` | Aircraft (with callsign label) |
| `X` | Crashed aircraft |
| `.` | Trail dot (previous position) |
| `=` | Runway |
| `9` / `27` | Runway heading numbers |
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
  - CRASH = crashed (red)
- **Line 2**: Current heading, altitude with arrow (↑ climbing, ↓ descending), speed
- **Line 3**: Pending commands (only shown when targets differ from current values)

**Click any flight strip** to auto-fill the callsign in the command input. Then you only need to type the command itself.

The strip panel scrolls automatically when there are many aircraft. Use the mouse wheel to scroll through strips.

### HUD

The top bar shows a stats table:

| Field | Meaning |
|-------|---------|
| SCORE | Number of successful landings |
| AIRCRAFT | Currently active aircraft count |
| TIME | Elapsed game time (MM:SS) |

Below the stats, the last 5 messages are displayed (commands confirmed, errors, aircraft entering/landing).

## Issuing Commands

Type commands in the input bar at the bottom and press Enter:

```
<CALLSIGN> <COMMAND> [<COMMAND>...]
```

### Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `H<0-359>` | Set heading (degrees, leading zeros OK) | `AA123 H270` |
| `A<1-40>` | Set altitude (thousands of feet) | `AA123 A3` |
| `S<1-5>` | Set speed | `AA123 S2` |
| `L` | Clear aircraft to land | `AA123 L` |

Commands can be chained in a single line:

```
AA123 H270 A3 S2
```

This sets AA123's heading to 270, altitude to 3000ft, and speed to 2 in one command.

### Quick Entry with Mouse

Instead of typing the full callsign, **click a flight strip** on the right panel. The callsign is auto-filled in the command input. Then just type the command:

```
Click strip for AA123 → input shows "AA123 "
Type: H270 A3
Press Enter
```

## Landing Aircraft

Landing is the core objective. To successfully land an aircraft, all four conditions must be met:

1. **Clearance**: You must issue the `L` command (e.g., `AA123 L`)
2. **Position**: Aircraft must be within 2 grid cells of the runway
3. **Heading**: Within +/-10 degrees of the runway heading
4. **Altitude**: Must be at altitude 1 (1000ft)

### Landing Strategy

A typical approach sequence:

1. Aircraft enters at the edge — note its heading, altitude, and position
2. Issue heading command to aim toward the runway: `AA123 H270`
3. Begin descent: `AA123 A5` then later `AA123 A1`
4. When close and aligned, clear to land: `AA123 L`
5. The aircraft lands automatically when all conditions are met

**Tips:**
- Start turning early — turns take ~9 seconds for 90 degrees
- Altitude changes are gradual — descend well before the runway
- Speed affects how fast aircraft cross the map; slower gives more time to align
- Watch the flight strip's arrow (↑/↓) to confirm altitude is changing
- The pending targets line (→ H270 A3) confirms your commands were received

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
5. **Turn early** — aircraft take ~9 seconds to turn 90 degrees at the game's turn rate
6. **Descend early** — altitude changes are gradual; start descending well before the runway
7. **Watch the pending line** — the "→ H270 A3" line on flight strips confirms your command was received
8. **Use separation** — keep aircraft at different altitudes until they need to descend for landing
9. **Enable plane trails** — visual path history helps you understand turn radius and approach angles
