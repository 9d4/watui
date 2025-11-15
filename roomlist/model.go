package roomlist

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type Room struct {
	Title       string
	LastMessage string
	Time        time.Time
}

type Model struct {
	Rooms           []Room
	cursor          int
	openedRoomIndex *int

	selectedItemColor lipgloss.AdaptiveColor
	inactiveItemColor lipgloss.AdaptiveColor
	openedItemColor   lipgloss.AdaptiveColor
}

func New() Model {
	m := Model{
		Rooms: []Room{
			{Title: "John", LastMessage: "Nanti malam jadi?", Time: time.Now()},
			{Title: "Foo", LastMessage: "Receipt received ✅", Time: time.Now().Add(-2 * time.Hour)},
			{Title: "C", LastMessage: "Lanjut besok aja deh", Time: time.Now().Add(2 * time.Hour)},
		},
		selectedItemColor: lipgloss.AdaptiveColor{Dark: fmt.Sprintf("%d", 0b010101), Light: "212"},
		inactiveItemColor: lipgloss.AdaptiveColor{Light: "243", Dark: "243"},
		openedItemColor:   lipgloss.AdaptiveColor{Light: "86", Dark: "86"},
	}

	return m
}

func (m Model) OpenedRoom() *Room {
	if m.openedRoomIndex == nil {
		return nil
	}

	idx := *m.openedRoomIndex
	if idx < 0 || idx >= len(m.Rooms) {
		return nil
	}

	room := m.Rooms[idx]
	return &room
}
