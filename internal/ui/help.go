package ui

import "strings"

// RenderCommandHelp builds the ATC command reference section (not keybindings — those are auto-generated).
func RenderCommandHelp() string {
	var sb strings.Builder

	sb.WriteString(HelpTitle.Render("  ATC HELP  ") + "\n\n")

	// Command syntax
	sb.WriteString(HelpTitle.Render("Commands") + "\n")
	commands := [][2]string{
		{"<CALLSIGN> H<heading>", "Set heading (0-359)"},
		{"<CALLSIGN> A<altitude>", "Set altitude (1-40, in 1000ft)"},
		{"<CALLSIGN> S<speed>", "Set speed (1-5)"},
		{"<CALLSIGN> L", "Clear to land"},
		{"Example:", "AA123 H270 A3 S2"},
	}
	for _, item := range commands {
		sb.WriteString("  " + HelpKey.Render(item[0]) + HelpDesc.Render(" - "+item[1]) + "\n")
	}
	sb.WriteRune('\n')

	// Landing requirements
	sb.WriteString(HelpTitle.Render("Landing Requirements") + "\n")
	reqs := [][2]string{
		{"Position", "Within 2 cells of runway"},
		{"Heading", "Within +/-10 of runway heading"},
		{"Altitude", "Must be 1 (1000ft)"},
		{"Clearance", "Must issue L command first"},
	}
	for _, item := range reqs {
		sb.WriteString("  " + HelpKey.Render(item[0]) + HelpDesc.Render(" - "+item[1]) + "\n")
	}

	return sb.String()
}
