package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderLogsContent(width, height int) string {
	container, ok := m.selectedLogsContainer()
	if !ok {
		return m.styles.errorText.Render("Selected container is no longer available.")
	}

	breadcrumb := clampSingleLine("Containers / "+container.Name+" / Logs", max(1, width-2))
	pageOuterWidth := m.logsPageOuterWidth(width)
	pageContentWidth := m.logsPageContentWidth(width)
	safeContentWidth := max(1, pageContentWidth-2)
	pageOuterHeight := m.logsPageOuterHeight(max(1, height))
	pageContentHeight := m.logsPageContentHeight(max(1, height))
	logsHeight := max(1, pageContentHeight-3)
	logs := m.logs.data.Logs
	maxOffset := max(0, len(logs)-logsHeight)
	start := clamp(m.logs.scrollY, 0, maxOffset)
	end := min(len(logs), start+logsHeight)
	headline := m.renderLogsHeader(safeContentWidth, breadcrumb, len(logs), start, end)
	logsPanel := m.renderLogsPanel(safeContentWidth, logsHeight)
	return m.styles.subpageFrame.
		Width(max(1, pageContentWidth)).
		Height(max(1, pageContentHeight)).
		MaxWidth(max(1, pageOuterWidth)).
		MaxHeight(max(1, pageOuterHeight)).
		Render(joinSections(headline, m.renderDivider(safeContentWidth), logsPanel))
}

func (m model) renderLogsHeader(width int, breadcrumb string, total, start, end int) string {
	follow := "off"
	if m.logs.follow {
		follow = "on"
	}
	first := 0
	last := 0
	if total > 0 {
		first = start + 1
		last = max(first, end)
	}
	left := m.styles.breadcrumb.Render(breadcrumb)
	followText := m.styles.followOff.Render(follow)
	if m.logs.follow {
		followText = m.styles.followOn.Render(follow)
	}
	right := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.styles.muted.Render("follow:"),
		followText,
		m.styles.muted.Render(fmt.Sprintf("  lines:(%d-%d/%d)", first, last, total)),
	)
	return m.renderRightPriorityLine(left, right, width)
}

func (m model) renderLogLoadingLine(width int) string {
	return clampSingleLine(m.styles.muted.Render("⟳ Loading logs..."), width)
}

func (m model) renderLogHistoryLoadingLine(width int) string {
	return clampSingleLine(m.styles.muted.Render("⇡ Loading older logs..."), width)
}

func (m model) renderLogsPanel(width, height int) string {
	contentWidth := max(1, width-2)
	if m.logs.initialLoad {
		lines := []string{m.renderLogLoadingLine(contentWidth)}
		for len(lines) < height {
			lines = append(lines, "")
		}
		return strings.Join(clipLines(lines, height), "\n")
	}

	logs := m.logs.data.Logs
	if len(logs) == 0 {
		lines := []string{clampSingleLine(m.styles.muted.Render("No logs found for this container."), contentWidth)}
		for len(lines) < height {
			lines = append(lines, "")
		}
		return strings.Join(clipLines(lines, height), "\n")
	}

	available := max(1, height)
	maxOffset := max(0, len(logs)-available)
	start := clamp(m.logs.scrollY, 0, maxOffset)
	end := min(len(logs), start+available)
	trimmed := make([]string, 0, height)
	if m.logs.historyLoad {
		trimmed = append(trimmed, m.renderLogHistoryLoadingLine(contentWidth))
	}
	for _, line := range logs[start:end] {
		if len(trimmed) >= height {
			break
		}
		window := logViewportLine(line, m.logs.scrollX, contentWidth)
		trimmed = append(trimmed, clipLogDisplayLine(window, contentWidth))
	}
	for len(trimmed) < height {
		trimmed = append(trimmed, "")
	}
	return strings.Join(clipLines(trimmed, height), "\n")
}

func (m model) logsPageOuterWidth(width int) int {
	return max(1, width)
}

func (m model) logsPageContentWidth(width int) int {
	return max(1, m.logsPageOuterWidth(width)-m.styles.subpageFrame.GetHorizontalFrameSize())
}

func (m model) logsPageOuterHeight(height int) int {
	return max(1, height)
}

func (m model) logsPageContentHeight(height int) int {
	return max(1, m.logsPageOuterHeight(height)-m.styles.subpageFrame.GetVerticalFrameSize())
}
