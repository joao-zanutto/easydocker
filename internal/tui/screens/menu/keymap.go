package menu

import (
	"easydocker/internal/tui/shared"

	"charm.land/bubbles/v2/key"
)

type MenuKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
}

func NewKeyMap() MenuKeyMap {
	return MenuKeyMap{
		Up:   shared.UpBinding("navigate"),
		Down: shared.DownBinding("navigate"),
		Select: shared.EnterBinding("select"),
		Back:   shared.EscBinding("back"),
	}
}
