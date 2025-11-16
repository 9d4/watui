package tui

import (
	"context"
	"fmt"
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
	case roomsLoadedMsg:
		if len(msg.rooms) > 0 {
			m.roomList = m.roomList.ReplaceRooms(msg.rooms)
			for _, room := range msg.rooms {
				m.chatTitles[room.ID] = room.Title
				if room.LastMessage != "" && m.chatSummaries[room.ID] == nil {
					line := formatMessageLine(room.Time, "Terakhir", room.LastMessage)
					m.chatSummaries[room.ID] = []string{line}
				}
			}
		}

		if msg.sync.Progress > 0 {
			appendCmd(m.syncProgress.SetPercent(float64(msg.sync.Progress) / 100))
		}

		if msg.sync.InProgress {
			m.historyReady = false
			if m.state == stateLoading {
				m.state = stateHistorySync
			}
			m.historyMessage = fmt.Sprintf("Melanjutkan sinkronisasi · %d%%", msg.sync.Progress)
		} else if msg.sync.Progress >= 100 {
			m.historyReady = true
			if m.state == stateLoading {
				m.state = stateChats
			}
		}

		if m.state == stateChats && msg.sync.InProgress {
			m.syncOverlay = syncOverlayState{
				active: true,
				label:  fmt.Sprintf("Sinkronisasi lanjutan · %d%%", msg.sync.Progress),
			}
		}

	case contactsLoadedMsg:
		for jid, name := range msg.names {
			m.applyContactName(jid, name)
		}

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

		appendCmd(m.loadContacts())
		appendCmd(m.syncContactsAppState())

		if m.cli.Store.ID == nil {
			m.state = stateWelcome
			m.statusMessage = "Tekan Enter untuk mulai pairing"
			m.historyReady = false
		} else {
			m.state = stateConnecting
			m.statusMessage = "Menghubungkan ke WhatsApp..."
			m.historyReady = true
			appendCmd(m.connectClient())
		}

	case qrCodeMsg:
		m.waQRCode = msg.Code
		m.state = statePairing
		m.qrStatus = ""
		m.historyReady = false
		m.syncOverlay = syncOverlayState{}
		appendCmd(m.waitEvents())

	case qrStatusMsg:
		appendCmd(m.waitEvents())

		switch msg.Status {
		case whatsmeow.QRChannelSuccess.Event:
			m.qrStatus = "QR berhasil dipindai, menunggu sinkronisasi"
			m.waQRCode = ""
			m.state = stateHistorySync
			m.historyMessage = "Menunggu history dari WhatsApp..."
			m.historyReady = false
			m.syncOverlay = syncOverlayState{}
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

		if m.state == stateWelcome {
			m.syncOverlay = syncOverlayState{}
		}

	case waEvent:
		appendCmd(m.waitEvents())

		switch evt := msg.evt.(type) {
		case *events.Connected:
			if m.historyReady {
				m.state = stateChats
				m.statusMessage = ""
			} else {
				m.state = stateHistorySync
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
				rooms := m.applyHistoryRooms(evt.Data)

				progress := float64(evt.Data.GetProgress()) / 100
				if progress > 1 {
					progress = 1
				}

				appendCmd(m.syncProgress.SetPercent(progress))
				appendCmd(m.persistHistory(evt.Data, rooms))

				var syncLabel string
				if evt.Data.SyncType != nil {
					syncLabel = strings.ToLower(evt.Data.GetSyncType().String())
				} else {
					syncLabel = "history"
				}

				info := fmt.Sprintf(
					"Sinkronisasi %s · %d chat",
					syncLabel,
					len(evt.Data.GetConversations()),
				)

				if m.historyReady || m.state == stateChats {
					if progress >= 1 {
						m.syncOverlay = syncOverlayState{}
					} else {
						m.syncOverlay = syncOverlayState{
							active: true,
							label:  info,
						}
					}
				} else {
					m.historyMessage = info
					m.state = stateHistorySync
				}

				if progress >= 1 {
					m.historyReady = true
					m.state = stateChats
					m.statusMessage = ""
					m.syncOverlay = syncOverlayState{}
				}
			}

		case *events.Message:
			if room := m.roomFromMessage(evt); room != nil {
				m.roomList = m.roomList.UpsertRoom(*room)
				m.chatTitles[room.ID] = room.Title
				appendCmd(m.persistRoom(*room))
			}

			m.pushDevLog(fmt.Sprintf(
				"%s ← %s (%s)",
				evt.Info.Chat.String(),
				evt.Info.Sender.String(),
				evt.Info.Type,
			))

		case *events.Contact:
			if evt.Action != nil {
				info := types.ContactInfo{
					FirstName: evt.Action.GetFirstName(),
					FullName:  evt.Action.GetFullName(),
				}
				name := resolveContactName(info, evt.JID.String())
				m.applyContactName(evt.JID.String(), name)
			}

		case *events.PushName:
			name := strings.TrimSpace(evt.NewPushName)
			if name != "" {
				m.applyContactName(evt.JID.String(), name)
			}
		}

	case errMsg:
		m.state = stateError
		m.statusMessage = msg.Error()

	case tea.FocusMsg:
		if m.cli != nil && m.state == stateChats {
			m.cli.SendPresence(types.PresenceAvailable)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width - 2
		m.height = msg.Height - 2
		m.roomList = m.roomList.SetViewportHeight(m.contentHeight())

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
				m.historyReady = false
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
