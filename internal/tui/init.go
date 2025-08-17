package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func fakeLoadLogin(ch chan any) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		ch <- loadingLogin{}
		time.Sleep(2 * time.Second)
		ch <- loggedIn{}

		return nil
	}
}

func waitLogin(ch chan any) tea.Cmd {
	return func() tea.Msg {
		a := <-ch
		return a
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.roomList.Init(),
		m.loading.Tick,
		fakeLoadLogin(m.loginEvent),
		waitLogin(m.loginEvent),
	)
}
