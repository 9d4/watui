package tui

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pairQrMsg:
		m.waQRCode = msg.Code
		cmds = append(cmds, m.waitEvents())
		cmds = append(cmds, tea.ClearScreen)
		return m, tea.Batch(cmds...)

	case loggedInMsg:
		m.state = stateIdle
		m.cli = msg.cli
		if msg.cli == nil {
			cmds = append(cmds, m.waitEvents())
			panic("cli is nil")
		}
		cmds = append(cmds, m.waitEvents())
		cmds = append(cmds, tea.ClearScreen)

		return m, tea.Batch(cmds...)

	case waEvent:
		cmds = append(cmds, m.waitEvents())

		switch evt := msg.evt.(type) {
		case *events.Connected:
			if m.state != stateIdle {
				break
			}

		case *events.LoggedOut:
			return m, func() tea.Msg {
				println("logging out")
				m.cli.Logout(context.Background())
				return tea.Quit()
			}

		case *events.HistorySync:
			println("history received")
			println(evt.Data.SyncType)
			println(evt.Data.Conversations)
			println(evt.Data.GetProgress())

		case *events.Message:
			log.Println("Received whatsapp message: ", evt.Info.Chat.String(), "-", evt.Info.Sender.String(), "-", evt.Info.Type)
		}

	case tea.FocusMsg:
		m.cli.SendPresence(types.PresenceAvailable)

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

	case stateIdle:
		m.roomList, cmd = m.roomList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
