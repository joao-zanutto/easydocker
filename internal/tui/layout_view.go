package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m model) renderDivider(width int) string {
	return m.styles.divider.Render(strings.Repeat("─", max(1, width-2)))
}

func (m model) detailLine(label, value string) string {
	return m.styles.label.Render(label+": ") + m.styles.value.Render(value)
}

func (m model) renderEdgeAlignedLine(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	if strings.TrimSpace(right) == "" {
		return constrainLine(left, width)
	}
	leftWidth := ansi.StringWidth(stripANSI(left))
	rightWidth := ansi.StringWidth(stripANSI(right))
	if leftWidth+rightWidth+1 > width {
		return constrainLine(left+" "+right, width)
	}
	return left + strings.Repeat(" ", width-leftWidth-rightWidth) + right
}

func (m model) renderCenteredLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	line = constrainLine(line, width)
	lineWidth := ansi.StringWidth(stripANSI(line))
	if lineWidth >= width {
		return line
	}
	leftPad := (width - lineWidth) / 2
	return strings.Repeat(" ", leftPad) + line
}

func clampSingleLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	flat := strings.ReplaceAll(line, "\n", " ")
	flat = strings.ReplaceAll(flat, "\r", " ")
	if ansi.StringWidth(stripANSI(flat)) <= width {
		return flat
	}
	return truncateANSI(flat, width)
}

func (m model) renderPinnedHeaderLine(leftStyle lipgloss.Style, leftText, right string, width int) string {
	if width <= 0 {
		return ""
	}
	if strings.TrimSpace(right) == "" {
		contentWidth := max(1, width-leftStyle.GetHorizontalFrameSize())
		return leftStyle.Render(constrainLine(leftText, contentWidth))
	}
	rightWidth := ansi.StringWidth(stripANSI(right))
	leftTotalWidth := max(1, width-rightWidth-1)
	leftContentWidth := max(1, leftTotalWidth-leftStyle.GetHorizontalFrameSize())
	left := leftStyle.Render(constrainLine(leftText, leftContentWidth))
	leftWidth := ansi.StringWidth(stripANSI(left))
	spacing := max(1, width-leftWidth-rightWidth)
	return left + strings.Repeat(" ", spacing) + right
}

func (m model) renderRightPriorityLine(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	right = clampSingleLine(right, width)
	rightWidth := ansi.StringWidth(stripANSI(right))
	if rightWidth >= width {
		return right
	}
	leftWidth := max(1, width-rightWidth-1)
	left = clampSingleLine(left, leftWidth)
	leftRenderedWidth := ansi.StringWidth(stripANSI(left))
	spacing := max(1, width-leftRenderedWidth-rightWidth)
	return left + strings.Repeat(" ", spacing) + right
}

func (m model) renderLogPosition(width, start, end, total int) string {
	position := fmt.Sprintf("lines %d-%d/%d", start+1, max(start+1, end), total)
	return constrainLine(m.styles.muted.Render(position), width)
}

func (m model) renderLogStatus(width, total, start, rows int) string {
	follow := "off"
	if m.logs.follow {
		follow = "on"
	}
	line := fmt.Sprintf("lines:%d  y:%d  rows:%d  x:%d  follow:%s", total, start+1, rows, m.logs.scrollX, follow)
	return constrainLine(m.styles.muted.Render(line), width)
}
