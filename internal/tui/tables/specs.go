package tables

// ContainerSchema defines column layout for containers table.
var ContainerSchema = []ColumnDef{
	{Header: "NAME", MinWidth: 10, Desired: proportionalWidth(10, 6)},
	{Header: "STATE", MinWidth: 12, Desired: fixedWidth(12)},
	{Header: "CPU", MinWidth: 7, Desired: fixedWidth(7)},
	{Header: "MEMORY", MinWidth: 20, Desired: fixedWidth(20)},
	{Header: "IMAGE", MinWidth: 12, Desired: proportionalWidth(12, 5)},
	{Header: "STATUS", MinWidth: 8, Desired: proportionalWidth(8, 4)},
}

// ImageSchema defines column layout for images table.
var ImageSchema = []ColumnDef{
	{Header: "REPOSITORY/TAGS", MinWidth: 24, Desired: proportionalWidth(24, 2)},
	{Header: "SIZE", MinWidth: 10, Desired: fixedWidth(10)},
	{Header: "CREATED", MinWidth: 12, Desired: fixedWidth(12)},
	{Header: "IMAGE ID", MinWidth: 12, Desired: proportionalWidth(12, 5)},
}

// NetworkSchema defines column layout for networks table.
var NetworkSchema = []ColumnDef{
	{Header: "NAME", MinWidth: 18, Desired: proportionalWidth(18, 4)},
	{Header: "DRIVER", MinWidth: 10, Desired: proportionalWidth(10, 6)},
	{Header: "SCOPE", MinWidth: 10, Desired: fixedWidth(10)},
	{Header: "ENDPOINTS", MinWidth: 10, Desired: fixedWidth(10)},
	{Header: "META", MinWidth: 18, Desired: proportionalWidth(18, 4)},
}

// VolumeSchema defines column layout for volumes table.
var VolumeSchema = []ColumnDef{
	{Header: "NAME", MinWidth: 18, Desired: proportionalWidth(18, 4)},
	{Header: "DRIVER", MinWidth: 10, Desired: proportionalWidth(10, 6)},
	{Header: "SCOPE", MinWidth: 10, Desired: fixedWidth(10)},
	{Header: "SIZE", MinWidth: 10, Desired: fixedWidth(10)},
	{Header: "REFS", MinWidth: 8, Desired: fixedWidth(8)},
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
