package util

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestFrameContentWidthAndHeight(t *testing.T) {
	frame := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2, 3, 4)

	if got := FrameContentWidth(30, frame); got != 22 {
		t.Fatalf("FrameContentWidth(30, frame) = %d, want 22", got)
	}
	if got := FrameContentHeight(20, frame); got != 14 {
		t.Fatalf("FrameContentHeight(20, frame) = %d, want 14", got)
	}
	if got := FrameContentWidth(1, frame); got != 1 {
		t.Fatalf("FrameContentWidth(1, frame) = %d, want 1", got)
	}
	if got := FrameContentHeight(1, frame); got != 1 {
		t.Fatalf("FrameContentHeight(1, frame) = %d, want 1", got)
	}
}

func TestMainAreaHeight(t *testing.T) {
	header := "line-1\nline-2"
	footer := "line-1"

	if got := MainAreaHeight(7, header, footer); got != 4 {
		t.Fatalf("MainAreaHeight(7, header, footer) = %d, want 4", got)
	}
	if got := MainAreaHeight(2, header, footer); got != 1 {
		t.Fatalf("MainAreaHeight should clamp to 1, got %d", got)
	}
}

func TestRenderInFrame(t *testing.T) {
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

	rendered := RenderInFrame(frame, "x", 24, 5)
	if got := lipgloss.Width(rendered); got > 24 {
		t.Fatalf("rendered width = %d, want <= %d", got, 24)
	}
	if got := lipgloss.Height(rendered); got > 5 {
		t.Fatalf("rendered height = %d, want <= %d", got, 5)
	}
}

func TestRenderInFrame_ClipsInnerLines(t *testing.T) {
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	content := "left side text that is definitely longer than the frame width"

	rendered := RenderInFrame(frame, content, 20, 4)
	if got := strings.Count(rendered, "\n"); got+1 > 4 {
		t.Fatalf("rendered lines = %d, want <= %d", got+1, 4)
	}
	for _, line := range strings.Split(rendered, "\n") {
		if lipgloss.Width(line) > 20 {
			t.Fatalf("line width = %d, want <= %d", lipgloss.Width(line), 20)
		}
	}
}
