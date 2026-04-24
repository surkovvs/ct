package ctapp

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/surkovvs/ct/ctifaces"
)

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

func extractTypeName(module any) string {
	return reflect.ValueOf(module).Type().Name()
}

func (r *groupSequentialNamer) genModuleName(module any) string {
	var modulePrefix string
	if np, ok := module.(ctifaces.NamePrefixer); ok {
		modulePrefix = np.GetModuleNamePrefix()
	} else {
		modulePrefix = extractTypeName(module)
	}

	return modulePrefix + "_" + strconv.Itoa(nameficator.getNextModuleNum())
}

func (r *groupSequentialNamer) genGroupName(module any) string {
	var groupName string
	if gi, ok := module.(ctifaces.GroupIdetifer); ok {
		groupName = gi.PreidentifyModuleGroup()
	} else {
		groupName = nameficator.getNextGroupName()
	}

	return groupName
}
