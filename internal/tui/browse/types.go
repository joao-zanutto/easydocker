package browse

import (
	"easydocker/internal/tui/components"
)

const FilterHeaderHeight = components.FilterHeaderHeight

type State struct {
	Filter components.FilterState
}

type FilterState = components.FilterState

func NewFilterState() FilterState {
	return components.NewFilterState()
}