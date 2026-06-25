package browse

import (
	"easydocker/internal/tui/ui/components"
)

type State struct {
	Filter components.FilterState
}

type FilterState = components.FilterState

func NewFilterState() FilterState {
	return components.NewFilterState()
}
