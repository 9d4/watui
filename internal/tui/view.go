package tui

import (
	"fmt"
	"strings"

	"github.com/9d4/watui/roomlist"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdp/qrterminal/v3"
)

var (
	appFrameStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("38"))

	leftPaneStyle = lipgloss.NewStyle().
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true)

	rightPaneStyle = lipgloss.NewStyle().
			Padding(0, 2)

	placeholderStyle = lipgloss.NewStyle().
				Bold(true)

	titleStyle  = lipgloss.NewStyle().Bold(true)
	subtleStyle = lipgloss.NewStyle().Faint(true)
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("10"))
	logOverlayStyle = lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			Foreground(lipgloss.Color("253")).
			Background(lipgloss.Color("60"))
	syncOverlayStyle = lipgloss.NewStyle().
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				Foreground(lipgloss.Color("229"))
)

func (m model) View() string {
	innerWidth, innerHeight := m.innerSize()
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 3 {
		innerHeight = 3
	}

	var baseContent string

	switch m.state {
	case stateLoading, stateConnecting:
		baseContent = m.loadingStatusView()

	case stateError:
		baseContent = m.errorView()

	case stateWelcome:
		baseContent = m.welcomeView()

	case statePairing:
		baseContent = m.pairingView()

	case stateHistorySync:
		baseContent = m.historySyncView()

	default:
		baseContent = m.loadingStatusView()
	}

	logBox := m.devLogBox()
	logHeight := lipgloss.Height(logBox)
	syncBox := m.syncOverlayBox()
	syncHeight := lipgloss.Height(syncBox)

	if logHeight > innerHeight {
		logHeight = innerHeight / 3
	}

	if syncHeight > innerHeight-logHeight {
		syncHeight = (innerHeight - logHeight) / 3
	}

	mainHeight := innerHeight - logHeight - syncHeight
	if mainHeight < 3 {
		mainHeight = 3
	}

	sections := make([]string, 0, 3)
	if logHeight > 0 {
		sections = append(sections,
			lipgloss.Place(innerWidth, logHeight, lipgloss.Right, lipgloss.Top, logBox),
		)
	}

	mainContent := baseContent
	mainAlignH := lipgloss.Center
	mainAlignV := lipgloss.Center
	if m.state == stateChats {
		mainAlignH = lipgloss.Left
		mainAlignV = lipgloss.Top
		mainContent = m.chatLayout(innerWidth, mainHeight)
	}

	sections = append(sections,
		lipgloss.Place(innerWidth, mainHeight, mainAlignH, mainAlignV, mainContent),
	)

	if syncHeight > 0 {
		sections = append(sections,
			lipgloss.Place(innerWidth, syncHeight, lipgloss.Right, lipgloss.Bottom, syncBox),
		)
	}

	final := lipgloss.JoinVertical(lipgloss.Top, sections...)

	frame := appFrameStyle
	width := m.width
	height := m.height
	frameWidth, frameHeight := appFrameStyle.GetFrameSize()
	if width == 0 {
		width = innerWidth + frameWidth
	}
	if height == 0 {
		height = innerHeight + frameHeight
	}
	frame = frame.Width(width).Height(height)

	return frame.Render(final)
}

func (m model) loadingStatusView() string {
	status := m.statusMessage
	if status == "" {
		status = "Menyiapkan..."
	}

	return fmt.Sprintf("%s %s", m.loading.View(), status)
}

