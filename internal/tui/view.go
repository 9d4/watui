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

	frame := appFrameStyle
	if m.width > 0 && m.height > 0 {
		frame = frame.Width(m.width).Height(m.height)
	}

	return frame.Render(content)
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
	leftWidth, rightWidth := m.computePaneWidths()

	left := leftPaneStyle.Width(leftWidth).Render(m.roomList.View())
	right := rightPaneStyle.Width(rightWidth).Render(m.chatPane(rightWidth))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) computePaneWidths() (int, int) {
	if m.width <= 0 {
		return 32, 48
	}

	left := m.width / 3
	if left < 24 {
		left = 24
	}

	right := m.width - left - 6
	if right < 24 {
		right = 24
	}

	return left, right
}

func (m model) chatPane(width int) string {
	room := m.roomList.OpenedRoom()
	if room == nil {
		h := m.height - 6
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
