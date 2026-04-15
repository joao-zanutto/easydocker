package tables

import (
	"strings"
	"testing"

	"easydocker/internal/core"
)

func TestTableColumnSchemas(t *testing.T) {
	tableWidth := 120

	tests := []struct {
		name       string
		columns    []ColumnDef
		wantHeader []string
	}{
		{
			name:       "containers",
			columns:    ContainerColumns(tableWidth),
			wantHeader: []string{"NAME", "STATE", "CPU", "MEMORY", "IMAGE", "STATUS"},
		},
		{
			name:       "images",
			columns:    ImageColumns(tableWidth),
			wantHeader: []string{"REPOSITORY/TAGS", "SIZE", "CREATED", "IMAGE ID"},
		},
		{
			name:       "networks",
			columns:    NetworkColumns(tableWidth),
			wantHeader: []string{"NAME", "DRIVER", "SCOPE", "ENDPOINTS", "META"},
		},
		{
			name:       "volumes",
			columns:    VolumeColumns(tableWidth),
			wantHeader: []string{"NAME", "DRIVER", "SCOPE", "SIZE", "REFS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.columns) != len(tt.wantHeader) {
				t.Fatalf("columns len = %d, want %d", len(tt.columns), len(tt.wantHeader))
			}

			totalWidth := 0
			for i, col := range tt.columns {
				if col.Header != tt.wantHeader[i] {
					t.Fatalf("header[%d] = %q, want %q", i, col.Header, tt.wantHeader[i])
				}
				if col.MinWidth <= 0 {
					t.Fatalf("minWidth[%d] = %d, want > 0", i, col.MinWidth)
				}
				totalWidth += col.MinWidth
			}
			totalWidth += (len(tt.columns) - 1) * 2
			if totalWidth > tableWidth {
				t.Fatalf("resolved columns width = %d, want <= %d", totalWidth, tableWidth)
			}
		})
	}
}

func TestContainerStateColumnWidth(t *testing.T) {
	if got := ContainerStateColumnWidth(nil); got != 0 {
		t.Fatalf("ContainerStateColumnWidth(nil) = %d, want 0", got)
	}
	if got := ContainerStateColumnWidth([]ColumnDef{{Header: "NAME", MinWidth: 10}}); got != 0 {
		t.Fatalf("ContainerStateColumnWidth(one column) = %d, want 0", got)
	}
	if got := ContainerStateColumnWidth([]ColumnDef{{Header: "NAME", MinWidth: 10}, {Header: "STATE", MinWidth: 12}}); got != 12 {
		t.Fatalf("ContainerStateColumnWidth(two columns) = %d, want 12", got)
	}
}

func TestSimpleResourceTableSpec_UsesSharedWidthAndFields(t *testing.T) {
	seenWidth := 0
	columns := []ColumnDef{{Header: "VALUE", MinWidth: 7, Desired: func(int) int { return 7 }}}
	spec := SimpleSpec(
		3,
		"No rows.",
		2,
		[]int{1, 2, 3},
		func(width int) []ColumnDef {
			seenWidth = width
			return columns
		},
		func(int) []string {
			return []string{"value"}
		},
	)

	if seenWidth != 3 {
		t.Fatalf("columns width arg = %d, want 3", seenWidth)
	}
	if spec.EmptyMessage != "No rows." {
		t.Fatalf("emptyMessage = %q, want %q", spec.EmptyMessage, "No rows.")
	}
	if spec.Cursor != 2 {
		t.Fatalf("cursor = %d, want 2", spec.Cursor)
	}
	if len(spec.Items) != 3 {
		t.Fatalf("items len = %d, want 3", len(spec.Items))
	}
	if len(spec.Columns) != 1 {
		t.Fatalf("columns len = %d, want 1", len(spec.Columns))
	}
	row := spec.RowBuilder(99)
	if len(row) != 1 || row[0] != "value" {
		t.Fatalf("rowBuilder output = %#v, want [\"value\"]", row)
	}
}

func TestContainerTableSpec_EmptyMessageHintByActiveTab(t *testing.T) {
	withHint := BuildContainerSpec(80, 0, nil, true, "")
	wantWithHint := "No containers found. Press a to switch between running and all containers."
	if withHint.EmptyMessage != wantWithHint {
		t.Fatalf("container empty message = %q, want %q", withHint.EmptyMessage, wantWithHint)
	}

	withoutHint := BuildContainerSpec(80, 0, nil, false, "")
	if withoutHint.EmptyMessage != "No containers found." {
		t.Fatalf("container empty message outside containers tab = %q, want %q", withoutHint.EmptyMessage, "No containers found.")
	}
}

func TestContainerTableRow_StateColoringByWidth(t *testing.T) {
	container := core.ContainerRow{
		Name:          "api",
		State:         "running",
		Healthy:       true,
		CPUPercent:    12.5,
		MemoryUsage:   "100 MiB",
		MemoryPercent: 25,
		MemoryLimit:   "400 MiB",
		Image:         "nginx:latest",
		Status:        "Up 2m",
	}

	colored := ContainerTableRow(container, 20, "")
	if len(colored) != 6 {
		t.Fatalf("container row len = %d, want 6", len(colored))
	}
	if !strings.Contains(colored[1], "\x1b[") {
		t.Fatalf("expected ANSI color in state column, got %q", colored[1])
	}

	narrow := ContainerTableRow(container, 1, "")
	if !strings.Contains(narrow[1], "\x1b[") {
		t.Fatalf("expected ANSI color in state column even when narrow, got %q", narrow[1])
	}
	if strings.Contains(narrow[1], "…") {
		t.Fatalf("expected narrow state to fallback to status circle, got %q", narrow[1])
	}
	if !strings.Contains(narrow[1], "●") {
		t.Fatalf("expected narrow state to include status circle, got %q", narrow[1])
	}

	medium := ContainerTableRow(container, 3, "")
	if strings.Contains(medium[1], "●\x1b[39m") {
		t.Fatalf("expected medium width to keep full state label before table truncation, got %q", medium[1])
	}
}
