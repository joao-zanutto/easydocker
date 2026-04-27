package browse

import (
	"strings"
	"testing"

	"easydocker/internal/core"
	"easydocker/internal/tui/util"

	"github.com/charmbracelet/lipgloss"
)

type fakeProvider struct{}

func (fakeProvider) DetailLine(label, value string, width int) string {
	line := label + ": " + value
	return util.ConstrainLine(line, width)
}
func (fakeProvider) RenderContainerState(container core.ContainerRow) string {
	return ContainerStateText(container)
}

func TestShouldRenderLoading(t *testing.T) {
	if !ShouldRenderLoading(true, core.Snapshot{}) {
		t.Fatalf("ShouldRenderLoading(true, empty) = false, want true")
	}
	if ShouldRenderLoading(false, core.Snapshot{}) {
		t.Fatalf("ShouldRenderLoading(false, empty) = true, want false")
	}
	if ShouldRenderLoading(true, core.Snapshot{Containers: []core.ContainerRow{{Name: "api"}}}) {
		t.Fatalf("ShouldRenderLoading(true, with resources) = true, want false")
	}
}

func TestListHeight(t *testing.T) {
	tests := []struct {
		h int
		w int
	}{
		{1, 1},
		{3, 1},
		{5, 3},
		{10, 6},
		{20, 12},
	}
	for _, tc := range tests {
		if got := ListHeight(tc.h); got != tc.w {
			t.Fatalf("ListHeight(%d) = %d, want %d", tc.h, got, tc.w)
		}
	}
}

func TestListHeightForContent_FilterShrinksTableFromTop(t *testing.T) {
	height := 20
	base := ListHeight(height)
	withFilter := ListHeightForContent(height, true)

	if withFilter != base-filterHeaderHeight {
		t.Fatalf("ListHeightForContent(%d, true) = %d, want %d", height, withFilter, base-filterHeaderHeight)
	}
}

func TestRenderFilterHeader_InputAndDivider(t *testing.T) {
	header := RenderFilterHeader("> abc", 24, lipgloss.NewStyle())
	lines := strings.Split(header, "\n")
	if len(lines) != 2 {
		t.Fatalf("filter header lines = %d, want 2", len(lines))
	}
	if util.DisplayWidth(lines[0]) != 24 {
		t.Fatalf("input line width = %d, want 24 (%q)", util.DisplayWidth(lines[0]), lines[0])
	}
	if util.DisplayWidth(lines[1]) != 22 {
		t.Fatalf("divider line width = %d, want 22 (%q)", util.DisplayWidth(lines[1]), lines[1])
	}
	if strings.Contains(lines[0], "┌") || strings.Contains(lines[0], "┐") || strings.Contains(lines[0], "└") || strings.Contains(lines[0], "┘") {
		t.Fatalf("input line should not render a box: %q", lines[0])
	}
}

func TestRenderContent_FilterKeepsDividerAnchored(t *testing.T) {
	vm := ViewModel{
		Loading: false,
		Snapshot: core.Snapshot{
			Containers: []core.ContainerRow{{Name: "api"}},
		},
		ActiveTab: 0,
		Width:     80,
		Height:    20,
		Styles: ViewStyles{
			Divider: lipgloss.NewStyle(),
			Muted:   lipgloss.NewStyle(),
			Section: lipgloss.NewStyle(),
		},
	}

	baseList := strings.Join(util.ClipAndPadLines([]string{"row"}, ListHeightForContent(vm.Height, false), ""), "\n")
	withoutFilter := RenderContent(vm, baseList, fakeProvider{})

	vm.FilterActive = true
	vm.FilterInput = "> abc"
	filterList := strings.Join(util.ClipAndPadLines([]string{"row"}, ListHeightForContent(vm.Height, true), ""), "\n")
	withFilter := RenderContent(vm, filterList, fakeProvider{})

	withoutDetails := lineIndex(withoutFilter, "Details")
	withDetails := lineIndex(withFilter, "Details")
	if withoutDetails == -1 || withDetails == -1 {
		t.Fatalf("could not find details section in rendered output")
	}
	if withoutDetails != withDetails {
		t.Fatalf("details line moved from %d to %d after enabling filter", withoutDetails, withDetails)
	}
	if strings.Contains(withFilter, "┌") || strings.Contains(withFilter, "└") {
		t.Fatalf("rendered output should not include boxed filter borders")
	}
	if countLinesContaining(withFilter, "─") < 2 {
		t.Fatalf("expected divider above table and between table/details")
	}
}

func TestRenderDetailEmptyAndSelected(t *testing.T) {
	muted := lipgloss.NewStyle()
	section := lipgloss.NewStyle()

	empty := RenderDetail(0, SelectionSet{}, "", fakeProvider{}, section, muted, 120, 6)
	if !strings.Contains(empty, "No container selected.") {
		t.Fatalf("empty detail missing message: %q", empty)
	}

	selected := RenderDetail(0, SelectionSet{
		Container: core.ContainerRow{
			ID:            "ctr-1",
			Name:          "api",
			Image:         "nginx:1.27",
			State:         "running",
			Healthy:       true,
			Status:        "Up 5m",
			CPUPercent:    12.5,
			MemoryUsage:   "100 MiB",
			MemoryPercent: 25,
			MemoryLimit:   "400 MiB",
			Ports:         "8080->80/tcp",
			Command:       "nginx",
		},
		HasContainer: true,
	}, "", fakeProvider{}, section, muted, 160, 20)

	for _, token := range []string{"api", "nginx:1.27", "Memory: 100 MiB (25.0%)", "8080->80/tcp", "ctr-1"} {
		if !strings.Contains(selected, token) {
			t.Fatalf("selected detail missing %q in %q", token, selected)
		}
	}
	if strings.Contains(selected, " / 400 MiB") {
		t.Fatalf("container detail should not include max memory in details, got %q", selected)
	}
}

