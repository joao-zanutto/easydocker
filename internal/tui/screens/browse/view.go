package browse

import (
	"fmt"
	"math"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

type ViewModel struct {
	Loading                 bool
	Snapshot                core.Snapshot
	ActiveTab               shared.Tab
	MetricsLoadingIndicator string
	Width                   int
	Height                  int
	Styles                  Styles
	Selections              SelectionSet
	Filter                  FilterState
}

type Styles struct {
	Divider lipgloss.Style
	Muted   lipgloss.Style
	Section lipgloss.Style
}

type SelectionSet struct {
	Container         core.ContainerRow
	HasContainer      bool
	ComposeProject    core.ComposeProject
	HasComposeProject bool
	Image             core.ImageRow
	HasImage          bool
	Network           core.NetworkRow
	HasNetwork        bool
	Volume            core.VolumeRow
	HasVolume         bool
}

type DetailProvider interface {
	DetailLine(label, value string, width int) string
	RenderContainerState(container core.ContainerRow) string
}

func RenderContent(vm ViewModel, list string, detailProvider DetailProvider) string {
	if ShouldRenderLoading(vm.Loading, vm.Snapshot) {
		return util.ConstrainLine(vm.Styles.Muted.Render("Loading Docker resources..."), vm.Width)
	}

	listHeight, detailHeight := ContentHeightsFromFilter(vm.Height)
	listLines := util.ClipAndPadLines(
		util.ConstrainLines(strings.Split(list, "\n"), vm.Width),
		listHeight,
		"",
	)
	listBlock := strings.Join(listLines, "\n")
	detail := RenderDetail(vm.ActiveTab, vm.Selections, vm.MetricsLoadingIndicator, detailProvider, vm.Styles.Section, vm.Styles.Muted, vm.Width, detailHeight)
	divider := vm.Styles.Divider.Render(strings.Repeat("─", max(1, vm.Width)))

	parts := []string{listBlock, divider, detail}
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
		listHeight = max(1, height)
	}
	return listHeight
}

func ListHeightForContent(height int) int {
	listHeight, _ := ContentHeightsFromFilter(height)
	return listHeight
}

func ContentHeightsFromFilter(height int) (int, int) {
	totalHeight := max(1, height)
	listHeight := ListHeight(totalHeight)

	detailHeight := totalHeight - listHeight - 1
	for detailHeight < 1 && listHeight > 1 {
		listHeight--
		detailHeight = totalHeight - listHeight - 1
	}
	if detailHeight < 1 {
		detailHeight = 1
	}

	return listHeight, detailHeight
}

func RenderDetail(activeTab shared.Tab, selections SelectionSet, loadingIndicator string, provider DetailProvider, sectionStyle, mutedStyle lipgloss.Style, width, height int) string {
	lines := append([]string{sectionStyle.Render("Details")}, activeDetailLines(activeTab, selections, loadingIndicator, provider, mutedStyle, width)...)
	return strings.Join(util.ClipAndPadLines(util.ConstrainLines(lines, width), height, ""), "\n")
}

func activeDetailLines(activeTab shared.Tab, selections SelectionSet, loadingIndicator string, provider DetailProvider, mutedStyle lipgloss.Style, width int) []string {
	switch activeTab {
	case shared.TabContainers:
		if selections.HasComposeProject {
			return detailLinesForSelection(selections.ComposeProject, selections.HasComposeProject, "No compose project selected.", composeProjectDetailLines, provider, mutedStyle, width)
		}
		builder := func(container core.ContainerRow, p DetailProvider, w int) []string {
			return containerDetailLines(container, loadingIndicator, p, w)
		}
		return detailLinesForSelection(selections.Container, selections.HasContainer, "No container selected.", builder, provider, mutedStyle, width)
	case shared.TabImages:
		return detailLinesForSelection(selections.Image, selections.HasImage, "No image selected.", imageDetailLines, provider, mutedStyle, width)
	case shared.TabNetworks:
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
		provider.DetailLine("CPU", util.ContainerCPUValue(container, loadingIndicator), width),
		provider.DetailLine("Memory", util.ContainerMemoryTableValue(container, loadingIndicator), width),
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
		provider.DetailLine("Created", network.Created, width),
		provider.DetailLine("Endpoints", fmt.Sprintf("%d", network.Endpoints), width),
		provider.DetailLine("ID", network.ID, width),
	}
}

func volumeDetailLines(volume core.VolumeRow, provider DetailProvider, width int) []string {
	return []string{
		provider.DetailLine("Name", volume.Name, width),
		provider.DetailLine("Created", volume.Created, width),
		provider.DetailLine("Size", volume.Size, width),
		provider.DetailLine("Mountpoint", volume.Mountpoint, width),
	}
}

func composeProjectDetailLines(project core.ComposeProject, provider DetailProvider, width int) []string {
	lines := []string{
		provider.DetailLine("Project", project.Name, width),
		provider.DetailLine("Working dir", project.WorkingDir, width),
		provider.DetailLine("Compose file", project.ConfigFiles, width),
		provider.DetailLine("Created at", project.Created, width),
		provider.DetailLine("CPU", util.FormatPercent(project.CPUPercent), width),
		provider.DetailLine("Memory", composeMemoryText(project), width),
	}
	lines = append(lines, composeProjectNetworkDetailLines(project, provider, width)...)
	return lines
}

func composeMemoryText(project core.ComposeProject) string {
	if project.MemoryUsage == "-" {
		return "-"
	}
	return fmt.Sprintf("%s (%.1f%%)", project.MemoryUsage, project.MemoryPercent)
}

func composeProjectNetworkDetailLines(project core.ComposeProject, provider DetailProvider, width int) []string {
	networks := composeProjectNetworks(project.Network)
	if len(networks) == 0 {
		return []string{provider.DetailLine("Networks", "", width)}
	}

	lines := []string{provider.DetailLine("Networks", "", width)}
	for _, network := range networks {
		lines = append(lines, util.ConstrainLine("  - "+network, width))
	}
	return lines
}

func composeProjectNetworks(networkField string) []string {
	values := strings.Split(networkField, ",")
	networks := make([]string, 0, len(values))
	for _, value := range values {
		network := strings.TrimSpace(value)
		if network == "" || network == "-" {
			continue
		}
		networks = append(networks, network)
	}
	return networks
}
