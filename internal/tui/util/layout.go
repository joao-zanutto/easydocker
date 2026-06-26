package util

import (
	"strings"

	"charm.land/lipgloss/v2"
)

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

func FrameContentWidth(total int, frame lipgloss.Style) int {
	return max(1, max(1, total)-frame.GetHorizontalFrameSize())
}

func FrameContentHeight(total int, frame lipgloss.Style) int {
	return max(1, max(1, total)-frame.GetVerticalFrameSize())
}

func MainAreaHeight(totalHeight int, header, footer string) int {
	return max(1, totalHeight-lipgloss.Height(header)-lipgloss.Height(footer))
}

func RenderInFrame(frame lipgloss.Style, content string, outerWidth, outerHeight int) string {
	innerWidth := max(1, outerWidth-frame.GetHorizontalFrameSize())
	contentHeight := FrameContentHeight(outerHeight, frame)
	clampedLines := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		clampedLines = append(clampedLines, ClampSingleLine(line, innerWidth))
	}
	content = strings.Join(clampedLines, "\n")
	return frame.
		Width(outerWidth).
		Height(contentHeight).
		MaxWidth(outerWidth).
		MaxHeight(outerHeight).
		Render(content)
}
