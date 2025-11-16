package roomlist

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m Model) View() string {
	var roomList strings.Builder

	start, visible := m.visibleRooms()

	for idx, item := range visible {
		i := start + idx
		timeStr := "-"
		if !item.Time.IsZero() {
			timeStr = item.Time.Format("02/01 15:04")
		}
		lastMessage := previewText(item.LastMessage, 48)

		switch {
		case m.openedRoomIndex != nil && *m.openedRoomIndex == i:
			roomList.WriteString(
				lipgloss.NewStyle().
					Bold(true).
					Render(timeStr) + " ",
			)

			roomList.WriteString(
				lipgloss.NewStyle().
					Foreground(m.openedItemColor).
					Bold(true).
					Render(item.Title) + "\n",
			)

		case i == m.cursor:
			roomList.WriteString(
				lipgloss.NewStyle().
					Bold(true).
					Render(timeStr) + " ",
			)

			roomList.WriteString(
				lipgloss.NewStyle().
					Foreground(m.selectedItemColor).
					Bold(true).
					Render(item.Title) + "\n",
			)

		default:
			roomList.WriteString(
				lipgloss.NewStyle().
					Faint(true).
					Render(timeStr) + " ",
			)

			roomList.WriteString(
				lipgloss.NewStyle().
					Faint(true).
					Render(item.Title) + "\n",
			)
		}

		roomList.WriteString(
			lipgloss.NewStyle().
				Faint(true).
				Render("  "+lastMessage) + "\n\n",
		)
	}

	return roomList.String()
}

func previewText(msg string, width int) string {
	if msg == "" {
		return "-"
	}

	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return "-"
	}

	if runewidth.StringWidth(msg) <= width {
		return msg
	}

	if width <= 1 {
		return runewidth.Truncate(msg, 1, "")
	}

	return runewidth.Truncate(msg, width-1, "") + "…"
}
