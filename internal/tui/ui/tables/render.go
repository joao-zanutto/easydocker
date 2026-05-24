package tables

import (
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

type Row []string

// allocateColumns distributes total width across desired column widths.
func allocateColumns(total int, desired []int) []int {
	if len(desired) == 0 {
		return []int{}
	}
	if total <= 0 {
		out := make([]int, len(desired))
		for i := range out {
			out[i] = 1
		}
		return out
	}

	out := make([]int, len(desired))
	sum := 0
	for i, width := range desired {
		if width < 1 {
			width = 1
		}
		out[i] = width
		sum += width
	}
	if sum == total {
		return out
	}
	if sum < total {
		out[len(out)-1] += total - sum
		return out
	}

	over := sum - total
	for over > 0 {
		changed := false
		for i := range out {
			if out[i] > 1 && over > 0 {
				out[i]--
				over--
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return out
}

type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Header:   lipgloss.NewStyle().Bold(true),
		Cell:     lipgloss.NewStyle(),
		Selected: lipgloss.NewStyle().Bold(true),
	}
}

// ResolveColumns computes final column widths based on available space.
func ResolveColumns(tableWidth int, defs []ColumnDef) []ColumnDef {
	desired := make([]int, 0, len(defs))
	for _, def := range defs {
		width := def.MinWidth
		if def.Desired != nil {
			width = max(width, def.Desired(tableWidth))
		}
		desired = append(desired, width)
	}

	widths := allocateColumns(max(1, tableWidth-((len(defs)-1)*2)), desired)
	resolved := make([]ColumnDef, 0, len(defs))
	for index, def := range defs {
		def.MinWidth = widths[index]
		resolved = append(resolved, def)
	}
	return resolved
}

// RowsFrom converts items to table rows using a builder function.
func RowsFrom[T any](items []T, rowBuilder func(T) []string) []Row {
	rows := make([]Row, 0, len(items))
	for _, item := range items {
		rows = append(rows, Row(rowBuilder(item)))
	}
	return rows
}

// RenderFromSpec renders a table from a spec, handling empty state.
func RenderFromSpec[T any](width, height int, spec Spec[T], styles Styles) string {
	rows := RowsFrom(spec.Items, spec.RowBuilder)
	return RenderOrEmpty(width, height, spec.EmptyMessage, spec.Columns, rows, spec.Cursor, styles)
}

// RenderOrEmpty renders a table or an empty message.
func RenderOrEmpty(width, height int, emptyMessage string, columns []ColumnDef, rows []Row, cursor int, styles Styles) string {
	if len(rows) == 0 {
		return util.ConstrainLine(emptyMessage, width)
	}
	return renderTable(styles, width, height, columns, rows, cursor)
}

// renderTable creates a rendered table with styled rows and cursor.
func renderTable(styles Styles, width, height int, defs []ColumnDef, rows []Row, cursor int) string {
	cols := make([]tableColumn, 0, len(defs))
	for _, def := range defs {
		cols = append(cols, tableColumn{Title: def.Header, Width: def.MinWidth})
	}
	privateRows := make([]tableRow, 0, len(rows))
	for _, row := range rows {
		privateRows = append(privateRows, tableRow(row))
	}
	privateStyles := tableStyles(styles)

	t := newTable(
		withColumns(cols),
		withRows(privateRows),
		withStyles(privateStyles),
		withWidth(max(1, width)),
		withHeight(max(2, height)),
	)
	if len(rows) > 0 {
		t.setCursor(util.Clamp(cursor, 0, len(rows)-1))
	}
	return t.view()
}
