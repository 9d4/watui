package tui

import (
	"github.com/9d4/watui/roomlist"
	"github.com/9d4/watui/wa"
	"github.com/charmbracelet/bubbles/spinner"
	"go.mau.fi/whatsmeow"
)

type sessionState int

const (
	stateInit sessionState = iota
	stateLogin
	stateIdle
)

type model struct {
	state    sessionState
	roomList roomlist.Model

	loading     spinner.Model
	loadingText string

	wa       *wa.Manager
	events   chan any
	waQRCode string

	cli *whatsmeow.Client
}

func New(wa *wa.Manager) model {
	return model{
		loading:     spinner.New(spinner.WithSpinner(spinner.Dot)),
		loadingText: "Loading",
		roomList:    roomlist.New(),

		wa:     wa,
		events: make(chan any),
	}
}

// Event wrapper from whatsmeow
type waEvent struct {
	evt any
}

type pairQrMsg struct {
	Code string
}

type loggedInMsg struct {
	cli *whatsmeow.Client
}
