package btable

import (
	"strings"
	"testing"
)

func TestRenderRow_ColoredCellDoesNotSpillIntoNextColumn(t *testing.T) {
	m := New(
		WithColumns([]Column{{Title: "STATE", Width: 4}, {Title: "STATUS", Width: 6}}),
		WithRows([]Row{{"\x1b[32mrunning\x1b[39m", "ok"}}),
		WithWidth(20),
		WithHeight(2),
	)

	row := m.renderRow(0)
	statusIndex := strings.Index(row, "ok")
	if statusIndex == -1 {
		t.Fatalf("expected rendered row to include status column, got %q", row)
	}

	resetIndex := strings.LastIndex(row[:statusIndex], "\x1b[39m")
	if resetIndex == -1 {
		t.Fatalf("expected foreground reset before next column text, got %q", row)
	}

	ellipsisIndex := strings.Index(row, "…")
	if ellipsisIndex == -1 {
		t.Fatalf("expected truncated state to include ellipsis, got %q", row)
	}
	if ellipsisIndex > resetIndex {
		t.Fatalf("expected ellipsis to be colorized before reset, got %q", row)
	}
}
