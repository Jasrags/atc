package radar

import (
	"strings"
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/runway"
)

func TestRenderEmptyGrid(t *testing.T) {
	rw := runway.New(10, 8, 270, 5)
	result := Render(20, 10, rw, nil)

	if !strings.Contains(result, "+") {
		t.Error("expected border characters in output")
	}

	lines := strings.Split(result, "\n")
	// top border + 10 rows + bottom border = 12 lines
	if len(lines) != 12 {
		t.Errorf("expected 12 lines, got %d", len(lines))
	}
}

func TestRenderWithAircraft(t *testing.T) {
	rw := runway.New(10, 8, 270, 3)
	planes := []aircraft.Aircraft{
		aircraft.New("AA1", 5, 3, 270, 5, 2),
	}

	result := Render(20, 10, rw, planes)

	if !strings.Contains(result, "@") {
		t.Error("expected aircraft symbol '@' in output")
	}
}

func TestRenderSidebarEmpty(t *testing.T) {
	result := RenderSidebar(nil)
	if !strings.Contains(result, "No aircraft") {
		t.Error("expected 'No aircraft' for empty sidebar")
	}
}

func TestRenderSidebarWithAircraft(t *testing.T) {
	planes := []aircraft.Aircraft{
		aircraft.New("UA456", 10, 10, 180, 8, 3),
	}
	result := RenderSidebar(planes)

	if !strings.Contains(result, "UA456") {
		t.Error("expected callsign in sidebar")
	}
	if !strings.Contains(result, "180") {
		t.Error("expected heading in sidebar")
	}
}
