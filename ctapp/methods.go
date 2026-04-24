//nolint:fatcontext // its ok, maybe (we will see)
package ctapp

import (
	"context"
	"reflect"

	"github.com/surkovvs/ct/ctapp/component"
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
	})
	if !comp.IsValid() {
		a.logger.Error(`module addition`,
			"application", a.name,
			`group`, groupName,
			`module`, moduleName,
			`unapplyed`, extractTypeName(module),
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

func (a *App) AddNamedModule(moduleName string, module any) {
	a.AddModuleToGroup(nameficator.genGroupName(module), moduleName, module)
}

func (a *App) AddNamedIngressModule(moduleName string, module any) {
	a.AddModuleToGroup(IngressGroup, moduleName, module)
}

func (a *App) AddNamedEgressModule(moduleName string, module any) {
	a.AddModuleToGroup(EgressGroup, moduleName, module)
}

func (a *App) AddModule(module any) {
	a.AddModuleToGroup(
		nameficator.genGroupName(module),
		nameficator.genModuleName(module),
		module,
	)
}

func (a *App) AddIngressModule(module any) {
	a.AddModuleToGroup(IngressGroup, nameficator.genModuleName(module), module)
}

func (a *App) AddEgressModule(module any) {
	a.AddModuleToGroup(EgressGroup, nameficator.genModuleName(module), module)
}
