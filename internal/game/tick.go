package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/collision"
	"github.com/Jasrags/atc/internal/config"
	"github.com/Jasrags/atc/internal/radar"
	"github.com/Jasrags/atc/internal/radio"
	tea "github.com/charmbracelet/bubbletea"
)

// tickEffect accumulates side effects from pure per-aircraft transformations.
// Each transformation returns a new Aircraft plus a tickEffect; the caller
// merges all effects into the Model after the pipeline completes.
type tickEffect struct {
	scoreDelta int
	messages   []radio.Message
	remove     bool // aircraft should be removed from the active map
}

// merge combines another tickEffect into this one.
func (e tickEffect) merge(other tickEffect) tickEffect {
	e.scoreDelta += other.scoreDelta
	e.messages = append(e.messages, other.messages...)
	e.remove = e.remove || other.remove
	return e
}

// tickLanding checks whether a Landing aircraft has reached a runway and
// transitions it to Landed. Returns the (possibly updated) aircraft and effect.
func (m Model) tickLanding(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if ac.State != aircraft.Landing {
		return ac, tickEffect{}
	}
	for i, rw := range m.runways {
		if ac.AssignedLandingRunway != "" && i < len(m.gameMap.Runways) {
			if !strings.Contains(m.gameMap.Runways[i].Name, ac.AssignedLandingRunway) {
				continue
			}
		}
		if rw.CanLand(ac.GridX(), ac.GridY(), ac.Heading, ac.Altitude) {
			next := ac
			next.State = aircraft.Landed
			return next, tickEffect{
				scoreDelta: 1,
				messages:   []radio.Message{radio.PilotMessage(elapsed, ac.Callsign, radio.FormatLanded(ac.Callsign))},
			}
		}
	}
	return ac, tickEffect{}
}

// tickAutoGroundArrival handles TRACON auto-ground: landed aircraft auto-taxi
// to the nearest available gate.
func (m Model) tickAutoGroundArrival(ac aircraft.Aircraft) (aircraft.Aircraft, tickEffect) {
	if ac.State != aircraft.Landed || m.gameConfig.Role != config.RoleTRACON {
		return ac, tickEffect{}
	}
	gate := m.findAvailableGate()
	if gate == "" {
		return ac, tickEffect{}
	}
	gateObj := m.gameMap.GateByID(gate)
	if gateObj == nil {
		return ac, tickEffect{}
	}
	node := m.gameMap.NodeByID(gateObj.NodeID)
	if node == nil {
		return ac, tickEffect{}
	}
	next := ac
	next.AssignedGate = gate
	next.State = aircraft.Taxiing
	next.TaxiPath = [][2]int{{ac.GridX(), ac.GridY()}, {node.X, node.Y}}
	next.TaxiPathIndex = 0
	return next, tickEffect{}
}

// tickTaxiComplete checks whether a taxiing aircraft with a gate assignment
// has finished its path. Returns an AtGate aircraft with a radio message.
func (m Model) tickTaxiComplete(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if ac.State != aircraft.Taxiing || len(ac.TaxiPath) != 0 || ac.AssignedGate == "" {
		return ac, tickEffect{}
	}
	next := ac
	next.State = aircraft.AtGate
	return next, tickEffect{
		messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
			fmt.Sprintf("%s at gate %s", ac.Callsign, ac.AssignedGate))},
	}
}

// tickAutoGroundDeparture handles TRACON auto-departure: automates pushback →
// taxi → hold short → on-runway transitions.
func (m Model) tickAutoGroundDeparture(ac aircraft.Aircraft) (aircraft.Aircraft, tickEffect) {
	if m.gameConfig.Role != config.RoleTRACON {
		return ac, tickEffect{}
	}
	switch ac.State {
	case aircraft.Pushback:
		if len(ac.TaxiPath) == 0 {
			hsNode := m.findNearestHoldShort(ac)
			if hsNode != nil {
				next := ac
				next.State = aircraft.Taxiing
				next.TaxiPath = [][2]int{{ac.GridX(), ac.GridY()}, {hsNode.X, hsNode.Y}}
				next.TaxiPathIndex = 0
				return next, tickEffect{}
			}
		}
	case aircraft.Taxiing:
		if len(ac.TaxiPath) == 0 && ac.AssignedGate == "" {
			next := ac
			next.State = aircraft.HoldShort
			return next, tickEffect{}
		}
	case aircraft.HoldShort:
		next := ac.ResetTickCount()
		next.State = aircraft.OnRunway
		next.TaxiPath = nil
		next.TaxiRoute = nil
		return next, tickEffect{}
	}
	return ac, tickEffect{}
}

