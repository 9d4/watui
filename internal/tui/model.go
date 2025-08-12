package tui

import "github.com/9d4/watui/roomlist"

type sessionState int

const (
	stateIdle sessionState = iota
)

type model struct {
	state    sessionState
	roomList roomlist.Model
}

func New() model {
	return model{
		roomList: roomlist.New(),
	}
}
