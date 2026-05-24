package browse

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"easydocker/internal/core"
	"easydocker/internal/tui/shared"
	"easydocker/internal/tui/ui/tables"
	"easydocker/internal/tui/util"
)

const cursorPageStep = 5

type BrowseData struct {
	ContainerListRows       []tables.ContainerListRow
	FilteredImages          []core.ImageRow
	FilteredNetworks        []core.NetworkRow
	FilteredVolumes         []core.VolumeRow
	MetricsLoadingIndicator string
}

type Model struct {
	Loading         bool
	Snapshot        core.Snapshot
	ActiveTab       shared.Tab
	ShowAll         bool
	Filter          FilterState
	ComposeExpanded map[string]bool

	ContainerCursor int
	ImageCursor     int
	NetworkCursor   int
	VolumeCursor    int

	Width  int
	Height int

	Data BrowseData

	DetailProvider DetailProvider
	ListRenderer   func(width, height int) string
}

func NewModel() Model {
	return Model{
		ActiveTab:       shared.TabContainers,
		ShowAll:         true,
		Filter:          NewFilterState(),
		ComposeExpanded: map[string]bool{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) View() string {
	if m.ListRenderer == nil {
		return ""
	}

	filterCopy := m.Filter
	if filterCopy.Active {
		filterCopy.Input.SetWidth(max(1, m.Width-util.DisplayWidth(filterCopy.Input.Prompt)))
	}

	vm := ViewModel{
		Loading:                 m.Loading,
		Snapshot:                m.Snapshot,
		ActiveTab:               m.ActiveTab,
		MetricsLoadingIndicator: m.Data.MetricsLoadingIndicator,
		Width:                   m.Width,
		Height:                  m.Height,
		Selections:              m.selections(),
		Filter:                  filterCopy,
	}

	list := m.ListRenderer(m.Width, ListHeightForContent(m.Height, m.Filter.Active))
	return RenderContent(vm, list, m.DetailProvider)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	keys := NewKeyMap()

	if m.Filter.Active {
		return m.handleFilterKey(msg, keys)
	}

	browseState := State{Filter: m.Filter}
	transition := Controller{}.HandleKey(&browseState, msg, keys)
	m.Filter = browseState.Filter

	if transition.OpenResource && m.cursorOnComposeRow() {
		transition.OpenResource = false
		transition.ToggleCompose = true
	}

	return m.applyTransition(transition)
}

func (m Model) handleFilterKey(msg tea.KeyPressMsg, keys KeyMap) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.OpenMenu):
		m.Filter.Active = false
		m.Filter.Input.Blur()
		m.Filter.Query = ""
		m.Filter.Input.SetValue("")
		m = m.clampCursors()
		return m, nil
	case msg.String() == "enter":
		m.Filter.Active = false
		m.Filter.Input.Blur()
		return m, nil
	case key.Matches(msg, keys.MoveUp):
		m = m.moveCursor(-1)
		return m, nil
	case key.Matches(msg, keys.MoveDown):
		m = m.moveCursor(1)
		return m, nil
	case key.Matches(msg, keys.PageUp):
		m = m.moveCursor(-cursorPageStep)
		return m, nil
	case key.Matches(msg, keys.PageDown):
		m = m.moveCursor(cursorPageStep)
		return m, nil
	default:
		var cmd tea.Cmd
		m.Filter.Input, cmd = m.Filter.Input.Update(msg)
		m.Filter.Query = m.Filter.Input.Value()
		m = m.clampCursors()
		return m, cmd
	}
}

func (m Model) applyTransition(transition Transition) (Model, tea.Cmd) {
	if transition.ChangeTab != 0 {
		m = m.moveActiveTab(int(transition.ChangeTab))
	}
	if transition.ActivateFilter {
		m.Filter.Active = true
		m.Filter.Input.Focus()
		m.Filter.Input.SetValue(m.Filter.Query)
	}
	if transition.ToggleScope {
		m.ShowAll = !m.ShowAll
		m = m.clampCursors()
	}
	if transition.OpenMenu {
		return m, func() tea.Msg { return TransitionMsg{OpenMenu: true} }
	}
	if transition.CursorMove != 0 {
		m = m.moveCursor(transition.CursorMove)
	}
	if transition.ToggleCompose {
		m = m.toggleCompose()
	}
	if transition.OpenResource || transition.OpenShell || transition.OpenInspect {
		return m, func() tea.Msg {
			return TransitionMsg{
				OpenResource: transition.OpenResource,
				OpenShell:    transition.OpenShell,
				OpenInspect:  transition.OpenInspect,
			}
		}
	}
	return m, nil
}

func (m Model) moveActiveTab(delta int) Model {
	m.ActiveTab = shared.MoveActiveTab(m.ActiveTab, delta, shared.TabContainers, shared.TabVolumes)
	if m.Filter.Query != "" {
		m.Filter.Query = ""
		m.Filter.Input.SetValue("")
	}
	return m
}

func (m Model) moveCursor(delta int) Model {
	count := m.ItemCountForTab(m.ActiveTab)
	c := m.cursors()
	_ = shared.MoveCursorForTab(&c, m.ActiveTab, delta, count)
	m.ContainerCursor = c.Container
	m.ImageCursor = c.Image
	m.NetworkCursor = c.Network
	m.VolumeCursor = c.Volume
	return m
}

