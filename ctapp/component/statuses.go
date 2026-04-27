package component

import "github.com/surkovvs/ct/ctapp/zorro"

type (
	Status string
	Stage  string
)

const (
	ready     zorro.Status = 4369  // 0001000100010001
	inProcess zorro.Status = 8738  // 0010001000100010
	done      zorro.Status = 17476 // 0100010001000100
	failed    zorro.Status = 34952 // 1000100010001000

	initMask        zorro.Mask = 15    // 0000000000001111
	runMask         zorro.Mask = 240   // 0000000011110000
	shutdownMask    zorro.Mask = 3840  // 0000111100000000
	healthcheckMask zorro.Mask = 61440 // 1111000000000000

	StageInit        Stage = "init"
	StageRun         Stage = "run"
	StageShutdown    Stage = "shutdown"
	StageHealthcheck Stage = "healthcheck"

	StatusUndefined Status = "undefined"
	StatusReady     Status = "ready"
	StatusInProcess Status = "in_process"
	StatusDone      Status = "done"
	StatusFailed    Status = "failed"
)

//nolint:gochecknoglobals // skip
var namedStatuses = map[uint64]Status{
	0: StatusUndefined,
	1: StatusReady,
	2: StatusInProcess,
	4: StatusDone,
	8: StatusFailed,
}

type statusProvider struct {
	provided zorro.Mask
	comp     Comp
}

func (r statusProvider) setInProcess() {
	r.comp.status.SetStatus(inProcess, r.provided)
}

func (r statusProvider) tryChangeStatus(prev zorro.Status, next zorro.Status) bool {
	return r.comp.status.TryChangeStatus(prev, next, r.provided)
}

func (r statusProvider) setDone() {
	r.comp.status.SetStatus(done, r.provided)
}

func (r statusProvider) setFailed() {
	r.comp.status.SetStatus(failed, r.provided)
}

func (r statusProvider) isReady() bool {
	return r.comp.status.GetStatus().CompareMasked(ready, r.provided)
}

func (r statusProvider) isInProcess() bool {
	return r.comp.status.GetStatus().CompareMasked(inProcess, r.provided)
}

func (r statusProvider) isDone() bool {
	return r.comp.status.GetStatus().CompareMasked(done, r.provided)
}

func (r statusProvider) isFailed() bool {
	return r.comp.status.GetStatus().CompareMasked(failed, r.provided)
}

func (r statusProvider) namedStatus() Status {
	bald := zorro.Status(r.comp.status.GetStatus().Querying(r.provided))
	s, ok := namedStatuses[bald.ShiftTrailingZeros(r.provided)]
	if !ok {
		return "unknown"
	}
	return s
}
