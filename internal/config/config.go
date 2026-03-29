package config

// Difficulty controls spawn rate, max aircraft, and speed range.
type Difficulty int

const (
	DifficultyEasy Difficulty = iota
	DifficultyNormal
	DifficultyHard
)

func (d Difficulty) String() string {
	switch d {
	case DifficultyEasy:
		return "Easy"
	case DifficultyNormal:
		return "Normal"
	case DifficultyHard:
		return "Hard"
	default:
		return "Normal"
	}
}

// DifficultyParams holds the tunable parameters for each difficulty level.
type DifficultyParams struct {
	IntervalMultiplier float64 // Multiplier on spawn interval (higher = slower spawns)
	MaxAircraft        int     // Cap on simultaneous aircraft
	MinSpeed           int     // Minimum aircraft speed on spawn
	MaxSpeed           int     // Maximum aircraft speed on spawn
}

// Params returns the gameplay parameters for this difficulty.
func (d Difficulty) Params() DifficultyParams {
	switch d {
	case DifficultyEasy:
		return DifficultyParams{
			IntervalMultiplier: 1.5,
			MaxAircraft:        8,
			MinSpeed:           1,
			MaxSpeed:           3,
		}
	case DifficultyHard:
		return DifficultyParams{
			IntervalMultiplier: 0.6,
			MaxAircraft:        25,
			MinSpeed:           2,
			MaxSpeed:           5,
		}
	default: // Normal
		return DifficultyParams{
			IntervalMultiplier: 1.0,
			MaxAircraft:        15,
			MinSpeed:           2,
			MaxSpeed:           4,
		}
	}
}

// Role defines the player's controller position, scoping commands and automation.
type Role int

const (
	RoleTRACON   Role = iota // Approach/departure — ground is automated
	RoleTower                // Local/ground control — approach is automated
	RoleCombined             // Both TRACON + Tower (hardest)
)

func (r Role) String() string {
	switch r {
	case RoleTRACON:
		return "TRACON"
	case RoleTower:
		return "Tower"
	case RoleCombined:
		return "Combined"
	default:
		return "TRACON"
	}
}

var (
	groundCmds       = map[string]bool{"PB": true, "TX": true, "HS": true, "CR": true, "GATE": true, "T": true}
	towerBlockedCmds = map[string]bool{"D": true, "HLD": true, "TL": true, "TR": true, "EX": true}
)

// IsCommandAllowed reports whether the given command keyword is permitted for this role.
func (r Role) IsCommandAllowed(cmd string) bool {
	switch r {
	case RoleTRACON:
		return !groundCmds[cmd]
	case RoleTower:
		return !towerBlockedCmds[cmd]
	case RoleCombined:
		return true
	}
	return true
}

// GameMode controls which types of aircraft traffic are active.
type GameMode int

const (
	ModeArrivalsOnly GameMode = iota
)

func (m GameMode) String() string {
	switch m {
	case ModeArrivalsOnly:
		return "Arrivals Only"
	default:
		return "Arrivals Only"
	}
}

// CallsignStyle controls the format of generated aircraft callsigns.
type CallsignStyle int

const (
	CallsignICAO  CallsignStyle = iota // AA123 format
	CallsignShort                      // A12 format
)

func (c CallsignStyle) String() string {
	switch c {
	case CallsignICAO:
		return "ICAO (AA123)"
	case CallsignShort:
		return "Short (A12)"
	default:
		return "ICAO (AA123)"
	}
}

// GameConfig holds all user-configurable game settings.
type GameConfig struct {
	MapID         string
	Role          Role
	Difficulty    Difficulty
	GameMode      GameMode
	CallsignStyle CallsignStyle
	PlaneTrails   bool
}

// DefaultConfig returns the default game configuration.
func DefaultConfig() GameConfig {
	return GameConfig{
		MapID:         "san",
		Role:          RoleTRACON,
		Difficulty:    DifficultyNormal,
		GameMode:      ModeArrivalsOnly,
		CallsignStyle: CallsignICAO,
		PlaneTrails:   false,
	}
}

// RoleOptions returns the display labels for available role selections.
func RoleOptions() []string {
	return []string{"TRACON", "Tower", "Combined"}
}

// RoleFromIndex converts a setup screen selection index to a Role.
// This is needed because the options list may not match the iota order.
func RoleFromIndex(idx int) Role {
	roles := []Role{RoleTRACON, RoleTower, RoleCombined}
	if idx >= 0 && idx < len(roles) {
		return roles[idx]
	}
	return RoleTRACON
}

// DifficultyOptions returns the display labels for difficulty selection.
func DifficultyOptions() []string {
	return []string{"Easy", "Normal", "Hard"}
}

// GameModeOptions returns the display labels for game mode selection.
func GameModeOptions() []string {
	return []string{"Arrivals Only"}
}

// CallsignStyleOptions returns the display labels for callsign style selection.
func CallsignStyleOptions() []string {
	return []string{"ICAO (AA123)", "Short (A12)"}
}

// PlaneTrailsOptions returns the display labels for plane trails toggle.
func PlaneTrailsOptions() []string {
	return []string{"On", "Off"}
}
