package tables

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/ui/theme"
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
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

// ContainerTableRow builds a single row for the containers table.
func ContainerTableRow(container core.ContainerRow, stateWidth int, loadingIndicator, treePrefix string) []string {
	state := core.ContainerStateText(container)
	if stateWidth > 0 && util.StripANSI(util.TruncateWithEllipsis(state, stateWidth)) == "…" {
		state = "●"
	}
	state = colorStateLabel(state, container.State)
	name := container.Name
	cpu := core.ContainerCPUValue(container, loadingIndicator)
	mem := core.ContainerMemoryTableValue(container, loadingIndicator)
	if treePrefix != "" {
		name = treePrefix + name
		cpu = childMetricGuidePrefix(treePrefix) + cpu
		mem = childMetricGuidePrefix(treePrefix) + mem
	}

	return []string{
		name,
		state,
		cpu,
		mem,
		container.Image,
		container.Status,
	}
}

func childMetricGuidePrefix(treePrefix string) string {
	guide := "├─ "
	if strings.HasPrefix(treePrefix, "└") {
		guide = "└─ "
	}
	return "\x1b[2m" + guide + "\x1b[22m"
}

func ComposeProjectTableRow(item ContainerListRow, loadingIndicator string) []string {
	prefix := "[+]"
	if item.ComposeExpanded {
		prefix = "[-]"
	}
	name := ansiBold(prefix + "  " + item.ComposeProject.Name)
	state := fmt.Sprintf("%d/%d running", item.ComposeProject.RunningCount, item.ComposeProject.ContainerCount)
	cpu := "-"
	if item.ComposeProject.CPUPercent > 0 {
		cpu = fmt.Sprintf("%.1f%%", item.ComposeProject.CPUPercent)
	} else if item.ComposeProject.RunningCount > 0 && loadingIndicator != "" && item.ComposeProject.MemoryUsage == "-" {
		cpu = loadingIndicator
	}
	mem := "-"
	if item.ComposeProject.MemoryUsage != "-" {
		mem = item.ComposeProject.MemoryUsage
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

func colorStateLabel(label, state string) string {
	s := lipgloss.NewStyle().Foreground(theme.ContainerStateColor(state)).Render(label)
	s = strings.TrimSuffix(s, "\x1b[0m")
	s = strings.TrimSuffix(s, "\x1b[m")
	return s + "\x1b[39m"
}
