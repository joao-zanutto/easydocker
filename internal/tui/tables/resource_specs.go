package tables

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/browse"
	"easydocker/internal/tui/util"
)

type ContainerListRowKind int

const (
	ContainerListRowContainer ContainerListRowKind = iota
	ContainerListRowComposeProject
)

type ContainerListRow struct {
	Kind            ContainerListRowKind
	Container       core.ContainerRow
	ComposeProject  core.ComposeProject
	ComposeExpanded bool
	TreePrefix      string
}

// ContainerColumns resolves container columns for a given table width.
func ContainerColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, ContainerSchema)
}

// ImageColumns resolves image columns for a given table width.
func ImageColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, ImageSchema)
}

// NetworkColumns resolves network columns for a given table width.
func NetworkColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, NetworkSchema)
}

// VolumeColumns resolves volume columns for a given table width.
func VolumeColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, VolumeSchema)
}

// BuildContainerListRows creates a flat row model that includes compose project rows.
func BuildContainerListRows(items []core.ContainerRow, composeExpanded map[string]bool) []ContainerListRow {
	projects := core.AggregateComposeProjects(items)
	projectByName := make(map[string]core.ComposeProject, len(projects))
	for _, project := range projects {
		projectByName[project.Name] = project
	}

	rows := make([]ContainerListRow, 0, len(items))
	emittedProjects := make(map[string]bool, len(projects))
	for _, container := range items {
		projectName := strings.TrimSpace(container.ComposeProject)
		if projectName == "" {
			rows = append(rows, ContainerListRow{Kind: ContainerListRowContainer, Container: container})
			continue
		}
		if emittedProjects[projectName] {
			continue
		}
		project, ok := projectByName[projectName]
		if !ok {
			continue
		}
		expanded := composeExpanded[projectName]
		rows = append(rows, ContainerListRow{
			Kind:            ContainerListRowComposeProject,
			ComposeProject:  project,
			ComposeExpanded: expanded,
		})
		emittedProjects[projectName] = true
		if !expanded {
			continue
		}

		childRows := make([]core.ContainerRow, len(project.Containers))
		copy(childRows, project.Containers)
		core.SortContainers(childRows)
		for index, child := range childRows {
			prefix := "├─ "
			if index == len(childRows)-1 {
				prefix = "└─ "
			}
			rows = append(rows, ContainerListRow{
				Kind:       ContainerListRowContainer,
				Container:  child,
				TreePrefix: prefix,
			})
		}
	}

	return rows
}

// BuildContainerSpec builds a complete containers table spec.
func BuildContainerSpec(width, cursor int, items []ContainerListRow, includeScopeHint bool, loadingIndicator string) Spec[ContainerListRow] {
	tableWidth := ContentWidth(width)
	columns := ContainerColumns(tableWidth)
	stateWidth := ContainerStateColumnWidth(columns)
	selectedContainerID := ""
	selectedComposeProject := ""
	selectedKind := ContainerListRowContainer
	hasSelectedKind := false
	if cursor >= 0 && cursor < len(items) {
		hasSelectedKind = true
		selectedKind = items[cursor].Kind
		if items[cursor].Kind == ContainerListRowContainer {
			selectedContainerID = items[cursor].Container.FullID
		} else {
			selectedComposeProject = items[cursor].ComposeProject.Name
		}
	}
	emptyMessage := "No containers found."
	if includeScopeHint {
		emptyMessage += " Press a to switch between running and all containers."
	}

	return Spec[ContainerListRow]{
		EmptyMessage: emptyMessage,
		Cursor:       cursor,
		Items:        items,
		Columns:      columns,
		RowBuilder: func(item ContainerListRow) []string {
			if item.Kind == ContainerListRowComposeProject {
				rowLoadingIndicator := ""
				if hasSelectedKind && selectedKind == ContainerListRowComposeProject && item.ComposeProject.Name == selectedComposeProject {
					rowLoadingIndicator = loadingIndicator
				}
				return ComposeProjectTableRow(item, rowLoadingIndicator)
			}
			container := item.Container
			rowLoadingIndicator := ""
			if hasSelectedKind && selectedKind == ContainerListRowContainer && selectedContainerID != "" && container.FullID == selectedContainerID {
				rowLoadingIndicator = loadingIndicator
			}
			return ContainerTableRow(container, stateWidth, rowLoadingIndicator, item.TreePrefix)
		},
	}
}

// BuildImageSpec builds a complete images table spec.
func BuildImageSpec(width, cursor int, items []core.ImageRow) Spec[core.ImageRow] {
	return SimpleSpec(width, "No images found.", cursor, items, ImageColumns, ImageTableRow)
}

