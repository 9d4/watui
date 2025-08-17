package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case loadingLogin:
		m.loadingText = "Checking login..."
		cmds = append(cmds, waitLogin(m.loginEvent))

	case loggedIn:
		m.loadingText = "Logged In"
		cmds = append(cmds, waitLogin(m.loginEvent))
	}

	switch m.state {
	case stateLoading:
		m.loading, cmd = m.loading.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.roomList, cmd = m.roomList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
