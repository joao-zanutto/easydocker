package tui

import (
	"easydocker/internal/tui/shared"
)

func (m model) shouldAnimateLogsLoadingIndicator() bool {
	return (m.screen == shared.Logs || m.screen == shared.Inspect) && (m.viewer.InitialLoad || m.viewer.HistoryLoad)
}

func (m model) shouldAnimateMetricsLoadingIndicator() bool {
	return !m.metricsLoaded && m.loadingStage != shared.StageIdle
}

func (m model) shouldReloadSnapshotOnTick() bool {
	return m.loadingStage == shared.StageIdle
}

func (m model) shouldLoadHistoryOnTick() bool {
	return m.screen == shared.Logs &&
		m.viewer.ContainerID != "" &&
		m.viewer.Viewport.AtTop() &&
		!m.viewer.InitialLoad &&
		!m.viewer.HistoryLoad &&
		!m.viewer.HistoryDone
}

func (m model) shouldPollLogsOnTick() bool {
	return m.screen == shared.Logs &&
		m.viewer.ContainerID != "" &&
		!m.viewer.Viewport.AtTop() &&
		!m.viewer.InitialLoad &&
		!m.viewer.HistoryLoad
}

func (m model) logsPollTail() int {
	if m.viewer.TailLines <= 0 {
		return InitialTail
	}
	return m.viewer.TailLines
}
