package tui

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pairQrMsg:
		m.waQRCode = msg.Code
		cmds = append(cmds, m.waitEvents())

	case loggedInMsg:
		m.state = stateIdle
		m.cli = msg.cli
		if msg.cli == nil {
			panic("cli is nil")
		}
		cmds = append(cmds, m.waitEvents())

	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "ctrl+q":
			if m.state != stateIdle {
				break
			}

			err := m.cli.Logout(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Logged out")
			return nil, tea.Quit
		}
	}

	switch m.state {
	case stateInit:
		m.loading, cmd = m.loading.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.roomList, cmd = m.roomList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
