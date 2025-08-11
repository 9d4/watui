package tui

type sessionState int

const (
	stateIdle sessionState = iota
)

type model struct {
	state sessionState
}

func New() model {
	return model{}
}
