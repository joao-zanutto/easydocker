package shared

import "time"

// SpinnerTickInterval is the common interval used for all spinner ticks.
const SpinnerTickInterval = time.Second / 7

// Stage represents the async loading lifecycle in the TUI.
type Stage int

const (
	StageIdle Stage = iota
	StageContainers
	StageResources
	StageMetrics
)

// Transition captures the minimal state mutation needed after a loading event.
type Transition struct {
	Loading bool
	Stage   Stage
	Err     error
}

func Begin(stage Stage) Transition {
	return Transition{Loading: true, Stage: stage}
}

func Fail(err error) Transition {
	return Transition{Loading: false, Stage: StageIdle, Err: err}
}

func Finish(err error) (Transition, bool) {
	return Transition{Loading: false, Stage: StageIdle, Err: err}, err == nil
}

func (s Stage) Next() Stage {
	switch s {
	case StageContainers:
		return StageResources
	case StageResources:
		return StageMetrics
	case StageMetrics:
		return StageIdle
	default:
		return StageIdle
	}
}
