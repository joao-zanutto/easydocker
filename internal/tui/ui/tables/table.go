package tables

import (
	"strings"

	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

type tableRow []string

type tableColumn struct {
	Title       string
	Width       int
	PinnedRight bool
}

type tableModel struct {
	cols       []tableColumn
	rows       []tableRow
	cursor     int
	styles     Styles
	hideHeader bool

	viewport viewport.Model
}

type tableOption func(*tableModel)

func newTable(opts ...tableOption) tableModel {
	m := tableModel{
		styles:   DefaultStyles(),
		viewport: viewport.New(viewport.WithWidth(0), viewport.WithHeight(20)),
	}
	for _, opt := range opts {
		opt(&m)
	}
	m.updateViewport()
	return m
}

func withColumns(cols []tableColumn) tableOption {
	return func(m *tableModel) {
		m.cols = cols
	}
}

func withRows(rows []tableRow) tableOption {
	return func(m *tableModel) {
		m.rows = rows
	}
}

func withHeight(h int) tableOption {
	return func(m *tableModel) {
		m.viewport.SetHeight(max(1, h))
	}
}

func withHideHeader(hide bool) tableOption {
	return func(m *tableModel) {
		m.hideHeader = hide
	}
}

func withWidth(w int) tableOption {
	return func(m *tableModel) {
		m.viewport.SetWidth(max(1, w))
	}
}

func withStyles(s Styles) tableOption {
	return func(m *tableModel) {
		m.styles = s
	}
}

func (m *tableModel) setCursor(n int) {
	if len(m.rows) == 0 {
		m.cursor = 0
		m.updateViewport()
		return
	}
	m.cursor = util.Clamp(n, 0, len(m.rows)-1)
	m.updateViewport()
}

func (m tableModel) view() string {
	body := m.viewport.View()
	var header string
	if !m.hideHeader && len(m.cols) > 0 {
		header = m.styles.Header.Render(m.headersView())
	}
	if body == "" {
		if header != "" {
			return header
		}
		return ""
	}
	if header != "" {
		return header + "\n" + body
	}
	return body
}

func (m *tableModel) updateViewport() {
	if m.viewport.Width() <= 0 {
		m.viewport.SetWidth(1)
	}
	if m.viewport.Height() <= 0 {
		m.viewport.SetHeight(1)
	}

	if len(m.rows) == 0 {
		m.viewport.SetContent("")
		m.viewport.SetYOffset(0)
		return
	}

	visible := max(1, m.viewport.Height())
	start, end := scrollWindow(len(m.rows), m.cursor, visible)
	renderedRows := make([]string, 0, max(0, end-start))
	for i := start; i < end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}
	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, renderedRows...))
	m.viewport.SetYOffset(0)
}

func (m tableModel) headersView() string {
	leftParts := make([]string, 0, len(m.cols))
	rightParts := make([]string, 0, len(m.cols))
	for _, col := range m.cols {
		if col.Width <= 0 {
			continue
		}
		cell := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true).Render(util.TruncateWithEllipsis(col.Title, col.Width))
		if col.PinnedRight {
			rightParts = append(rightParts, cell)
		} else {
			leftParts = append(leftParts, cell)
		}
	}

	leftLine := joinWithGap(leftParts, "  ")
	if len(rightParts) == 0 {
		return leftLine
	}
	rightLine := joinWithGap(rightParts, "  ")
	padding := max(0, m.viewport.Width()-util.DisplayWidth(leftLine)-util.DisplayWidth(rightLine))
	return leftLine + strings.Repeat(" ", padding) + rightLine
}

func (m tableModel) renderRow(rowIndex int) string {
	leftParts := make([]string, 0, len(m.cols))
	rightParts := make([]string, 0, len(m.cols))
	for colIndex, col := range m.cols {
		if col.Width <= 0 {
			continue
		}
		value := ""
		if colIndex < len(m.rows[rowIndex]) {
			value = m.rows[rowIndex][colIndex]
		}
		cell := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true).Render(util.TruncateWithEllipsis(value, col.Width))
		if col.PinnedRight {
			rightParts = append(rightParts, cell)
		} else {
			leftParts = append(leftParts, cell)
		}
	}

	leftLine := joinWithGap(leftParts, "  ")
	if len(rightParts) == 0 {
		if rowIndex == m.cursor {
			return m.styles.Selected.Render(leftLine)
		}
		return m.styles.Cell.Render(leftLine)
	}
	rightLine := joinWithGap(rightParts, "  ")
	padding := max(0, m.viewport.Width()-util.DisplayWidth(leftLine)-util.DisplayWidth(rightLine))
	line := leftLine + strings.Repeat(" ", padding) + rightLine
	if rowIndex == m.cursor {
		return m.styles.Selected.Render(line)
	}
	return m.styles.Cell.Render(line)
}

func joinWithGap(parts []string, gap string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	out := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			out = append(out, gap)
		}
		out = append(out, part)
	}
	return strings.Join(out, "")
}

func scrollWindow(total, cursor, height int) (int, int) {
	if total <= 0 || height <= 0 {
		return 0, 0
	}
	if height >= total {
		return 0, total
	}
	cursor = util.Clamp(cursor, 0, total-1)
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
	}
	if start < 0 {
		start = 0
	}
	return start, end
}