// tickTowerAutoHandoff removes departing aircraft at altitude >= 3 in Tower mode,
// scoring +1 for a successful handoff to TRACON.
func (m Model) tickTowerAutoHandoff(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if m.gameConfig.Role != config.RoleTower {
		return ac, tickEffect{}
	}
	if ac.State != aircraft.Departing || ac.Altitude < 3 {
		return ac, tickEffect{}
	}
	return ac, tickEffect{
		scoreDelta: 1,
		messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
			fmt.Sprintf("%s contact departure, good day", ac.Callsign))},
		remove: true,
	}
}

// tickPatience advances the patience timer for airborne aircraft and generates
// nag messages or removal when patience expires.
func (m Model) tickPatience(ac aircraft.Aircraft, elapsed time.Duration) (aircraft.Aircraft, tickEffect) {
	if !ac.State.IsAirborne() || ac.PatienceMax == 0 {
		return ac, tickEffect{}
	}
	next := ac
	next.PatienceTicks++

	nagThreshold := next.PatienceMax + next.PatienceNagCount*aircraft.PatienceNagEvery
	if next.PatienceTicks < nagThreshold {
		return next, tickEffect{}
	}

	next.PatienceNagCount++
	switch {
	case next.PatienceNagCount >= aircraft.PatienceLeaveAt:
		return next, tickEffect{
			scoreDelta: -1,
			messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s leaving your airspace, good day", ac.Callsign))},
			remove: true,
		}
	case next.PatienceNagCount >= aircraft.PatiencePenaltyAt:
		return next, tickEffect{
			messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s requesting ANY instructions!", ac.Callsign))},
		}
	default:
		return next, tickEffect{
			messages: []radio.Message{radio.PilotMessage(elapsed, ac.Callsign,
				fmt.Sprintf("%s still waiting for vectors", ac.Callsign))},
		}
	}
}

func (m Model) handleTick(msg tickMsg) (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying {
		return m, nil
	}

	// Time frozen: skip physics and spawning, but keep tick loop alive for rendering.
	if m.timeFrozen {
		return m, tickCmd()
	}

	elapsed := m.stopwatch.Elapsed()

	// Speed multiplier: run physics N times per render frame.
	ticks := m.speedMultiplier
	if ticks < 1 {
		ticks = 1
	}

	for tick := 0; tick < ticks; tick++ {
		m = m.tickPhysics(elapsed)
		if m.screen == screenGameOver {
			return m, nil
		}
	}

	// Phase 5: spawning (once per frame regardless of speed)
	if !m.spawnerPaused && m.spawner.ShouldSpawn(elapsed, len(m.aircraft)) {
		countBefore := len(m.aircraft)
		if m.spawnDeparture && len(m.gameMap.Gates) > 0 {
			m = m.trySpawnDeparture(elapsed)
		} else {
			m = m.spawnArrival(elapsed)
		}
		if len(m.aircraft) > countBefore {
			m.spawnDeparture = !m.spawnDeparture
		}
	}

	// Update flight strip viewport content
	m.stripViewport.SetContent(radar.RenderFlightStrips(m.sortedAircraft(), m.gameConfig.Role))

	return m, tickCmd()
}

