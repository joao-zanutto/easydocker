package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type tableColumnDef struct {
	header   string
	minWidth int
	desired  func(tableWidth int) int
}

type tableCell struct {
	value         string
	style         lipgloss.Style
	hasStyle      bool
	selectedStyle lipgloss.Style
	hasSelected   bool
}

func (m model) renderResourceList(width, height int) string {
	switch m.activeTab {
	case tabContainers:
		return m.renderContainersTable(width, height)
	case tabImages:
		return m.renderImagesTable(width, height)
	case tabNetworks:
		return m.renderNetworksTable(width, height)
	default:
		return m.renderVolumesTable(width, height)
	}
}

func (m model) renderContainersTable(width, height int) string {
	containers := m.filteredContainers()
	if len(containers) == 0 {
		message := "No containers found."
		if m.activeTab == tabContainers {
			message += " Press a to switch between running and all containers."
		}
		return constrainLine(m.styles.muted.Render(message), width)
	}

	tableWidth := max(1, width-4)
	columns := m.containerTableColumns(tableWidth)
	head := m.renderTableHeader(tableWidth, columns)

	return m.renderWindowedTable(head, tableWidth, height, len(containers), m.containerCursor, func(index int, selected bool) string {
		container := containers[index]
		return m.renderTableCells(tableWidth, columns, m.containerTableCells(container), selected)
	})
}

func (m model) renderImagesTable(width, height int) string {
	if len(m.snapshot.Images) == 0 {
		return constrainLine(m.styles.muted.Render("No images found."), width)
	}

	tableWidth := max(1, width-4)
	columns := m.imageTableColumns(tableWidth)
	head := m.renderTableHeader(tableWidth, columns)

	return m.renderWindowedTable(head, tableWidth, height, len(m.snapshot.Images), m.imageCursor, func(index int, selected bool) string {
		image := m.snapshot.Images[index]
		return m.renderTableCells(tableWidth, columns, m.imageTableCells(image), selected)
	})
}

func (m model) renderNetworksTable(width, height int) string {
	if len(m.snapshot.Networks) == 0 {
		return constrainLine(m.styles.muted.Render("No networks found."), width)
	}

	tableWidth := max(1, width-4)
	columns := m.networkTableColumns(tableWidth)
	head := m.renderTableHeader(tableWidth, columns)

	return m.renderWindowedTable(head, tableWidth, height, len(m.snapshot.Networks), m.networkCursor, func(index int, selected bool) string {
		network := m.snapshot.Networks[index]
		return m.renderTableCells(tableWidth, columns, m.networkTableCells(network), selected)
	})
}

func (m model) renderVolumesTable(width, height int) string {
	if len(m.snapshot.Volumes) == 0 {
		return constrainLine(m.styles.muted.Render("No volumes found."), width)
	}

	tableWidth := max(1, width-4)
	columns := m.volumeTableColumns(tableWidth)
	head := m.renderTableHeader(tableWidth, columns)

	return m.renderWindowedTable(head, tableWidth, height, len(m.snapshot.Volumes), m.volumeCursor, func(index int, selected bool) string {
		volume := m.snapshot.Volumes[index]
		return m.renderTableCells(tableWidth, columns, m.volumeTableCells(volume), selected)
	})
}

func (m model) resolveTableColumns(tableWidth int, defs []tableColumnDef) []tableColumnDef {
	desired := make([]int, 0, len(defs))
	for _, def := range defs {
		width := def.minWidth
		if def.desired != nil {
			width = max(width, def.desired(tableWidth))
		}
		desired = append(desired, width)
	}

	widths := allocateColumns(max(1, tableWidth-((len(defs)-1)*2)), desired)
	resolved := make([]tableColumnDef, 0, len(defs))
	for index, def := range defs {
		def.minWidth = widths[index]
		resolved = append(resolved, def)
	}
	return resolved
}

func (m model) renderTableHeader(tableWidth int, columns []tableColumnDef) string {
	parts := make([]string, 0, len(columns)*2)
	for index, column := range columns {
		if index > 0 {
			parts = append(parts, "  ")
		}
		parts = append(parts, fit(column.header, column.minWidth))
	}
	return constrainLine(strings.Join(parts, ""), tableWidth)
}

func (m model) renderWindowedTable(header string, tableWidth, height, totalRows, cursor int, rowRenderer func(index int, selected bool) string) string {
	lines := []string{m.styles.headerRow.Render(header)}
	visibleRows := max(1, height-1)
	start, end := scrollWindow(totalRows, cursor, visibleRows)
	for index := start; index < end; index++ {
		lines = append(lines, rowRenderer(index, index == cursor))
	}
	return strings.Join(lines, "\n")
}

func (m model) renderTableRow(line string, width int, selected bool) string {
	padded := constrainLine(line, width)
	if selected {
		return m.styles.activeRow.Width(width).MaxWidth(width).Render(padded)
	}
	return m.styles.row.Width(width).MaxWidth(width).Render(padded)
}

func (m model) renderTableCells(width int, columns []tableColumnDef, cells []tableCell, selected bool) string {
	defaultSelected := lipgloss.NewStyle().Bold(true).Background(m.styles.activeBG)
	selectedSep := lipgloss.NewStyle().Background(m.styles.activeBG).Render("  ")
	parts := make([]string, 0, len(cells)*2)
	for i, cell := range cells {
		if i > 0 {
			if selected {
				parts = append(parts, selectedSep)
			} else {
				parts = append(parts, "  ")
			}
		}

		value := fit(cell.value, columns[i].minWidth)
		if selected {
			style := cell.selectedStyle
			if !cell.hasSelected {
				style = defaultSelected
			}
			parts = append(parts, style.Render(value))
			continue
		}

		if cell.hasStyle {
			parts = append(parts, cell.style.Render(value))
			continue
		}
		parts = append(parts, value)
	}

	line := strings.Join(parts, "")
	if selected {
		return constrainLine(line, width)
	}
	return m.renderTableRow(line, width, false)
}
