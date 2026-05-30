package browse

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestControllerHandleKey(t *testing.T) {
	km := NewKeyMap()
	tests := []struct {
		name  string
		state State
		msg   tea.KeyPressMsg
		want  Transition
	}{
		{
			name: "down arrow moves cursor down",
			msg:  tea.KeyPressMsg{Code: tea.KeyDown},
			want: Transition{CursorMove: 1},
		},
		{
			name: "up arrow moves cursor up",
			msg:  tea.KeyPressMsg{Code: tea.KeyUp},
			want: Transition{CursorMove: -1},
		},
		{
			name: "right arrow switches tab right",
			msg:  tea.KeyPressMsg{Code: tea.KeyRight},
			want: Transition{ChangeTab: 1},
		},
		{
			name: "left arrow switches tab left",
			msg:  tea.KeyPressMsg{Code: tea.KeyLeft},
			want: Transition{ChangeTab: -1},
		},
		{
			name: "page down moves cursor by page step",
			msg:  tea.KeyPressMsg{Code: tea.KeyPgDown},
			want: Transition{CursorMove: 5},
		},
		{
			name: "page up moves cursor by page step",
			msg:  tea.KeyPressMsg{Code: tea.KeyPgUp},
			want: Transition{CursorMove: -5},
		},
		{
			name: "enter opens resource",
			msg:  tea.KeyPressMsg{Code: tea.KeyEnter},
			want: Transition{OpenResource: true},
		},
		{
			name: "escape opens menu",
			msg:  tea.KeyPressMsg{Code: tea.KeyEscape},
			want: Transition{OpenMenu: true},
		},
		{
			name: "a toggles scope",
			msg:  tea.KeyPressMsg{Code: 'a', Text: "a"},
			want: Transition{ToggleScope: true},
		},
		{
			name: "forward slash activates filter",
			msg:  tea.KeyPressMsg{Code: '/', Text: "/"},
			want: Transition{ActivateFilter: true},
		},
		{
			name: "s opens shell",
			msg:  tea.KeyPressMsg{Code: 's', Text: "s"},
			want: Transition{OpenShell: true},
		},
		{
			name: "i opens inspect",
			msg:  tea.KeyPressMsg{Code: 'i', Text: "i"},
			want: Transition{OpenInspect: true},
		},
		{
			name: "unmapped key returns empty transition",
			msg:  tea.KeyPressMsg{Code: 'x', Text: "x"},
			want: Transition{},
		},
	}
	ctrl := Controller{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctrl.HandleKey(&tt.state, tt.msg, km)
			if got != tt.want {
				t.Errorf("HandleKey() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
