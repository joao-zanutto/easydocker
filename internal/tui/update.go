package tui

import (
	"fmt"
	"strings"
	"time"

	"easydocker/internal/core"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg.String())
	case containersResultMsg:
		m.err = msg.err
		if msg.err != nil {
			m.loading = false
			m.loadingStage = loadStageIdle
			return m, nil
		}
		m.snapshot.Containers = msg.containers
		m.snapshot.Timestamp = time.Now()
		m.loading = true
		m.loadingStage = loadStageResources
		m.clampCursors()
		return m, m.loadResourcesCmd()
	case resourcesResultMsg:
		m.err = msg.err
		if msg.err != nil {
			m.loading = false
			m.loadingStage = loadStageIdle
			return m, nil
		}
		m.snapshot.Images = msg.snapshot.Images
		m.snapshot.Networks = msg.snapshot.Networks
		m.snapshot.Volumes = msg.snapshot.Volumes
		m.snapshot.TotalLimit = msg.snapshot.TotalLimit
		m.snapshot.Timestamp = time.Now()
		m.loading = true
		m.loadingStage = loadStageMetrics
		m.clampCursors()
		return m, m.loadMetricsCmd(m.snapshot.Containers)
	case metricsResultMsg:
		m.err = msg.err
		m.loading = false
		m.loadingStage = loadStageIdle
		if msg.err != nil {
			return m, nil
		}
		m.snapshot.Containers = core.ApplyMetricsToContainers(m.snapshot.Containers, msg.metricsByID)
		m.snapshot.TotalCPU = msg.totalCPU
		m.snapshot.TotalMem = msg.totalMem
		m.snapshot.Timestamp = time.Now()
		m.clampCursors()
		return m, nil
	case loadResultMsg:
		m.loading = false
		m.loadingStage = loadStageIdle
		m.err = msg.err
		if msg.err == nil {
			m.snapshot = msg.snapshot
			if m.screen == screenModeLogs {
				if index, ok := m.findContainerIndexByID(m.logs.containerID); ok {
					m.containerCursor = index
				} else {
					m.exitLogsMode()
					m.err = fmt.Errorf("selected container is no longer available")
				}
			}
			m.clampCursors()
		}
		return m, nil
	case logsResultMsg:
		return m, m.handleLogsResult(msg)
	case tickMsg:
		cmds := []tea.Cmd{tickCmd()}
		if m.loadingStage == loadStageIdle {
			cmds = append(cmds, m.loadDockerCmd())
		}
		if m.screen == screenModeLogs && m.logs.containerID != "" {
			tail := m.logs.tailLines
			if tail <= 0 {
				tail = logsInitialTail
			}
			cmds = append(cmds, m.loadLogsDataCmd(m.logs.containerID, m.logs.sessionID, m.logs.data.CPUHistory, m.logs.data.MemHistory, tail, "poll"))
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q", "tab":
		return m, nil
	}

	if m.screen == screenModeLogs {
		return m, m.handleLogsKey(key)
	}

	switch key {
	case "right", "l":
		m.activeTab = min(tabVolumes, m.activeTab+1)
	case "left", "h":
		m.activeTab = max(tabContainers, m.activeTab-1)
	case "a":
		if m.activeTab == tabContainers {
			m.showAll = !m.showAll
			m.clampCursors()
		}
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "pgup":
		m.moveCursor(-5)
	case "pgdown":
		m.moveCursor(5)
	case "enter":
		if m.activeTab == tabContainers {
			if container, ok := m.selectedContainer(); ok {
				return m, m.enterLogsMode(container)
			}
		}
	case "esc", "backspace":
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) moveCursor(delta int) {
	switch m.activeTab {
	case tabContainers:
		m.containerCursor = clamp(m.containerCursor+delta, 0, max(0, len(m.filteredContainers())-1))
	case tabImages:
		m.imageCursor = clamp(m.imageCursor+delta, 0, max(0, len(m.snapshot.Images)-1))
	case tabNetworks:
		m.networkCursor = clamp(m.networkCursor+delta, 0, max(0, len(m.snapshot.Networks)-1))
	case tabVolumes:
		m.volumeCursor = clamp(m.volumeCursor+delta, 0, max(0, len(m.snapshot.Volumes)-1))
	}
}

func (m *model) clampCursors() {
	m.containerCursor = clamp(m.containerCursor, 0, max(0, len(m.filteredContainers())-1))
	m.imageCursor = clamp(m.imageCursor, 0, max(0, len(m.snapshot.Images)-1))
	m.networkCursor = clamp(m.networkCursor, 0, max(0, len(m.snapshot.Networks)-1))
	m.volumeCursor = clamp(m.volumeCursor, 0, max(0, len(m.snapshot.Volumes)-1))
}

func (m model) filteredContainers() []core.ContainerRow {
	return core.FilterContainersByScope(m.snapshot.Containers, m.showAll)
}

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

func (m model) findContainerIndexByID(id string) (int, bool) {
	for index, container := range m.filteredContainers() {
		if container.FullID == id {
			return index, true
		}
	}
	return 0, false
}

func (m model) selectedContainer() (core.ContainerRow, bool) {
	containers := m.filteredContainers()
	if len(containers) == 0 {
		return core.ContainerRow{}, false
	}
	return containers[m.containerCursor], true
}

func (m model) selectedImage() (core.ImageRow, bool) {
	if len(m.snapshot.Images) == 0 {
		return core.ImageRow{}, false
	}
	return m.snapshot.Images[m.imageCursor], true
}

func (m model) selectedNetwork() (core.NetworkRow, bool) {
	if len(m.snapshot.Networks) == 0 {
		return core.NetworkRow{}, false
	}
	return m.snapshot.Networks[m.networkCursor], true
}

func (m model) selectedVolume() (core.VolumeRow, bool) {
	if len(m.snapshot.Volumes) == 0 {
		return core.VolumeRow{}, false
	}
	return m.snapshot.Volumes[m.volumeCursor], true
}

func (m model) stateStyle(state string) lipgloss.Style {
	switch strings.ToLower(state) {
	case "running":
		return m.styles.stateRun
	case "paused", "restarting", "created":
		return m.styles.stateWarn
	case "exited", "stopped":
		return m.styles.stateStop
	case "dead":
		return m.styles.stateDead
	default:
		return m.styles.stateOther
	}
}
