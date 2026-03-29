package game

import (
	"testing"
)

func newDevModel() Model {
	m := NewModel(true)
	m.width, m.height = 100, 50
	started, _ := m.startGame()
	return started
}

func TestDevCommandRequiresFlag(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50
	started, _ := m.startGame()

	result := started.processDevCommand("help")
	// Should show error about --dev flag
	messages := result.radioLog.All()
	last := messages[len(messages)-1]
	if last.Text != "dev commands require --dev flag" {
		t.Errorf("expected dev flag error, got %q", last.Text)
	}
}

func TestDevHelp(t *testing.T) {
	m := newDevModel()
	result := m.processDevCommand("help")
	messages := result.radioLog.All()
	// Should have multiple help lines
	if len(messages) < 5 {
		t.Errorf("expected help messages, got %d", len(messages))
	}
}

func TestDevSpawnArrival(t *testing.T) {
	m := newDevModel()
	before := len(m.aircraft)
	result := m.processDevCommand("spawn")
	if len(result.aircraft) != before+1 {
		t.Errorf("expected %d aircraft after spawn, got %d", before+1, len(result.aircraft))
	}
}

func TestDevSpawnDeparture(t *testing.T) {
	m := newDevModel()
	before := len(m.aircraft)
	result := m.processDevCommand("spawn dep")
	// May or may not spawn depending on available gates, but shouldn't panic
	if len(result.aircraft) < before {
		t.Error("aircraft count should not decrease after spawn dep")
	}
}

func TestDevClear(t *testing.T) {
	m := newDevModel()
	m = m.processDevCommand("spawn")
	m = m.processDevCommand("spawn")
	if len(m.aircraft) < 2 {
		t.Skip("couldn't spawn enough aircraft")
	}
	result := m.processDevCommand("clear")
	if len(result.aircraft) != 0 {
		t.Errorf("expected 0 aircraft after clear, got %d", len(result.aircraft))
	}
}

func TestDevGodToggle(t *testing.T) {
	m := newDevModel()
	if m.godMode {
		t.Error("god mode should start off")
	}
	m = m.processDevCommand("god")
	if !m.godMode {
		t.Error("god mode should be on after toggle")
	}
	m = m.processDevCommand("god")
	if m.godMode {
		t.Error("god mode should be off after second toggle")
	}
}

func TestDevPauseToggle(t *testing.T) {
	m := newDevModel()
	if m.spawnerPaused {
		t.Error("spawner should start unpaused")
	}
	m = m.processDevCommand("pause")
	if !m.spawnerPaused {
		t.Error("spawner should be paused after toggle")
	}
	m = m.processDevCommand("pause")
	if m.spawnerPaused {
		t.Error("spawner should be unpaused after second toggle")
	}
}

func TestDevSpeed(t *testing.T) {
	m := newDevModel()
	if m.speedMultiplier != 1 {
		t.Errorf("expected speed 1, got %d", m.speedMultiplier)
	}
	m = m.processDevCommand("speed 3")
	if m.speedMultiplier != 3 {
		t.Errorf("expected speed 3, got %d", m.speedMultiplier)
	}
}

func TestDevSpeedInvalid(t *testing.T) {
	m := newDevModel()
	m = m.processDevCommand("speed 0")
	if m.speedMultiplier != 1 {
		t.Error("speed 0 should be rejected")
	}
	m = m.processDevCommand("speed 13")
	if m.speedMultiplier != 1 {
		t.Error("speed 13 should be rejected")
	}
}

func TestDevUnknownCommand(t *testing.T) {
	m := newDevModel()
	result := m.processDevCommand("foo")
	messages := result.radioLog.All()
	last := messages[len(messages)-1]
	if last.Text != "unknown dev command: /foo (try /help)" {
		t.Errorf("expected unknown command error, got %q", last.Text)
	}
}

func TestDevStatusString(t *testing.T) {
	m := NewModel(false)
	if m.DevStatus() != "" {
		t.Error("non-dev model should have empty DevStatus")
	}

	m = NewModel(true)
	if m.DevStatus() != "[DEV]" {
		t.Errorf("expected [DEV], got %q", m.DevStatus())
	}

	m.godMode = true
	m.spawnerPaused = true
	status := m.DevStatus()
	if status != "[DEV GOD NOSPAWN]" {
		t.Errorf("expected [DEV GOD NOSPAWN], got %q", status)
	}
}
