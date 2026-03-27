package gamemap

// All returns every registered map.
func All() []Map {
	return []Map{
		SanDiego(),
		Chicago(),
		Tutorial(),
	}
}

// ByID looks up a map by its ID. Returns the tutorial map if not found.
func ByID(id string) Map {
	for _, m := range All() {
		if m.ID == id {
			return m
		}
	}
	return Tutorial()
}

// SanDiego models the SAN TRACON approach area.
func SanDiego() Map {
	return Map{
		ID:          "san",
		Name:        "San Diego TRACON",
		Description: "SAN approach — runway 9/27, coastal terrain",
		Width:       80,
		Height:      35,
		Runways: []Runway{
			{Name: "9/27", X: 35, Y: 18, Heading: 270, Length: 7},
		},
		Fixes: []Fix{
			// Waypoints
			{Name: "MAFAN", X: 15, Y: 3, Type: FixWaypoint},
			{Name: "BOKNE", X: 20, Y: 9, Type: FixWaypoint},
			{Name: "TORIE", X: 30, Y: 14, Type: FixWaypoint},
			{Name: "SARGS", X: 18, Y: 20, Type: FixWaypoint},
			{Name: "SWATT", X: 58, Y: 19, Type: FixWaypoint},

			// Airports / VORs
			{Name: "NKX", X: 38, Y: 13, Type: FixVOR},
			{Name: "MZB", X: 32, Y: 18, Type: FixVOR},
			{Name: "SAN", X: 48, Y: 19, Type: FixAirport},
			{Name: "NZY", X: 36, Y: 22, Type: FixVOR},
			{Name: "PGY", X: 58, Y: 25, Type: FixVOR},
			{Name: "NRS", X: 48, Y: 28, Type: FixVOR},
			{Name: "TIJ", X: 60, Y: 27, Type: FixVOR},
			{Name: "UN", X: 55, Y: 30, Type: FixVOR},

			// Intersection fixes
			{Name: "LOWMA", X: 30, Y: 24, Type: FixIntersection},
			{Name: "LYNDI", X: 63, Y: 22, Type: FixIntersection},
			{Name: "LUCKI", X: 73, Y: 20, Type: FixIntersection},
		},
	}
}

// Chicago models a simplified ORD approach area.
func Chicago() Map {
	return Map{
		ID:          "ord",
		Name:        "Chicago O'Hare",
		Description: "ORD approach — parallel runways 10L/28R and 10R/28L",
		Width:       80,
		Height:      35,
		Runways: []Runway{
			{Name: "10L/28R", X: 40, Y: 15, Heading: 280, Length: 7},
			{Name: "10R/28L", X: 40, Y: 20, Heading: 280, Length: 7},
		},
		Fixes: []Fix{
			{Name: "PLANO", X: 10, Y: 5, Type: FixWaypoint},
			{Name: "DUPAGE", X: 20, Y: 10, Type: FixVOR},
			{Name: "CMSKY", X: 55, Y: 6, Type: FixWaypoint},
			{Name: "MOBLE", X: 65, Y: 12, Type: FixWaypoint},
			{Name: "ORD", X: 42, Y: 17, Type: FixAirport},
			{Name: "MIDWY", X: 38, Y: 27, Type: FixAirport},
			{Name: "PEKNY", X: 12, Y: 20, Type: FixIntersection},
			{Name: "BRAVE", X: 68, Y: 25, Type: FixWaypoint},
			{Name: "GLENW", X: 25, Y: 28, Type: FixIntersection},
			{Name: "JOT", X: 15, Y: 32, Type: FixVOR},
		},
	}
}

// Tutorial is a simple map for learning the game.
func Tutorial() Map {
	return Map{
		ID:          "tutorial",
		Name:        "Tutorial",
		Description: "Small map with one runway — learn the basics",
		Width:       60,
		Height:      30,
		Runways: []Runway{
			{Name: "9/27", X: 30, Y: 25, Heading: 270, Length: 5},
		},
		Fixes: []Fix{
			{Name: "NORTH", X: 30, Y: 5, Type: FixWaypoint},
			{Name: "EAST", X: 50, Y: 15, Type: FixWaypoint},
			{Name: "SOUTH", X: 30, Y: 28, Type: FixWaypoint},
			{Name: "WEST", X: 10, Y: 15, Type: FixWaypoint},
			{Name: "CTR", X: 30, Y: 15, Type: FixVOR},
		},
	}
}
