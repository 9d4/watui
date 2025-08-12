package tui

import tea "github.com/charmbracelet/bubbletea"

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.state {
	case stateIdle:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			switch key {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	m.roomList, cmd = m.roomList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
