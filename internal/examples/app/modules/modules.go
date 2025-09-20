//nolint:depguard,gochecknoglobals // mock modules
package modules

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

var (
	ErrModule     = errors.New("module error")
	tickFrequency = time.Millisecond * 500
)

type ElemCfg struct {
	TotalDur         time.Duration
	DelayOnStopByCtx time.Duration
	WantFail         bool
}

type mimic struct {
	name    string
	hcCfg   ElemCfg
	initCfg ElemCfg
	runCfg  ElemCfg
	sdCfg   ElemCfg
}

func (m mimic) init(ctx context.Context, module any) error {
	return executing(ctx, module, m.name, m.initCfg, "init")
}

func (m mimic) run(ctx context.Context, module any) error {
	return executing(ctx, module, m.name, m.runCfg, "run")
}

func (m mimic) shutdown(ctx context.Context, module any) error {
	return executing(ctx, module, m.name, m.sdCfg, "shutdown")
}

func (m mimic) healthcheck(ctx context.Context, module any) error {
	return executing(ctx, module, m.name, m.hcCfg, "healthcheck")
}

func executing(ctx context.Context, module any, name string, cfg ElemCfg, stage string) error {
	log.Printf("--- internal module log: %s (%s): %s %s\n",
		name, reflect.ValueOf(module).Type().Name(), stage, "started")
	defer log.Printf("--- internal module log: %s (%s): %s %s\n",
		name, reflect.ValueOf(module).Type().Name(), stage, "stopped")
	exCtx := context.Background()
	if cfg.TotalDur != 0 {
		ctxTo, cancel := context.WithTimeout(ctx, cfg.TotalDur)
		defer cancel()
		exCtx = ctxTo
	}
	ticker := time.NewTicker(tickFrequency)
EndlessCycle:
	for {
		select {
		case <-ctx.Done():
			log.Printf("--- internal module log: %s [%s]: %s\n",
				reflect.ValueOf(module).Type().Name(), name, "stage "+stage+" ended by external ctx")
			if cfg.DelayOnStopByCtx != 0 {
				time.Sleep(cfg.DelayOnStopByCtx)
			}
			return fmt.Errorf("contex done: %w", ctx.Err())
		case <-exCtx.Done():
			break EndlessCycle
		case <-ticker.C:
			log.Printf("--- internal module log: %s [%s]: %s\n",
				reflect.ValueOf(module).Type().Name(), name, "executing "+stage)
		}
	}
	if cfg.WantFail {
		return fmt.Errorf("%s [%s]:%w", reflect.ValueOf(module).Type().Name(), name, ErrModule)
	}
	return nil
}

type (
	ModuleHcInitRunSd struct {
		m mimic
	}
	ModuleHcInitRunSdCfg struct {
		Name        string
		Healthcheck ElemCfg
		Init        ElemCfg
		Run         ElemCfg
		Shutdown    ElemCfg
	}
)

func NewModuleHcInitRunSd(cfg ModuleHcInitRunSdCfg) ModuleHcInitRunSd {
	return ModuleHcInitRunSd{
		m: mimic{
			name:    cfg.Name,
			initCfg: cfg.Init,
			runCfg:  cfg.Run,
			sdCfg:   cfg.Shutdown,
			hcCfg:   cfg.Healthcheck,
		},
	}
}

func (m ModuleHcInitRunSd) Healthcheck(ctx context.Context) error {
	return m.m.healthcheck(ctx, m)
}

func (m ModuleHcInitRunSd) Init(ctx context.Context) error {
	return m.m.init(ctx, m)
}

func (m ModuleHcInitRunSd) Run(ctx context.Context) error {
	return m.m.run(ctx, m)
}

func (m ModuleHcInitRunSd) Shutdown(ctx context.Context) error {
	return m.m.shutdown(ctx, m)
}

type (
	ModuleInitRunSd struct {
		m mimic
	}
	ModuleInitRunSdCfg struct {
		Name     string
		Init     ElemCfg
		Run      ElemCfg
		Shutdown ElemCfg
	}
)

func NewModuleInitRunSd(cfg ModuleInitRunSdCfg) ModuleInitRunSd {
	return ModuleInitRunSd{
		m: mimic{
			name:    cfg.Name,
			initCfg: cfg.Init,
			runCfg:  cfg.Run,
			sdCfg:   cfg.Shutdown,
		},
	}
}

