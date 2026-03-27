package ui

import (
	"strings"
	"testing"
)

func TestRenderSetupContainsSections(t *testing.T) {
	sections := []SetupSection{
		{Title: "Map", Options: []string{"Tutorial", "San Diego"}, Selected: 0},
		{Title: "Difficulty", Options: []string{"Easy", "Normal", "Hard"}, Selected: 1},
	}

	result := RenderSetup(sections, 0)

	if !strings.Contains(result, "Map") {
		t.Error("expected 'Map' section title")
	}
	if !strings.Contains(result, "Difficulty") {
		t.Error("expected 'Difficulty' section title")
	}
	if !strings.Contains(result, "Tutorial") {
		t.Error("expected 'Tutorial' option")
	}
	if !strings.Contains(result, "Normal") {
		t.Error("expected 'Normal' option")
	}
}

func TestRenderSetupSelectedMarker(t *testing.T) {
	sections := []SetupSection{
		{Title: "Test", Options: []string{"A", "B"}, Selected: 1},
	}

	result := RenderSetup(sections, 0)

	if !strings.Contains(result, "> B") {
		t.Error("expected '> B' for selected option")
	}
}

func TestRenderSetupFooter(t *testing.T) {
	sections := []SetupSection{
		{Title: "Test", Options: []string{"A"}, Selected: 0},
	}

	result := RenderSetup(sections, 0)

	if !strings.Contains(result, "Tab") {
		t.Error("expected 'Tab' in footer")
	}
	if !strings.Contains(result, "Enter") {
		t.Error("expected 'Enter' in footer")
	}
	if !strings.Contains(result, "Esc") {
		t.Error("expected 'Esc' in footer")
	}
}

func TestRenderSetupGameSetupTitle(t *testing.T) {
	sections := []SetupSection{
		{Title: "Test", Options: []string{"A"}, Selected: 0},
	}

	result := RenderSetup(sections, 0)

	if !strings.Contains(result, "GAME SETUP") {
		t.Error("expected 'GAME SETUP' title")
	}
}
