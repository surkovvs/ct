package ctapp

import "fmt"

//nolint:gochecknoglobals // as planned
var nameficator groupSequentialNamer

type groupSequentialNamer struct {
	moduleCounter int
	groupCounter  int
}

func (r *groupSequentialNamer) getNextGroupName() string {
	r.groupCounter++
	return fmt.Sprintf("group_%d", r.groupCounter)
}

func (r *groupSequentialNamer) getNextModuleNum() int {
	r.moduleCounter++
	return r.moduleCounter
}
