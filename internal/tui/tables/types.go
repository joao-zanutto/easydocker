package tables

// ColumnDef defines a table column specification.
type ColumnDef struct {
	Header   string
	MinWidth int
	Desired  func(tableWidth int) int
}

// Spec defines table data and rendering properties for a resource type.
type Spec[T any] struct {
	EmptyMessage string
	Cursor       int
	Items        []T
	Columns      []ColumnDef
	RowBuilder   func(T) []string
}
