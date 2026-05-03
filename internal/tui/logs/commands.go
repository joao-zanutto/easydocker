package logs

import (
	"easydocker/internal/core"

	tea "charm.land/bubbletea/v2"
)

func LoadLogsDataCmd(service *core.Service, containerID string, sessionID int, previousCPU, previousMem []float64, tail int, src Source) tea.Cmd {
	return func() tea.Msg {
		data, err := service.LoadContainerLiveData(containerID, previousCPU, previousMem, tail)
		return ResultMsg{
			ContainerID: containerID,
			SessionID:   sessionID,
			Data:        data,
			Err:         err,
			Tail:        tail,
			Src:         src,
		}
	}
}
