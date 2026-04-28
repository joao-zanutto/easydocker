package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/loading"
	"easydocker/internal/tui/logs"
	"easydocker/internal/tui/theme"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	// Browse filter mode state
	browseFilterActive bool
	browseFilterInput  textinput.Model
	browseFilterQuery  string
	composeExpanded    map[string]bool
}

func New(service *core.Service) tea.Model {
	metricsSpinner := spinner.New(spinner.WithSpinner(spinner.Points))
	containerSpinner := spinner.New(spinner.WithSpinner(spinner.Points))
	logsSpinner := spinner.New(spinner.WithSpinner(spinner.Dot))

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Prompt = "🔎︎ "
	filterInput.Placeholder = ""
	filterInput.CharLimit = 200

	return model{
		service:           service,
		activeTab:         tabContainers,
		showAll:           true,
		loading:           true,
		screen:            screenModeBrowse,
		loadingStage:      loadStageContainers,
		logs:              logs.NewState(),
		styles:            defaultStyles(),
		metricsSpinner:    metricsSpinner,
		containerSpinner:  containerSpinner,
		logsSpinner:       logsSpinner,
		browseFilterInput: filterInput,
		composeExpanded:   map[string]bool{},
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.loadContainersCmd(), tickCmd()}
	if m.shouldAnimateMetricsLoadingIndicator() {
		cmds = append(cmds, m.metricsSpinner.Tick, m.containerSpinner.Tick)
	}
	return tea.Batch(cmds...)
}

func defaultStyles() theme.Set {
	return theme.Default()
}
