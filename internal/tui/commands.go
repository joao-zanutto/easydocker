package tui

import (
	"context"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"

	tea "charm.land/bubbletea/v2"
)

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func loadContainersCmd(svc core.ServiceInterface) tea.Cmd {
	return func() tea.Msg {
		containers, err := svc.LoadContainerRows(context.Background())
		return containersResultMsg{containers: containers, err: err}
	}
}

func loadResourcesCmd(svc core.ServiceInterface) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := svc.LoadSupportingResources(context.Background())
		return resourcesResultMsg{snapshot: snapshot, err: err}
	}
}

func loadMetricsCmd(svc core.ServiceInterface, rows []core.ContainerRow) tea.Cmd {
	return func() tea.Msg {
		metricsByID, totalCPU, totalMem, err := svc.LoadContainerMetrics(context.Background(), rows)
		return metricsResultMsg{metricsByID: metricsByID, totalCPU: totalCPU, totalMem: totalMem, err: err}
	}
}

func loadDockerCmd(svc core.ServiceInterface) tea.Cmd {
	return func() tea.Msg {
		snapshot, err := svc.LoadSnapshot(context.Background())
		return loadResultMsg{snapshot: snapshot, err: err}
	}
}

func shellCmd(svc core.ServiceInterface, containerID string) tea.Cmd {
	return shared.ShellCmd(svc, containerID, shellDoneMsg{})
}
