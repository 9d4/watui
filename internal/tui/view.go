package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdp/qrterminal/v3"
)

func (m model) View() string {
	switch m.state {
	case stateInit:
		if m.waQRCode != "" {
			var str strings.Builder
			str.WriteString(m.loading.View() + "Scan the code below \n\n")
			qrterminal.GenerateHalfBlock(m.waQRCode, qrterminal.L, &str)
			return str.String()
		}

		return m.loading.View() + m.loadingText
	}

	leftBox := m.roomList.View()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Top, leftBox),
	)
}
