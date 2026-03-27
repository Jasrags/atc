package ui

import "strings"

// RenderHelp builds the help overlay content.
func RenderHelp() string {
	sections := []struct {
		title string
		items [][2]string
	}{
		{
			title: "Commands",
			items: [][2]string{
				{"<CALLSIGN> H<heading>", "Set heading (0-359)"},
				{"<CALLSIGN> A<altitude>", "Set altitude (1-40, in 1000ft)"},
				{"<CALLSIGN> S<speed>", "Set speed (1-5)"},
				{"<CALLSIGN> L", "Clear to land"},
				{"Example:", "AA123 H270 A3 S2"},
			},
		},
		{
			title: "Keys",
			items: [][2]string{
				{"Enter", "Submit command"},
				{"P", "Pause / Resume"},
				{"?", "Toggle help"},
				{"Esc", "Back to menu"},
				{"R (game over)", "Restart"},
				{"Ctrl+C", "Quit"},
			},
		},
		{
			title: "Landing Requirements",
			items: [][2]string{
				{"Position", "Within 2 cells of runway"},
				{"Heading", "Within +/-10 of runway heading"},
				{"Altitude", "Must be 1 (1000ft)"},
				{"Clearance", "Must issue L command first"},
			},
		},
	}

	var sb strings.Builder
	sb.WriteString(HelpTitle.Render("  ATC HELP  ") + "\n\n")

	for _, section := range sections {
		sb.WriteString(HelpTitle.Render(section.title) + "\n")
		for _, item := range section.items {
			key := HelpKey.Render(item[0])
			desc := HelpDesc.Render(" - " + item[1])
			sb.WriteString("  " + key + desc + "\n")
		}
		sb.WriteRune('\n')
	}

	sb.WriteString(Dim.Render("  Press ? or Esc to go back") + "\n")

	return HelpBox.Render(sb.String())
}
