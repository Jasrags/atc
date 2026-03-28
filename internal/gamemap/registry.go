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
//
// Ground layout (runway at Y=26, taxiway A at Y=28, gates at Y=30):
//
//	==RE9===RWY 9/27===RE27==   Y=26  Runway
//	  |                  |
//	 HS9               HS27     Y=27  Hold-short points
//	  |                  |
//	--A1----A2----A3----A4--    Y=28  Taxiway A
//	  |     |     |     |
//	  B1    B2    B3    B4      Y=29  Taxiway B connectors
//	  |     |     |     |
//	 [G1]  [G2]  [G3]  [G4]    Y=30  Gates
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
		TaxiNodes: []TaxiNode{
			// Taxiway A (east-west, south of runway)
			{ID: "A1", X: 44, Y: 28, Type: NodeIntersection},
			{ID: "A2", X: 49, Y: 28, Type: NodeIntersection},
			{ID: "A3", X: 55, Y: 28, Type: NodeIntersection},
			{ID: "A4", X: 60, Y: 28, Type: NodeIntersection},
			// Hold-short points
			{ID: "HS9", X: 44, Y: 27, Type: NodeHoldShort, Runway: "9"},
			{ID: "HS27", X: 60, Y: 27, Type: NodeHoldShort, Runway: "27"},
			// Runway entry/exit
			{ID: "RE9", X: 44, Y: 26, Type: NodeRunwayEntry, Runway: "9"},
			{ID: "RE27", X: 60, Y: 26, Type: NodeRunwayEntry, Runway: "27"},
			// Gates
			{ID: "G1", X: 44, Y: 30, Type: NodeGate},
			{ID: "G2", X: 49, Y: 30, Type: NodeGate},
			{ID: "G3", X: 55, Y: 30, Type: NodeGate},
			{ID: "G4", X: 60, Y: 30, Type: NodeGate},
		},
		TaxiEdges: []TaxiEdge{
			// Taxiway A segments
			{From: "A1", To: "A2", Taxiway: "A"},
			{From: "A2", To: "A3", Taxiway: "A"},
			{From: "A3", To: "A4", Taxiway: "A"},
			// Taxiway B (connectors to gates)
			{From: "G1", To: "A1", Taxiway: "B"},
			{From: "G2", To: "A2", Taxiway: "B"},
			{From: "G3", To: "A3", Taxiway: "B"},
			{From: "G4", To: "A4", Taxiway: "B"},
			// Taxiway C (west connector to runway)
			{From: "A1", To: "HS9", Taxiway: "C"},
			{From: "HS9", To: "RE9", Taxiway: "C"},
			// Taxiway D (east connector to runway)
			{From: "A4", To: "HS27", Taxiway: "D"},
			{From: "HS27", To: "RE27", Taxiway: "D"},
		},
		Gates: []Gate{
			{ID: "G1", NodeID: "G1"},
			{ID: "G2", NodeID: "G2"},
			{ID: "G3", NodeID: "G3"},
			{ID: "G4", NodeID: "G4"},
		},
	}
}

