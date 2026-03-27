package game

import (
	"testing"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	tea "github.com/charmbracelet/bubbletea"
)

// newPlayingModel returns a model in the playing state for tests.
func newPlayingModel() Model {
	m := NewModel()
	m.width, m.height = 100, 50
	started, _ := m.startGame()
	return started
}

func TestNewModelStartsAtMenu(t *testing.T) {
	m := NewModel()
	if m.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", m.screen)
	}
}

func TestMenuNavigation(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50

	// Down
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	model := result.(Model)
	if model.menuSelected != 1 {
		t.Errorf("expected selected=1 after down, got %d", model.menuSelected)
	}

	// Up
	result, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	model = result.(Model)
	if model.menuSelected != 0 {
		t.Errorf("expected selected=0 after up, got %d", model.menuSelected)
	}

	// Don't go above 0
	result, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	model = result.(Model)
	if model.menuSelected != 0 {
		t.Errorf("expected selected=0 (clamped), got %d", model.menuSelected)
	}
}

func TestMenuNewGameGoesToMapSelect(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50

	// Press enter on "New Game" -> map select
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	if model.menuScreen != menuMapSelect {
		t.Errorf("expected menuMapSelect, got %d", model.menuScreen)
	}
}

func TestMapSelectStartsGame(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50
	m.menuScreen = menuMapSelect

	// Press enter on first map
	result, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
	if cmd == nil {
		t.Error("expected init commands after starting game")
	}
}

func TestMapSelectEscGoesBack(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50
	m.menuScreen = menuMapSelect

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.menuScreen != menuMain {
		t.Errorf("expected menuMain, got %d", model.menuScreen)
	}
}

func TestMenuHelp(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model := result.(Model)

	if model.screen != screenHelp {
		t.Errorf("expected screenHelp, got %d", model.screen)
	}
}

func TestMenuQuit(t *testing.T) {
	m := NewModel()

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestHelpReturnToMenu(t *testing.T) {
	m := NewModel()
	m.screen = screenHelp

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestHelpReturnToPlaying(t *testing.T) {
	m := NewModel()
	m.screen = screenHelp
	m.started = true // game was in progress

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := result.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
}

func TestPlayingEscGoesToMenu(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50
	// Start a game first
	result, _ := m.startGame()
	m = result

	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := res.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestPlayingPause(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50
	result, _ := m.startGame()
	m = result

	res, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model := res.(Model)

	if model.screen != screenPaused {
		t.Errorf("expected screenPaused, got %d", model.screen)
	}
	if cmd != nil {
		t.Error("expected no tick command while pausing")
	}
}

func TestPauseResume(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50
	m.screen = screenPaused
	m.started = true

	result, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model := result.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
	if cmd == nil {
		t.Error("expected tick command after unpausing")
	}
}

func TestPauseQuit(t *testing.T) {
	m := NewModel()
	m.screen = screenPaused

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestPauseEscGoesToMenu(t *testing.T) {
	m := NewModel()
	m.screen = screenPaused

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestPauseStopsTick(t *testing.T) {
	m := newPlayingModel()
	m.screen = screenPaused

	ac := aircraft.New("AA1", 30, 15, 90, 5, 3)
	m.aircraft["AA1"] = ac

	result, cmd := m.Update(tickMsg(time.Now()))
	model := result.(Model)

	if cmd != nil {
		t.Error("expected no tick command while paused")
	}
	if model.aircraft["AA1"].X != ac.X {
		t.Error("aircraft should not move while paused")
	}
}

func TestTickAdvancesAircraft(t *testing.T) {
	m := newPlayingModel()

	ac := aircraft.New("AA123", 30, 15, 90, 5, 3)
	m.aircraft["AA123"] = ac

	now := time.Now()
	result, cmd := m.Update(tickMsg(now))
	model := result.(Model)

	if cmd == nil {
		t.Error("expected tick command to be returned")
	}
	moved := model.aircraft["AA123"]
	if moved.X == ac.X && moved.Y == ac.Y {
		t.Error("aircraft should have moved after tick")
	}
}

func TestCollisionEndsGame(t *testing.T) {
	m := newPlayingModel()

	m.aircraft["AA1"] = aircraft.New("AA1", 30, 15, 0, 5, 1)
	m.aircraft["AA2"] = aircraft.New("AA2", 30, 15, 0, 5, 1)

	result, cmd := m.Update(tickMsg(time.Now()))
	model := result.(Model)

	if model.screen != screenGameOver {
		t.Errorf("expected screenGameOver, got %d", model.screen)
	}
	if cmd != nil {
		t.Error("expected no tick command after game over")
	}
}

func TestLandingScores(t *testing.T) {
	m := newPlayingModel()

	rw := m.runways[0]
	ac := aircraft.New("AA1", float64(rw.X), float64(rw.Y), rw.Heading, 1, 1)
	ac.State = aircraft.Landing
	m.aircraft["AA1"] = ac

	result, _ := m.Update(tickMsg(time.Now()))
	model := result.(Model)

	if model.score != 1 {
		t.Errorf("score = %d, want 1", model.score)
	}
	if _, exists := model.aircraft["AA1"]; exists {
		t.Error("landed aircraft should be removed")
	}
}

func TestGameOverRestart(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50
	m.screen = screenGameOver
	m.score = 10

	result, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := result.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
	if model.score != 0 {
		t.Errorf("score should be 0 after restart, got %d", model.score)
	}
	if cmd == nil {
		t.Error("expected init command after restart")
	}
}

func TestGameOverEscGoesToMenu(t *testing.T) {
	m := NewModel()
	m.screen = screenGameOver

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestViewMenuScreen(t *testing.T) {
	m := NewModel()
	m.width, m.height = 100, 50

	view := m.View()
	if view == "" {
		t.Error("expected menu view output")
	}
}

func TestViewMinSize(t *testing.T) {
	m := NewModel()
	m.width, m.height = 30, 10

	view := m.View()
	if view == "" {
		t.Error("expected view output even for small terminal")
	}
}
