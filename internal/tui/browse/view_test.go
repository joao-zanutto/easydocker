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

	for _, token := range []string{"api", "nginx:1.27", "8080->80/tcp", "ctr-1"} {
		if !strings.Contains(selected, token) {
			t.Fatalf("selected detail missing %q in %q", token, selected)
		}
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
