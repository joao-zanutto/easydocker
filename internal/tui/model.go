package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/loading"
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	tabContainers = iota
	tabImages
	tabNetworks
	tabVolumes

	pollInterval = time.Second

	loadStageIdle       = int(loading.StageIdle)
	loadStageContainers = int(loading.StageContainers)
	loadStageResources  = int(loading.StageResources)
	loadStageMetrics    = int(loading.StageMetrics)
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
	logs            logs.State
	loadingStage    int
	styles          theme.Set
}

func New(service *core.Service) tea.Model {
	return model{
		service:      service,
		activeTab:    tabContainers,
		showAll:      true,
		loading:      true,
		screen:       screenModeBrowse,
		loadingStage: loadStageContainers,
		logs:         logs.NewState(),
		styles:       defaultStyles(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loadContainersCmd(), tickCmd())
}

func defaultStyles() theme.Set {
	return theme.Default()
}
