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
			got, gotOverlap := mergePolledLogs(tt.previous, tt.polled, 0)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergePolledLogs() = %v, want %v", got, tt.want)
			}
			if gotOverlap != tt.wantOverlap {
				t.Errorf("overlap = %v, want %v", gotOverlap, tt.wantOverlap)
			}
		})
	}

	t.Run("max lines trims merged result", func(t *testing.T) {
		got, gotOverlap := mergePolledLogs([]string{"1", "2", "3"}, []string{"3", "4", "5"}, 3)
		if !reflect.DeepEqual(got, []string{"3", "4", "5"}) {
			t.Errorf("mergePolledLogs() = %v, want %v", got, []string{"3", "4", "5"})
		}
		if !gotOverlap {
			t.Error("expected overlap true")
		}
	})
}

func TestHistoryLoading(t *testing.T) {
	t.Run("CanLoadHistory returns true when at top and not loading", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.HistoryDone = false
		state.HistoryLoad = false
		if !CanLoadHistory(state) {
			t.Fatal("expected CanLoadHistory to return true")
		}
	})

	t.Run("CanLoadHistory returns false when not at top", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetHeight(2)
		state.Viewport.SetContent("line1\nline2\nline3")
		state.Viewport.SetYOffset(1)
		state.HistoryDone = false
		state.HistoryLoad = false
		if CanLoadHistory(state) {
			t.Fatal("expected CanLoadHistory to return false")
		}
	})

	t.Run("StartHistoryLoad sets correct state", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.TailLines = 200
		StartHistoryLoad(&state, 400)
		if !state.HistoryLoad {
			t.Fatal("expected HistoryLoad to be true")
		}
		if state.TailLines != 400 {
			t.Fatalf("TailLines = %d, want 400", state.TailLines)
		}
	})

	t.Run("ApplyHistory prepends new lines and adjusts viewport", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"base1", "base2", "base3"}
		state.Viewport.SetYOffset(2)
		state.Follow = false
		state.HistoryLoad = true
		olderLogs := []string{"older1", "older2"}
		newData := append(olderLogs, state.Data...)
		ApplyHistoryWithMerge(&state, newData)
		if state.HistoryLoad {
			t.Fatal("expected HistoryLoad to be false")
		}
		if len(state.Data) != 5 {
			t.Fatalf("expected 5 lines, got %d", len(state.Data))
		}
		if state.Data[0] != "older1" {
			t.Fatalf("expected first line to be older1, got %s", state.Data[0])
		}
	})

	t.Run("ApplyHistory marks done after three unchanged responses", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.Follow = false
		state.HistoryBaseLen = 3
		for attempt := 0; attempt < 3; attempt++ {
			state.HistoryLoad = true
			ApplyHistoryWithMerge(&state, state.Data)
		}
		if !state.HistoryDone {
			t.Fatal("expected HistoryDone to be true")
		}
	})

	t.Run("ApplyPoll merges overlapping logs", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		polled := []string{"line2", "line3", "line4", "line5"}
		merged, _ := ApplyPollWithMerge(&state, polled)
		state.Data = merged
		if len(state.Data) != 5 {
			t.Fatalf("expected 5 lines, got %d", len(state.Data))
		}
		if state.Data[3] != "line4" {
			t.Fatalf("expected line4 at index 3, got %s", state.Data[3])
		}
	})

	t.Run("ApplyPoll replaces non-overlapping logs", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"old1", "old2"}
		polled := []string{"new1", "new2"}
		merged, _ := ApplyPollWithMerge(&state, polled)
		state.Data = merged
		if len(state.Data) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(state.Data))
		}
		if state.Data[0] != "new1" {
			t.Fatalf("expected new1, got %s", state.Data[0])
		}
	})
}

func TestHistoryKeyHandling(t *testing.T) {
	t.Run("Home key triggers history load", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.Follow = false
		state.HistoryDone = false
		viewer.Controller{}.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyHome}, viewer.NewKeyMap())
		trans := HistoryLoadRequest(&state)
		if trans == nil {
			t.Fatal("expected Load request")
		}
		if trans.Src != viewer.SourceHistory {
			t.Fatalf("expected SourceHistory, got %v", trans.Src)
		}
	})

	t.Run("PgUp key triggers history load", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(1)
		state.Follow = false
		viewer.Controller{}.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyPgUp}, viewer.NewKeyMap())
		trans := HistoryLoadRequest(&state)
		if trans == nil {
			t.Fatal("expected Load request")
		}
		if trans.Src != viewer.SourceHistory {
			t.Fatalf("expected SourceHistory, got %v", trans.Src)
		}
	})

	t.Run("End key triggers history check", func(t *testing.T) {
		state := NewLogsState()
		state.Data = []string{"line1", "line2", "line3"}
		state.Viewport.SetYOffset(0)
		state.Follow = false
		state.HistoryDone = false
		viewer.Controller{}.HandleKey(&state, tea.KeyPressMsg{Code: tea.KeyEnd}, viewer.NewKeyMap())
		trans := HistoryLoadRequest(&state)
		if trans == nil {
			t.Fatal("expected Load request for End key")
		}
	})
}
