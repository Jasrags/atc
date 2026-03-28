package gamemap

import "testing"

func TestRunwayNumber(t *testing.T) {
	tests := []struct {
		heading int
		want    int
	}{
		{270, 27},
		{90, 9},
		{280, 28},
		{100, 10},
		{360, 36},
		{0, 36},
		{5, 1},
		{355, 36},
	}
	for _, tt := range tests {
		got := RunwayNumber(tt.heading)
		if got != tt.want {
			t.Errorf("RunwayNumber(%d) = %d, want %d", tt.heading, got, tt.want)
		}
	}
}

func TestOppositeHeading(t *testing.T) {
	tests := []struct {
		heading int
		want    int
	}{
		{270, 90},
		{90, 270},
		{0, 180},
		{180, 0},
	}
	for _, tt := range tests {
		rw := Runway{Heading: tt.heading}
		got := rw.OppositeHeading()
		if got != tt.want {
			t.Errorf("OppositeHeading(%d) = %d, want %d", tt.heading, got, tt.want)
		}
	}
}

func TestPrimaryRunwayFallback(t *testing.T) {
	m := Map{Width: 80, Height: 35}
	rw := m.PrimaryRunway()
	if rw.Heading != 270 {
		t.Errorf("fallback heading = %d, want 270", rw.Heading)
	}
}

func TestAllMaps(t *testing.T) {
	maps := All()
	if len(maps) < 3 {
		t.Fatalf("expected at least 3 maps, got %d", len(maps))
	}

	ids := make(map[string]bool)
	for _, m := range maps {
		if m.ID == "" {
			t.Error("map has empty ID")
		}
		if ids[m.ID] {
			t.Errorf("duplicate map ID: %s", m.ID)
		}
		ids[m.ID] = true

		if m.Width <= 0 || m.Height <= 0 {
			t.Errorf("map %s has invalid dimensions: %dx%d", m.ID, m.Width, m.Height)
		}
		if len(m.Runways) == 0 {
			t.Errorf("map %s has no runways", m.ID)
		}

		// Verify all fixes are within bounds
		for _, f := range m.Fixes {
			if f.X < 0 || f.X >= m.Width || f.Y < 0 || f.Y >= m.Height {
				t.Errorf("map %s: fix %s at (%d,%d) out of bounds %dx%d",
					m.ID, f.Name, f.X, f.Y, m.Width, m.Height)
			}
		}

		// Verify all runways are within bounds
		for _, r := range m.Runways {
			if r.X < 0 || r.X >= m.Width || r.Y < 0 || r.Y >= m.Height {
				t.Errorf("map %s: runway %s at (%d,%d) out of bounds",
					m.ID, r.Name, r.X, r.Y)
			}
		}
	}
}

func TestByID(t *testing.T) {
	m := ByID("san")
	if m.ID != "san" {
		t.Errorf("expected san, got %s", m.ID)
	}

	m = ByID("nonexistent")
	if m.ID != "tutorial" {
		t.Errorf("expected tutorial fallback, got %s", m.ID)
	}
}

func TestNodeByID(t *testing.T) {
	m := Tutorial()
	node := m.NodeByID("A1")
	if node == nil {
		t.Fatal("expected to find node A1")
	}
	if node.X != 40 || node.Y != 31 {
		t.Errorf("A1 position = (%d,%d), want (40,31)", node.X, node.Y)
	}

	if m.NodeByID("NOPE") != nil {
		t.Error("expected nil for nonexistent node")
	}
}

func TestGateByID(t *testing.T) {
	m := Tutorial()
	gate := m.GateByID("G2")
	if gate == nil {
		t.Fatal("expected to find gate G2")
	}
	if gate.NodeID != "G2" {
		t.Errorf("G2 nodeID = %s, want G2", gate.NodeID)
	}

	if m.GateByID("NOPE") != nil {
		t.Error("expected nil for nonexistent gate")
	}
}

func TestNeighbors(t *testing.T) {
	m := Tutorial()
	edges := m.Neighbors("A2")
	if len(edges) == 0 {
		t.Fatal("expected neighbors for A2")
	}
	// A2 connects to A1 (taxiway A), A3 (taxiway A), G2 (taxiway B)
	if len(edges) != 3 {
		t.Errorf("A2 should have 3 neighbors, got %d", len(edges))
	}
}

func TestResolveTaxiRoute(t *testing.T) {
	m := Tutorial()

	// Gate G1 -> taxiway B to A1 -> taxiway A to A4 -> taxiway D to HS27
	path, err := m.ResolveTaxiRoute("G1", []string{"B", "A", "D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) == 0 {
		t.Fatal("expected non-empty path")
	}
	if path[0] != "G1" {
		t.Errorf("path should start at G1, got %s", path[0])
	}
	// Should end at HS27 (walking along D from A4 -> HS27 -> RE27)
	last := path[len(path)-1]
	if last != "RE27" {
		t.Errorf("path should end at RE27, got %s", last)
	}
}

func TestResolveTaxiRouteInvalid(t *testing.T) {
	m := Tutorial()

	_, err := m.ResolveTaxiRoute("G1", []string{})
	if err == nil {
		t.Error("expected error for empty route")
	}

	_, err = m.ResolveTaxiRoute("G1", []string{"Z"})
	if err == nil {
		t.Error("expected error for nonexistent taxiway")
	}
}

func TestAllMapsHaveTaxiways(t *testing.T) {
	for _, m := range All() {
		if len(m.TaxiNodes) == 0 {
			t.Errorf("map %s has no taxi nodes", m.ID)
		}
		if len(m.TaxiEdges) == 0 {
			t.Errorf("map %s has no taxi edges", m.ID)
		}
		if len(m.Gates) == 0 {
			t.Errorf("map %s has no gates", m.ID)
		}

		// Verify all taxi nodes are within bounds
		for _, n := range m.TaxiNodes {
			if n.X < 0 || n.X >= m.Width || n.Y < 0 || n.Y >= m.Height {
				t.Errorf("map %s: taxi node %s at (%d,%d) out of bounds",
					m.ID, n.ID, n.X, n.Y)
			}
		}

		// Verify all edges reference existing nodes
		nodeIDs := make(map[string]bool)
		for _, n := range m.TaxiNodes {
			nodeIDs[n.ID] = true
		}
		for _, e := range m.TaxiEdges {
			if !nodeIDs[e.From] {
				t.Errorf("map %s: edge references unknown node %s", m.ID, e.From)
			}
			if !nodeIDs[e.To] {
				t.Errorf("map %s: edge references unknown node %s", m.ID, e.To)
			}
		}

		// Verify all gate NodeIDs exist
		for _, g := range m.Gates {
			if !nodeIDs[g.NodeID] {
				t.Errorf("map %s: gate %s references unknown node %s", m.ID, g.ID, g.NodeID)
			}
		}
	}
}

func TestFixSymbols(t *testing.T) {
	tests := []struct {
		ft   FixType
		want string
	}{
		{FixWaypoint, "△"},
		{FixAirport, "◎"},
		{FixVOR, "◉"},
		{FixIntersection, "✦"},
	}
	for _, tt := range tests {
		got := tt.ft.Symbol()
		if got != tt.want {
			t.Errorf("FixType(%d).Symbol() = %q, want %q", tt.ft, got, tt.want)
		}
	}
}