// BuildNetworkSpec builds a complete networks table spec.
func BuildNetworkSpec(width, cursor int, items []core.NetworkRow) Spec[core.NetworkRow] {
	return SimpleSpec(width, "No networks found.", cursor, items, NetworkColumns, NetworkTableRow)
}

// BuildVolumeSpec builds a complete volumes table spec.
func BuildVolumeSpec(width, cursor int, items []core.VolumeRow) Spec[core.VolumeRow] {
	return SimpleSpec(width, "No volumes found.", cursor, items, VolumeColumns, VolumeTableRow)
}

// ContainerTableRow builds a single row for the containers table.
func ContainerTableRow(container core.ContainerRow, stateWidth int, loadingIndicator, treePrefix string) []string {
	state := browse.ContainerStateText(container)
	if stateWidth > 0 && util.StripANSI(util.TruncateWithEllipsis(state, stateWidth)) == "…" {
		state = "●"
	}
	state = colorStateLabel(state, container.State)
	name := container.Name
	if treePrefix != "" {
		name = treePrefix + name
	}

	return []string{
		name,
		state,
		browse.ContainerCPUValue(container, loadingIndicator),
		browse.ContainerMemoryTableValue(container, loadingIndicator),
		container.Image,
		container.Status,
	}
}

func ComposeProjectTableRow(item ContainerListRow, loadingIndicator string) []string {
	prefix := "[+]"
	if item.ComposeExpanded {
		prefix = "[-]"
	}
	name := ansiBold(prefix + " " + item.ComposeProject.Name)
	state := fmt.Sprintf("%d/%d running", item.ComposeProject.RunningCount, item.ComposeProject.ContainerCount)
	cpu := "-"
	if item.ComposeProject.CPUPercent > 0 {
		cpu = fmt.Sprintf("%.1f%%", item.ComposeProject.CPUPercent)
	} else if item.ComposeProject.RunningCount > 0 && loadingIndicator != "" && item.ComposeProject.MemoryUsage == "-" {
		cpu = loadingIndicator
	}
	mem := "-"
	if item.ComposeProject.MemoryUsage != "-" {
		mem = fmt.Sprintf("%s (%.1f%%)", item.ComposeProject.MemoryUsage, item.ComposeProject.MemoryPercent)
	} else if item.ComposeProject.RunningCount > 0 && loadingIndicator != "" {
		mem = loadingIndicator
	}
	status := item.ComposeProject.Created
	if status == "" {
		status = "-"
	}
	return []string{name, ansiBold(state), ansiBold(cpu), ansiBold(mem), ansiBold("-"), ansiBold(status)}
}

func ansiBold(value string) string {
	if value == "" {
		return ""
	}
	return "\x1b[1m" + value + "\x1b[22m"
}

// ImageTableRow builds a single row for the images table.
func ImageTableRow(image core.ImageRow) []string {
	repository, tags := splitImageTags(image.Tags)
	return []string{repository, tags, image.Size, image.Created, image.ID}
}

func splitImageTags(formattedTags string) (string, string) {
	if formattedTags == "" {
		return "-", "-"
	}

	repositories := make([]string, 0, 4)
	tags := make([]string, 0, 4)
	for _, reference := range strings.Split(formattedTags, ", ") {
		repository, tag := splitImageTagReference(reference)
		repositories = append(repositories, repository)
		tags = append(tags, tag)
	}

	return strings.Join(repositories, ", "), strings.Join(tags, ", ")
}

func splitImageTagReference(reference string) (string, string) {
	if reference == "<none>:<none>" {
		return "<none>", "<none>"
	}

	lastColon := strings.LastIndex(reference, ":")
	if lastColon <= 0 || lastColon == len(reference)-1 {
		return reference, "-"
	}

	return reference[:lastColon], reference[lastColon+1:]
}

// NetworkTableRow builds a single row for the networks table.
func NetworkTableRow(network core.NetworkRow) []string {
	return []string{
		network.Name,
		network.Driver,
		network.Scope,
		fmt.Sprintf("%d", network.Endpoints),
		fmt.Sprintf("internal:%s attach:%s", network.Internal, network.Attachable),
	}
}

// VolumeTableRow builds a single row for the volumes table.
func VolumeTableRow(volume core.VolumeRow) []string {
	return []string{
		volume.Name,
		volume.Driver,
		volume.Scope,
		volume.Size,
		util.RefCountText(volume.RefCount),
	}
}

func colorStateLabel(label, state string) string {
	code := "37"
	switch strings.ToLower(state) {
	case "running":
		code = "32"
	case "paused", "restarting", "created":
		code = "33"
	case "exited", "stopped":
		code = "31"
	case "dead":
		code = "91"
	default:
		code = "36"
	}
	// Reset only the foreground so selected-row background remains intact.
	return "\x1b[" + code + "m" + label + "\x1b[39m"
}
