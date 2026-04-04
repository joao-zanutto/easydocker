package tui

import (
	"fmt"
	"math"
	"strings"

	"easydocker/internal/core"
)

func (m model) renderBrowseContent(width, height int) string {
	if m.loading && len(m.snapshot.Containers) == 0 && len(m.snapshot.Images) == 0 && len(m.snapshot.Networks) == 0 && len(m.snapshot.Volumes) == 0 {
		return constrainLine(m.styles.muted.Render("Loading Docker resources..."), width)
	}

	listHeight := int(math.Round(float64(height) * 0.6))
	if listHeight < 3 {
		listHeight = 3
	}
	if listHeight > height-2 {
		listHeight = max(1, height-2)
	}
	detailHeight := max(1, height-listHeight-1)

	list := m.renderResourceList(width, listHeight)
	divider := m.renderDivider(width)
	detail := m.renderDetail(width, detailHeight)
	return joinSections(list, divider, detail)
}

func (m model) renderDetail(width, height int) string {
	lines := []string{m.styles.section.Render("Details")}

	switch m.activeTab {
	case tabContainers:
		if container, ok := m.selectedContainer(); ok {
			lines = append(lines,
				m.detailLine("Name", container.Name),
				m.detailLine("Image", container.Image),
				m.detailLine("State", m.stateStyle(container.State).Render(m.renderContainerState(container))),
				m.detailLine("Status", container.Status),
				m.detailLine("CPU", m.renderCPUValue(container.CPUPercent)),
				m.detailLine("Memory", formatMemoryUsage(container.MemoryUsage, container.MemoryPercent, container.MemoryLimit)),
				m.detailLine("Ports", container.Ports),
				m.detailLine("Command", container.Command),
				m.detailLine("ID", container.ID),
			)
		} else {
			lines = append(lines, m.styles.muted.Render("No container selected."))
		}
	case tabImages:
		if image, ok := m.selectedImage(); ok {
			lines = append(lines,
				m.detailLine("Tags", image.Tags),
				m.detailLine("Size", image.Size),
				m.detailLine("Created", image.Created),
				m.detailLine("Containers", fmt.Sprintf("%d", image.Containers)),
				m.detailLine("ID", image.ID),
			)
		} else {
			lines = append(lines, m.styles.muted.Render("No image selected."))
		}
	case tabNetworks:
		if network, ok := m.selectedNetwork(); ok {
			lines = append(lines,
				m.detailLine("Name", network.Name),
				m.detailLine("Driver", network.Driver),
				m.detailLine("Scope", network.Scope),
				m.detailLine("Created", network.Created),
				m.detailLine("Internal", network.Internal),
				m.detailLine("Attachable", network.Attachable),
				m.detailLine("Endpoints", fmt.Sprintf("%d", network.Endpoints)),
				m.detailLine("ID", network.ID),
			)
		} else {
			lines = append(lines, m.styles.muted.Render("No network selected."))
		}
	default:
		if volume, ok := m.selectedVolume(); ok {
			lines = append(lines,
				m.detailLine("Name", volume.Name),
				m.detailLine("Driver", volume.Driver),
				m.detailLine("Scope", volume.Scope),
				m.detailLine("Created", volume.Created),
				m.detailLine("Size", volume.Size),
				m.detailLine("References", refCountText(volume.RefCount)),
				m.detailLine("Mountpoint", volume.Mountpoint),
			)
		} else {
			lines = append(lines, m.styles.muted.Render("No volume selected."))
		}
	}

	return strings.Join(clipLines(constrainLines(lines, width), height), "\n")
}

func (m model) renderContainerMemorySummary(container core.ContainerRow) string {
	if container.MemoryUsage == "-" {
		return "-"
	}
	return fmt.Sprintf("%s (%s)", container.MemoryUsage, renderPercent(container.MemoryPercent))
}

func (m model) renderCPUValue(value float64) string {
	if value < 0.05 {
		return "-"
	}
	return renderPercent(value)
}

func (m model) renderContainerState(container core.ContainerRow) string {
	if strings.EqualFold(container.State, "running") && container.Healthy {
		return "● healthy"
	}
	return "● " + container.State
}
