package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/9d4/watui/chatstore"
	"github.com/9d4/watui/roomlist"
	tea "github.com/charmbracelet/bubbletea"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	waHistorySync "go.mau.fi/whatsmeow/proto/waHistorySync"
	waWeb "go.mau.fi/whatsmeow/proto/waWeb"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (m *model) applyHistoryRooms(data *waHistorySync.HistorySync) []roomlist.Room {
	if data == nil {
		return nil
	}

	pushnames := make(map[string]string)
	for _, pn := range data.GetPushnames() {
		if pn == nil {
			continue
		}
		if id := pn.GetID(); id != "" {
			pushnames[id] = pn.GetPushname()
		}
	}

	var rooms []roomlist.Room
	for _, conv := range data.GetConversations() {
		room := m.roomFromConversation(conv, pushnames)
		if room == nil {
			continue
		}

		rooms = append(rooms, *room)
		m.chatTitles[room.ID] = room.Title
		m.roomList = m.roomList.UpsertRoom(*room)
	}

	return rooms
}

func (m *model) roomFromConversation(conv *waHistorySync.Conversation, pushnames map[string]string) *roomlist.Room {
	if conv == nil {
		return nil
	}

	jid := conv.GetID()
	if jid == "" {
		return nil
	}

	parsed, err := types.ParseJID(jid)
	if err != nil {
		return nil
	}

	title := conv.GetName()
	if title == "" {
		title = pushnames[jid]
	}
	if title == "" {
		title = parsed.String()
	}

	lastMessage := conversationSummary(conv)
	lastTs := conv.GetLastMsgTimestamp()
	var ts time.Time
	if lastTs > 0 {
		ts = time.Unix(int64(lastTs), 0)
	}

	return &roomlist.Room{
		ID:          parsed.String(),
		Title:       title,
		LastMessage: lastMessage,
		Time:        ts,
		UnreadCount: int(conv.GetUnreadCount()),
	}
}

func (m *model) roomFromMessage(evt *events.Message) *roomlist.Room {
	if evt == nil {
		return nil
	}

	jid := evt.Info.Chat
	if jid.User == "" && jid.Server == "" {
		return nil
	}

	ts := evt.Info.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	title := m.chatTitles[jid.String()]
	if title == "" {
		title = jid.String()
	}

	summary := summarizeMessage(evt.Message)
	if summary == "" {
		summary = fmt.Sprintf("Pesan %s", evt.Info.Type)
	}

	room := roomlist.Room{
		ID:          jid.String(),
		Title:       title,
		LastMessage: summary,
		Time:        ts,
	}

	return &room
}

func (m model) persistHistory(data *waHistorySync.HistorySync, rooms []roomlist.Room) tea.Cmd {
	if m.store == nil || data == nil {
		return nil
	}

	progress := int(data.GetProgress())
	chunk := int(data.GetChunkOrder())
	syncType := ""
	if data.SyncType != nil {
		syncType = data.GetSyncType().String()
	}

	state := chatstore.SyncState{
		Progress:   progress,
		ChunkOrder: chunk,
		SyncType:   syncType,
	}

	return func() tea.Msg {
		ctx := context.Background()
		if err := m.store.PersistHistory(ctx, rooms, state); err != nil {
			return errMsg{err: fmt.Errorf("gagal menyimpan history: %w", err)}
		}
		return nil
	}
}

func (m model) persistRoom(room roomlist.Room) tea.Cmd {
	if m.store == nil || room.ID == "" {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		if err := m.store.UpsertRoom(ctx, room); err != nil {
			return errMsg{err: fmt.Errorf("gagal menyimpan chat: %w", err)}
		}
		return nil
	}
}

func conversationSummary(conv *waHistorySync.Conversation) string {
	msgs := conv.GetMessages()
	for i := len(msgs) - 1; i >= 0; i-- {
		if summary := summarizeWebMessage(msgs[i].GetMessage()); summary != "" {
			return summary
		}
	}

	return "-"
}

func summarizeWebMessage(info *waWeb.WebMessageInfo) string {
	if info == nil {
		return ""
	}
	return summarizeMessage(info.GetMessage())
}

func summarizeMessage(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}

	switch {
	case msg.GetConversation() != "":
		return msg.GetConversation()
	case msg.GetExtendedTextMessage() != nil:
		return msg.GetExtendedTextMessage().GetText()
	case msg.GetImageMessage() != nil:
		return "📷 Foto"
	case msg.GetVideoMessage() != nil:
		return "🎥 Video"
	case msg.GetAudioMessage() != nil:
		return "🎵 Audio"
	case msg.GetDocumentMessage() != nil:
		return fmt.Sprintf("📄 %s", msg.GetDocumentMessage().GetTitle())
	case msg.GetButtonsMessage() != nil:
		return msg.GetButtonsMessage().GetContentText()
	case msg.GetButtonsResponseMessage() != nil:
		return msg.GetButtonsResponseMessage().GetSelectedDisplayText()
	case msg.GetListResponseMessage() != nil:
		return msg.GetListResponseMessage().GetTitle()
	case msg.GetStickerMessage() != nil:
		return "💠 Stiker"
	case msg.GetContactMessage() != nil:
		return msg.GetContactMessage().GetDisplayName()
	case msg.GetLocationMessage() != nil:
		loc := msg.GetLocationMessage()
		return fmt.Sprintf("📍 Lokasi %.3f, %.3f", loc.GetDegreesLatitude(), loc.GetDegreesLongitude())
	case msg.GetLiveLocationMessage() != nil:
		return "📍 Lokasi realtime"
	case msg.GetTemplateButtonReplyMessage() != nil:
		return msg.GetTemplateButtonReplyMessage().GetSelectedDisplayText()
	case msg.GetInteractiveResponseMessage() != nil:
		return msg.GetInteractiveResponseMessage().GetNativeFlowResponseMessage().GetName()
	default:
		return "Pesan baru"
	}
}
