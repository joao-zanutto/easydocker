package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

func (m model) handleContainersResultMsg(msg containersResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setLoadError(msg.err)
		return m, nil
	}

	m.snapshot.Containers = core.PreserveRunningContainerMetrics(msg.containers, m.snapshot.Containers)
	m.beginLoadingStage(loadStageResources)
	return m, loadResourcesCmd(m.service)
}

func (m model) handleResourcesResultMsg(msg resourcesResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setLoadError(msg.err)
		return m, nil
	}

	m.snapshot.Images = msg.snapshot.Images
	m.snapshot.Networks = msg.snapshot.Networks
	m.snapshot.Volumes = msg.snapshot.Volumes
	m.snapshot.TotalLimit = msg.snapshot.TotalLimit
	m.beginLoadingStage(loadStageMetrics)
	return m, loadMetricsCmd(m.service, m.snapshot.Containers)
}

func (m model) handleMetricsResultMsg(msg metricsResultMsg) (tea.Model, tea.Cmd) {
	if !m.finishLoadingStage(msg.err) {
		return m, nil
	}

	m.snapshot.Containers = core.ApplyMetricsToContainers(m.snapshot.Containers, msg.metricsByID)
	m.snapshot.TotalCPU = msg.totalCPU
	m.snapshot.TotalMem = msg.totalMem
	m.snapshot.Timestamp = time.Now()
	m.metricsLoaded = true
	m.clampCursors()
	return m, nil
}

func (m model) handleLoadResultMsg(msg loadResultMsg) (tea.Model, tea.Cmd) {
	if !m.finishLoadingStage(msg.err) {
		return m, nil
	}

	previousContainers := m.snapshot.Containers
	m.snapshot = msg.snapshot
	m.snapshot.Containers = core.PreserveRunningContainerMetrics(m.snapshot.Containers, previousContainers)
	if err := m.reconcileLogsSelection(); err != nil {
		m.err = err
	}
	m.clampCursors()
	return m, nil
}

func (m *model) applyLoadingTransition(transition shared.Transition) {
	m.loading = transition.Loading
	m.loadingStage = int(transition.Stage)
	m.err = transition.Err
}

func (m *model) setLoadError(err error) {
	m.applyLoadingTransition(shared.Fail(err))
}

func (m *model) beginLoadingStage(stage int) {
	m.applyLoadingTransition(shared.Begin(shared.Stage(stage)))
	m.snapshot.Timestamp = time.Now()
	m.clampCursors()
}

func (m *model) finishLoadingStage(err error) bool {
	transition, ok := shared.Finish(err)
	m.applyLoadingTransition(transition)
	return ok
}

func (m model) shouldReloadSnapshotOnTick() bool {
	return m.loadingStage == loadStageIdle
}
