package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

func (m *model) handleContainersResultMsg(msg containersResultMsg) (tea.Model, tea.Cmd) {
	m.dataDirty = true
	if msg.err != nil {
		m.setLoadError(msg.err)
		return m, nil
	}

	m.browse.Snapshot.Containers = core.PreserveRunningContainerMetrics(msg.containers, m.browse.Snapshot.Containers)
	m.loadingStage = m.loadingStage.Next()
	return m, loadResourcesCmd(m.service)
}

func (m *model) handleResourcesResultMsg(msg resourcesResultMsg) (tea.Model, tea.Cmd) {
	m.dataDirty = true
	if msg.err != nil {
		m.setLoadError(msg.err)
		return m, nil
	}

	m.browse.Snapshot.Images = core.ApplyImageContainerCounts(msg.snapshot.Images, m.browse.Snapshot.Containers)
	m.browse.Snapshot.Volumes = msg.snapshot.Volumes
	m.browse.Snapshot.TotalLimit = msg.snapshot.TotalLimit

	m.browse.Snapshot.Networks = core.ApplyNetworkEndpointCounts(msg.snapshot.Networks, m.browse.Snapshot.Containers)

	m.loadingStage = m.loadingStage.Next()
	return m, loadMetricsCmd(m.service, m.browse.Snapshot.Containers)
}

func (m *model) handleMetricsResultMsg(msg metricsResultMsg) (tea.Model, tea.Cmd) {
	m.dataDirty = true
	if !m.finishLoadingStage(msg.err) {
		return m, nil
	}

	m.browse.Snapshot.Containers = core.ApplyMetricsToContainers(m.browse.Snapshot.Containers, msg.metricsByID)
	m.browse.Snapshot.TotalCPU = msg.totalCPU
	m.browse.Snapshot.TotalMem = msg.totalMem
	m.browse.Snapshot.Timestamp = time.Now()
	m.metricsLoaded = true
	return m, tickCmd()
}

func (m *model) handleLoadResultMsg(msg loadResultMsg) (tea.Model, tea.Cmd) {
	m.snapshotInflight = false
	m.dataDirty = true
	if !m.finishLoadingStage(msg.err) {
		return m, nil
	}

	previousContainers := m.browse.Snapshot.Containers
	m.browse.Snapshot = msg.snapshot
	m.browse.Snapshot.Containers = core.PreserveRunningContainerMetrics(m.browse.Snapshot.Containers, previousContainers)
	recomputeSnapshotTotals(&m.browse.Snapshot)
	if err := m.reconcileLogsSelection(); err != nil {
		m.err = err
	}
	return m, nil
}

func recomputeSnapshotTotals(snapshot *core.Snapshot) {
	var totalCPU float64
	var totalMem uint64
	for i := range snapshot.Containers {
		c := &snapshot.Containers[i]
		if c.State != core.StateRunning {
			continue
		}
		if c.CPUPercent > 0 {
			totalCPU += c.CPUPercent
		}
		totalMem += c.MemoryUsageBytes
	}
	snapshot.TotalCPU = totalCPU
	snapshot.TotalMem = totalMem
}

func (m *model) handleViewerContentMsg(msg viewer.ContentMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewer, cmd = m.viewer.Update(msg)
	return m, cmd
}

func (m *model) handleInspectResultMsg(msg inspectResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.viewer.Vp.InitialLoad = false
		return m, nil
	}
	m.viewer.ApplyInitial(msg.data)
	m.viewer.Vp.SyncFromData(m.viewer.VisibleWidth(), m.viewer.VisibleRows())
	return m, nil
}

func (m *model) applyLoadingTransition(transition shared.Transition) {
	m.loading = transition.Loading
	m.loadingStage = transition.Stage
	m.err = transition.Err
}

func (m *model) setLoadError(err error) {
	m.applyLoadingTransition(shared.Fail(err))
}

func (m *model) finishLoadingStage(err error) bool {
	transition, ok := shared.Finish(err)
	m.applyLoadingTransition(transition)
	return ok
}
