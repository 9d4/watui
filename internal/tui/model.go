package tui

import (
	"github.com/9d4/watui/roomlist"
	"github.com/9d4/watui/wa"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"go.mau.fi/whatsmeow"
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
	events   chan any
	waQRCode string

	qrStatus string

	width        int
	height       int
	devMode      bool
	devLogs      []string
	historyReady bool
	syncOverlay  syncOverlayState
	chatTitles   map[string]string

	cli *whatsmeow.Client
}

type syncOverlayState struct {
	active bool
	label  string
}

func New(wa *wa.Manager, devMode bool) model {
	return model{
		state:         stateLoading,
		loading:       spinner.New(spinner.WithSpinner(spinner.Dot)),
		syncProgress:  progress.New(progress.WithDefaultGradient()),
		roomList:      roomlist.New(),
		statusMessage: "Menyiapkan WhatsApp session...",
		devMode:       devMode,
		wa:            wa,
		events:        make(chan any),
		chatTitles:    make(map[string]string),
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
