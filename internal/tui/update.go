package tui

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	appendCmd := func(c tea.Cmd) {
		if c != nil {
			cmds = append(cmds, c)
		}
	}

	switch msg := msg.(type) {
	case clientReadyMsg:
		m.cli = msg.cli
		if m.cli == nil {
			m.state = stateError
			m.statusMessage = "client tidak tersedia"
			break
		}

		m.cli.AddEventHandler(func(evt any) {
			m.events <- waEvent{evt}
		})

		if m.cli.Store.ID == nil {
			m.state = stateWelcome
			m.statusMessage = "Tekan Enter untuk mulai pairing"
		} else {
			m.state = stateConnecting
			m.statusMessage = "Menghubungkan ke WhatsApp..."
			appendCmd(m.connectClient())
		}

	case qrCodeMsg:
		m.waQRCode = msg.Code
		m.state = statePairing
		m.qrStatus = ""
		appendCmd(m.waitEvents())

	case qrStatusMsg:
		appendCmd(m.waitEvents())

		switch msg.Status {
		case whatsmeow.QRChannelSuccess.Event:
			m.qrStatus = "QR berhasil dipindai, menunggu sinkronisasi"
			m.waQRCode = ""
			m.state = stateHistorySync
			m.historyMessage = "Menunggu history dari WhatsApp..."
			appendCmd(m.syncProgress.SetPercent(0))
		case whatsmeow.QRChannelTimeout.Event:
			m.qrStatus = "Sesi pairing habis, coba ulangi"
			m.state = stateWelcome
		case whatsmeow.QRChannelErrUnexpectedEvent.Event:
			m.qrStatus = "Status pairing tidak terduga, coba ulangi"
			m.state = stateWelcome
		case whatsmeow.QRChannelClientOutdated.Event:
			m.qrStatus = "Client kedaluwarsa, update whatsmeow"
			m.state = stateWelcome
		case whatsmeow.QRChannelScannedWithoutMultidevice.Event:
			m.qrStatus = "Aktifkan multi-device di HP terlebih dahulu"
			m.state = stateWelcome
		case whatsmeow.QRChannelEventError:
			m.qrStatus = fmt.Sprintf("Pairing gagal: %v", msg.Err)
			m.state = stateWelcome
		default:
			m.qrStatus = fmt.Sprintf("Status pairing: %s", msg.Status)
		}

	case waEvent:
		appendCmd(m.waitEvents())

		switch evt := msg.evt.(type) {
		case *events.Connected:
			if m.state != stateHistorySync {
				m.state = stateChats
				m.statusMessage = ""
			}

		case *events.LoggedOut:
			return m, func() tea.Msg {
				if m.cli != nil {
					_ = m.cli.Logout(context.Background())
				}
				return tea.Quit()
			}

		case *events.HistorySync:
			if evt.Data != nil {
				progress := float64(evt.Data.GetProgress()) / 100
				if progress > 1 {
					progress = 1
				}

				if progress >= 1 {
					m.state = stateChats
					m.statusMessage = ""
					appendCmd(m.syncProgress.SetPercent(1))
					break
				}

				appendCmd(m.syncProgress.SetPercent(progress))

				var syncLabel string
				if evt.Data.SyncType != nil {
					syncLabel = strings.ToLower(evt.Data.GetSyncType().String())
				} else {
					syncLabel = "history"
				}

				m.historyMessage = fmt.Sprintf(
					"Sinkronisasi %s · %d chat",
					syncLabel,
					len(evt.Data.GetConversations()),
				)

				m.state = stateHistorySync
			}

		case *events.Message:
			log.Println("Received whatsapp message: ", evt.Info.Chat.String(), "-", evt.Info.Sender.String(), "-", evt.Info.Type)
		}

	case errMsg:
		m.state = stateError
		m.statusMessage = msg.Error()

	case tea.FocusMsg:
		if m.cli != nil && m.state == stateChats {
			m.cli.SendPresence(types.PresenceAvailable)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "ctrl+q":
			if m.cli == nil {
				break
			}

			return m, func() tea.Msg {
				_ = m.cli.Logout(context.Background())
				return tea.Quit()
			}

		case "enter":
			if m.state == stateWelcome {
				m.state = statePairing
				m.statusMessage = "Menghubungkan..."
				m.qrStatus = ""
				appendCmd(m.startPairing())
			}
		}
	}

	switch m.state {
	case stateChats:
		m.roomList, cmd = m.roomList.Update(msg)
		appendCmd(cmd)
	default:
		m.loading, cmd = m.loading.Update(msg)
		appendCmd(cmd)
	}

	var progressModel tea.Model
	progressModel, cmd = m.syncProgress.Update(msg)
	if prog, ok := progressModel.(progress.Model); ok {
		m.syncProgress = prog
	}
	appendCmd(cmd)

	return m, tea.Batch(cmds...)
}
