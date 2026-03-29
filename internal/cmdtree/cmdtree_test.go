package cmdtree

import (
	"testing"

	"github.com/Jasrags/atc/internal/aircraft"
	"github.com/Jasrags/atc/internal/config"
)

func TestResolveEmpty(t *testing.T) {
	tree := Resolve("", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseIdle {
		t.Errorf("expected PhaseIdle, got %d", tree.Phase)
	}
	if len(tree.Options) != 0 {
		t.Error("expected no options for empty input")
	}
}

func TestResolveCallsignNoSpace(t *testing.T) {
	tree := Resolve("AA123", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseIdle {
		t.Errorf("expected PhaseIdle while still typing callsign, got %d", tree.Phase)
	}
}

func TestResolveCallsignWithSpace(t *testing.T) {
	tree := Resolve("AA123 ", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseCallsign {
		t.Errorf("expected PhaseCallsign, got %d", tree.Phase)
	}
	if len(tree.Options) == 0 {
		t.Fatal("expected command options")
	}
	// Should have H, A, S, L for approaching aircraft
	labels := optionValues(tree.Options)
	for _, want := range []string{"H", "A", "S", "L"} {
		if !contains(labels, want) {
			t.Errorf("expected option %q in %v", want, labels)
		}
	}
}

func TestResolveCallsignLanding(t *testing.T) {
	tree := Resolve("AA123 ", aircraft.Landing, config.RoleCombined)
	if tree.Phase != PhaseCallsign {
		t.Errorf("expected PhaseCallsign, got %d", tree.Phase)
	}
	labels := optionValues(tree.Options)
	if !contains(labels, "GA") {
		t.Errorf("expected GA for landing aircraft, got %v", labels)
	}
	if contains(labels, "H") {
		t.Error("should not offer H for landing aircraft")
	}
}

func TestResolveCommandPrefix(t *testing.T) {
	tree := Resolve("AA123 H", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseValue {
		t.Errorf("expected PhaseValue, got %d", tree.Phase)
	}
	if len(tree.Options) == 0 {
		t.Fatal("expected heading value options")
	}
	// Should have compass rose values
	labels := optionValues(tree.Options)
	if !contains(labels, "090") {
		t.Errorf("expected 090 in heading options, got %v", labels)
	}
	if !contains(labels, "270") {
		t.Errorf("expected 270 in heading options, got %v", labels)
	}
}

func TestResolveAltitudePrefix(t *testing.T) {
	tree := Resolve("AA123 A", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseValue {
		t.Errorf("expected PhaseValue, got %d", tree.Phase)
	}
	labels := optionValues(tree.Options)
	if !contains(labels, "3") {
		t.Errorf("expected 3 in altitude options, got %v", labels)
	}
}

func TestResolveSpeedPrefix(t *testing.T) {
	tree := Resolve("AA123 S", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseValue {
		t.Errorf("expected PhaseValue, got %d", tree.Phase)
	}
	labels := optionValues(tree.Options)
	if !contains(labels, "3") {
		t.Errorf("expected 3 in speed options, got %v", labels)
	}
	if len(tree.Options) != 5 {
		t.Errorf("expected 5 speed options, got %d", len(tree.Options))
	}
}

func TestResolveChainAfterCommand(t *testing.T) {
	tree := Resolve("AA123 H270 ", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseChain {
		t.Errorf("expected PhaseChain, got %d", tree.Phase)
	}
	labels := optionValues(tree.Options)
	// H already used, should not appear
	if contains(labels, "H") {
		t.Error("H should not appear in chain options (already used)")
	}
	// A and S should still be available
	if !contains(labels, "A") {
		t.Errorf("expected A in chain options, got %v", labels)
	}
	if !contains(labels, "S") {
		t.Errorf("expected S in chain options, got %v", labels)
	}
	// Should have a send option
	hasSend := false
	for _, o := range tree.Options {
		if o.IsSubmit {
			hasSend = true
		}
	}
	if !hasSend {
		t.Error("expected Send option in chain")
	}
}

func TestResolveChainAfterLand(t *testing.T) {
	tree := Resolve("AA123 L ", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseChain {
		t.Errorf("expected PhaseChain, got %d", tree.Phase)
	}
	labels := optionValues(tree.Options)
	if contains(labels, "L") {
		t.Error("L should not appear in chain options (already used)")
	}
}

func TestResolveChainAfterMultiple(t *testing.T) {
	tree := Resolve("AA123 H270 A3 ", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseChain {
		t.Errorf("expected PhaseChain, got %d", tree.Phase)
	}
	labels := optionValues(tree.Options)
	if contains(labels, "H") {
		t.Error("H should not appear (already used)")
	}
	if contains(labels, "A") {
		t.Error("A should not appear (already used)")
	}
	if !contains(labels, "S") {
		t.Error("S should still be available")
	}
}

func TestResolveMidTypeValue(t *testing.T) {
	// User is typing a heading value — no tree
	tree := Resolve("AA123 H27", aircraft.Approaching, config.RoleCombined)
	if tree.Phase != PhaseIdle {
		t.Errorf("expected PhaseIdle while typing value, got %d", tree.Phase)
	}
}

func TestResolveHeading360Maps000(t *testing.T) {
	tree := Resolve("AA123 H", aircraft.Approaching, config.RoleCombined)
	for _, opt := range tree.Options {
		if opt.Label == "360" {
			if opt.Value != "000" {
				t.Errorf("heading 360 should map to value 000, got %s", opt.Value)
			}
			return
		}
	}
	t.Error("expected heading 360 option")
}

func optionValues(opts []Option) []string {
	values := make([]string, len(opts))
	for i, o := range opts {
		values[i] = o.Value
	}
	return values
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
