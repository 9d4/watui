package roomlist

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

type Room struct {
	Title       string
	LastMessage string
	Time        time.Time
}

type Model struct {
	rooms             []Room
	cursor            int
	selectedRoomIndex *int
	openedRoomIndex   *int

	selectedItemColor lipgloss.AdaptiveColor
	inactiveItemColor lipgloss.AdaptiveColor
	openedItemColor   lipgloss.AdaptiveColor
}

func New() Model {
	m := Model{
		rooms: []Room{
			{Title: "John", Time: time.Now()},
			{Title: "Foo", Time: time.Now().Add(time.Hour)},
			{Title: "C", Time: time.Now().Add(2 * time.Hour)},
		},
		selectedItemColor: lipgloss.AdaptiveColor{Light: "212", Dark: "212"},
		inactiveItemColor: lipgloss.AdaptiveColor{Light: "243", Dark: "243"},
		openedItemColor:   lipgloss.AdaptiveColor{Light: "86", Dark: "86"},
	}

	return m
}
