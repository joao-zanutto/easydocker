package tui

import (
	"easydocker/internal/core"
	"easydocker/internal/tui/shared"
)

type testModelBuilder struct {
	m model
}

func newTestModel() testModelBuilder {
	m := unwrapModel(New(nil, nil, ""))
	return testModelBuilder{m: *m}
}

func (b testModelBuilder) withSize(width, height int) testModelBuilder {
	b.m.width = width
	b.m.height = height
	return b
}

func (b testModelBuilder) withLoading(loading bool, stage shared.Stage) testModelBuilder {
	b.m.loading = loading
	b.m.loadingStage = stage
	return b
}

func (b testModelBuilder) withContainers(containers ...core.ContainerRow) testModelBuilder {
	b.m.browse.Snapshot.Containers = containers
	b.m.browse.ActiveTab = tabContainers
	b.m.browse.ShowAll = true
	return b
}

func (b testModelBuilder) build() model {
	return b.m
}