// tickPhysics runs one cycle of aircraft movement, collision, separation, and
// state transitions. Extracted so the speed multiplier can call it N times.
func (m Model) tickPhysics(elapsed time.Duration) Model {
	// Phase 1: advance physics (each Tick/GroundTick/TakeoffTick is already pure)
	newAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		var next aircraft.Aircraft
		switch {
		case ac.State == aircraft.OnRunway:
			next = ac.TakeoffTick()
			if next.State == aircraft.Departing && ac.State == aircraft.OnRunway {
				hdg := m.runwayHeading(ac.AssignedRunway)
				if hdg == 0 {
					hdg = m.gameMap.PrimaryRunway().Heading
				}
				next = next.WithHeading(hdg)
			}
		case ac.State.IsGround() && ac.State != aircraft.Landed:
			next = ac.GroundTick()
		default:
			next = ac.Tick()
		}
		// Departing aircraft that leave airspace = successful departure.
		// In Tower mode, departures are handed off at altitude 3 (tickTowerAutoHandoff),
		// so they should not reach the edge — but handle it gracefully if they do.
		if next.State == aircraft.Departing && next.IsOffScreen(m.gameMap.Width, m.gameMap.Height) {
			if m.gameConfig.Role != config.RoleTower {
				m.score++
			}
			m = m.addRadio(radio.PilotMessage(elapsed, next.Callsign,
				fmt.Sprintf("%s leaving airspace, good day", next.Callsign)))
			continue
		}
		// Other airborne aircraft that leave = just removed
		if next.State.IsAirborne() && next.IsOffScreen(m.gameMap.Width, m.gameMap.Height) {
			continue
		}
		newAircraft[k] = next
	}
	m.aircraft = newAircraft

	// Phase 2: collision detection (immutable — build new map with crashed state)
	collisions := collision.Check(m.aircraft)
	if len(collisions) > 0 {
		for _, c := range collisions {
			m = m.addRadio(radio.SystemMessage(elapsed,
				radio.FormatCollision(c.Callsign1, c.Callsign2), radio.Emergency))
		}
		if !m.godMode {
			crashed := make(map[string]aircraft.Aircraft, len(m.aircraft))
			for k, ac := range m.aircraft {
				crashed[k] = ac
			}
			for _, c := range collisions {
				crashed[c.Callsign1] = crashed[c.Callsign1].WithState(aircraft.Crashed)
				crashed[c.Callsign2] = crashed[c.Callsign2].WithState(aircraft.Crashed)
			}
			m.aircraft = crashed
			m.screen = screenGameOver
			return m
		}
	}

	// Phase 3: separation violations
	violations := collision.CheckSeparation(m.aircraft)
	currentPairs := make(map[string]bool)
	for _, v := range violations {
		pairKey := v.Callsign1 + ":" + v.Callsign2
		currentPairs[pairKey] = true
		m.score -= collision.ViolationPenalty
		if !m.activeViolations[pairKey] {
			m.nearMisses++
			m = m.addRadio(radio.SystemMessage(elapsed,
				fmt.Sprintf("TRAFFIC ALERT: %s and %s — loss of separation (%.1f cells)",
					v.Callsign1, v.Callsign2, v.Distance), radio.Urgent))
		}
	}
	if m.score < 0 {
		m.score = 0
	}
	m.activeViolations = currentPairs

	// Phase 4: per-aircraft state pipeline (pure transformations)
	activeAircraft := make(map[string]aircraft.Aircraft, len(m.aircraft))
	for k, ac := range m.aircraft {
		var fx tickEffect
		var combined tickEffect

		ac, fx = m.tickLanding(ac, elapsed)
		combined = combined.merge(fx)

		ac, fx = m.tickAutoGroundArrival(ac)
		combined = combined.merge(fx)

		ac, fx = m.tickTaxiComplete(ac, elapsed)
		combined = combined.merge(fx)

		ac, fx = m.tickAutoGroundDeparture(ac)
		combined = combined.merge(fx)

		ac, fx = m.tickTowerAutoHandoff(ac, elapsed)
		combined = combined.merge(fx)

		ac, fx = m.tickPatience(ac, elapsed)
		combined = combined.merge(fx)

		// Apply accumulated effects
		m.score += combined.scoreDelta
		for _, msg := range combined.messages {
			m = m.addRadio(msg)
		}

		// Remove aircraft that left or completed arrival
		if combined.remove || ac.State == aircraft.AtGate {
			continue
		}
		activeAircraft[k] = ac
	}
	if m.score < 0 {
		m.score = 0
	}
	m.aircraft = activeAircraft

	return m
}
