package game

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
)

func TestMain(m *testing.M) {
	zone.NewGlobal()
	os.Exit(m.Run())
}

func newPlayingModel() Model {
	m := NewModel(false)
	m.width, m.height = 100, 50
	cfg := config.DefaultConfig()
	cfg.Role = config.RoleCombined // tests use Combined for explicit control
	m.gameConfig = cfg
	started, _ := m.startGame()
	return started
}

func TestNewModelStartsAtMenu(t *testing.T) {
	m := NewModel(false)
	if m.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", m.screen)
	}
}

func TestMenuNavigation(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	model := result.(Model)
	if model.menuSelected != 1 {
		t.Errorf("expected selected=1 after down, got %d", model.menuSelected)
	}

	result, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	model = result.(Model)
	if model.menuSelected != 0 {
		t.Errorf("expected selected=0 after up, got %d", model.menuSelected)
	}
}

func TestMenuNewGameGoesToSetup(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	if model.menuScreen != menuSetup {
		t.Errorf("expected menuSetup, got %d", model.menuScreen)
	}
}

func TestSetupTabCyclesFocus(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50
	m.menuScreen = menuSetup
	m.setupFocus = 0

	for i := 0; i < setupSections; i++ {
		result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
		m = result.(Model)
		expected := (i + 1) % setupSections
		if m.setupFocus != expected {
			t.Errorf("after %d tabs, expected focus %d, got %d", i+1, expected, m.setupFocus)
		}
	}
}

func TestSetupShiftTabCyclesBack(t *testing.T) {
	m := NewModel(false)
	m.menuScreen = menuSetup
	m.setupFocus = 0

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	model := result.(Model)
	if model.setupFocus != setupSections-1 {
		t.Errorf("expected focus %d, got %d", setupSections-1, model.setupFocus)
	}
}

func TestSetupUpDownChangesSelection(t *testing.T) {
	m := NewModel(false)
	m.menuScreen = menuSetup
	m.setupFocus = setupDiff         // Difficulty: Easy/Normal/Hard
	m.setupSelections[setupDiff] = 1 // Normal

	// Down -> Hard
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	model := result.(Model)
	if model.setupSelections[setupDiff] != 2 {
		t.Errorf("expected 2 (Hard), got %d", model.setupSelections[setupDiff])
	}

	// Down again -> clamped at 2
	result, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	model = result.(Model)
	if model.setupSelections[setupDiff] != 2 {
		t.Errorf("expected 2 (clamped), got %d", model.setupSelections[setupDiff])
	}

	// Up -> Normal
	result, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	model = result.(Model)
	if model.setupSelections[setupDiff] != 1 {
		t.Errorf("expected 1 (Normal), got %d", model.setupSelections[setupDiff])
	}
}

func TestSetupStartsGame(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50
	m.menuScreen = menuSetup

	result, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
	if cmd == nil {
		t.Error("expected init commands after starting game")
	}
}

func TestSetupBuildsConfig(t *testing.T) {
	m := NewModel(false)
	m.menuScreen = menuSetup
	m.setupSelections[setupDiff] = 2     // Hard
	m.setupSelections[setupCallsign] = 1 // Short
	m.setupSelections[setupTrails] = 0   // On

	cfg := m.buildConfigFromSetup()
	if cfg.Difficulty != config.DifficultyHard {
		t.Errorf("expected Hard, got %v", cfg.Difficulty)
	}
	if cfg.CallsignStyle != config.CallsignShort {
		t.Errorf("expected Short, got %v", cfg.CallsignStyle)
	}
	if !cfg.PlaneTrails {
		t.Error("expected PlaneTrails true")
	}
}

func TestSetupEscGoesBack(t *testing.T) {
	m := NewModel(false)
	m.menuScreen = menuSetup

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.menuScreen != menuMain {
		t.Errorf("expected menuMain, got %d", model.menuScreen)
	}
}

func TestMenuHelp(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model := result.(Model)

	if model.screen != screenHelp {
		t.Errorf("expected screenHelp, got %d", model.screen)
	}
}

