package browse

import (
	"fmt"
	"math"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/util"

	"github.com/charmbracelet/lipgloss"
)

const filterHeaderHeight = 2

type ViewModel struct {
	Loading                 bool
	Snapshot                core.Snapshot
	ActiveTab               int
	MetricsLoadingIndicator string
	Width                   int
	Height                  int
	Styles                  ViewStyles
	Selections              SelectionSet
	// Filter mode state
	FilterActive bool
	FilterQuery  string
	FilterInput  string
}

type ViewStyles struct {
	Divider lipgloss.Style
	Muted   lipgloss.Style
	Section lipgloss.Style
}

type SelectionSet struct {
	Container    core.ContainerRow
	HasContainer bool
	Image        core.ImageRow
	HasImage     bool
	Network      core.NetworkRow
	HasNetwork   bool
	Volume       core.VolumeRow
	HasVolume    bool
}

type DetailProvider interface {
	DetailLine(label, value string, width int) string
	RenderContainerState(container core.ContainerRow) string
}

func RenderContent(vm ViewModel, list string, detailProvider DetailProvider) string {
	if ShouldRenderLoading(vm.Loading, vm.Snapshot) {
		return util.ConstrainLine(vm.Styles.Muted.Render("Loading Docker resources..."), vm.Width)
	}

	filterHeight, listHeight, detailHeight := contentHeights(vm.Height, vm.FilterActive)
	listLines := util.ClipAndPadLines(
		util.ConstrainLines(strings.Split(list, "\n"), vm.Width),
		listHeight,
		"",
	)
	listBlock := strings.Join(listLines, "\n")
	detail := RenderDetail(vm.ActiveTab, vm.Selections, vm.MetricsLoadingIndicator, detailProvider, vm.Styles.Section, vm.Styles.Muted, vm.Width, detailHeight)
	divider := vm.Styles.Divider.Render(strings.Repeat("─", max(1, vm.Width-2)))

	var parts []string
	if filterHeight > 0 {
		parts = append(parts, RenderFilterHeader(vm.FilterInput, vm.Width, vm.Styles.Divider))
	}
	parts = append(parts, listBlock, divider, detail)
	return util.JoinSections(parts...)
}

func ShouldRenderLoading(loading bool, snapshot core.Snapshot) bool {
	return loading && !HasResources(snapshot)
}

func HasResources(snapshot core.Snapshot) bool {
	return len(snapshot.Containers) > 0 ||
		len(snapshot.Images) > 0 ||
		len(snapshot.Networks) > 0 ||
		len(snapshot.Volumes) > 0
}

func ListHeight(height int) int {
	listHeight := int(math.Round(float64(height) * 0.6))
	if listHeight < 3 {
		listHeight = 3
	}
	if listHeight > height-2 {
		listHeight = max(1, height-2)
	}
	return listHeight
}

// ListHeightForContent computes table height while preserving a fixed divider position
// when the filter container is shown above the table.
func ListHeightForContent(height int, filterActive bool) int {
	_, listHeight, _ := contentHeights(height, filterActive)
	return listHeight
}

func contentHeights(height int, filterActive bool) (int, int, int) {
	totalHeight := max(1, height)
	filterHeight := 0
	if filterActive {
		filterHeight = filterHeaderHeight
		// Keep room for list + divider + detail.
		maxFilterHeight := max(0, totalHeight-3)
		if filterHeight > maxFilterHeight {
			filterHeight = maxFilterHeight
		}
	}

	listHeight := ListHeight(totalHeight)
	if filterHeight > 0 {
		// Shrink table from the top to preserve divider/bottom anchoring.
		listHeight = max(1, listHeight-filterHeight)
	}

	detailHeight := totalHeight - filterHeight - listHeight - 1
	for detailHeight < 1 && listHeight > 1 {
		listHeight--
		detailHeight = totalHeight - filterHeight - listHeight - 1
	}
	if detailHeight < 1 {
		detailHeight = 1
	}

	return filterHeight, listHeight, detailHeight
}

func RenderDetail(activeTab int, selections SelectionSet, loadingIndicator string, provider DetailProvider, sectionStyle, mutedStyle lipgloss.Style, width, height int) string {
	lines := append([]string{sectionStyle.Render("Details")}, activeDetailLines(activeTab, selections, loadingIndicator, provider, mutedStyle, width)...)
	return strings.Join(util.ClipLines(util.ConstrainLines(lines, width), height), "\n")
}

func activeDetailLines(activeTab int, selections SelectionSet, loadingIndicator string, provider DetailProvider, mutedStyle lipgloss.Style, width int) []string {
	switch activeTab {
	case 0:
		builder := func(container core.ContainerRow, p DetailProvider, w int) []string {
			return containerDetailLines(container, loadingIndicator, p, w)
		}
		return detailLinesForSelection(selections.Container, selections.HasContainer, "No container selected.", builder, provider, mutedStyle, width)
	case 1:
		return detailLinesForSelection(selections.Image, selections.HasImage, "No image selected.", imageDetailLines, provider, mutedStyle, width)
	case 2:
		return detailLinesForSelection(selections.Network, selections.HasNetwork, "No network selected.", networkDetailLines, provider, mutedStyle, width)
	default:
		return detailLinesForSelection(selections.Volume, selections.HasVolume, "No volume selected.", volumeDetailLines, provider, mutedStyle, width)
	}
}

