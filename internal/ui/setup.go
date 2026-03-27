package ui

import (
	"fmt"
	"strings"
)

// SetupSection represents one configurable section on the setup screen.
type SetupSection struct {
	Title    string
	Options  []string
	Selected int
}

// RenderSetup builds the combined game setup screen.
func RenderSetup(sections []SetupSection, focusedSection int) string {
	var sb strings.Builder

	sb.WriteString(HelpTitle.Render("  GAME SETUP  ") + "\n\n")

	for i, section := range sections {
		focused := i == focusedSection

		// Section title
		titleStyle := Dim
		if focused {
			titleStyle = HUDScore
		}
		sb.WriteString(titleStyle.Render(section.Title) + "\n")

		// Options
		for j, opt := range section.Options {
			selected := j == section.Selected
			line := renderSetupOption(opt, selected, focused)
			sb.WriteString(line + "\n")
		}

		// Spacing between sections (except last)
		if i < len(sections)-1 {
			sb.WriteRune('\n')
		}
	}

	sb.WriteString("\n")
	sb.WriteString(Dim.Render(fmt.Sprintf(
		"  Tab: next section  |  %s/%s: select  |  Enter: start  |  Esc: back",
		HelpKey.Render("↑"), HelpKey.Render("↓"))) + "\n")

	return HelpBox.Render(sb.String())
}

func renderSetupOption(label string, selected, focused bool) string {
	if selected && focused {
		return HUDScore.Render("> " + label)
	}
	if selected {
		return HelpDesc.Render("> " + label)
	}
	return Dim.Render("  " + label)
}
