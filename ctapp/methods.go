//nolint:fatcontext // its ok, maybe (we will see)
package ctapp

import (
	"context"
	"reflect"
	"strconv"

	"github.com/surkovvs/ct/ctapp/component"
	"github.com/surkovvs/ct/ctifaces"
)

func (a *App) Start(ctx context.Context) {
	go a.accompaniment()
	a.execution.runCtx, a.execution.initRunCancel = context.WithCancel(ctx)
	if a.execution.initTimeout != nil {
		var cancel context.CancelFunc
		a.execution.initCtx, cancel = context.WithTimeout(a.execution.runCtx, *a.execution.initTimeout)
		defer cancel()
	} else {
		a.execution.initCtx = a.execution.runCtx
	}
	a.logger.Debug(`app started`, `application`, a.name)
	go a.exec()
	<-a.shutdown.shutdownDone
}

func (a *App) AddModuleToGroup(groupName, moduleName string, module any) {
	comp := component.DefineComponent(component.Define{
		GroupName: groupName,
		CompName:  moduleName,
		Component: module,
		ExecWg:    a.execution.wg,
	})
	if !comp.IsValid() {
		a.logger.Error(`module addition`,
			"application", a.name,
			`group`, groupName,
			`module`, moduleName,
			`unapplyed`, reflect.ValueOf(module).Type().Name(),
			`error`, "module does not implement any of valid methods")
		return
	}
	if err := a.storage.AddComponent(groupName, comp); err != nil {
		a.logger.Error(`module addition`,
			"application", a.name,
			`group`, groupName,
			`module`, moduleName,
			`unapplyed`, reflect.ValueOf(module).Elem().Type().Elem().Name(),
			`error`, err)
	}
}

func (a *App) AddModule(moduleName string, module any) {
	groupName := nameficator.getNextGroupName()
	a.AddModuleToGroup(groupName, moduleName, module)
}

func (a *App) AddBackgroundModule(moduleName string, module any) {
	a.AddModuleToGroup(BackgroundGroup, moduleName, module)
}

func (a *App) AddBackgroundSyncModule(moduleName string, module any) {
	a.AddModuleToGroup(BackgroundSyncGroup, moduleName, module)
}

func (a *App) AddUnnamedModule(module any) {
	var groupName, modulePrefix string

	if gi, ok := module.(ctifaces.GroupIdetifer); ok {
		groupName = gi.PreidentifyModuleGroup()
	} else {
		groupName = nameficator.getNextGroupName()
	}

	if np, ok := module.(ctifaces.NamePrefixer); ok {
		modulePrefix = np.GetModuleNamePrefix()
	} else {
		modulePrefix = reflect.ValueOf(module).Type().Name()
	}
	moduleName := modulePrefix + "_" + strconv.Itoa(nameficator.getNextModuleNum())

	a.AddModuleToGroup(groupName, moduleName, module)
}

func (a *App) AddUnnamedBackgroundModule(module any) {
	a.AddModuleToGroup(BackgroundGroup, reflect.ValueOf(module).Type().Name(), module)
}

func (a *App) AddUnnamedBackgroundSyncModule(module any) {
	a.AddModuleToGroup(BackgroundSyncGroup, reflect.ValueOf(module).Type().Name(), module)
}
