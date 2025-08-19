package tui

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"go.mau.fi/whatsmeow"
)

func (m model) run() tea.Cmd {
	d, err := m.wa.C.GetFirstDevice(context.Background())
	if err != nil {
		log.Fatalf("error getting device: %w", err)
	}

	cli := whatsmeow.NewClient(d, m.wa.WaLog())

	return func() tea.Msg {
		if cli.Store.ID == nil {
			qrChan, _ := cli.GetQRChannel(context.Background())

			err = cli.Connect()
			if err != nil {
				log.Fatal(err)
			}

			for evt := range qrChan {
				if evt.Event == "code" {
					m.events <- pairQrMsg{Code: evt.Code}
				} else {
					m.events <- pairQrMsg{}
					m.events <- loggedInMsg{cli: cli}
				}
			}
		}

		if !cli.IsConnected() {
			err := cli.Connect()
			if err != nil {
				log.Fatal(err)
			}
		}

		m.events <- loggedInMsg{cli: cli}
		return nil
	}
}

func (m model) waitEvents() tea.Cmd {
	return func() tea.Msg {
		return <-m.events
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.roomList.Init(),
		m.loading.Tick,
		m.run(),
		m.waitEvents(),
	)
}
