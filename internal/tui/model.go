package tui

import (
	"time"

	"easydocker/internal/core"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	tabContainers = iota
	tabImages
	tabNetworks
	tabVolumes

	pollInterval = time.Second

	loadStageIdle = iota
	loadStageContainers
	loadStageResources
	loadStageMetrics
)

type screenMode int

const (
	screenModeBrowse screenMode = iota
	screenModeLogs
)

type tickMsg time.Time

type containersResultMsg struct {
	containers []core.ContainerRow
	err        error
}

type resourcesResultMsg struct {
	snapshot core.Snapshot
	err      error
}

type metricsResultMsg struct {
	metricsByID map[string]core.ContainerMetrics
	totalCPU    float64
	totalMem    uint64
	err         error
}

type loadResultMsg struct {
	snapshot core.Snapshot
	err      error
}

type logsResultMsg struct {
	containerID string
	sessionID   int
	data        core.ContainerLiveData
	err         error
	tail        int
	src         string
}

type logsState struct {
	containerID string
	sessionID   int
	data        core.ContainerLiveData
	tailLines   int
	initialLoad bool
	historyDone bool
	historyLoad bool
	scrollX     int
	scrollY     int
	follow      bool
}

type model struct {
	service         *core.Service
	width           int
	height          int
	activeTab       int
	showAll         bool
	loading         bool
	err             error
	snapshot        core.Snapshot
	containerCursor int
	imageCursor     int
	networkCursor   int
	volumeCursor    int
	screen          screenMode
	logs            logsState
	loadingStage    int
	styles          styles
}

func New(service *core.Service) tea.Model {
	return model{
		service:      service,
		activeTab:    tabContainers,
		showAll:      true,
		loading:      true,
		screen:       screenModeBrowse,
		loadingStage: loadStageContainers,
		logs:         logsState{follow: true},
		styles:       defaultStyles(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loadContainersCmd(), tickCmd())
}