func (m model) errorView() string {
	if m.statusMessage == "" {
		return "Terjadi kesalahan yang tidak diketahui."
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(m.statusMessage)
}

func (m model) welcomeView() string {
	var sections []string

	title := lipgloss.NewStyle().
		Bold(true).
		Height(3).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("watui")
	desc := "Connect WhatsAppmu langsung dari terminal tanpa ribet."
	button := buttonStyle.Render(" Continue ")
	hint := "Tekan Enter untuk melanjutkan."

	sections = append(sections, title, desc, "", button, hint)

	if m.qrStatus != "" {
		sections = append(sections, "", subtleStyle.Render(m.qrStatus))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) pairingView() string {
	if m.waQRCode == "" {
		return m.loadingStatusView()
	}

	var builder strings.Builder
	builder.WriteString(m.loading.View())
	builder.WriteString(" Scan kode di bawah menggunakan WhatsApp\n\n")
	qrterminal.GenerateHalfBlock(m.waQRCode, qrterminal.L, &builder)

	if m.qrStatus != "" {
		builder.WriteString("\n" + subtleStyle.Render(m.qrStatus))
	}

	return builder.String()
}

func (m model) historySyncView() string {
	var builder strings.Builder
	builder.WriteString(m.loading.View())
	builder.WriteString(" Sinkronisasi riwayat\n\n")
	if m.historyMessage != "" {
		builder.WriteString(m.historyMessage)
	} else {
		builder.WriteString("Menunggu data WhatsApp...")
	}
	builder.WriteString("\n\n")
	builder.WriteString(m.syncProgress.View())
	return builder.String()
}

func (m model) chatLayout(width, height int) string {
	leftWidth, rightWidth := m.computePaneWidths(width)

	leftPane := leftPaneStyle.Width(leftWidth).Height(height).Render(m.roomList.View())
	rightPane := rightPaneStyle.Width(rightWidth).Height(height).Render(m.chatPane(rightWidth, height))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (m model) computePaneWidths(total int) (int, int) {
	if total <= 0 {
		return 32, 48
	}

	left := total / 3
	if left < 24 {
		left = 24
	}

	right := total - left
	if right < 32 {
		right = 32
	}

	return left, right
}

func (m model) chatPane(width, height int) string {
	room := m.activeRoom()
	if room == nil {
		h := height
		if h < 6 {
			h = 6
		}
		return placeholderStyle.
			Width(width).
			Height(h).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render("watui")
	}

	timeLabel := "-"
	if !room.Time.IsZero() {
		timeLabel = room.Time.Format("02 Jan 15:04")
	}

	meta := fmt.Sprintf("%s · %s", room.ID, timeLabel)
	unread := "Tidak ada pesan baru"
	if room.UnreadCount > 0 {
		unread = fmt.Sprintf("%d pesan belum dibaca", room.UnreadCount)
	}

	lastMsg := "Belum ada pesan"
	if room.LastMessage != "" {
		lastMsg = room.LastMessage
	}

	history := m.chatSummaries[room.ID]
	var historyBuilder strings.Builder
	if len(history) == 0 {
		historyBuilder.WriteString("Belum ada riwayat pesan.\n")
	} else {
		start := 0
		if len(history) > 10 {
			start = len(history) - 10
		}
		for _, line := range history[start:] {
			historyBuilder.WriteString("• " + line + "\n")
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(room.Title),
		subtleStyle.Render(meta),
		subtleStyle.Render(unread),
		"",
		"Pesan terakhir:",
		lastMsg,
		"",
		"Riwayat terbaru:",
		historyBuilder.String(),
	)
}

func (m model) activeRoom() *roomlist.Room {
	if room := m.roomList.OpenedRoom(); room != nil {
		return room
	}
	return m.roomList.CursorRoom()
}

func (m model) innerSize() (int, int) {
	frameWidth, frameHeight := appFrameStyle.GetFrameSize()

	width := m.width
	height := m.height

	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	width -= frameWidth
	height -= frameHeight

	if width < 40 {
		width = 40
	}
	if height < 12 {
		height = 12
	}

	return width, height
}

func (m model) devLogBox() string {
	if !m.devMode || len(m.devLogs) == 0 {
		return ""
	}

	var b strings.Builder
	for _, item := range m.devLogs {
		b.WriteString("• ")
		b.WriteString(item)
		b.WriteString("\n")
	}

	return logOverlayStyle.Render(strings.TrimSuffix(b.String(), "\n"))
}

func (m model) syncOverlayBox() string {
	if !m.syncOverlay.active {
		return ""
	}

	label := m.syncOverlay.label
	if label == "" {
		label = "Sinkronisasi data"
	}

	return syncOverlayStyle.Render(label + "\n" + m.syncProgress.View())
}
