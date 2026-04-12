package core

import "testing"

func TestFilterContainersByScope_ShowAllReturnsOriginalSlice(t *testing.T) {
	containers := []ContainerRow{{Name: "one"}, {Name: "two"}}

	got := FilterContainersByScope(containers, true)
	if len(got) != len(containers) {
		t.Fatalf("FilterContainersByScope(..., true) len = %d, want %d", len(got), len(containers))
	}
	if &got[0] != &containers[0] {
		t.Fatalf("FilterContainersByScope(..., true) should return original slice")
	}
}

func TestFilterContainersByScope_RunningOnlyCaseInsensitive(t *testing.T) {
	containers := []ContainerRow{
		{Name: "run-1", State: "running"},
		{Name: "run-2", State: "RUNNING"},
		{Name: "stopped", State: "exited"},
		{Name: "other", State: "created"},
	}

	got := FilterContainersByScope(containers, false)
	if len(got) != 2 {
		t.Fatalf("FilterContainersByScope(..., false) len = %d, want 2", len(got))
	}
	if got[0].Name != "run-1" || got[1].Name != "run-2" {
		t.Fatalf("FilterContainersByScope(..., false) = %#v, want running rows only", got)
	}
}
