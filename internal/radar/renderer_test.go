package radar

import (
	"strings"
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/gamemap"
)

func testMap() gamemap.Map {
	return gamemap.Map{
		ID:     "test",
		Name:   "Test",
		Width:  20,
		Height: 10,
		Runways: []gamemap.Runway{
			{Name: "9/27", X: 10, Y: 8, Heading: 270, Length: 3},
		},
		Fixes: []gamemap.Fix{
			{Name: "TST", X: 5, Y: 3, Type: gamemap.FixVOR},
		},
	}
}

func TestRenderEmptyGrid(t *testing.T) {
	result := Render(testMap(), nil)

	if !strings.Contains(result, "+") {
		t.Error("expected border characters in output")
	}

	lines := strings.Split(result, "\n")
	// top border + 10 rows + bottom border = 12 lines
	if len(lines) != 12 {
		t.Errorf("expected 12 lines, got %d", len(lines))
	}
}

func TestRenderWithFixes(t *testing.T) {
	result := Render(testMap(), nil)

	if !strings.Contains(result, "TST") {
		t.Error("expected fix label 'TST' in output")
	}
}

func TestRenderWithAircraft(t *testing.T) {
	planes := []aircraft.Aircraft{
		aircraft.New("AA1", 5, 3, 270, 5, 2),
	}

	result := Render(testMap(), planes)

	if !strings.Contains(result, "@") {
		t.Error("expected aircraft symbol '@' in output")
	}
}

func TestRenderFlightStripsEmpty(t *testing.T) {
	result := RenderFlightStrips(nil)
	if !strings.Contains(result, "No aircraft") {
		t.Error("expected 'No aircraft' for empty strips")
	}
}

func TestRenderFlightStripsWithAircraft(t *testing.T) {
	planes := []aircraft.Aircraft{
		aircraft.New("UA456", 10, 10, 180, 8, 3),
	}
	result := RenderFlightStrips(planes)

	if !strings.Contains(result, "UA456") {
		t.Error("expected callsign in flight strip")
	}
	if !strings.Contains(result, "180") {
		t.Error("expected heading in flight strip")
	}
	if !strings.Contains(result, "APPR") {
		t.Error("expected state in flight strip")
	}
}

func TestRenderFlightStripsShowsTargets(t *testing.T) {
	ac := aircraft.New("AA1", 30, 15, 90, 10, 3)
	ac.TargetHeading = 270
	ac.TargetAltitude = 3
	planes := []aircraft.Aircraft{ac}

	result := RenderFlightStrips(planes)

	if !strings.Contains(result, "H270") {
		t.Error("expected target heading H270 in flight strip")
	}
	if !strings.Contains(result, "A3") {
		t.Error("expected target altitude A3 in flight strip")
	}
}

func TestRenderFlightStripsAltitudeArrow(t *testing.T) {
	ac := aircraft.New("AA1", 30, 15, 90, 10, 3)
	ac.TargetAltitude = 5 // descending
	planes := []aircraft.Aircraft{ac}

	result := RenderFlightStrips(planes)

	if !strings.Contains(result, "↓") {
		t.Error("expected descend arrow in flight strip")
	}
}

func TestRenderRunwayNumbers(t *testing.T) {
	gm := gamemap.Map{
		Width: 30, Height: 5,
		Runways: []gamemap.Runway{
			{Name: "9/27", X: 15, Y: 2, Heading: 270, Length: 5},
		},
	}
	result := Render(gm, nil)

	if !strings.Contains(result, "9") {
		t.Error("expected runway number 9 in output")
	}
	if !strings.Contains(result, "27") {
		t.Error("expected runway number 27 in output")
	}
}
