package tui

import "github.com/charmbracelet/lipgloss"

func (m model) View() string {
	leftBox := m.roomList.View()

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Render(
			lipgloss.JoinHorizontal(lipgloss.Top, leftBox),
		),
	)
}