func TestMenuQuit(t *testing.T) {
	m := NewModel(false)

	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestHelpReturnToMenu(t *testing.T) {
	m := NewModel(false)
	m.screen = screenHelp

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestHelpReturnToPlaying(t *testing.T) {
	m := NewModel(false)
	m.screen = screenHelp
	m.started = true

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := result.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
}

func TestPlayingEscGoesToMenu(t *testing.T) {
	m := newPlayingModel()

	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := res.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestPlayingFreeze(t *testing.T) {
	m := newPlayingModel()

	res, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model := res.(Model)

	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying (freeze stays on playing screen), got %d", model.screen)
	}
	if !model.timeFrozen {
		t.Error("expected timeFrozen=true after pressing p")
	}
	if cmd == nil {
		t.Error("expected stopwatch stop command when freezing")
	}
}

func TestPlayingFreezeBlockedWhileTyping(t *testing.T) {
	m := newPlayingModel()
	m.input.SetValue("AA123 ")

	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model := res.(Model)

	if model.timeFrozen {
		t.Error("pressing 'p' while typing should NOT freeze")
	}
	if !strings.Contains(model.input.Value(), "p") {
		t.Error("'p' should be forwarded to input when typing")
	}
}

func TestPlayingHelpBlockedWhileTyping(t *testing.T) {
	m := newPlayingModel()
	m.input.SetValue("AA123 ")

	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := res.(Model)

	if model.screen != screenPlaying {
		t.Error("pressing '?' while typing should NOT open help")
	}
}

func TestFreezeResume(t *testing.T) {
	m := newPlayingModel()
	m.timeFrozen = true

	result, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model := result.(Model)

	if model.timeFrozen {
		t.Error("expected timeFrozen=false after pressing p again")
	}
	if model.screen != screenPlaying {
		t.Errorf("expected screenPlaying, got %d", model.screen)
	}
	if cmd == nil {
		t.Error("expected tick+stopwatch command after unfreezing")
	}
}

func TestFreezeStopsTick(t *testing.T) {
	m := newPlayingModel()
	m.timeFrozen = true

	ac := aircraft.New("AA1", 30, 15, 90, 5, 3)
	m.aircraft["AA1"] = ac

	result, cmd := m.Update(tickMsg(time.Now()))
	model := result.(Model)

	if cmd == nil {
		t.Error("expected tick command to keep render loop alive while frozen")
	}
	if model.aircraft["AA1"].X != ac.X {
		t.Error("aircraft should not move while frozen")
	}
}

func TestSpeedUpDown(t *testing.T) {
	m := newPlayingModel()
	if m.speedMultiplier != 1 {
		t.Fatalf("expected initial speed 1, got %d", m.speedMultiplier)
	}

	// Speed up
	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model := res.(Model)
	if model.speedMultiplier != 2 {
		t.Errorf("expected speed 2 after ], got %d", model.speedMultiplier)
	}

	// Speed down back to 1
	res, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = res.(Model)
	if model.speedMultiplier != 1 {
		t.Errorf("expected speed 1 after [, got %d", model.speedMultiplier)
	}

	// Speed down clamps at 1
	res, _ = model.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = res.(Model)
	if model.speedMultiplier != 1 {
		t.Errorf("expected speed clamped at 1, got %d", model.speedMultiplier)
	}
}

func TestSpeedClampsAt12(t *testing.T) {
	m := newPlayingModel()
	m.speedMultiplier = 12

	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model := res.(Model)
	if model.speedMultiplier != 12 {
		t.Errorf("expected speed clamped at 12, got %d", model.speedMultiplier)
	}
}

func TestSpeedBlockedWhileTyping(t *testing.T) {
	m := newPlayingModel()
	m.input.SetValue("AA123 ")

	res, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model := res.(Model)
	if model.speedMultiplier != 1 {
		t.Error("speed should not change while typing")
	}
}

func TestNewGameResetsTimeControls(t *testing.T) {
	m := newPlayingModel()
	m.timeFrozen = true
	m.speedMultiplier = 8

	// Go to menu and start new game
	m.screen = screenMenu
	m.menuScreen = menuSetup
	started, _ := m.startGame()

	if started.timeFrozen {
		t.Error("expected timeFrozen=false after new game")
	}
	if started.speedMultiplier != 1 {
		t.Errorf("expected speed=1 after new game, got %d", started.speedMultiplier)
	}
}

func TestTickAdvancesAircraft(t *testing.T) {
	m := newPlayingModel()

	ac := aircraft.New("AA123", 30, 15, 90, 5, 3)
	m.aircraft["AA123"] = ac

	result, cmd := m.Update(tickMsg(time.Now()))
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
	// Landed aircraft stay in the map for gate assignment
	ac2, exists := model.aircraft["AA1"]
	if !exists {
		t.Fatal("landed aircraft should remain in map for gate taxi")
	}
	if ac2.State != aircraft.Landed {
		t.Errorf("state = %v, want Landed", ac2.State)
	}
}

func TestGameOverRestart(t *testing.T) {
	m := newPlayingModel()
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
	m := NewModel(false)
	m.screen = screenGameOver

	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	if model.screen != screenMenu {
		t.Errorf("expected screenMenu, got %d", model.screen)
	}
}

func TestViewMenuScreen(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50

	view := m.View()
	if view == "" {
		t.Error("expected menu view output")
	}
}

func TestViewSetupScreen(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 100, 50
	m.menuScreen = menuSetup

	view := m.View()
	if view == "" {
		t.Error("expected setup view output")
	}
}

func TestProcessCommandValid(t *testing.T) {
	m := newPlayingModel()
	ac := aircraft.New("AA123", 30, 15, 90, 10, 3)
	m.aircraft["AA123"] = ac

	m = m.processCommand("AA123 H270 A3")

	updated := m.aircraft["AA123"]
	if updated.TargetHeading != 270 {
		t.Errorf("expected target heading 270, got %d", updated.TargetHeading)
	}
	if updated.TargetAltitude != 3 {
		t.Errorf("expected target altitude 3, got %d", updated.TargetAltitude)
	}
}

func TestProcessCommandInvalid(t *testing.T) {
	m := newPlayingModel()
	logCount := m.radioLog.Len()

	m = m.processCommand("INVALID GARBAGE")

	if m.radioLog.Len() <= logCount {
		t.Error("expected error message for invalid command")
	}
}

func TestProcessCommandUnknownCallsign(t *testing.T) {
	m := newPlayingModel()
	logCount := m.radioLog.Len()

	m = m.processCommand("XX999 H270")

	if m.radioLog.Len() <= logCount {
		t.Error("expected error message for unknown callsign")
	}
}

func TestViewMinSize(t *testing.T) {
	m := NewModel(false)
	m.width, m.height = 30, 10

	view := m.View()
	if view == "" {
		t.Error("expected view output even for small terminal")
	}
}
