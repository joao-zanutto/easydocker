package util

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// AllocateColumns distributes total width across desired column widths.
func AllocateColumns(total int, desired []int) []int {
	if len(desired) == 0 {
		return []int{}
	}
	if total <= 0 {
		out := make([]int, len(desired))
		for i := range out {
			out[i] = 1
		}
		return out
	}

	out := make([]int, len(desired))
	sum := 0
	for i, width := range desired {
		if width < 1 {
			width = 1
		}
		out[i] = width
		sum += width
	}
	if sum == total {
		return out
	}
	if sum < total {
		out[len(out)-1] += total - sum
		return out
	}

	over := sum - total
	for over > 0 {
		changed := false
		for i := range out {
			if out[i] > 1 && over > 0 {
				out[i]--
				over--
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return out
}

// FrameLayout describes the dimensions of a framed region.
type FrameLayout struct {
	OuterWidth    int
	OuterHeight   int
	ContentWidth  int
	ContentHeight int
}

// FrameContentWidth returns available width inside a frame after accounting for borders/padding.
func FrameContentWidth(total int, frame lipgloss.Style) int {
	return max(1, max(1, total)-frame.GetHorizontalFrameSize())
}

// FrameContentHeight returns available height inside a frame after accounting for borders/padding.
func FrameContentHeight(total int, frame lipgloss.Style) int {
	return max(1, max(1, total)-frame.GetVerticalFrameSize())
}

// MainAreaHeight returns height available for main content (total - header - footer).
func MainAreaHeight(totalHeight int, header, footer string) int {
	return max(1, totalHeight-lipgloss.Height(header)-lipgloss.Height(footer))
}

// ComputeFrameLayout calculates content dimensions within a frame.
func ComputeFrameLayout(outerWidth, outerHeight int, frame lipgloss.Style) FrameLayout {
	width := max(1, outerWidth)
	height := max(1, outerHeight)
	return FrameLayout{
		OuterWidth:    width,
		OuterHeight:   height,
		ContentWidth:  FrameContentWidth(width, frame),
		ContentHeight: FrameContentHeight(height, frame),
	}
}

// RenderFramedContent wraps content in a frame with calculated dimensions.
func RenderFramedContent(frame lipgloss.Style, layout FrameLayout, content string) string {
	innerWidth := max(1, layout.OuterWidth-frame.GetHorizontalFrameSize())
	clampedLines := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		clampedLines = append(clampedLines, ClampSingleLine(line, innerWidth))
	}
	content = strings.Join(clampedLines, "\n")
	return frame.
		Width(layout.OuterWidth).
		Height(layout.ContentHeight).
		MaxWidth(layout.OuterWidth).
		MaxHeight(layout.OuterHeight).
		Render(content)
}