func (m ModuleInitRunSd) Init(ctx context.Context) error {
	return m.m.init(ctx, m)
}

func (m ModuleInitRunSd) Run(ctx context.Context) error {
	return m.m.run(ctx, m)
}

func (m ModuleInitRunSd) Shutdown(ctx context.Context) error {
	return m.m.shutdown(ctx, m)
}

type (
	ModuleRunSd struct {
		m mimic
	}
	ModuleRunSdCfg struct {
		Name     string
		Run      ElemCfg
		Shutdown ElemCfg
	}
)

func NewModuleRunSd(cfg ModuleRunSdCfg) ModuleRunSd {
	return ModuleRunSd{
		m: mimic{
			name:   cfg.Name,
			runCfg: cfg.Run,
			sdCfg:  cfg.Shutdown,
		},
	}
}

func (m ModuleRunSd) Run(ctx context.Context) error {
	return m.m.run(ctx, m)
}

func (m ModuleRunSd) Shutdown(ctx context.Context) error {
	return m.m.shutdown(ctx, m)
}

type (
	ModuleHcInitRun struct {
		m mimic
	}
	ModuleHcInitRunCfg struct {
		Name        string
		Healthcheck ElemCfg
		Init        ElemCfg
		Run         ElemCfg
	}
)

func NewModuleHcInitRun(cfg ModuleHcInitRunCfg) ModuleHcInitRun {
	return ModuleHcInitRun{
		m: mimic{
			name:    cfg.Name,
			initCfg: cfg.Init,
			runCfg:  cfg.Run,
			hcCfg:   cfg.Healthcheck,
		},
	}
}

func (m ModuleHcInitRun) Healthcheck(ctx context.Context) error {
	return m.m.healthcheck(ctx, m)
}

func (m ModuleHcInitRun) Init(ctx context.Context) error {
	return m.m.init(ctx, m)
}

func (m ModuleHcInitRun) Run(ctx context.Context) error {
	return m.m.run(ctx, m)
}

type (
	ModuleInitRun struct {
		m mimic
	}
	ModuleInitRunCfg struct {
		Name string
		Init ElemCfg
		Run  ElemCfg
	}
)

func NewModuleInitRun(cfg ModuleInitRunCfg) ModuleInitRun {
	return ModuleInitRun{
		m: mimic{
			name:    cfg.Name,
			initCfg: cfg.Init,
			runCfg:  cfg.Run,
		},
	}
}

func (m ModuleInitRun) Init(ctx context.Context) error {
	return m.m.init(ctx, m)
}

func (m ModuleInitRun) Run(ctx context.Context) error {
	return m.m.run(ctx, m)
}

type (
	ModuleInitSd struct {
		m mimic
	}
	ModuleInitSdCfg struct {
		Name     string
		Init     ElemCfg
		Shutdown ElemCfg
	}
)

func NewModuleInitSd(cfg ModuleInitSdCfg) ModuleInitSd {
	return ModuleInitSd{
		m: mimic{
			name:    cfg.Name,
			initCfg: cfg.Init,
			sdCfg:   cfg.Shutdown,
		},
	}
}

func (m ModuleInitSd) Init(ctx context.Context) error {
	return m.m.init(ctx, m)
}

func (m ModuleInitSd) Shutdown(ctx context.Context) error {
	return m.m.shutdown(ctx, m)
}

type (
	ModuleHcInitSd struct {
		m mimic
	}
	ModuleHcInitSdCfg struct {
		Name        string
		Healthcheck ElemCfg
		Init        ElemCfg
		Shutdown    ElemCfg
	}
)

func NewModuleHcInitSd(cfg ModuleHcInitSdCfg) ModuleHcInitSd {
	return ModuleHcInitSd{
		m: mimic{
			name:    cfg.Name,
			initCfg: cfg.Init,
			sdCfg:   cfg.Shutdown,
			hcCfg:   cfg.Healthcheck,
		},
	}
}

func (m ModuleHcInitSd) Healthcheck(ctx context.Context) error {
	return m.m.healthcheck(ctx, m)
}

func (m ModuleHcInitSd) Init(ctx context.Context) error {
	return m.m.init(ctx, m)
}

func (m ModuleHcInitSd) Shutdown(ctx context.Context) error {
	return m.m.shutdown(ctx, m)
}
