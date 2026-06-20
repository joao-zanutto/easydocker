package tui

import (
	"easydocker/internal/tui/shared"
)

func (m model) shouldAnimateMetricsLoadingIndicator() bool {
	return !m.metricsLoaded && m.loadingStage != shared.StageIdle
}

func (m *model) shouldReloadSnapshotOnTick() bool {
	return m.loadingStage == shared.StageIdle
}

func (m model) shouldLoadHistoryOnTick() bool {
	return m.screen == shared.LogViewer &&
		m.viewer.ContainerID != "" &&
		m.viewer.Vp.AtTop() &&
		!m.viewer.Vp.InitialLoad &&
		!m.viewer.Logs.HistoryLoad &&
		!m.viewer.Logs.HistoryDone
}

func (m model) shouldPollLogsOnTick() bool {
	return m.screen == shared.LogViewer &&
		m.viewer.ContainerID != "" &&
		!m.viewer.Vp.AtTop() &&
		!m.viewer.Vp.InitialLoad &&
		!m.viewer.Logs.HistoryLoad
}

func (m *model) logsPollTail() int {
	if m.viewer.Logs.TailLines <= 0 {
		return InitialTail
	}
	return m.viewer.Logs.TailLines
}
