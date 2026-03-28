package cmdtree

import (
	"strings"

	"github.com/Jasrags/atc/internal/ui"
	zone "github.com/lrstanley/bubblezone"
)

// Render builds the command tree display as a row of clickable options.
// Returns empty string if there are no options to show.
func Render(tree Tree) string {
	if len(tree.Options) == 0 {
		return ""
	}

	var sb strings.Builder

	switch tree.Phase {
	case PhaseCallsign:
		sb.WriteString(ui.Dim.Render("Commands: "))
	case PhaseValue:
		sb.WriteString(ui.Dim.Render("Values: "))
	case PhaseChain:
		sb.WriteString(ui.Dim.Render("Add: "))
	}

	for i, opt := range tree.Options {
		if i > 0 {
			sb.WriteString("  ")
		}
		var styled string
		if opt.IsSubmit {
			styled = ui.CmdTreeSubmit.Render("[" + opt.Label + "]")
		} else {
			styled = ui.CmdTreeOption.Render("[" + opt.Label + "]")
		}
		sb.WriteString(zone.Mark(opt.ZoneID, styled))
	}

	return sb.String()
}
