package tui

import (
	"reflect"
	"testing"

	"easydocker/internal/tui/screens/viewer"

	tea "charm.land/bubbletea/v2"
)

func TestMergePolledLogs(t *testing.T) {
	tests := []struct {
		name        string
		previous    []string
		polled      []string
		want        []string
		wantOverlap bool
	}{
		{"empty previous returns polled", nil, []string{"a", "b"}, []string{"a", "b"}, true},
		{"empty polled keeps previous", []string{"a", "b"}, nil, []string{"a", "b"}, true},
		{"overlap appends only new suffix", []string{"1", "2", "3"}, []string{"3", "4", "5"}, []string{"1", "2", "3", "4", "5"}, true},
		{"identical slices stay stable", []string{"1", "2"}, []string{"1", "2"}, []string{"1", "2"}, true},
		{"smaller polled suffix keeps previous", []string{"1", "2", "3", "4"}, []string{"3", "4"}, []string{"1", "2", "3", "4"}, true},
		{"carriage returns are normalized", []string{"1\r", "2\r", "3\r"}, []string{"3", "4"}, []string{"1", "2", "3", "4"}, true},
		{"disjoint poll replaces log buffer", []string{"1", "2"}, []string{"8", "9"}, []string{"8", "9"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOverlap := viewer.MergePolledLogs(tt.previous, tt.polled, 0)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergePolledLogs() = %v, want %v", got, tt.want)
			}
			if gotOverlap != tt.wantOverlap {
				t.Errorf("overlap = %v, want %v", gotOverlap, tt.wantOverlap)
			}
		})
	}

	t.Run("max lines trims merged result", func(t *testing.T) {
		got, gotOverlap := viewer.MergePolledLogs([]string{"1", "2", "3"}, []string{"3", "4", "5"}, 3)
		if !reflect.DeepEqual(got, []string{"3", "4", "5"}) {
			t.Errorf("MergePolledLogs() = %v, want %v", got, []string{"3", "4", "5"})
		}
		if !gotOverlap {
			t.Error("expected overlap true")
		}
	})
}

func TestViewerStateHistory(t *testing.T) {
	t.Run("at top triggers history condition", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		if !state.Viewport.AtTop() {
			t.Fatal("expected viewport to be at top")
		}
	})

	t.Run("not at top when scrolled", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetHeight(2)
		state.Viewport.SetContent("line1\nline2\nline3")
		state.Viewport.SetYOffset(1)
		if state.Viewport.AtTop() {
			t.Fatal("expected viewport not to be at top")
		}
	})

	t.Run("set follow jumps to bottom", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetContent("line1\nline2\nline3")
		state.SetFollow(true)
		if !state.Viewport.AtBottom() {
			t.Fatal("expected follow to scroll to bottom")
		}
	})

	t.Run("unfollow stops following", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetContent("line1\nline2\nline3")
		state.SetFollow(false)
		if state.Follow {
			t.Fatal("expected Follow to be false")
		}
	})

	t.Run("SyncFromData applies content to viewport", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line-a", "line-b", "line-c"}
		state.SyncFromData(80, 10)
		if state.Viewport.View() == "" {
			t.Fatal("expected viewport to have content after SyncFromData")
		}
	})
}

func TestViewerControllerHistoryKey(t *testing.T) {
	t.Run("Home key moves to top, disables follow", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.Follow = false

		trans := viewer.Controller{}.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyHome}, viewer.NewKeyMap())
		if state.Viewport.AtTop() && state.Follow == false && trans.Load == nil {
			// Expected: home moves to top, does not trigger history load directly
		} else {
			t.Fatalf("unexpected state after Home: AtTop=%v, Follow=%v, Load=%v", state.Viewport.AtTop(), state.Follow, trans.Load)
		}
	})

	t.Run("End key enables follow", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.Follow = false

		viewer.Controller{}.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyEnd}, viewer.NewKeyMap())
		if !state.Follow {
			t.Fatal("expected follow to be enabled after End key")
		}
	})

	t.Run("PgUp at top returns transition without load", func(t *testing.T) {
		state := viewer.NewState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.Follow = false

		trans := viewer.Controller{}.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyPgUp}, viewer.NewKeyMap())
		if trans.Load != nil {
			t.Fatal("expected no load request from PgUp in current architecture; history loading is tick-driven")
		}
	})
}
