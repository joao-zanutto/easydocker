package viewer

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestControllerHandleKey(t *testing.T) {
	km := NewKeyMap()
	tests := []struct {
		name      string
		setup     func(*Viewport)
		msg       tea.KeyPressMsg
		wantTrans Transition
		check     func(*testing.T, *Viewport)
	}{
		{
			name: "up scrolls up one line",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c", "d", "e", "f", "g"}
				vp.SetHeight(3)
				vp.SetContent("a\nb\nc\nd\ne\nf\ng")
				vp.SetYOffset(4)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyUp},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.YOffset() != 3 {
					t.Errorf("YOffset = %d, want 3", vp.YOffset())
				}
				if vp.Follow {
					t.Error("Follow should be false after scroll up")
				}
			},
		},
		{
			name: "down scrolls down one line",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c", "d", "e", "f", "g"}
				vp.SetHeight(3)
				vp.SetContent("a\nb\nc\nd\ne\nf\ng")
				vp.SetYOffset(3)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyDown},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.YOffset() != 4 {
					t.Errorf("YOffset = %d, want 4", vp.YOffset())
				}
			},
		},
		{
			name: "down at bottom re-enables follow",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c", "d", "e"}
				vp.SetHeight(4)
				vp.SetContent("a\nb\nc\nd\ne")
				vp.SetYOffset(1)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyDown},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if !vp.Follow {
					t.Error("Follow should be true when scrolling down at bottom")
				}
			},
		},
		{
			name: "right scrolls horizontally",
			setup: func(vp *Viewport) {
				vp.WrapLines = false
				vp.SetWidth(10)
				vp.SetHeight(1)
				vp.SetContent("very long line here")
				vp.SetYOffset(0)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyRight},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.XOffset() == 0 {
					t.Error("expected XOffset > 0 after right scroll")
				}
				if vp.Follow {
					t.Error("Follow should be false after horizontal scroll")
				}
			},
		},
		{
			name: "left scrolls horizontally",
			setup: func(vp *Viewport) {
				vp.WrapLines = false
				vp.SetWidth(10)
				vp.SetHeight(1)
				vp.SetContent("very long line here")
				vp.SetXOffset(8)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyLeft},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.XOffset() == 8 {
					t.Error("expected XOffset < 8 after left scroll")
				}
			},
		},
		{
			name: "right scroll ignored when wrapped",
			setup: func(vp *Viewport) {
				vp.WrapLines = true
				vp.SetWidth(10)
				vp.SetHeight(1)
				vp.SetContent("very long line here")
				vp.SetXOffset(0)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyRight},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.XOffset() != 0 {
					t.Errorf("XOffset should remain 0 when wrapped, got %d", vp.XOffset())
				}
			},
		},
		{
			name: "page up scrolls up one page",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c", "d", "e", "f", "g"}
				vp.SetHeight(3)
				vp.SetContent("a\nb\nc\nd\ne\nf\ng")
				vp.SetYOffset(4)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyPgUp},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.YOffset() >= 4 {
					t.Errorf("YOffset should decrease after page up, got %d", vp.YOffset())
				}
				if vp.Follow {
					t.Error("Follow should be false after page up")
				}
			},
		},
		{
			name: "page down scrolls down one page",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c", "d", "e", "f", "g"}
				vp.SetHeight(3)
				vp.SetContent("a\nb\nc\nd\ne\nf\ng")
				vp.SetYOffset(0)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyPgDown},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.YOffset() == 0 {
					t.Error("YOffset should increase after page down")
				}
			},
		},
		{
			name: "home goes to top and disables follow",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c"}
				vp.SetHeight(2)
				vp.SetContent("a\nb\nc")
				vp.SetYOffset(1)
				vp.Follow = true
				vp.SetXOffset(8)
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyHome},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if !vp.AtTop() {
					t.Error("expected AtTop after Home")
				}
				if vp.Follow {
					t.Error("Follow should be false after Home")
				}
				if vp.XOffset() != 0 {
					t.Errorf("XOffset should be 0 after Home, got %d", vp.XOffset())
				}
			},
		},
		{
			name: "end enables follow",
			setup: func(vp *Viewport) {
				vp.Data = []string{"a", "b", "c"}
				vp.SetHeight(2)
				vp.Follow = false
			},
			msg:       tea.KeyPressMsg{Code: tea.KeyEnd},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if !vp.Follow {
					t.Error("Follow should be true after End")
				}
			},
		},
		{
			name: "w toggles wrap on",
			setup: func(vp *Viewport) {
				vp.WrapLines = false
			},
			msg:       tea.KeyPressMsg{Code: 'w', Text: "w"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if !vp.WrapLines {
					t.Error("WrapLines should be true after toggle")
				}
			},
		},
		{
			name: "w toggles wrap off",
			setup: func(vp *Viewport) {
				vp.WrapLines = true
			},
			msg:       tea.KeyPressMsg{Code: 'w', Text: "w"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.WrapLines {
					t.Error("WrapLines should be false after toggle")
				}
			},
		},
		{
			name: "f toggles follow on",
			setup: func(vp *Viewport) {
				vp.Follow = false
			},
			msg:       tea.KeyPressMsg{Code: 'f', Text: "f"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if !vp.Follow {
					t.Error("Follow should be true after toggle")
				}
			},
		},
		{
			name: "f toggles follow off",
			setup: func(vp *Viewport) {
				vp.Follow = true
			},
			msg:       tea.KeyPressMsg{Code: 'f', Text: "f"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.Follow {
					t.Error("Follow should be false after toggle")
				}
			},
		},
		{
			name: "slash opens filter",
			setup: func(vp *Viewport) {
				vp.Filter.Active = false
			},
			msg:       tea.KeyPressMsg{Code: '/', Text: "/"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if !vp.Filter.Active {
					t.Error("Filter should be active after slash")
				}
			},
		},
		{
			name: "slash closes filter and clears query",
			setup: func(vp *Viewport) {
				vp.Filter.Active = true
				vp.Filter.Input.Focus()
				vp.Filter.Query = "search"
				vp.Filter.Input.SetValue("search")
			},
			msg:       tea.KeyPressMsg{Code: '/', Text: "/"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
				if vp.Filter.Active {
					t.Error("Filter should be inactive after second slash")
				}
				if vp.Filter.Query != "" {
					t.Errorf("Filter query should be cleared, got %q", vp.Filter.Query)
				}
			},
		},
		{
			name: "s returns shell transition",
			setup: func(vp *Viewport) {
			},
			msg:       tea.KeyPressMsg{Code: 's', Text: "s"},
			wantTrans: Transition{LaunchShell: true},
			check: func(t *testing.T, vp *Viewport) {
			},
		},
		{
			name: "unknown key returns empty transition",
			setup: func(vp *Viewport) {
			},
			msg:       tea.KeyPressMsg{Code: 'x', Text: "x"},
			wantTrans: Transition{},
			check: func(t *testing.T, vp *Viewport) {
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := NewViewport()
			tt.setup(vp)
			got := Controller{}.HandleKey(vp, tt.msg, km)
			if got != tt.wantTrans {
				t.Errorf("HandleKey() transition = %+v, want %+v", got, tt.wantTrans)
			}
			tt.check(t, vp)
		})
	}
}
