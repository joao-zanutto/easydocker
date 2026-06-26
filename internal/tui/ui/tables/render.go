package tables

import (
	"easydocker/internal/tui/util"

	"charm.land/lipgloss/v2"
)

type Row []string

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
	firstPinned := len(defs)
	for i, def := range defs {
		if def.PinnedRight {
			firstPinned = i
			break
		}
	}

	nonPinnedCount := firstPinned
	pinnedCount := len(defs) - firstPinned
	gapCount := max(0, nonPinnedCount-1) + max(0, pinnedCount-1)
	if nonPinnedCount > 0 && pinnedCount > 0 {
		gapCount++ // section gap between left and right pinned groups
	}
	netWidth := max(1, tableWidth-gapCount*2)

	desired := make([]int, len(defs))
	for i, def := range defs {
		w := def.Width
		if def.Desired != nil {
			w = max(w, def.Desired(tableWidth))
		}
		desired[i] = w
	}

	nonPinnedDesired := make([]int, firstPinned)
	nonPinnedMin := 0
	nonPinnedSum := 0
	for i := 0; i < firstPinned; i++ {
		nonPinnedDesired[i] = desired[i]
		nonPinnedMin += defs[i].Width
		nonPinnedSum += desired[i]
	}

	pinnedDesired := desired[firstPinned:]
	pinnedMin := 0
	pinnedSum := 0
	for i := firstPinned; i < len(defs); i++ {
		pinnedMin += defs[i].Width
		pinnedSum += desired[i]
	}

	remaining := netWidth - pinnedMin - nonPinnedMin

	var nonPinnedWidths, pinnedWidths []int

	if remaining >= 0 {
		// Enough for everyone's minimum. Give to non-pinned up to
		// desired, then to pinned up to desired, then surplus to
		// non-pinned.
		nonPinnedExtra := nonPinnedSum - nonPinnedMin
		give := min(remaining, nonPinnedExtra)
		nonPinnedWidths = util.AllocateColumns(nonPinnedMin+give, nonPinnedDesired)
		remaining -= give

		pinnedExtra := pinnedSum - pinnedMin
		give = min(remaining, pinnedExtra)
		pinnedWidths = util.AllocateColumns(pinnedMin+give, pinnedDesired)
		remaining -= give

		if remaining > 0 {
			nonPinnedWidths = util.AllocateColumns(nonPinnedMin+nonPinnedExtra+remaining, nonPinnedDesired)
		}
	} else {
		// Deficit: pinned keep their mins, left absorb the shortfall.
		pinnedWidths = util.AllocateColumns(pinnedMin, pinnedDesired)
		nonPinnedWidths = util.AllocateColumns(max(1, netWidth-pinnedMin), nonPinnedDesired)
	}

	resolved := make([]ColumnDef, 0, len(defs))
	for i, def := range defs {
		if def.PinnedRight {
			def.Width = pinnedWidths[i-firstPinned]
		} else {
			def.Width = nonPinnedWidths[i]
		}
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
	return RenderOrEmpty(width, height, spec.EmptyMessage, spec.Columns, rows, spec.Cursor, styles, spec.HideHeader)
}

// RenderOrEmpty renders a table or an empty message.
// When rows are empty but the header should be visible (hideHeader=false),
// the header is rendered above the empty message.
func RenderOrEmpty(width, height int, emptyMessage string, columns []ColumnDef, rows []Row, cursor int, styles Styles, hideHeader bool) string {
	if len(rows) == 0 {
		if hideHeader || len(columns) == 0 {
			return util.ConstrainLine(emptyMessage, width)
		}
		resolved := ResolveColumns(max(1, width), columns)
		t := newTable()
		t.cols = resolved
		t.styles = styles
		t.viewport.SetWidth(max(1, width))
		t.viewport.SetHeight(0)
		t.hideHeader = false
		t.updateViewport()
		header := t.view()
		if header != "" {
			return header + "\n" + util.ConstrainLine(emptyMessage, width)
		}
		return util.ConstrainLine(emptyMessage, width)
	}
	return renderTable(styles, width, height, columns, rows, cursor, hideHeader)
}

// renderTable creates a rendered table with styled rows and cursor.
func renderTable(styles Styles, width, height int, defs []ColumnDef, rows []Row, cursor int, hideHeader bool) string {
	privateRows := make([]tableRow, 0, len(rows))
	for _, row := range rows {
		privateRows = append(privateRows, tableRow(row))
	}
	viewportHeight := max(1, height)
	if !hideHeader {
		viewportHeight = max(1, height-1)
	}
	t := newTable()
	t.cols = defs
	t.rows = privateRows
	t.styles = styles
	t.viewport.SetWidth(max(1, width))
	t.viewport.SetHeight(viewportHeight)
	t.hideHeader = hideHeader
	t.updateViewport()
	if len(rows) > 0 {
		t.setCursor(util.Clamp(cursor, 0, len(rows)-1))
	}
	return t.view()
}
