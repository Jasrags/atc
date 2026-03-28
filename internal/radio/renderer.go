package radio

import (
	"fmt"
	"strings"
	"time"

	"github.com/Jasrags/atc/internal/ui"
)

// RenderLog formats the radio log for display in the radio viewport.
func RenderLog(messages []Message) string {
	if len(messages) == 0 {
		return ui.Dim.Render("  No radio traffic")
	}

	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(renderMessage(msg))
		sb.WriteRune('\n')
	}
	return sb.String()
}

func renderMessage(msg Message) string {
	timestamp := formatTimestamp(msg.Time)
	dimTime := ui.Dim.Render(timestamp)

	var line string
	switch msg.Direction {
	case Inbound:
		line = fmt.Sprintf("%s %s: %s", dimTime, ui.RadioInbound.Render(msg.From), msg.Text)
	case Outbound:
		prefix := fmt.Sprintf("ATC → %s", msg.To)
		line = fmt.Sprintf("%s %s: %s", dimTime, ui.RadioOutbound.Render(prefix), msg.Text)
	}

	switch msg.Priority {
	case Urgent:
		return ui.RadioUrgent.Render(line)
	case Emergency:
		return ui.RadioEmergency.Render(line)
	default:
		return line
	}
}

func formatTimestamp(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
