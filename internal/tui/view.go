package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	leftBox := m.roomList.View()

	switch m.state {
	case stateLoading:
		return m.loading.View() + m.loadingText
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Top, leftBox),
	)
}
