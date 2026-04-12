package tables

import (
	"strconv"
	"strings"
	"testing"
)

func TestTableContentWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{name: "negative", width: -2, want: 1},
		{name: "zero", width: 0, want: 1},
		{name: "small", width: 3, want: 3},
		{name: "wide", width: 20, want: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContentWidth(tt.width); got != tt.want {
				t.Fatalf("ContentWidth(%d) = %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}

func TestRenderTableOrEmpty_EmptyRows(t *testing.T) {
	spec := Spec[int]{
		EmptyMessage: "No rows available.",
		Columns:      []ColumnDef{{Header: "NAME", MinWidth: 10, Desired: func(int) int { return 10 }}},
		Items:        nil,
		RowBuilder:   func(int) []string { return nil },
	}

	view := RenderFromSpec(80, 6, spec, DefaultStyles())
	if view == "" {
		t.Fatalf("RenderFromSpec() returned empty view")
	}
	if !strings.Contains(view, "No rows available.") {
		t.Fatalf("expected empty message to be rendered, got %q", view)
	}
}

func TestRenderFromSpec_CursorClampBehavior(t *testing.T) {
	columns := []ColumnDef{{Header: "VALUE", MinWidth: 10, Desired: func(int) int { return 10 }}}

	specClamped := Spec[int]{
		EmptyMessage: "No rows",
		Cursor:       99,
		Items:        []int{1, 2},
		Columns:      columns,
		RowBuilder: func(v int) []string {
			return []string{strconv.Itoa(v)}
		},
	}
	specExpected := specClamped
	specExpected.Cursor = 1

	viewClamped := RenderFromSpec(80, 6, specClamped, DefaultStyles())
	viewExpected := RenderFromSpec(80, 6, specExpected, DefaultStyles())

	if viewClamped != viewExpected {
		t.Fatalf("expected clamped cursor view to match explicit cursor view")
	}
}
