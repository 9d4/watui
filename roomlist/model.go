package roomlist

import (
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type Room struct {
	ID          string
	Title       string
	LastMessage string
	Time        time.Time
	UnreadCount int
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
	return Model{
		selectedItemColor: lipgloss.AdaptiveColor{Dark: fmt.Sprintf("%d", 0b010101), Light: "212"},
		inactiveItemColor: lipgloss.AdaptiveColor{Light: "243", Dark: "243"},
		openedItemColor:   lipgloss.AdaptiveColor{Light: "86", Dark: "86"},
	}
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

func (m Model) CursorRoom() *Room {
	if len(m.Rooms) == 0 {
		return nil
	}

	if m.cursor < 0 {
		return nil
	}
	if m.cursor >= len(m.Rooms) {
		return nil
	}

	room := m.Rooms[m.cursor]
	return &room
}

func (m Model) ReplaceRooms(rooms []Room) Model {
	slices.SortFunc(rooms, func(a, b Room) int {
		switch {
		case a.Time.After(b.Time):
			return -1
		case a.Time.Before(b.Time):
			return 1
		default:
			return 0
		}
	})

	m.Rooms = rooms
	if len(m.Rooms) == 0 {
		m.cursor = 0
		m.openedRoomIndex = nil
		return m
	}

	if m.cursor >= len(m.Rooms) {
		m.cursor = len(m.Rooms) - 1
	}

	if m.openedRoomIndex != nil && *m.openedRoomIndex >= len(m.Rooms) {
		m.openedRoomIndex = nil
	}

	return m
}

func (m Model) UpsertRoom(room Room) Model {
	found := false
	var openedJID string
	var openedHas bool
	if m.openedRoomIndex != nil && *m.openedRoomIndex < len(m.Rooms) {
		openedJID = m.Rooms[*m.openedRoomIndex].ID
		openedHas = true
	}

	cursorJID := ""
	if len(m.Rooms) > 0 && m.cursor < len(m.Rooms) {
		cursorJID = m.Rooms[m.cursor].ID
	}

	for i := range m.Rooms {
		if m.Rooms[i].ID == room.ID {
			m.Rooms[i] = room
			found = true
			break
		}
	}

	if !found {
		m.Rooms = append(m.Rooms, room)
	}

	m = m.ReplaceRooms(m.Rooms)

	if cursorJID != "" {
		for i, r := range m.Rooms {
			if r.ID == cursorJID {
				m.cursor = i
				break
			}
		}
	}

	if openedHas {
		for i, r := range m.Rooms {
			if r.ID == openedJID {
				idx := i
				m.openedRoomIndex = &idx
				break
			}
		}
	}

	return m
}
