package game

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/gamemap"
	"github.com/Jasrags/atc/internal/radio"
)

// trySpawnDeparture attempts to spawn a departure aircraft at an unoccupied gate.
func (m Model) trySpawnDeparture(elapsed time.Duration) Model {
	// Build list of available gates (not occupied by another aircraft)
	occupiedGates := make(map[string]bool)
	for _, ac := range m.aircraft {
		if ac.AssignedGate != "" && ac.State.IsGround() {
			occupiedGates[ac.AssignedGate] = true
		}
	}

	var available []struct{ ID string; X, Y int }
	for _, g := range m.gameMap.Gates {
		if !occupiedGates[g.ID] {
			node := m.gameMap.NodeByID(g.NodeID)
			if node != nil {
				available = append(available, struct{ ID string; X, Y int }{g.ID, node.X, node.Y})
			}
		}
	}

	if len(available) == 0 {
		return m
	}

	ac, ok := m.spawner.SpawnDeparture(available)
	if !ok {
		return m
	}
	if _, exists := m.aircraft[ac.Callsign]; exists {
		return m
	}

	// In TRACON mode, auto-pushback departures immediately
	if m.gameConfig.Role == config.RoleTRACON {
		ac.State = aircraft.Pushback
	}

	m.aircraft[ac.Callsign] = ac
	m = m.addRadio(radio.PilotMessage(elapsed, ac.Callsign,
		fmt.Sprintf("%s at gate %s, requesting pushback", ac.Callsign, ac.AssignedGate)))
	return m
}

// findNearestHoldShort returns the nearest hold-short node to the aircraft, or nil if none.
func (m Model) findNearestHoldShort(ac aircraft.Aircraft) *gamemap.TaxiNode {
	gx, gy := ac.GridX(), ac.GridY()
	bestDist := math.MaxFloat64
	var bestNode *gamemap.TaxiNode
	for i := range m.gameMap.TaxiNodes {
		n := &m.gameMap.TaxiNodes[i]
		if n.Type != gamemap.NodeHoldShort {
			continue
		}
		dx := float64(n.X - gx)
		dy := float64(n.Y - gy)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			bestNode = n
		}
	}
	return bestNode
}

// findAvailableGate returns the ID of an unoccupied gate, or "" if none available.
func (m Model) findAvailableGate() string {
	occupied := make(map[string]bool)
	for _, ac := range m.aircraft {
		if ac.AssignedGate != "" {
			occupied[ac.AssignedGate] = true
		}
	}
	for _, g := range m.gameMap.Gates {
		if !occupied[g.ID] {
			return g.ID
		}
	}
	return ""
}

// runwayHeading returns the heading of the named runway, or 0 if not found.
func (m Model) runwayHeading(name string) int {
	for _, rw := range m.gameMap.Runways {
		// Match by runway number (e.g., "27" matches heading 270)
		num := gamemap.RunwayNumber(rw.Heading)
		if fmt.Sprintf("%d", num) == name {
			return rw.Heading
		}
		// Also check opposite end
		oppNum := gamemap.RunwayNumber(rw.OppositeHeading())
		if fmt.Sprintf("%d", oppNum) == name {
			return rw.OppositeHeading()
		}
	}
	return 0
}

// findNearestNode returns the ID of the closest taxi node to the aircraft's position.
func (m Model) findNearestNode(ac aircraft.Aircraft) string {
	gx, gy := ac.GridX(), ac.GridY()
	bestDist := math.MaxFloat64
	bestID := ""
	for _, node := range m.gameMap.TaxiNodes {
		dx := float64(node.X - gx)
		dy := float64(node.Y - gy)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			bestID = node.ID
		}
	}
	return bestID
}

// nodeIDsToPositions converts a list of node IDs into grid positions.
func (m Model) nodeIDsToPositions(nodeIDs []string) [][2]int {
	positions := make([][2]int, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		node := m.gameMap.NodeByID(id)
		if node != nil {
			positions = append(positions, [2]int{node.X, node.Y})
		}
	}
	return positions
}

func (m Model) sortedAircraft() []aircraft.Aircraft {
	planes := make([]aircraft.Aircraft, 0, len(m.aircraft))
	for _, ac := range m.aircraft {
		planes = append(planes, ac)
	}
	sort.Slice(planes, func(i, j int) bool {
		return planes[i].Callsign < planes[j].Callsign
	})
	return planes
}