func detailLinesForSelection[T any](item T, ok bool, emptyMessage string, buildLines func(T, DetailProvider, int) []string, provider DetailProvider, mutedStyle lipgloss.Style, width int) []string {
	if ok {
		return buildLines(item, provider, width)
	}
	return []string{mutedStyle.Render(emptyMessage)}
}

func containerDetailLines(container core.ContainerRow, loadingIndicator string, provider DetailProvider, width int) []string {
	return []string{
		provider.DetailLine("Name", container.Name, width),
		provider.DetailLine("Image", container.Image, width),
		provider.DetailLine("State", provider.RenderContainerState(container), width),
		provider.DetailLine("Status", container.Status, width),
		provider.DetailLine("CPU", ContainerCPUValue(container, loadingIndicator), width),
		provider.DetailLine("Memory", ContainerMemorySummary(container, loadingIndicator), width),
		provider.DetailLine("Ports", container.Ports, width),
		provider.DetailLine("Command", container.Command, width),
		provider.DetailLine("ID", container.ID, width),
	}
}

func imageDetailLines(image core.ImageRow, provider DetailProvider, width int) []string {
	return []string{
		provider.DetailLine("Tags", image.Tags, width),
		provider.DetailLine("Size", image.Size, width),
		provider.DetailLine("Created", image.Created, width),
		provider.DetailLine("Containers", fmt.Sprintf("%d", image.Containers), width),
		provider.DetailLine("ID", image.ID, width),
	}
}

func networkDetailLines(network core.NetworkRow, provider DetailProvider, width int) []string {
	return []string{
		provider.DetailLine("Name", network.Name, width),
		provider.DetailLine("Driver", network.Driver, width),
		provider.DetailLine("Scope", network.Scope, width),
		provider.DetailLine("Created", network.Created, width),
		provider.DetailLine("Internal", network.Internal, width),
		provider.DetailLine("Attachable", network.Attachable, width),
		provider.DetailLine("Endpoints", fmt.Sprintf("%d", network.Endpoints), width),
		provider.DetailLine("ID", network.ID, width),
	}
}

func volumeDetailLines(volume core.VolumeRow, provider DetailProvider, width int) []string {
	return []string{
		provider.DetailLine("Name", volume.Name, width),
		provider.DetailLine("Driver", volume.Driver, width),
		provider.DetailLine("Scope", volume.Scope, width),
		provider.DetailLine("Created", volume.Created, width),
		provider.DetailLine("Size", volume.Size, width),
		provider.DetailLine("References", util.RefCountText(volume.RefCount), width),
		provider.DetailLine("Mountpoint", volume.Mountpoint, width),
	}
}

func ContainerCPUValue(container core.ContainerRow, loadingIndicator string) string {
	if container.CPUPercent < 0 {
		if strings.EqualFold(container.State, "running") {
			return metricsLoadingValue(loadingIndicator)
		}
		return "-"
	}
	if container.CPUPercent < 0.05 {
		if strings.EqualFold(container.State, "running") {
			return "0.0%"
		}
		return "-"
	}
	return fmt.Sprintf("%.1f%%", container.CPUPercent)
}

func ContainerMemorySummary(container core.ContainerRow, loadingIndicator string) string {
	if container.MemoryUsage == "-" || strings.EqualFold(container.MemoryUsage, "loading") {
		if strings.EqualFold(container.State, "running") {
			return metricsLoadingValue(loadingIndicator)
		}
		return "-"
	}
	if container.MemoryLimit != "" && container.MemoryLimit != "-" {
		return fmt.Sprintf("%s / %s (%.1f%%)", container.MemoryUsage, container.MemoryLimit, container.MemoryPercent)
	}
	return fmt.Sprintf("%s (%.1f%%)", container.MemoryUsage, container.MemoryPercent)
}

func ContainerMemoryTableValue(container core.ContainerRow, loadingIndicator string) string {
	if container.MemoryUsage == "-" || strings.EqualFold(container.MemoryUsage, "loading") {
		if strings.EqualFold(container.State, "running") {
			return metricsLoadingValue(loadingIndicator)
		}
		return "-"
	}
	return fmt.Sprintf("%s (%.1f%%)", container.MemoryUsage, container.MemoryPercent)
}

func metricsLoadingValue(loadingIndicator string) string {
	if strings.TrimSpace(loadingIndicator) == "" {
		return "-"
	}
	return loadingIndicator
}

// RenderFilterHeader renders a plain filter input line followed by a divider.
func RenderFilterHeader(input string, width int, dividerStyle lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	inputLine := padVisibleWidth(input, width)
	divider := dividerStyle.Render(strings.Repeat("─", max(1, width-2)))
	return util.JoinSections(inputLine, divider)
}

func padVisibleWidth(line string, width int) string {
	constrained := util.ClampSingleLine(line, width)
	padding := width - util.DisplayWidth(constrained)
	if padding <= 0 {
		return constrained
	}
	return constrained + strings.Repeat(" ", padding)
}

func ContainerStateText(container core.ContainerRow) string {
	if container.Healthy && container.State == "running" {
		return "● healthy"
	}
	return "● " + container.State
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
