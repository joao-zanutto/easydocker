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

func TestComputeFrameLayout(t *testing.T) {
	frame := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2, 3, 4)

	got := ComputeFrameLayout(30, 20, frame)
	if got.OuterWidth != 30 || got.OuterHeight != 20 {
		t.Fatalf("outer size = (%d,%d), want (30,20)", got.OuterWidth, got.OuterHeight)
	}
	if got.ContentWidth != 22 || got.ContentHeight != 14 {
		t.Fatalf("content size = (%d,%d), want (22,14)", got.ContentWidth, got.ContentHeight)
	}

	small := ComputeFrameLayout(0, -5, frame)
	if small.OuterWidth != 1 || small.OuterHeight != 1 {
		t.Fatalf("small outer size = (%d,%d), want (1,1)", small.OuterWidth, small.OuterHeight)
	}
	if small.ContentWidth != 1 || small.ContentHeight != 1 {
		t.Fatalf("small content size = (%d,%d), want (1,1)", small.ContentWidth, small.ContentHeight)
	}
}

func TestRenderFramedContent(t *testing.T) {
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	layout := ComputeFrameLayout(24, 5, frame)

	rendered := RenderFramedContent(frame, layout, "x")
	if got := lipgloss.Width(rendered); got > layout.OuterWidth {
		t.Fatalf("rendered width = %d, want <= %d", got, layout.OuterWidth)
	} else if got < layout.ContentWidth {
		t.Fatalf("rendered width = %d, want >= %d", got, layout.ContentWidth)
	}
	if got := lipgloss.Height(rendered); got > layout.OuterHeight {
		t.Fatalf("rendered height = %d, want <= %d", got, layout.OuterHeight)
	} else if got < layout.ContentHeight {
		t.Fatalf("rendered height = %d, want >= %d", got, layout.ContentHeight)
	}
}

func TestRenderFramedContent_ClipsInnerLines(t *testing.T) {
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	layout := ComputeFrameLayout(20, 4, frame)
	content := "left side text that is definitely longer than the frame width"

	rendered := RenderFramedContent(frame, layout, content)
	if got := strings.Count(rendered, "\n"); got+1 > layout.OuterHeight {
		t.Fatalf("rendered lines = %d, want <= %d", got+1, layout.OuterHeight)
	}
	for _, line := range strings.Split(rendered, "\n") {
		if lipgloss.Width(line) > layout.OuterWidth {
			t.Fatalf("line width = %d, want <= %d", lipgloss.Width(line), layout.OuterWidth)
		}
	}
}
