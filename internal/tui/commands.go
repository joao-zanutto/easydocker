package tui

import (
	"io"
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/logs"

	tea "charm.land/bubbletea/v2"
)

func tickCmd() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) loadContainersCmd() tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		containers, err := svc.LoadContainerRows()
		return containersResultMsg{containers: containers, err: err}
	}
}

func (m model) loadResourcesCmd() tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		snapshot, err := svc.LoadSupportingResources()
		return resourcesResultMsg{snapshot: snapshot, err: err}
	}
}

func (m model) loadMetricsCmd(rows []core.ContainerRow) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		metricsByID, totalCPU, totalMem, err := svc.LoadContainerMetrics(rows)
		return metricsResultMsg{metricsByID: metricsByID, totalCPU: totalCPU, totalMem: totalMem, err: err}
	}
}

func (m model) loadDockerCmd() tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		snapshot, err := svc.LoadSnapshot()
		return loadResultMsg{snapshot: snapshot, err: err}
	}
}

type execShellCommand struct {
	service     *core.Service
	containerID string
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
}

func (e *execShellCommand) SetStdin(r io.Reader)  { e.stdin = r }
func (e *execShellCommand) SetStdout(w io.Writer) { e.stdout = w }
func (e *execShellCommand) SetStderr(w io.Writer) { e.stderr = w }

func (e *execShellCommand) Run() error {
	// Enter alternate screen buffer
	_, _ = io.WriteString(e.stdout, "\033[?1049h")
	defer func() {
		// Exit alternate screen buffer
		_, _ = io.WriteString(e.stdout, "\033[?1049l")
	}()
	return e.service.ExecShell(e.containerID, e.stdin, e.stdout, e.stderr)
}

func (m model) execTerminalCmd(containerID string) tea.Cmd {
	return tea.Exec(
		&execShellCommand{service: m.service, containerID: containerID},
		func(err error) tea.Msg { return execDoneMsg{err: err} },
	)
}

func (m model) loadLogsDataCmd(containerID string, sessionID int, previousCPU, previousMem []float64, tail int, src logs.Source) tea.Cmd {
	svc := m.service
	return func() tea.Msg {
		data, err := svc.LoadContainerLiveData(containerID, previousCPU, previousMem, tail)
		return logs.ResultMsg{ContainerID: containerID, SessionID: sessionID, Data: data, Err: err, Tail: tail, Src: src}
	}
}
