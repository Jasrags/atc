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
		Width:       120,
		Height:      50,
		Runways: []Runway{
			{Name: "9/27", X: 52, Y: 26, Heading: 270, Length: 9},
		},
		Fixes: []Fix{
			// Waypoints
			{Name: "MAFAN", X: 22, Y: 4, Type: FixWaypoint},
			{Name: "BOKNE", X: 30, Y: 13, Type: FixWaypoint},
			{Name: "TORIE", X: 45, Y: 20, Type: FixWaypoint},
			{Name: "SARGS", X: 27, Y: 29, Type: FixWaypoint},
			{Name: "SWATT", X: 87, Y: 27, Type: FixWaypoint},

			// Airports / VORs
			{Name: "NKX", X: 57, Y: 18, Type: FixVOR},
			{Name: "MZB", X: 48, Y: 26, Type: FixVOR},
			{Name: "SAN", X: 72, Y: 27, Type: FixAirport},
			{Name: "NZY", X: 54, Y: 32, Type: FixVOR},
			{Name: "PGY", X: 87, Y: 36, Type: FixVOR},
			{Name: "NRS", X: 72, Y: 40, Type: FixVOR},
			{Name: "TIJ", X: 90, Y: 39, Type: FixVOR},
			{Name: "UN", X: 82, Y: 43, Type: FixVOR},

			// Intersection fixes
			{Name: "LOWMA", X: 45, Y: 35, Type: FixIntersection},
			{Name: "LYNDI", X: 95, Y: 32, Type: FixIntersection},
			{Name: "LUCKI", X: 110, Y: 29, Type: FixIntersection},
		},
	}
}

// Chicago models a simplified ORD approach area.
func Chicago() Map {
	return Map{
		ID:          "ord",
		Name:        "Chicago O'Hare",
		Description: "ORD approach — parallel runways 10L/28R and 10R/28L",
		Width:       120,
		Height:      50,
		Runways: []Runway{
			{Name: "10L/28R", X: 60, Y: 21, Heading: 280, Length: 9},
			{Name: "10R/28L", X: 60, Y: 29, Heading: 280, Length: 9},
		},
		Fixes: []Fix{
			{Name: "PLANO", X: 15, Y: 7, Type: FixWaypoint},
			{Name: "DUPAGE", X: 30, Y: 14, Type: FixVOR},
			{Name: "CMSKY", X: 82, Y: 8, Type: FixWaypoint},
			{Name: "MOBLE", X: 97, Y: 17, Type: FixWaypoint},
			{Name: "ORD", X: 63, Y: 25, Type: FixAirport},
			{Name: "MIDWY", X: 57, Y: 39, Type: FixAirport},
			{Name: "PEKNY", X: 18, Y: 29, Type: FixIntersection},
			{Name: "BRAVE", X: 102, Y: 36, Type: FixWaypoint},
			{Name: "GLENW", X: 37, Y: 40, Type: FixIntersection},
			{Name: "JOT", X: 22, Y: 46, Type: FixVOR},
		},
	}
}

// Tutorial is a simple map for learning the game.
func Tutorial() Map {
	return Map{
		ID:          "tutorial",
		Name:        "Tutorial",
		Description: "Small map with one runway — learn the basics",
		Width:       90,
		Height:      40,
		Runways: []Runway{
			{Name: "9/27", X: 45, Y: 33, Heading: 270, Length: 7},
		},
		Fixes: []Fix{
			{Name: "NORTH", X: 45, Y: 7, Type: FixWaypoint},
			{Name: "EAST", X: 75, Y: 20, Type: FixWaypoint},
			{Name: "SOUTH", X: 45, Y: 37, Type: FixWaypoint},
			{Name: "WEST", X: 15, Y: 20, Type: FixWaypoint},
			{Name: "CTR", X: 45, Y: 20, Type: FixVOR},
		},
	}
}
