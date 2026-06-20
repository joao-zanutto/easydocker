package chrome

import (
	"strings"
	"testing"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

func TestTotalsText(t *testing.T) {
	snapshot := core.Snapshot{TotalCPU: 12.3, TotalMem: 1024, TotalLimit: 2048}
	got := totalsText(snapshot, shared.StageIdle, true, "", totalsFull)
	if !strings.Contains(got, "CPU") || !strings.Contains(got, "MEM") {
		t.Fatalf("totalsText(full) = %q, want CPU/MEM", got)
	}

	gotCPU := totalsText(snapshot, shared.StageIdle, true, "", totalsCPU)
	if !strings.Contains(gotCPU, "CPU") {
		t.Fatalf("totalsText(cpu) = %q, want CPU", gotCPU)
	}
	if strings.Contains(gotCPU, "MEM") {
		t.Fatalf("totalsText(cpu) = %q, should not contain MEM", gotCPU)
	}

	loading := totalsText(snapshot, shared.StageMetrics, false, "⠋", totalsFull)
	if !strings.Contains(loading, "CPU") || !strings.Contains(loading, "⠋") {
		t.Fatalf("totalsText(loading) = %q, want CPU with spinner", loading)
	}

	stale := totalsText(snapshot, shared.StageMetrics, true, "⠋", totalsFull)
	if strings.Contains(stale, "⠋") {
		t.Fatalf("totalsText(post-metrics) = %q, should not have spinner", stale)
	}
}

type testFooterKeyMap struct {
	bindings []key.Binding
}

func (m testFooterKeyMap) ShortHelp() []key.Binding {
	return m.bindings
}

func (m testFooterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.bindings}
}

func TestScopeLabel(t *testing.T) {
	tests := []struct {
		name    string
		showAll bool
		want    string
	}{
		{name: "all", showAll: true, want: "all"},
		{name: "running", showAll: false, want: "running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScopeLabel(tt.showAll); got != tt.want {
				t.Fatalf("ScopeLabel(%v) = %q, want %q", tt.showAll, got, tt.want)
			}
		})
	}
}

func TestRenderHeaderAndFooter(t *testing.T) {
	header := RenderHeader(HeaderInput{
		Width:            220,
		Title:            "EasyDocker",
		LoadingStageText: "loading resources",
		ActiveTab:        0,
		ShowAll:          true,
		Err:              nil,
		Tabs: []TabSpec{
			{Tab: 0, Icon: "🐳", Name: "Containers", Count: 2},
			{Tab: 1, Icon: "💿", Name: "Images", Count: 2},
		},
		Styles: HeaderStyles{
			Header:    lipgloss.NewStyle(),
			Title:     lipgloss.NewStyle(),
			TitleMeta: lipgloss.NewStyle(),
			Badge:     lipgloss.NewStyle(),
			ErrorText: lipgloss.NewStyle(),
		},
		RenderTab:     func(tab shared.Tab, label string) string { return label },
		Snapshot:      core.Snapshot{TotalCPU: 12.3, TotalMem: 1024, TotalLimit: 2048},
		MetricsLoaded: true,
	})
	if !strings.Contains(util.StripANSI(header), "EasyDocker") {
		t.Fatalf("expected header to contain title, got %q", header)
	}

	footer := RenderFooter(FooterInput{
		Width: 220,
		KeyMap: testFooterKeyMap{bindings: []key.Binding{
			key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
			key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "switch tabs")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle running/all")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "logs")),
		}},
		Styles: FooterStyles{
			Footer:  lipgloss.NewStyle(),
			Key:     lipgloss.NewStyle(),
			KeyText: lipgloss.NewStyle(),
		},
	})
	for _, token := range []string{"navigate", "switch tabs", "toggle running/all", "logs"} {
		if !strings.Contains(util.StripANSI(footer), token) {
			t.Fatalf("expected footer to contain %q, got %q", token, footer)
		}
	}
}
