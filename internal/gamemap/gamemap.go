package gamemap

import "fmt"

// FixType distinguishes different kinds of navigation points on the radar.
type FixType int

const (
	FixWaypoint    FixType = iota // △ named waypoint/fix
	FixAirport                    // ◎ airport
	FixVOR                        // ◉ VOR navigation aid
	FixIntersection               // ✦ intersection fix
)

func (f FixType) Symbol() string {
	switch f {
	case FixWaypoint:
		return "△"
	case FixAirport:
		return "◎"
	case FixVOR:
		return "◉"
	case FixIntersection:
		return "✦"
	default:
		return "·"
	}
}

// Fix represents a named navigation point on the radar.
type Fix struct {
	Name string
	X    int
	Y    int
	Type FixType
}

// Runway represents an airport runway.
type Runway struct {
	Name      string // e.g. "9/27"
	X         int    // Center X position
	Y         int    // Center Y position
	Heading   int    // Primary approach heading (0-359)
	Length    int    // Visual length in grid cells
}

// OppositeHeading returns the reciprocal runway heading.
func (r Runway) OppositeHeading() int {
	return (r.Heading + 180) % 360
}

// RunwayNumber returns the runway number for a given heading (heading / 10).
func RunwayNumber(heading int) int {
	n := (heading + 5) / 10
	if n == 0 {
		n = 36
	}
	return n
}

// TaxiNodeType distinguishes different kinds of taxiway nodes.
type TaxiNodeType int

const (
	NodeIntersection TaxiNodeType = iota // Junction of two or more taxiways
	NodeHoldShort                        // Hold-short point before a runway
	NodeGate                             // Aircraft parking gate
	NodeRunwayEntry                      // Entry/exit point onto a runway
)

// TaxiNode represents a point on the airport surface graph.
type TaxiNode struct {
	ID      string       // Unique identifier (e.g., "A1", "HS27", "G1")
	X       int          // Grid X position
	Y       int          // Grid Y position
	Type    TaxiNodeType // Node type
	Runway  string       // Associated runway (for HoldShort/RunwayEntry nodes)
}

// TaxiEdge represents a taxiway segment connecting two nodes.
type TaxiEdge struct {
	From    string // Source node ID
	To      string // Destination node ID
	Taxiway string // Taxiway name (e.g., "A", "B", "C1")
}

// Gate represents an aircraft parking position.
type Gate struct {
	ID     string // Gate identifier (e.g., "G1", "G2")
	NodeID string // TaxiNode ID where this gate is located
}

// Map defines a complete game map with all its features.
type Map struct {
	ID          string
	Name        string
	Description string
	Width       int
	Height      int
	Runways     []Runway
	Fixes       []Fix
	TaxiNodes   []TaxiNode
	TaxiEdges   []TaxiEdge
	Gates       []Gate
}

// NodeByID returns the TaxiNode with the given ID, or nil if not found.
func (m Map) NodeByID(id string) *TaxiNode {
	for i := range m.TaxiNodes {
		if m.TaxiNodes[i].ID == id {
			return &m.TaxiNodes[i]
		}
	}
	return nil
}

// GateByID returns the Gate with the given ID, or nil if not found.
func (m Map) GateByID(id string) *Gate {
	for i := range m.Gates {
		if m.Gates[i].ID == id {
			return &m.Gates[i]
		}
	}
	return nil
}

// Neighbors returns all node IDs reachable from the given node, with their taxiway names.
func (m Map) Neighbors(nodeID string) []TaxiEdge {
	var result []TaxiEdge
	for _, e := range m.TaxiEdges {
		if e.From == nodeID || e.To == nodeID {
			result = append(result, e)
		}
	}
	return result
}

// ResolveTaxiRoute takes a list of taxiway names and a start node ID,
// and returns the ordered list of node IDs to traverse.
// Returns an error if the route is not connected or a taxiway name is not found.
//
// Assumption: each named taxiway forms a simple linear path (no forks).
// The walk greedily follows each taxiway until no unvisited neighbor remains.
// If a taxiway branches, the result is non-deterministic.
func (m Map) ResolveTaxiRoute(startNodeID string, taxiways []string) ([]string, error) {
	if len(taxiways) == 0 {
		return nil, fmt.Errorf("empty taxi route")
	}

	// Build adjacency map: nodeID -> []TaxiEdge
	adj := make(map[string][]TaxiEdge)
	for _, e := range m.TaxiEdges {
		adj[e.From] = append(adj[e.From], e)
		adj[e.To] = append(adj[e.To], TaxiEdge{From: e.To, To: e.From, Taxiway: e.Taxiway})
	}

	path := []string{startNodeID}
	current := startNodeID

	for _, twName := range taxiways {
		// Find all edges from current node on this taxiway
		found := false
		for _, edge := range adj[current] {
			if edge.Taxiway == twName {
				// Walk along this taxiway as far as it goes
				visited := map[string]bool{current: true}
				next := edge.To
				for {
					path = append(path, next)
					visited[next] = true
					// Continue along same taxiway if there's exactly one unvisited neighbor on it
					var continued bool
					for _, e2 := range adj[next] {
						if e2.Taxiway == twName && !visited[e2.To] {
							next = e2.To
							continued = true
							break
						}
					}
					if !continued {
						break
					}
				}
				current = next
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("taxiway %s not reachable from current position", twName)
		}
	}

	return path, nil
}

// PrimaryRunway returns the first runway, which is the main landing target.
func (m Map) PrimaryRunway() Runway {
	if len(m.Runways) == 0 {
		return Runway{X: m.Width / 2, Y: m.Height - 5, Heading: 270, Length: 5}
	}
	return m.Runways[0]
}
