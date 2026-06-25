package tui

import (
	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	"charm.land/bubbles/v2/spinner"
)

type testModelBuilder struct {
	m model
}

func newTestModel() testModelBuilder {
	m := unwrapModel(model{
		dataDirty:    true,
		screen:       shared.Main,
		loadingStage: shared.StageContainers,
		styles:       defaultStyles(),
		spinner:      spinner.New(spinner.WithSpinner(spinner.Points)),
		browse:       browse.NewModel(),
		viewer:       viewer.NewModel(),
		menu:         menu.NewMenuState(),
		help:         menu.NewHelpState(0, 0),
	})
	return testModelBuilder{m: *m}
}

func (b testModelBuilder) withSize(width, height int) testModelBuilder {
	b.m.width = width
	b.m.height = height
	return b
}

func (b testModelBuilder) withLoading(stage shared.Stage) testModelBuilder {
	b.m.loadingStage = stage
	return b
}

func (b testModelBuilder) withContainers(containers ...core.ContainerRow) testModelBuilder {
	b.m.browse.Snapshot.Containers = containers
	b.m.browse.ActiveTab = shared.TabContainers
	b.m.browse.ShowAll = true
	return b
}

func (b testModelBuilder) build() model {
	return b.m
}
