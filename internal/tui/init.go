package tui

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"go.mau.fi/whatsmeow"
)

func (m model) initClient() tea.Cmd {
	return func() tea.Msg {
		d, err := m.wa.C.GetFirstDevice(context.Background())
		if err != nil {
			return errMsg{err: fmt.Errorf("gagal membaca device: %w", err)}
		}

		return clientReadyMsg{cli: whatsmeow.NewClient(d, m.wa.WaLog())}
	}
}

func (m model) startPairing() tea.Cmd {
	return func() tea.Msg {
		if m.cli == nil {
			return errMsg{err: errors.New("client belum siap")}
		}

		qrChan, err := m.cli.GetQRChannel(context.Background())
		if err != nil {
			return errMsg{err: fmt.Errorf("gagal membuat qr channel: %w", err)}
		}

		go func() {
			for evt := range qrChan {
				switch evt.Event {
				case whatsmeow.QRChannelEventCode:
					m.events <- qrCodeMsg{Code: evt.Code}
				case whatsmeow.QRChannelEventError:
					m.events <- qrStatusMsg{Status: evt.Event, Err: evt.Error}
				default:
					m.events <- qrStatusMsg{Status: evt.Event}
				}
			}
		}()

		if err := m.cli.Connect(); err != nil {
			return errMsg{err: fmt.Errorf("gagal connect: %w", err)}
		}

		return nil
	}
}

func (m model) connectClient() tea.Cmd {
	return func() tea.Msg {
		if m.cli == nil {
			return errMsg{err: errors.New("client belum siap")}
		}

		if m.cli.IsConnected() {
			return nil
		}

		if err := m.cli.Connect(); err != nil {
			return errMsg{err: fmt.Errorf("gagal connect: %w", err)}
		}

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
		m.initClient(),
		m.waitEvents(),
	)
}
