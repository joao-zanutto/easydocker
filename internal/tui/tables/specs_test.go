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
			wantHeader: []string{"REPOSITORY", "TAGS", "SIZE", "CREATED", "IMAGE ID"},
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

	colored := ContainerTableRow(container, 20, "", "")
	if len(colored) != 6 {
		t.Fatalf("container row len = %d, want 6", len(colored))
	}
	if !strings.Contains(colored[1], "\x1b[") {
		t.Fatalf("expected ANSI color in state column, got %q", colored[1])
	}

	narrow := ContainerTableRow(container, 1, "", "")
	if !strings.Contains(narrow[1], "\x1b[") {
		t.Fatalf("expected ANSI color in state column even when narrow, got %q", narrow[1])
	}
	if strings.Contains(narrow[1], "…") {
		t.Fatalf("expected narrow state to fallback to status circle, got %q", narrow[1])
	}
	if !strings.Contains(narrow[1], "●") {
		t.Fatalf("expected narrow state to include status circle, got %q", narrow[1])
	}

	medium := ContainerTableRow(container, 3, "", "")
	if strings.Contains(medium[1], "●\x1b[39m") {
		t.Fatalf("expected medium width to keep full state label before table truncation, got %q", medium[1])
	}
}

func TestBuildContainerSpec_LoadingIndicatorOnlyOnSelectedRow(t *testing.T) {
	items := []core.ContainerRow{
		{FullID: "ctr-1", Name: "api", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"},
		{FullID: "ctr-2", Name: "worker", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"},
	}

	rows := BuildContainerListRows(items, map[string]bool{})
	spec := BuildContainerSpec(80, 1, rows, true, "⠋")
	first := spec.RowBuilder(rows[0])
	second := spec.RowBuilder(rows[1])

	if first[2] != "-" || first[3] != "-" {
		t.Fatalf("non-selected row should not use loading indicator, got cpu=%q mem=%q", first[2], first[3])
	}
	if second[2] != "⠋" || second[3] != "⠋" {
		t.Fatalf("selected row should use loading indicator, got cpu=%q mem=%q", second[2], second[3])
	}
}

func TestBuildContainerSpec_LoadingIndicatorOnlyOnHoveredRowAcrossKinds(t *testing.T) {
	items := []core.ContainerRow{
		{FullID: "ctr-standalone", Name: "standalone", State: "running", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"},
		{FullID: "ctr-compose", Name: "api", State: "running", ComposeProject: "shop", CPUPercent: -1, MemoryUsage: "-", MemoryLimit: "-"},
	}

	rows := BuildContainerListRows(items, map[string]bool{})
	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(rows))
	}

	composeIndex := -1
	containerIndex := -1
	for i, row := range rows {
		switch row.Kind {
		case ContainerListRowComposeProject:
			composeIndex = i
		case ContainerListRowContainer:
			if row.Container.FullID == "ctr-standalone" {
				containerIndex = i
			}
		}
	}
	if composeIndex == -1 || containerIndex == -1 {
		t.Fatalf("expected one compose and one standalone row, got %#v", rows)
	}

	composeSelected := BuildContainerSpec(80, composeIndex, rows, true, "⠋")
	composeRow := composeSelected.RowBuilder(rows[composeIndex])
	containerRow := composeSelected.RowBuilder(rows[containerIndex])
	if composeRow[2] != "\x1b[1m⠋\x1b[22m" || composeRow[3] != "\x1b[1m⠋\x1b[22m" {
		t.Fatalf("selected compose row should show spinner, got cpu=%q mem=%q", composeRow[2], composeRow[3])
	}
	if containerRow[2] != "-" || containerRow[3] != "-" {
		t.Fatalf("non-selected container row should not show spinner, got cpu=%q mem=%q", containerRow[2], containerRow[3])
	}

	containerSelected := BuildContainerSpec(80, containerIndex, rows, true, "⠋")
	composeRow = containerSelected.RowBuilder(rows[composeIndex])
	containerRow = containerSelected.RowBuilder(rows[containerIndex])
	if composeRow[2] != "\x1b[1m-\x1b[22m" || composeRow[3] != "\x1b[1m-\x1b[22m" {
		t.Fatalf("non-selected compose row should not show spinner, got cpu=%q mem=%q", composeRow[2], composeRow[3])
	}
	if containerRow[2] != "⠋" || containerRow[3] != "⠋" {
		t.Fatalf("selected container row should show spinner, got cpu=%q mem=%q", containerRow[2], containerRow[3])
	}
}

func TestBuildContainerListRows_ComposeGroupingAndExpansion(t *testing.T) {
	items := []core.ContainerRow{
		{FullID: "c1", Name: "standalone", State: "running"},
		{FullID: "c2", Name: "z-run", State: "running", ComposeProject: "shop"},
		{FullID: "c3", Name: "a-exit", State: "exited", ComposeProject: "shop"},
	}

	collapsed := BuildContainerListRows(items, map[string]bool{})
	if len(collapsed) != 2 {
		t.Fatalf("collapsed rows len = %d, want 2", len(collapsed))
	}
	if collapsed[0].Kind != ContainerListRowContainer || collapsed[0].Container.FullID != "c1" {
		t.Fatalf("collapsed first row = %#v, want standalone container", collapsed[0])
	}
	if collapsed[1].Kind != ContainerListRowComposeProject {
		t.Fatalf("collapsed second row kind = %v, want compose project", collapsed[1].Kind)
	}
	if collapsed[1].ComposeProject.Name != "shop" || collapsed[1].ComposeProject.ContainerCount != 2 || collapsed[1].ComposeProject.RunningCount != 1 {
		t.Fatalf("collapsed project summary = %#v, want shop with 2 containers and 1 running", collapsed[1])
	}

	expanded := BuildContainerListRows(items, map[string]bool{"shop": true})
	if len(expanded) != 4 {
		t.Fatalf("expanded rows len = %d, want 4", len(expanded))
	}
	if expanded[1].Kind != ContainerListRowComposeProject || !expanded[1].ComposeExpanded {
		t.Fatalf("expanded project row = %#v, want expanded compose project", expanded[1])
	}
	if expanded[2].Kind != ContainerListRowContainer || expanded[2].Container.FullID != "c2" {
		t.Fatalf("expanded first child row = %#v, want c2", expanded[2])
	}
	if expanded[2].TreePrefix != "├─ " {
		t.Fatalf("expanded first child prefix = %q, want %q", expanded[2].TreePrefix, "├─ ")
	}
	if expanded[3].Kind != ContainerListRowContainer || expanded[3].Container.FullID != "c3" {
		t.Fatalf("expanded second child row = %#v, want c3", expanded[3])
	}
	if expanded[3].TreePrefix != "└─ " {
		t.Fatalf("expanded second child prefix = %q, want %q", expanded[3].TreePrefix, "└─ ")
	}
}

func TestBuildContainerSpec_LoadingIndicatorOnlyOnHoveredComposeRow(t *testing.T) {
	items := []core.ContainerRow{
		{FullID: "shop-1", Name: "shop-api", State: "running", ComposeProject: "shop"},
		{FullID: "blog-1", Name: "blog-api", State: "running", ComposeProject: "blog"},
	}

	rows := BuildContainerListRows(items, map[string]bool{})
	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(rows))
	}

	spec := BuildContainerSpec(80, 0, rows, true, "⠋")
	first := spec.RowBuilder(rows[0])
	second := spec.RowBuilder(rows[1])

	if first[2] != "\x1b[1m⠋\x1b[22m" || first[3] != "\x1b[1m⠋\x1b[22m" {
		t.Fatalf("hovered compose row should show loading indicator, got cpu=%q mem=%q", first[2], first[3])
	}
	if second[2] != "\x1b[1m-\x1b[22m" || second[3] != "\x1b[1m-\x1b[22m" {
		t.Fatalf("non-hovered compose row should not show loading indicator, got cpu=%q mem=%q", second[2], second[3])
	}
}

