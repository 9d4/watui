package roomlist

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "j", "down":
			if len(m.Rooms)-1 != m.cursor {
				m.cursor++
			}
			m.pendingGoTop = false

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			m.pendingGoTop = false

		case "enter":
			if m.openedRoomIndex == nil {
				m.openedRoomIndex = new(int)
			}

			*m.openedRoomIndex = m.cursor
			m.pendingGoTop = false

		case "g":
			if m.pendingGoTop {
				m.cursor = 0
				m.pendingGoTop = false
			} else {
				m.pendingGoTop = true
			}

		case "G":
			if len(m.Rooms) > 0 {
				m.cursor = len(m.Rooms) - 1
			} else {
				m.cursor = 0
			}
			m.pendingGoTop = false

		default:
			m.pendingGoTop = false

		case "esc":
			m.openedRoomIndex = nil
			m.pendingGoTop = false
		}
	}

	m.ensureCursorVisible()

	return m, nil
}
