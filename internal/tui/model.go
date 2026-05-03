package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/browse"
	"easydocker/internal/tui/loading"
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/theme"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
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

type execDoneMsg struct{ err error }

type model struct {
	service          *core.Service
	width            int
	height           int
	activeTab        int
	showAll          bool
	loading          bool
	err              error
	snapshot         core.Snapshot
	containerCursor  int
	imageCursor      int
	networkCursor    int
	volumeCursor     int
	screen           screenMode
	logs             logs.State
	loadingStage     int
	styles           theme.Set
	metricsLoaded    bool
	metricsSpinner   spinner.Model
	containerSpinner spinner.Model
	logsSpinner      spinner.Model
	browseFilter     browse.FilterState
	composeExpanded  map[string]bool
}

func New(service *core.Service) tea.Model {
	metricsSpinner := spinner.New(spinner.WithSpinner(spinner.Points))
	containerSpinner := spinner.New(spinner.WithSpinner(spinner.Points))
	logsSpinner := spinner.New(spinner.WithSpinner(spinner.Dot))

	return model{
		service:          service,
		activeTab:        tabContainers,
		showAll:          true,
		loading:          true,
		screen:           screenModeBrowse,
		loadingStage:     loadStageContainers,
		logs:             logs.NewState(),
		styles:           defaultStyles(),
		metricsSpinner:   metricsSpinner,
		containerSpinner: containerSpinner,
		logsSpinner:      logsSpinner,
		browseFilter:     browse.NewFilterState(),
		composeExpanded:  map[string]bool{},
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadContainersCmd(), tickCmd()}
	if m.shouldAnimateMetricsLoadingIndicator() {
		spinnerTickInterval := time.Second / 7
		cmds = append(cmds,
			tea.Tick(spinnerTickInterval, func(t time.Time) tea.Msg {
				return spinner.TickMsg{Time: t, ID: m.metricsSpinner.ID()}
			}),
			tea.Tick(spinnerTickInterval, func(t time.Time) tea.Msg {
				return spinner.TickMsg{Time: t, ID: m.containerSpinner.ID()}
			}))
	}
	return tea.Batch(cmds...)
}

func defaultStyles() theme.Set {
	return theme.Default()
}