// Chicago models a simplified ORD approach area.
//
// Ground layout (two parallel runways, taxiway A between them, gates south):
//
//	==RE10L===RWY 10L/28R===RE28R==   Y=21
//	   |                      |
//	  HS10L                 HS28R      Y=22
//	   |                      |
//	---A1------A2------A3----A4---     Y=25  Taxiway A (between runways)
//	   |                      |
//	  HS10R                 HS28L      Y=28
//	   |                      |
//	==RE10R===RWY 10R/28L===RE28L==   Y=29
//	   |       |       |      |
//	---E1-----E2------E3-----E4---    Y=32  Taxiway E (south of runways)
//	   |       |       |      |
//	  [G1]   [G2]    [G3]   [G4]      Y=34  Gates
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
		TaxiNodes: []TaxiNode{
			// Taxiway A (between the two runways)
			{ID: "A1", X: 52, Y: 25, Type: NodeIntersection},
			{ID: "A2", X: 57, Y: 25, Type: NodeIntersection},
			{ID: "A3", X: 63, Y: 25, Type: NodeIntersection},
			{ID: "A4", X: 68, Y: 25, Type: NodeIntersection},
			// Hold-short for north runway (10L/28R)
			{ID: "HS10L", X: 52, Y: 22, Type: NodeHoldShort, Runway: "10L"},
			{ID: "HS28R", X: 68, Y: 22, Type: NodeHoldShort, Runway: "28R"},
			// Runway entry north
			{ID: "RE10L", X: 52, Y: 21, Type: NodeRunwayEntry, Runway: "10L"},
			{ID: "RE28R", X: 68, Y: 21, Type: NodeRunwayEntry, Runway: "28R"},
			// Hold-short for south runway (10R/28L)
			{ID: "HS10R", X: 52, Y: 28, Type: NodeHoldShort, Runway: "10R"},
			{ID: "HS28L", X: 68, Y: 28, Type: NodeHoldShort, Runway: "28L"},
			// Runway entry south
			{ID: "RE10R", X: 52, Y: 29, Type: NodeRunwayEntry, Runway: "10R"},
			{ID: "RE28L", X: 68, Y: 29, Type: NodeRunwayEntry, Runway: "28L"},
			// Taxiway E (south of south runway)
			{ID: "E1", X: 52, Y: 32, Type: NodeIntersection},
			{ID: "E2", X: 57, Y: 32, Type: NodeIntersection},
			{ID: "E3", X: 63, Y: 32, Type: NodeIntersection},
			{ID: "E4", X: 68, Y: 32, Type: NodeIntersection},
			// Gates
			{ID: "G1", X: 52, Y: 34, Type: NodeGate},
			{ID: "G2", X: 57, Y: 34, Type: NodeGate},
			{ID: "G3", X: 63, Y: 34, Type: NodeGate},
			{ID: "G4", X: 68, Y: 34, Type: NodeGate},
		},
		TaxiEdges: []TaxiEdge{
			// Taxiway A segments
			{From: "A1", To: "A2", Taxiway: "A"},
			{From: "A2", To: "A3", Taxiway: "A"},
			{From: "A3", To: "A4", Taxiway: "A"},
			// Taxiway B (north connectors: A -> hold-short -> runway north)
			{From: "A1", To: "HS10L", Taxiway: "B"},
			{From: "HS10L", To: "RE10L", Taxiway: "B"},
			{From: "A4", To: "HS28R", Taxiway: "B"},
			{From: "HS28R", To: "RE28R", Taxiway: "B"},
			// Taxiway C (south connectors: A -> hold-short -> runway south)
			{From: "A1", To: "HS10R", Taxiway: "C"},
			{From: "HS10R", To: "RE10R", Taxiway: "C"},
			{From: "A4", To: "HS28L", Taxiway: "C"},
			{From: "HS28L", To: "RE28L", Taxiway: "C"},
			// Taxiway D (south runway to taxiway E)
			{From: "RE10R", To: "E1", Taxiway: "D"},
			{From: "RE28L", To: "E4", Taxiway: "D"},
			// Taxiway E segments
			{From: "E1", To: "E2", Taxiway: "E"},
			{From: "E2", To: "E3", Taxiway: "E"},
			{From: "E3", To: "E4", Taxiway: "E"},
			// Taxiway F (connectors to gates)
			{From: "E1", To: "G1", Taxiway: "F"},
			{From: "E2", To: "G2", Taxiway: "F"},
			{From: "E3", To: "G3", Taxiway: "F"},
			{From: "E4", To: "G4", Taxiway: "F"},
		},
		Gates: []Gate{
			{ID: "G1", NodeID: "G1"},
			{ID: "G2", NodeID: "G2"},
			{ID: "G3", NodeID: "G3"},
			{ID: "G4", NodeID: "G4"},
		},
	}
}

// Tutorial is a simple map for learning the game.
//
// Ground layout (runway at Y=33, taxiway A at Y=31, gates at Y=29):
//
//	[G1] [G2] [G3]         Y=29
//	  |    |    |
//	--A1---A2---A3---A4--   Y=31  Taxiway A
//	  |              |
//	 HS9           HS27     Y=32  Hold-short points
//	  |              |
//	==RE9=RWY 9/27=RE27==   Y=33  Runway
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
		TaxiNodes: []TaxiNode{
			// Taxiway A (east-west parallel to runway)
			{ID: "A1", X: 40, Y: 31, Type: NodeIntersection},
			{ID: "A2", X: 45, Y: 31, Type: NodeIntersection},
			{ID: "A3", X: 50, Y: 31, Type: NodeIntersection},
			{ID: "A4", X: 55, Y: 31, Type: NodeIntersection},
			// Hold-short points
			{ID: "HS9", X: 40, Y: 32, Type: NodeHoldShort, Runway: "9"},
			{ID: "HS27", X: 55, Y: 32, Type: NodeHoldShort, Runway: "27"},
			// Runway entry/exit
			{ID: "RE9", X: 40, Y: 33, Type: NodeRunwayEntry, Runway: "9"},
			{ID: "RE27", X: 55, Y: 33, Type: NodeRunwayEntry, Runway: "27"},
			// Gates
			{ID: "G1", X: 40, Y: 29, Type: NodeGate},
			{ID: "G2", X: 45, Y: 29, Type: NodeGate},
			{ID: "G3", X: 50, Y: 29, Type: NodeGate},
		},
		TaxiEdges: []TaxiEdge{
			// Taxiway A segments
			{From: "A1", To: "A2", Taxiway: "A"},
			{From: "A2", To: "A3", Taxiway: "A"},
			{From: "A3", To: "A4", Taxiway: "A"},
			// Taxiway B (north-south connectors to gates)
			{From: "G1", To: "A1", Taxiway: "B"},
			{From: "G2", To: "A2", Taxiway: "B"},
			{From: "G3", To: "A3", Taxiway: "B"},
			// Taxiway C (connectors to hold-short / runway)
			{From: "A1", To: "HS9", Taxiway: "C"},
			{From: "HS9", To: "RE9", Taxiway: "C"},
			{From: "A4", To: "HS27", Taxiway: "D"},
			{From: "HS27", To: "RE27", Taxiway: "D"},
		},
		Gates: []Gate{
			{ID: "G1", NodeID: "G1"},
			{ID: "G2", NodeID: "G2"},
			{ID: "G3", NodeID: "G3"},
		},
	}
}
