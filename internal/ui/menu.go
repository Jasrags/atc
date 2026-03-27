package ui

import "strings"

// MenuItem represents a selectable menu option.
type MenuItem struct {
	Label string
	Key   string
}

// RenderMenu builds the main menu screen.
func RenderMenu(items []MenuItem, selected int) string {
	var sb strings.Builder

	// Title art
	title := `
    ___   _______ ______
   /   | /_  __/ // ___/
  / /| |  / / / // /
 / ___ | / / / // /___
/_/  |_|/_/ /_/ \____/
`
	sb.WriteString(HUDTitle.Render(title) + "\n\n")
	sb.WriteString(HelpDesc.Render("  Terminal Air Traffic Control") + "\n\n")

	for i, item := range items {
		cursor := "  "
		style := HelpDesc
		if i == selected {
			cursor = "> "
			style = HUDScore
		}
		sb.WriteString(style.Render(cursor+item.Label) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(Dim.Render("  Use arrow keys to select, Enter to confirm") + "\n")

	return HelpBox.Render(sb.String())
}

// RenderMapSelect builds the map selection screen.
func RenderMapSelect(maps []MenuItem, selected int) string {
	var sb strings.Builder

	sb.WriteString(HelpTitle.Render("  SELECT MAP  ") + "\n\n")

	for i, item := range maps {
		cursor := "  "
		style := HelpDesc
		if i == selected {
			cursor = "> "
			style = HUDScore
		}
		sb.WriteString(style.Render(cursor+item.Label) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(Dim.Render("  Enter to start, Esc to go back") + "\n")

	return HelpBox.Render(sb.String())
}