func TestRenderDetailComposeProjectSelected(t *testing.T) {
	muted := lipgloss.NewStyle()
	section := lipgloss.NewStyle()

	selected := RenderDetail(0, SelectionSet{
		ComposeProject: core.ComposeProject{
			Name:           "shop",
			ContainerCount: 2,
			RunningCount:   1,
			HealthyCount:   1,
			Network:        "shop_default, shop_shared",
			WorkingDir:     "/srv/shop",
			ConfigFiles:    "compose.yaml",
			Created:        "just now",
			CPUPercent:     12.5,
			MemoryUsage:    "150 B",
			MemoryLimit:    "300 B",
			MemoryPercent:  50,
			Services:       []string{"api", "worker"},
			Containers:     []core.ContainerRow{{Name: "api"}, {Name: "worker"}},
		},
		HasComposeProject: true,
	}, "", fakeProvider{}, section, muted, 160, 20)

	for _, token := range []string{"Project: shop", "Working dir: /srv/shop", "Compose file: compose.yaml", "Created at: just now", "CPU: 12.5%", "Memory: 150 B (50.0%)", "Networks: - shop_default", "  - shop_shared"} {
		if !strings.Contains(selected, token) {
			t.Fatalf("compose project detail missing %q in %q", token, selected)
		}
	}
	if strings.Contains(selected, "Config files") {
		t.Fatalf("compose project detail should use Compose file label, got %q", selected)
	}
	if strings.Contains(selected, " / 300 B") {
		t.Fatalf("compose project detail should not include max memory in details, got %q", selected)
	}
	if project := strings.Index(selected, "Project: shop"); project == -1 {
		t.Fatalf("compose project detail missing project line: %q", selected)
	} else if workingDir := strings.Index(selected, "Working dir: /srv/shop"); workingDir == -1 || workingDir < project {
		t.Fatalf("compose project detail order should place Working dir after Project, got %q", selected)
	}
}

func TestRenderDetailClipsLongImageLineAtNarrowWidth(t *testing.T) {
	muted := lipgloss.NewStyle()
	section := lipgloss.NewStyle()
	image := "429457073918.dkr.ecr.us-east-1.amazonaws.com/jetstream-robot:2"

	got := RenderDetail(0, SelectionSet{
		Container: core.ContainerRow{
			Name:  "api",
			Image: image,
			State: "running",
		},
		HasContainer: true,
	}, "", fakeProvider{}, section, muted, 40, 10)

	if strings.Contains(got, image) {
		t.Fatalf("narrow detail still contains the full image string: %q", got)
	}
	if strings.Count(got, "Image:") != 1 {
		t.Fatalf("expected one image line, got %q", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if util.DisplayWidth(line) > 40 {
			t.Fatalf("line exceeds width 40: %q", line)
		}
	}
}

func TestContainerCPUValue_RunningNeverDash(t *testing.T) {
	loading := ContainerCPUValue(core.ContainerRow{State: "running", CPUPercent: -1}, "⠋")
	if loading != "⠋" {
		t.Fatalf("running loading cpu = %q, want spinner icon", loading)
	}

	idle := ContainerCPUValue(core.ContainerRow{State: "running", CPUPercent: 0}, "⠋")
	if idle != "0.0%" {
		t.Fatalf("running zero cpu = %q, want 0.0%%", idle)
	}

	afterInitial := ContainerCPUValue(core.ContainerRow{State: "running", CPUPercent: -1}, "")
	if afterInitial != "-" {
		t.Fatalf("running loading cpu with no indicator = %q, want -", afterInitial)
	}
}

func TestContainerMemoryTableValue_OmitsLimit(t *testing.T) {
	running := core.ContainerRow{State: "running", MemoryUsage: "128 MiB", MemoryLimit: "2 GiB", MemoryPercent: 6.25}
	got := ContainerMemoryTableValue(running, "⠋")
	if got != "128 MiB (6.2%)" {
		t.Fatalf("table memory value = %q, want %q", got, "128 MiB (6.2%)")
	}

	loading := ContainerMemoryTableValue(core.ContainerRow{State: "running", MemoryUsage: "-"}, "⠋")
	if loading != "⠋" {
		t.Fatalf("running placeholder memory value = %q, want spinner icon", loading)
	}

	afterInitial := ContainerMemoryTableValue(core.ContainerRow{State: "running", MemoryUsage: "-"}, "")
	if afterInitial != "-" {
		t.Fatalf("running placeholder memory value with no indicator = %q, want -", afterInitial)
	}
}

func lineIndex(view, token string) int {
	for index, line := range strings.Split(view, "\n") {
		if strings.Contains(line, token) {
			return index
		}
	}
	return -1
}

func countLinesContaining(view, token string) int {
	count := 0
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, token) {
			count++
		}
	}
	return count
}
