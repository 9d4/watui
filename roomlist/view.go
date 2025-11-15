package roomlist

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var roomList strings.Builder

	for i, item := range m.Rooms {
		time := item.Time.Format("02/01 15:04")
		lastMessage := item.LastMessage
		if lastMessage == "" {
			lastMessage = "-"
		}

		switch {
		case m.openedRoomIndex != nil && *m.openedRoomIndex == i:
			roomList.WriteString(
				lipgloss.NewStyle().
					Bold(true).
					Render(time) + " ",
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
					Render(time) + " ",
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
					Render(time) + " ",
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
