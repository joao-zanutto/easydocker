package tables

// ContainerSchema defines column layout for containers table.
var ContainerSchema = []ColumnDef{
	{Header: "NAME", MinWidth: 10, Desired: proportionalWidth(24, 3)},
	{Header: "STATE", MinWidth: 12, Desired: fixedWidth(12)},
	{Header: "CPU", MinWidth: 9, Desired: fixedWidth(9)},
	{Header: "MEMORY", MinWidth: 12, Desired: fixedWidth(12)},
	{Header: "IMAGE", MinWidth: 12, Desired: proportionalWidth(22, 4)},
	{Header: "STATUS", MinWidth: 24, Desired: fixedWidth(24), PinnedRight: true},
}

// ImageSchema defines column layout for images table.
var ImageSchema = []ColumnDef{
	{Header: "REPOSITORY", MinWidth: 18, Desired: proportionalWidth(24, 2)},
	{Header: "TAGS", MinWidth: 16, Desired: proportionalWidth(16, 4)},
	{Header: "SIZE", MinWidth: 10, Desired: fixedWidth(10), PinnedRight: true},
	{Header: "CREATED", MinWidth: 9, Desired: fixedWidth(9), PinnedRight: true},
	{Header: "IMAGE ID", MinWidth: 16, Desired: fixedWidth(16), PinnedRight: true},
}

// NetworkSchema defines column layout for networks table.
var NetworkSchema = []ColumnDef{
	{Header: "NAME", MinWidth: 18, Desired: proportionalWidth(18, 4)},
	{Header: "DRIVER", MinWidth: 8, Desired: fixedWidth(8), PinnedRight: true},
	{Header: "ENDPOINTS", MinWidth: 9, Desired: fixedWidth(9), PinnedRight: true},
	{Header: "CREATED", MinWidth: 10, Desired: fixedWidth(10), PinnedRight: true},
}

// VolumeSchema defines column layout for volumes table.
var VolumeSchema = []ColumnDef{
	{Header: "NAME", MinWidth: 18, Desired: proportionalWidth(22, 3)},
	{Header: "MOUNTPOINT", MinWidth: 14, Desired: proportionalWidth(14, 3)},
	{Header: "SIZE", MinWidth: 10, Desired: fixedWidth(10), PinnedRight: true},
	{Header: "CREATED", MinWidth: 10, Desired: fixedWidth(10), PinnedRight: true},
}

// fixedWidth returns a width function that always returns the same value.
func fixedWidth(width int) func(int) int {
	return func(int) int {
		return width
	}
}

// proportionalWidth returns a width function that scales with table width.
func proportionalWidth(minWidth, divisor int) func(int) int {
	return func(tableWidth int) int {
		return max(minWidth, tableWidth/divisor)
	}
}

// ContainerStateColumnWidth extracts the STATE column width.
func ContainerStateColumnWidth(columns []ColumnDef) int {
	if len(columns) <= 1 {
		return 0
	}
	return columns[1].MinWidth
}

// SimpleSpec builds a resource table spec with standard parameters.
func SimpleSpec[T any](
	width int,
	emptyMessage string,
	cursor int,
	items []T,
	columnsForWidth func(int) []ColumnDef,
	rowBuilder func(T) []string,
) Spec[T] {
	tableWidth := ContentWidth(width)
	return Spec[T]{
		EmptyMessage: emptyMessage,
		Cursor:       cursor,
		Items:        items,
		Columns:      columnsForWidth(tableWidth),
		RowBuilder:   rowBuilder,
	}
}

// ContentWidth returns the normalized table width provided by the caller.
func ContentWidth(width int) int {
	return max(1, width)
}
