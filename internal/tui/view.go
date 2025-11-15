package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdp/qrterminal/v3"
)

var (
	appFrameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

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
	var content string

	switch m.state {
	case stateLoading, stateConnecting:
		content = m.loadingStatusView()

	case stateError:
		content = m.errorView()

	case stateWelcome:
		content = m.welcomeView()

	case statePairing:
		content = m.pairingView()

	case stateHistorySync:
		content = m.historySyncView()

	case stateChats:
		content = m.chatLayout()

	default:
		content = m.loadingStatusView()
	}

	innerWidth := m.contentWidth()
	innerHeight := m.contentHeight()

	if m.state != stateChats {
		content = lipgloss.Place(
			innerWidth,
			innerHeight,
			lipgloss.Center,
			lipgloss.Center,
			content,
		)
	}

	var sections []string
	if logs := m.devLogView(innerWidth); logs != "" {
		sections = append(sections, logs)
	}

	sections = append(sections, content)

	if overlay := m.syncOverlayView(innerWidth); overlay != "" {
		sections = append(sections, overlay)
	}

	final := lipgloss.JoinVertical(lipgloss.Top, sections...)

	frame := appFrameStyle
	if m.width > 0 && m.height > 0 {
		frame = frame.Width(m.width).Height(m.height)
	}

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

func (m model) chatLayout() string {
	contentWidth := m.contentWidth()
	leftWidth, rightWidth := m.computePaneWidths(contentWidth)

	left := leftPaneStyle.Width(leftWidth).Render(m.roomList.View())
	right := rightPaneStyle.Width(rightWidth).Render(m.chatPane(rightWidth))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
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

func (m model) chatPane(width int) string {
	room := m.roomList.OpenedRoom()
	if room == nil {
		h := m.contentHeight()
		if h < 8 {
			h = 8
		}
		return placeholderStyle.
			Width(width).
			Height(h).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render("watui")
	}

	header := titleStyle.Render(room.Title)
	timeInfo := subtleStyle.Render(room.Time.Format(time.RFC822))
	body := subtleStyle.Render("Belum ada percakapan yang dimuat untuk room ini.")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("%s  %s", header, timeInfo),
		"",
		body,
	)
}

func (m model) contentWidth() int {
	if m.width <= 0 {
		return 80
	}

	w := m.width - 6
	if w < 40 {
		return 40
	}
	return w
}

func (m model) contentHeight() int {
	if m.height <= 0 {
		return 24
	}

	h := m.height - 4
	if h < 12 {
		return 12
	}
	return h
}

func (m model) devLogView(width int) string {
	if !m.devMode || len(m.devLogs) == 0 {
		return ""
	}

	var b strings.Builder
	for _, item := range m.devLogs {
		b.WriteString("• ")
		b.WriteString(item)
		b.WriteString("\n")
	}

	box := logOverlayStyle.Render(strings.TrimSuffix(b.String(), "\n"))
	return lipgloss.PlaceHorizontal(width, lipgloss.Right, box)
}

func (m model) syncOverlayView(width int) string {
	if !m.syncOverlay.active {
		return ""
	}

	label := m.syncOverlay.label
	if label == "" {
		label = "Sinkronisasi data"
	}

	box := syncOverlayStyle.Render(
		label + "\n" + m.syncProgress.View(),
	)

	return lipgloss.PlaceHorizontal(width, lipgloss.Right, box)
}
