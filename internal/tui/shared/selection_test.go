package shared

import "testing"

func TestMoveActiveTab(t *testing.T) {
	if got := MoveActiveTab(0, 1, TabContainers, TabVolumes); got != 1 {
		t.Fatalf("MoveActiveTab(0,+1) = %d, want 1", got)
	}
	if got := MoveActiveTab(0, -1, TabContainers, TabVolumes); got != 0 {
		t.Fatalf("MoveActiveTab lower clamp = %d, want 0", got)
	}
	if got := MoveActiveTab(3, 1, TabContainers, TabVolumes); got != 3 {
		t.Fatalf("MoveActiveTab upper clamp = %d, want 3", got)
	}
}

func TestCursorAPIs(t *testing.T) {
	c := Cursors{Container: 1, Image: 2, Network: 3, Volume: 4}

	if got, ok := CursorForTab(c, TabContainers); !ok || got != 1 {
		t.Fatalf("CursorForTab(container) = (%d,%v), want (1,true)", got, ok)
	}
	if ok := SetCursorForTab(&c, TabNetworks, 9); !ok || c.Network != 9 {
		t.Fatalf("SetCursorForTab(network) failed")
	}

	_ = MoveCursorForTab(&c, TabContainers, 10, 3)
	if c.Container != 2 {
		t.Fatalf("MoveCursorForTab clamp upper = %d, want 2", c.Container)
	}

	_ = ClampCursorForTab(&c, TabImages, 0)
	if c.Image != 0 {
		t.Fatalf("ClampCursorForTab empty list = %d, want 0", c.Image)
	}

}

func TestClampAllCursors(t *testing.T) {
	c := Cursors{Container: 10, Image: 10, Network: 10, Volume: 10}
	tabs := []Tab{TabContainers, TabImages, TabNetworks, TabVolumes}
	counts := map[Tab]int{TabContainers: 1, TabImages: 2, TabNetworks: 3, TabVolumes: 4}

	ClampAllCursors(&c, tabs, func(tab Tab) int { return counts[tab] })

	if c.Container != 0 || c.Image != 1 || c.Network != 2 || c.Volume != 3 {
		t.Fatalf("unexpected clamped cursors: %#v", c)
	}
}
