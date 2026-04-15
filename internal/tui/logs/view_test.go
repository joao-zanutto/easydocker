package logs

import (
	"strings"
	"testing"

	"easydocker/internal/tui/util"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderPanel_InitialLoadPadsToHeight(t *testing.T) {
	state := NewState()
	state.InitialLoad = true

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 4)
	lines := strings.Split(panel, "\n")
	if len(lines) != 4 {
		t.Fatalf("lines len = %d, want 4", len(lines))
	}
	if !strings.Contains(lines[0], "Loading logs") {
		t.Fatalf("expected loading line, got %q", lines[0])
	}
}

func TestRenderPanel_EmptyLogsPadsToHeight(t *testing.T) {
	state := NewState()
	state.InitialLoad = false
	state.Data.Logs = nil

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 3)
	lines := strings.Split(panel, "\n")
	if len(lines) != 3 {
		t.Fatalf("lines len = %d, want 3", len(lines))
	}
	if !strings.Contains(lines[0], "No logs found") {
		t.Fatalf("expected empty logs hint, got %q", lines[0])
	}
}

func TestRenderPanel_HistoryLoadPrependsIndicator(t *testing.T) {
	state := NewState()
	state.Data.Logs = []string{"line-1", "line-2", "line-3"}
	state.SyncViewport([]string{"line-1", "line-2", "line-3"}, 78, 4)
	state.HistoryLoad = true

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 4)
	lines := strings.Split(panel, "\n")
	if len(lines) != 4 {
		t.Fatalf("lines len = %d, want 4", len(lines))
	}
	if !strings.Contains(lines[0], "Loading older logs") {
		t.Fatalf("expected history loading line first, got %q", lines[0])
	}
}

func TestRenderContent_RendersFrameAndBreadcrumb(t *testing.T) {
	state := NewState()
	state.Data.Logs = []string{"line-1", "line-2"}
	state.SyncViewport([]string{"line-1", "line-2"}, 60, 4)

	view := RenderContent(ViewModel{
		State:         state,
		ContainerName: "api",
		Width:         80,
		Height:        8,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	})

	if view == "" {
		t.Fatalf("RenderContent returned empty string")
	}
	if !strings.Contains(view, "api") || !strings.Contains(view, "Logs") {
		t.Fatalf("expected breadcrumb in view, got %q", view)
	}
}

func TestRenderContent_KeepsLineCountIndicatorOnOneLine(t *testing.T) {
	state := NewState()
	logs := make([]string, 45)
	for i := range logs {
		logs[i] = strings.Repeat("x", 10)
	}
	state.Data.Logs = logs
	state.SyncViewport(logs, 60, 8)

	header := RenderHeader(ViewModel{
		State:         state,
		ContainerName: "server-postgres",
		Width:         76,
		Height:        12,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	}, "Containers / server-postgres / Logs", len(state.Data.Logs), 37, 45)

	view := RenderContent(ViewModel{
		State:         state,
		ContainerName: "server-postgres",
		Width:         80,
		Height:        12,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	})

	if !strings.Contains(view, "lines:(38-45/45)") {
		t.Fatalf("expected intact line-count indicator in view; header width=%d header=%q view=%q", util.DisplayWidth(header), header, view)
	}
	if strings.Contains(view, "lines:(38-45/…") {
		t.Fatalf("line-count indicator was unexpectedly ellipsized: %q", view)
	}
	if strings.Contains(view, "\n│ 45/45)") {
		t.Fatalf("line-count indicator was split across lines: %q", view)
	}
}

func TestVisibleRowsForContent_FilterAdjustsHeight(t *testing.T) {
	base := VisibleRowsForContent(20, false)
	withFilter := VisibleRowsForContent(20, true)
	if withFilter != base-filterHeaderHeight {
		t.Fatalf("VisibleRowsForContent(20,true) = %d, want %d", withFilter, base-filterHeaderHeight)
	}
}

func TestRenderContent_FilterHeaderBelowTitleAndFilteredCounts(t *testing.T) {
	state := NewState()
	state.FilterActive = true
	state.FilterQuery = "match"
	state.FilterInput.SetValue("match")
	state.Data.Logs = []string{"nope", "match one", "match two", "other"}
	state.SyncViewportFromData(60, 6)

	view := RenderContent(ViewModel{
		State:         state,
		ContainerName: "api",
		Width:         80,
		Height:        12,
		Styles: ViewStyles{
			Breadcrumb:   lipgloss.NewStyle().Bold(true),
			FollowOn:     lipgloss.NewStyle().Bold(true),
			FollowOff:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			Divider:      lipgloss.NewStyle(),
			SubpageFrame: lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		},
	})

	if !strings.Contains(view, "Containers / api / Logs") {
		t.Fatalf("missing logs title line: %q", view)
	}
	if !strings.Contains(view, "🔎︎ match") {
		t.Fatalf("missing filter input line under header: %q", view)
	}
	if !hasDividerBetween(view, "Containers / api / Logs", "🔎︎ match") {
		t.Fatalf("expected divider between logs title and filter input, got: %q", view)
	}
	if !strings.Contains(view, "lines:(1-2/2)") {
		t.Fatalf("line counts should use filtered logs total, got: %q", view)
	}
	if strings.Contains(view, "nope") || strings.Contains(view, "other") {
		t.Fatalf("non-matching lines should be hidden, got: %q", view)
	}
}

func TestRenderPanel_FilterNoMatchesMessage(t *testing.T) {
	state := NewState()
	state.InitialLoad = false
	state.FilterQuery = "missing"
	state.Data.Logs = []string{"line-1", "line-2"}
	state.SyncViewportFromData(60, 4)

	panel := RenderPanel(ViewModel{State: state, Styles: ViewStyles{Muted: lipgloss.NewStyle()}}, 80, 3)
	lines := strings.Split(panel, "\n")
	if len(lines) != 3 {
		t.Fatalf("lines len = %d, want 3", len(lines))
	}
	if !strings.Contains(lines[0], "No log lines match current filter.") {
		t.Fatalf("expected filter-empty message, got %q", lines[0])
	}
}

func TestRenderDivider_FillsRequestedWidth(t *testing.T) {
	line := renderDivider(lipgloss.NewStyle(), 24)
	if util.DisplayWidth(line) != 24 {
		t.Fatalf("divider width = %d, want 24", util.DisplayWidth(line))
	}
}

func TestDynamicInputWidth_UsesFullLineMinusPrompt(t *testing.T) {
	if got := dynamicInputWidth("🔎︎ ", 20); got != 17 {
		t.Fatalf("dynamicInputWidth = %d, want 17", got)
	}
	if got := dynamicInputWidth("", 8); got != 8 {
		t.Fatalf("dynamicInputWidth with empty prompt = %d, want 8", got)
	}
}

func hasDividerBetween(view, upperToken, lowerToken string) bool {
	lines := strings.Split(view, "\n")
	upper := -1
	lower := -1
	for i, line := range lines {
		if upper == -1 && strings.Contains(line, upperToken) {
			upper = i
		}
		if strings.Contains(line, lowerToken) {
			lower = i
			break
		}
	}
	if upper == -1 || lower == -1 || lower <= upper+1 {
		return false
	}
	for i := upper + 1; i < lower; i++ {
		if strings.Contains(lines[i], "─") || strings.Contains(lines[i], "━") {
			return true
		}
	}
	return false
}
