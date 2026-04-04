package tui

import (
	"fmt"

	"easydocker/internal/core"
)

func (m model) containerTableColumns(tableWidth int) []tableColumnDef {
	return m.resolveTableColumns(tableWidth, []tableColumnDef{
		{header: "NAME", minWidth: 10, desired: func(tableWidth int) int { return max(10, tableWidth/6) }},
		{header: "STATE", minWidth: 12, desired: func(int) int { return 12 }},
		{header: "CPU", minWidth: 7, desired: func(int) int { return 7 }},
		{header: "MEMORY", minWidth: 20, desired: func(int) int { return 20 }},
		{header: "IMAGE", minWidth: 12, desired: func(tableWidth int) int { return max(12, tableWidth/5) }},
		{header: "STATUS", minWidth: 8, desired: func(tableWidth int) int { return max(8, tableWidth/4) }},
	})
}

func (m model) imageTableColumns(tableWidth int) []tableColumnDef {
	return m.resolveTableColumns(tableWidth, []tableColumnDef{
		{header: "REPOSITORY/TAGS", minWidth: 24, desired: func(tableWidth int) int { return max(24, tableWidth/2) }},
		{header: "SIZE", minWidth: 10, desired: func(int) int { return 10 }},
		{header: "CREATED", minWidth: 12, desired: func(int) int { return 12 }},
		{header: "IMAGE ID", minWidth: 12, desired: func(tableWidth int) int { return max(12, tableWidth/5) }},
	})
}

func (m model) networkTableColumns(tableWidth int) []tableColumnDef {
	return m.resolveTableColumns(tableWidth, []tableColumnDef{
		{header: "NAME", minWidth: 18, desired: func(tableWidth int) int { return max(18, tableWidth/4) }},
		{header: "DRIVER", minWidth: 10, desired: func(tableWidth int) int { return max(10, tableWidth/6) }},
		{header: "SCOPE", minWidth: 10, desired: func(int) int { return 10 }},
		{header: "ENDPOINTS", minWidth: 10, desired: func(int) int { return 10 }},
		{header: "META", minWidth: 18, desired: func(tableWidth int) int { return max(18, tableWidth/4) }},
	})
}

func (m model) volumeTableColumns(tableWidth int) []tableColumnDef {
	return m.resolveTableColumns(tableWidth, []tableColumnDef{
		{header: "NAME", minWidth: 18, desired: func(tableWidth int) int { return max(18, tableWidth/4) }},
		{header: "DRIVER", minWidth: 10, desired: func(tableWidth int) int { return max(10, tableWidth/6) }},
		{header: "SCOPE", minWidth: 10, desired: func(int) int { return 10 }},
		{header: "SIZE", minWidth: 10, desired: func(int) int { return 10 }},
		{header: "REFS", minWidth: 8, desired: func(int) int { return 8 }},
	})
}

func (m model) containerTableCells(container core.ContainerRow) []tableCell {
	return []tableCell{
		{value: container.Name},
		{value: m.renderContainerState(container), style: m.stateStyle(container.State), hasStyle: true, selectedStyle: m.stateStyle(container.State).Bold(true).Background(m.styles.activeBG), hasSelected: true},
		{value: m.renderCPUValue(container.CPUPercent)},
		{value: m.renderContainerMemorySummary(container)},
		{value: container.Image},
		{value: container.Status},
	}
}

func (m model) imageTableCells(image core.ImageRow) []tableCell {
	return []tableCell{{value: image.Tags}, {value: image.Size}, {value: image.Created}, {value: image.ID}}
}

func (m model) networkTableCells(network core.NetworkRow) []tableCell {
	return []tableCell{
		{value: network.Name},
		{value: network.Driver},
		{value: network.Scope},
		{value: fmt.Sprintf("%d", network.Endpoints)},
		{value: fmt.Sprintf("internal:%s attach:%s", network.Internal, network.Attachable)},
	}
}

func (m model) volumeTableCells(volume core.VolumeRow) []tableCell {
	return []tableCell{
		{value: volume.Name},
		{value: volume.Driver},
		{value: volume.Scope},
		{value: volume.Size},
		{value: refCountText(volume.RefCount)},
	}
}
