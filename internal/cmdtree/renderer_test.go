package cmdtree

import (
	"os"
	"strings"
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
	zone "github.com/lrstanley/bubblezone"
)

func TestMain(m *testing.M) {
	zone.NewGlobal()
	os.Exit(m.Run())
}

func TestRenderEmpty(t *testing.T) {
	tree := Tree{Phase: PhaseIdle}
	result := Render(tree)
	if result != "" {
		t.Errorf("expected empty string for idle tree, got %q", result)
	}
}

func TestRenderCallsignPhase(t *testing.T) {
	tree := Resolve("AA123 ", aircraft.Approaching)
	result := Render(tree)
	if !strings.Contains(result, "Commands") {
		t.Error("expected 'Commands' label in callsign phase")
	}
	if !strings.Contains(result, "H") {
		t.Error("expected H option in output")
	}
}

func TestRenderValuePhase(t *testing.T) {
	tree := Resolve("AA123 H", aircraft.Approaching)
	result := Render(tree)
	if !strings.Contains(result, "Values") {
		t.Error("expected 'Values' label in value phase")
	}
	if !strings.Contains(result, "270") {
		t.Error("expected 270 in heading values")
	}
}

func TestRenderChainPhase(t *testing.T) {
	tree := Resolve("AA123 H270 ", aircraft.Approaching)
	result := Render(tree)
	if !strings.Contains(result, "Add") {
		t.Error("expected 'Add' label in chain phase")
	}
	if !strings.Contains(result, "Enter") {
		t.Error("expected Enter/Send option in chain phase")
	}
}

func TestRenderOptionsClickable(t *testing.T) {
	tree := Resolve("AA123 ", aircraft.Approaching)
	result := Render(tree)
	// Each option should be wrapped in brackets
	if !strings.Contains(result, "[") {
		t.Error("expected bracketed options")
	}
}
