package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/9d4/watui/chatstore"
	"github.com/9d4/watui/roomlist"
	"github.com/9d4/watui/wa"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
)

type sessionState int

const (
	stateLoading sessionState = iota
	stateWelcome
	statePairing
	stateHistorySync
	stateConnecting
	stateChats
	stateError
)

type model struct {
	state    sessionState
	roomList roomlist.Model

	loading        spinner.Model
	syncProgress   progress.Model
	statusMessage  string
	historyMessage string

	wa       *wa.Manager
	store    *chatstore.Store
	events   chan any
	waQRCode string

	qrStatus string

	width         int
	height        int
	devMode       bool
	devLogs       []string
	historyReady  bool
	syncOverlay   syncOverlayState
	chatTitles    map[string]string
	chatSummaries map[string][]string
	contactNames  map[string]string

	cli *whatsmeow.Client
}

type syncOverlayState struct {
	active bool
	label  string
}

func New(wa *wa.Manager, store *chatstore.Store, devMode bool) model {
	return model{
		state:         stateLoading,
		loading:       spinner.New(spinner.WithSpinner(spinner.Dot)),
		syncProgress:  progress.New(progress.WithDefaultGradient()),
		roomList:      roomlist.New(),
		statusMessage: "Menyiapkan WhatsApp session...",
		devMode:       devMode,
		wa:            wa,
		store:         store,
		events:        make(chan any),
		chatTitles:    make(map[string]string),
		chatSummaries: make(map[string][]string),
		contactNames:  make(map[string]string),
	}
}

// Event wrapper from whatsmeow
type waEvent struct {
	evt any
}

type clientReadyMsg struct {
	cli *whatsmeow.Client
}

type qrCodeMsg struct {
	Code string
}

type qrStatusMsg struct {
	Status string
	Err    error
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (m *model) pushDevLog(entry string) {
	if !m.devMode || entry == "" {
		return
	}

	const maxLogs = 5
	m.devLogs = append(m.devLogs, entry)
	if len(m.devLogs) > maxLogs {
		m.devLogs = m.devLogs[len(m.devLogs)-maxLogs:]
	}
}

type roomsLoadedMsg struct {
	rooms []roomlist.Room
	sync  chatstore.SyncState
}

type contactsLoadedMsg struct {
	names map[string]string
}

func (m model) loadStoredRooms() tea.Cmd {
	if m.store == nil {
		return nil
	}

	return func() tea.Msg {
		rooms, syncState, err := m.store.LoadAll(context.Background())
		if err != nil {
			return errMsg{err: fmt.Errorf("gagal memuat chat: %w", err)}
		}
		return roomsLoadedMsg{rooms: rooms, sync: syncState}
	}
}

func (m model) loadContacts() tea.Cmd {
	if m.cli == nil || m.cli.Store == nil || m.cli.Store.Contacts == nil {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		contacts, err := m.cli.Store.Contacts.GetAllContacts(ctx)
		if err != nil {
			return errMsg{err: fmt.Errorf("gagal memuat kontak: %w", err)}
		}

		names := make(map[string]string, len(contacts))
		for jid, info := range contacts {
			if name := resolveContactName(info, jid.String()); name != "" {
				names[jid.String()] = name
			}
		}

		return contactsLoadedMsg{names: names}
	}
}

func (m model) syncContactsAppState() tea.Cmd {
	if m.cli == nil {
		return nil
	}

	return func() tea.Msg {
		err := m.cli.FetchAppState(context.Background(), appstate.WAPatchCriticalUnblockLow, false, true)
		if err != nil {
			return errMsg{err: fmt.Errorf("gagal sync kontak: %w", err)}
		}
		return m.loadContacts()()
	}
}

func resolveContactName(info types.ContactInfo, fallback string) string {
	switch {
	case info.FullName != "":
		return strings.TrimSpace(info.FullName)
	case info.FirstName != "":
		return strings.TrimSpace(info.FirstName)
	case info.BusinessName != "":
		return strings.TrimSpace(info.BusinessName)
	case info.PushName != "":
		return strings.TrimSpace(info.PushName)
	default:
		return fallback
	}
}

func (m *model) applyContactName(jid, name string) {
	if name == "" {
		return
	}
	m.contactNames[jid] = name
	m.chatTitles[jid] = name
	m.roomList = m.roomList.UpdateTitle(jid, name)
}

func (m *model) resolveTitle(jid, current string) string {
	name := strings.TrimSpace(current)
	switch {
	case name != "" && name != jid:
		return name
	case m.contactNames[jid] != "":
		return m.contactNames[jid]
	case m.chatTitles[jid] != "":
		return m.chatTitles[jid]
	default:
		return jid
	}
}
