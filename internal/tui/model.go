package tui

import (
	"time"

	"easydocker/internal/core"
	"easydocker/internal/tui/screens/browse"
	"easydocker/internal/tui/screens/menu"
	"easydocker/internal/tui/screens/viewer"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/theme"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

const (
	tabContainers = shared.TabContainers
	tabImages     = shared.TabImages
	tabNetworks   = shared.TabNetworks
	tabVolumes    = shared.TabVolumes

	pollInterval = time.Second

	loadStageIdle       = int(shared.StageIdle)
	loadStageContainers = int(shared.StageContainers)
	loadStageResources  = int(shared.StageResources)
	loadStageMetrics    = int(shared.StageMetrics)
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
	err        error
}

type loadResultMsg struct {
	snapshot core.Snapshot
	err      error
}

type shellDoneMsg struct{ err error }

type inspectResultMsg struct {
	resourceType int
	resourceID   string
	resourceName string
	data         []string
	err          error
}

type model struct {
	service *core.Service
	err     error

	width  int
	height int

	browse browse.Model
	viewer viewer.Model

	screen         shared.Screen
	previousScreen shared.Screen

	loading       bool
	loadingStage  int
	metricsLoaded bool
	metricsSpinner   spinner.Model
	containerSpinner spinner.Model

	styles theme.Set
	menu   menu.MenuState
	help   menu.HelpState
}

func New(service *core.Service) tea.Model {
	metricsSpinner := spinner.New(spinner.WithSpinner(spinner.Points))
	containerSpinner := spinner.New(spinner.WithSpinner(spinner.Points))

	bm := browse.NewModel()
	vm := viewer.NewModel()
	vm.Spinner = spinner.New(spinner.WithSpinner(spinner.Dot))

	return model{
		service:          service,
		loading:          true,
		screen:           shared.Browse,
		loadingStage:     loadStageContainers,
		styles:           defaultStyles(),
		metricsSpinner:   metricsSpinner,
		containerSpinner: containerSpinner,
		browse:           bm,
		viewer:           vm,
		menu:             menu.NewMenuState(),
		help:             menu.NewHelpState(0, 0),
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{loadContainersCmd(m.service), tickCmd()}

	spinnerTickInterval := time.Second / 7
	cmds = append(cmds,
		tea.Tick(spinnerTickInterval, func(t time.Time) tea.Msg {
			return spinner.TickMsg{Time: t, ID: m.metricsSpinner.ID()}
		}),
		tea.Tick(spinnerTickInterval, func(t time.Time) tea.Msg {
			return spinner.TickMsg{Time: t, ID: m.containerSpinner.ID()}
		}),
		tea.Tick(spinnerTickInterval, func(t time.Time) tea.Msg {
			return spinner.TickMsg{Time: t, ID: m.viewer.Spinner.ID()}
		}),
	)

	return tea.Batch(cmds...)
}

func defaultStyles() theme.Set {
	return theme.Default()
}
