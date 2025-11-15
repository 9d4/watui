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

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "enter":
			if m.openedRoomIndex == nil {
				m.openedRoomIndex = new(int)
			}

			*m.openedRoomIndex = m.cursor

		case "esc":
			m.openedRoomIndex = nil
		}
	}

	return m, nil
}
