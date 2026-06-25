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
)

type tickMsg struct{}

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

type shellDoneMsg struct{ err error }

type inspectResultMsg struct {
	data []string
	err  error
}

type model struct {
	service core.ServiceInterface
	err     error

	width  int
	height int

	browse browse.Model
	viewer viewer.Model

	screen      shared.Screen
	screenStack []shared.Screen

	dataDirty        bool
	loading          bool
	loadingStage     shared.Stage
	metricsLoaded    bool
	snapshotInflight bool
	spinner          spinner.Model

	lastResizeTime time.Time
	tickStarted    bool

	styles theme.Set
	menu   menu.MenuState
	help   menu.HelpState

	appliedConfig []string

	logTailLines int
}

func New(service core.ServiceInterface, appliedConfig []string, logTailLines int) tea.Model {
	s := spinner.New(spinner.WithSpinner(spinner.Points))

	bm := browse.NewModel()
	vm := viewer.NewModel()

	return model{
		service:        service,
		dataDirty:      true,
		loading:        true,
		screen:         shared.Main,
		loadingStage:   shared.StageContainers,
		styles:         defaultStyles(),
		spinner:        s,
		browse:         bm,
		viewer:         vm,
		menu:           menu.NewMenuState(),
		help:           menu.NewHelpState(0, 0),
		appliedConfig: appliedConfig,
		logTailLines:  logTailLines,
		tickStarted:    false,
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{loadContainersCmd(m.service)}

	cmds = append(cmds,
		tea.Tick(shared.SpinnerTickInterval, func(t time.Time) tea.Msg {
			return spinner.TickMsg{Time: t, ID: m.spinner.ID()}
		}),
	)

	return tea.Batch(cmds...)
}

func defaultStyles() theme.Set {
	return theme.Default()
}

func (m *model) pushScreen(s shared.Screen) {
	m.screenStack = append(m.screenStack, s)
}

func (m *model) popScreen() shared.Screen {
	if len(m.screenStack) == 0 {
		return shared.Main
	}
	s := m.screenStack[len(m.screenStack)-1]
	m.screenStack = m.screenStack[:len(m.screenStack)-1]
	return s
}

func (m *model) initViewer(
	screen shared.Screen,
	contentType viewer.ContentType,
	resourceType core.ResourceType,
	containerID, resourceName, loadingMsg, emptyMsg string,
) {
	m.pushScreen(m.screen)
	m.screen = screen
	m.viewer.Width = m.width
	m.viewer.Height = max(1, m.height-4)
	m.viewer.ContainerID = containerID
	m.viewer.ResourceType = resourceType
	m.viewer.ContainerName = resourceName
	m.viewer.LoadingMsg = loadingMsg
	m.viewer.EmptyMsg = emptyMsg
	m.viewer.Breadcrumb = ""
	m.viewer.Vp.InitialLoad = true
	m.viewer.Vp.Data = nil
	m.viewer.Vp.SetXOffset(0)
	m.viewer.Vp.Filter.Active = false
	m.viewer.Vp.Filter.Query = ""
	m.viewer.Vp.Filter.Input.SetValue("")
	m.viewer.Vp.ContentType = contentType
	m.viewer.Styles = viewer.Styles{
		Breadcrumb:   m.styles.Viewer.Breadcrumb,
		FollowOn:     m.styles.Viewer.FollowOn,
		FollowOff:    m.styles.Viewer.FollowOff,
		Muted:        m.styles.Browse.Muted,
		Divider:      m.styles.Browse.Divider,
		SubpageFrame: m.styles.Viewer.SubpageFrame,
		Key:          m.styles.Chrome.Key,
		KeyText:      m.styles.Chrome.KeyText,
	}
	m.err = nil
}
