package shared

import (
	"errors"
	"testing"
)

func TestFail(t *testing.T) {
	err := errors.New("boom")
	transition := Fail(err)
	if transition.Loading {
		t.Fatalf("Loading = true, want false")
	}
	if transition.Stage != StageIdle {
		t.Fatalf("Stage = %d, want %d", transition.Stage, StageIdle)
	}
	if transition.Err != err {
		t.Fatalf("Err = %v, want %v", transition.Err, err)
	}
}

func TestFinish(t *testing.T) {
	err := errors.New("failed")
	transition, ok := Finish(err)
	if ok {
		t.Fatalf("ok = true, want false")
	}
	if transition.Loading {
		t.Fatalf("Loading = true, want false")
	}
	if transition.Stage != StageIdle {
		t.Fatalf("Stage = %d, want %d", transition.Stage, StageIdle)
	}
	if transition.Err != err {
		t.Fatalf("Err = %v, want %v", transition.Err, err)
	}

	transition, ok = Finish(nil)
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if transition.Err != nil {
		t.Fatalf("Err = %v, want nil", transition.Err)
	}
}