func TestComposeProjectTableRow_ShowsCollapsedState(t *testing.T) {
	row := ComposeProjectTableRow(ContainerListRow{
		Kind: ContainerListRowComposeProject,
		ComposeProject: core.ComposeProject{
			Name:           "shop",
			ContainerCount: 3,
			RunningCount:   2,
			Created:        "just now",
			CPUPercent:     12.5,
			MemoryUsage:    "100 MiB",
			MemoryLimit:    "400 MiB",
			MemoryPercent:  25,
		},
		ComposeExpanded: false,
	}, "")

	if row[0] != "\x1b[1m[+] shop\x1b[22m" {
		t.Fatalf("compose name column = %q, want bold [+] prefix", row[0])
	}
	if row[1] != "\x1b[1m2/3 running\x1b[22m" {
		t.Fatalf("compose state column = %q, want bold state", row[1])
	}
	if row[2] != "\x1b[1m12.5%\x1b[22m" || row[3] != "\x1b[1m100 MiB (25.0%)\x1b[22m" {
		t.Fatalf("compose metrics columns = %#v, want cpu/memory aggregation", row[2:4])
	}
	if row[4] != "\x1b[1m-\x1b[22m" || row[5] != "\x1b[1mjust now\x1b[22m" {
		t.Fatalf("compose image/status columns = %#v, want image dash and created time", row[4:6])
	}
}

func TestImageTableRow_SplitsRepositoryAndTags(t *testing.T) {
	row := ImageTableRow(core.ImageRow{Tags: "nginx:latest, redis:alpine", Size: "1.0 KiB", Created: "just now", ID: "img-1"})

	if len(row) != 5 {
		t.Fatalf("image row len = %d, want 5", len(row))
	}
	if row[0] != "nginx, redis" {
		t.Fatalf("repository column = %q, want %q", row[0], "nginx, redis")
	}
	if row[1] != "latest, alpine" {
		t.Fatalf("tags column = %q, want %q", row[1], "latest, alpine")
	}
	if row[2] != "1.0 KiB" || row[3] != "just now" || row[4] != "img-1" {
		t.Fatalf("tail columns = %#v, want size/created/id preserved", row[2:])
	}
}
