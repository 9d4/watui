package tui

import (
	"github.com/9d4/watui/roomlist"
	"github.com/charmbracelet/bubbles/spinner"
)

type sessionState int

const (
	stateLoading sessionState = iota
	stateLogin
	stateIdle
)

type model struct {
	state       sessionState
	roomList    roomlist.Model
	loading     spinner.Model
	loadingText string
	loginEvent  chan any
}

func New() model {
	return model{
		loading:     spinner.New(spinner.WithSpinner(spinner.Dot)),
		loadingText: "Loading",
		roomList:    roomlist.New(),
		loginEvent:  make(chan any),
	}
}

type (
	loadingLogin struct{}
	loggedIn     struct{}
)
