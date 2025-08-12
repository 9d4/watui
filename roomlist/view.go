package roomlist

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var roomList strings.Builder

	width := 50 // total width for each row
	nameStyle := lipgloss.NewStyle().Bold(true)
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	timeStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("#666666"))
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	// Background styles for states
	selectedRow := lipgloss.NewStyle().
		Background(lipgloss.Color("#ff0000")).
		// Width(width).
		Padding(0, 1)

	unselectedRow := lipgloss.NewStyle().
		Background(lipgloss.Color("#1e1e1e")).
		Padding(0, 1)

	openedRow := lipgloss.NewStyle().
		Background(lipgloss.Color("#2a2a2a")).
		Padding(0, 1)

	for i, item := range m.rooms {
		// Pick background based on state
		rowStyle := unselectedRow
		if i == m.cursor {
			rowStyle = selectedRow
		} else if m.openedRoomIndex != nil && *m.openedRoomIndex == i {
			rowStyle = openedRow
		}

		// First line: Title + time
		name := nameStyle.Render(item.Title)
		timeStr := timeStyle.Render(item.Time.Format("15:04"))
		line1 := lipgloss.JoinHorizontal(
			lipgloss.Top,
			name,
			strings.Repeat(" ", width-lipgloss.Width(name)-lipgloss.Width(timeStr)),
			timeStr,
		)

		// Second line: last message preview
		lastMsg := truncate(item.LastMessage, width)
		line2 := msgStyle.Render(lastMsg)

		// Combine and apply background style
		row := lipgloss.JoinVertical(lipgloss.Left, line1, line2)
		roomList.WriteString(rowStyle.Render(row))
		roomList.WriteString("\n")

		// Add separator line between rooms
		if i < len(m.rooms)-1 {
			roomList.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
			roomList.WriteString("\n")
		}
	}

	return roomList.String()
}

// Helper to truncate strings without breaking UI
func truncate(s string, max int) string {
	if lipgloss.Width(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