func (m Model) clampCursors() Model {
	tabs := []shared.Tab{shared.TabContainers, shared.TabImages, shared.TabNetworks, shared.TabVolumes}
	c := m.cursors()
	shared.ClampAllCursors(&c, tabs, m.ItemCountForTab)
	m.ContainerCursor = c.Container
	m.ImageCursor = c.Image
	m.NetworkCursor = c.Network
	m.VolumeCursor = c.Volume
	return m
}

func (m Model) cursors() shared.Cursors {
	return shared.Cursors{
		Container: m.ContainerCursor,
		Image:     m.ImageCursor,
		Network:   m.NetworkCursor,
		Volume:    m.VolumeCursor,
	}
}

func (m Model) ItemCountForTab(tab shared.Tab) int {
	switch tab {
	case shared.TabContainers:
		if len(m.Data.ContainerListRows) > 0 {
			return len(m.Data.ContainerListRows)
		}
		return len(m.Snapshot.Containers)
	case shared.TabImages:
		if len(m.Data.FilteredImages) > 0 {
			return len(m.Data.FilteredImages)
		}
		return len(m.Snapshot.Images)
	case shared.TabNetworks:
		if len(m.Data.FilteredNetworks) > 0 {
			return len(m.Data.FilteredNetworks)
		}
		return len(m.Snapshot.Networks)
	case shared.TabVolumes:
		if len(m.Data.FilteredVolumes) > 0 {
			return len(m.Data.FilteredVolumes)
		}
		return len(m.Snapshot.Volumes)
	default:
		return 0
	}
}

func (m Model) selections() SelectionSet {
	container, hasContainer := m.SelectedContainer()
	composeProject, hasComposeProject := m.SelectedComposeProject()
	image, hasImage := m.SelectedImage()
	network, hasNetwork := m.SelectedNetwork()
	volume, hasVolume := m.SelectedVolume()
	return SelectionSet{
		Container:         container,
		HasContainer:      hasContainer,
		ComposeProject:    composeProject,
		HasComposeProject: hasComposeProject,
		Image:             image,
		HasImage:          hasImage,
		Network:           network,
		HasNetwork:        hasNetwork,
		Volume:            volume,
		HasVolume:         hasVolume,
	}
}

func (m Model) SelectedContainer() (core.ContainerRow, bool) {
	if len(m.Data.ContainerListRows) > 0 {
		row, ok := selectedAt(m.Data.ContainerListRows, m.ContainerCursor)
		if !ok || row.Kind != tables.ContainerListRowContainer {
			return core.ContainerRow{}, false
		}
		return row.Container, true
	}
	return selectedAt(m.Snapshot.Containers, m.ContainerCursor)
}

func (m Model) SelectedComposeProject() (core.ComposeProject, bool) {
	if len(m.Data.ContainerListRows) == 0 {
		return core.ComposeProject{}, false
	}
	row, ok := selectedAt(m.Data.ContainerListRows, m.ContainerCursor)
	if !ok || row.Kind != tables.ContainerListRowComposeProject {
		return core.ComposeProject{}, false
	}
	return row.ComposeProject, true
}

func (m Model) SelectedImage() (core.ImageRow, bool) {
	if len(m.Data.FilteredImages) > 0 {
		return selectedAt(m.Data.FilteredImages, m.ImageCursor)
	}
	return selectedAt(m.Snapshot.Images, m.ImageCursor)
}

func (m Model) SelectedNetwork() (core.NetworkRow, bool) {
	if len(m.Data.FilteredNetworks) > 0 {
		return selectedAt(m.Data.FilteredNetworks, m.NetworkCursor)
	}
	return selectedAt(m.Snapshot.Networks, m.NetworkCursor)
}

func (m Model) SelectedVolume() (core.VolumeRow, bool) {
	if len(m.Data.FilteredVolumes) > 0 {
		return selectedAt(m.Data.FilteredVolumes, m.VolumeCursor)
	}
	return selectedAt(m.Snapshot.Volumes, m.VolumeCursor)
}

func (m Model) cursorOnComposeRow() bool {
	if m.ActiveTab != shared.TabContainers {
		return false
	}
	if len(m.Data.ContainerListRows) == 0 {
		return false
	}
	row, ok := selectedAt(m.Data.ContainerListRows, m.ContainerCursor)
	return ok && row.Kind == tables.ContainerListRowComposeProject
}

func (m Model) toggleCompose() Model {
	if len(m.Data.ContainerListRows) == 0 {
		return m
	}
	row, ok := selectedAt(m.Data.ContainerListRows, m.ContainerCursor)
	if !ok || row.Kind != tables.ContainerListRowComposeProject {
		return m
	}
	projectName := row.ComposeProject.Name
	if m.ComposeExpanded == nil {
		m.ComposeExpanded = map[string]bool{}
	}
	m.ComposeExpanded[projectName] = !m.ComposeExpanded[projectName]
	return m
}

func selectedAt[T any](items []T, cursor int) (T, bool) {
	var zero T
	if len(items) == 0 || cursor < 0 || cursor >= len(items) {
		return zero, false
	}
	return items[cursor], true
}

type TransitionMsg struct {
	OpenResource bool
	OpenShell    bool
	OpenInspect  bool
	OpenMenu     bool
}
