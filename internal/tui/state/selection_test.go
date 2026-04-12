package state

import "testing"

func TestMoveActiveTab(t *testing.T) {
	if got := MoveActiveTab(0, 1, 0, 3); got != 1 {
		t.Fatalf("MoveActiveTab(0,+1) = %d, want 1", got)
	}
	if got := MoveActiveTab(0, -1, 0, 3); got != 0 {
		t.Fatalf("MoveActiveTab lower clamp = %d, want 0", got)
	}
	if got := MoveActiveTab(3, 1, 0, 3); got != 3 {
		t.Fatalf("MoveActiveTab upper clamp = %d, want 3", got)
	}
}

func TestToggleContainerScope(t *testing.T) {
	showAll, toggled := ToggleContainerScope(0, 0, true)
	if !toggled || showAll {
		t.Fatalf("expected toggle on containers tab")
	}

	showAll, toggled = ToggleContainerScope(1, 0, true)
	if toggled || !showAll {
		t.Fatalf("expected no toggle outside containers tab")
	}
}

func TestCursorAPIs(t *testing.T) {
	c := Cursors{Container: 1, Image: 2, Network: 3, Volume: 4}

	if got, ok := CursorForTab(c, 0); !ok || got != 1 {
		t.Fatalf("CursorForTab(container) = (%d,%v), want (1,true)", got, ok)
	}
	if ok := SetCursorForTab(&c, 2, 9); !ok || c.Network != 9 {
		t.Fatalf("SetCursorForTab(network) failed")
	}

	_ = MoveCursorForTab(&c, 0, 10, 3)
	if c.Container != 2 {
		t.Fatalf("MoveCursorForTab clamp upper = %d, want 2", c.Container)
	}

	_ = ClampCursorForTab(&c, 1, 0)
	if c.Image != 0 {
		t.Fatalf("ClampCursorForTab empty list = %d, want 0", c.Image)
	}

	if ok := ReconcileCursorForTab(&c, 3, 7, false); ok {
		t.Fatalf("ReconcileCursorForTab should fail when not found")
	}
	if ok := ReconcileCursorForTab(&c, 3, 7, true); !ok || c.Volume != 7 {
		t.Fatalf("ReconcileCursorForTab should set cursor when found")
	}
}

func TestClampAllCursors(t *testing.T) {
	c := Cursors{Container: 10, Image: 10, Network: 10, Volume: 10}
	tabs := []int{0, 1, 2, 3}
	counts := map[int]int{0: 1, 1: 2, 2: 3, 3: 4}

	ClampAllCursors(&c, tabs, func(tab int) int { return counts[tab] })

	if c.Container != 0 || c.Image != 1 || c.Network != 2 || c.Volume != 3 {
		t.Fatalf("unexpected clamped cursors: %#v", c)
	}
}