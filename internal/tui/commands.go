package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadContainersCmd(svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		containers, err := svc.LoadContainerRows()
		return containersResultMsg{containers: containers, err: err}
	}
}

func loadResourcesCmd(svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := svc.LoadSupportingResources()
		return resourcesResultMsg{snapshot: snapshot, err: err}
	}
}

func loadMetricsCmd(svc *core.Service, rows []core.ContainerRow) tea.Cmd {
	return func() tea.Msg {
		metricsByID, totalCPU, totalMem, err := svc.LoadContainerMetrics(rows)
		return metricsResultMsg{metricsByID: metricsByID, totalCPU: totalCPU, totalMem: totalMem, err: err}
	}
}

func loadDockerCmd(svc *core.Service) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := svc.LoadSnapshot()
		return loadResultMsg{snapshot: snapshot, err: err}
	}
}

func shellCmd(svc *core.Service, containerID string) tea.Cmd {
	return shared.ShellCmd(svc, containerID, shellDoneMsg{})
}
