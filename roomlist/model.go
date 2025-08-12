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

	selectedItemColor   lipgloss.AdaptiveColor
	unselectedItemColor lipgloss.AdaptiveColor
	inactiveItemColor   lipgloss.AdaptiveColor
	separatorColor      lipgloss.AdaptiveColor
}

func New() Model {
	m := Model{
		rooms: []Room{
			{Title: "A", LastMessage: "Be good", Time: time.Now()},
			{Title: "B", LastMessage: "Be Nice Be Nice Be Nice Be Nice Be Nice Be Nice Be Nice Be Nice", Time: time.Now().Add(time.Hour)},
			{Title: "C", LastMessage: "Nice boy", Time: time.Now().Add(2 * time.Hour)},
		},
		selectedItemColor:   lipgloss.AdaptiveColor{Light: "212", Dark: "212"},
		unselectedItemColor: lipgloss.AdaptiveColor{Light: "ffffff", Dark: "#000000"},
		inactiveItemColor:   lipgloss.AdaptiveColor{Light: "243", Dark: "243"},
	}

	return m
}
