package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading EasyDocker..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	mainHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if mainHeight < 1 {
		mainHeight = 1
	}
	main := m.renderMain(mainHeight)
	used := lipgloss.Height(header) + lipgloss.Height(main) + lipgloss.Height(footer)
	spacer := ""
	if used < m.height {
		spacer = strings.Repeat("\n", m.height-used)
	}

	return m.styles.page.Render(lipgloss.JoinVertical(lipgloss.Left, header, main, spacer, footer))
}

func (m model) renderMain(height int) string {
	totalWidth := max(1, m.width)
	totalHeight := max(1, height)
	if m.screen == screenModeLogs && m.activeTab == tabContainers {
		return m.renderLogsContent(totalWidth, totalHeight)
	}

	innerWidth := max(1, totalWidth-m.styles.mainFrame.GetHorizontalFrameSize())
	innerHeight := max(1, totalHeight-m.styles.mainFrame.GetVerticalFrameSize())
	content := m.renderBrowseContent(innerWidth, innerHeight)
	return m.styles.mainFrame.Width(innerWidth).Height(innerHeight).MaxWidth(totalWidth).MaxHeight(totalHeight).Render(content)
}
