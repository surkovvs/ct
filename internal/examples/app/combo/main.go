//nolint:mnd,funlen,gochecknoglobals // example
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	ctapp "github.com/surkovvs/ct/ctapp"
	"github.com/surkovvs/ct/internal/examples/app/modules"
)

var (
	appName         = "example"
	initTimeout     = time.Millisecond * 2000
	shutdownTimeout = time.Millisecond * 2000
)

func main() {
	// log.SetOutput(io.Discard)

	app := ctapp.New(
		ctapp.WithConfig(ctapp.Config{
			Silient:         false,
			TolerantMode:    true,
			Name:            &appName,
			InitTimeout:     &initTimeout,
			ShutdownTimeout: &shutdownTimeout,
		}),
		ctapp.WithLogger(slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
				// Level: slog.LevelWarn,
			}),
		)),
		ctapp.WithProvidedSigs(os.Interrupt),
	)

	app.AddBackgroundModule("module_bg_1", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "mock_bg_1",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 750,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 3000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 1000,
			WantFail: false,
		},
	}))

	app.AddBackgroundModule("module_bg_2", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_bg_2",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 750,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 1000,
			WantFail: false,
		},
	}))

	app.AddBackgroundModule("module_bg_3", modules.NewModuleHcInitSd(modules.ModuleHcInitSdCfg{
		Name: "mock_bg_3",
		Healthcheck: modules.ElemCfg{
			TotalDur: time.Millisecond * 750,
			WantFail: true,
		},
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 750,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 1000,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("example_group_1", "module_1:1", modules.NewModuleHcInitRun(modules.ModuleHcInitRunCfg{
		Name: "mock_1:1",
		Healthcheck: modules.ElemCfg{
			TotalDur: time.Millisecond * 1500,
			WantFail: false,
		},
		Init: modules.ElemCfg{
			TotalDur:         time.Millisecond * 500,
			DelayOnStopByCtx: time.Millisecond * 300,
			WantFail:         false,
		},
		Run: modules.ElemCfg{
			DelayOnStopByCtx: time.Millisecond * 500,
		},
	}))

	app.AddModuleToGroup("example_group_1", "module_1:2", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "mock_1:2",
		Init: modules.ElemCfg{
			TotalDur:         time.Millisecond * 500,
			DelayOnStopByCtx: time.Millisecond * 300,
			WantFail:         false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 1600,
			WantFail: false,
		},
		Run: modules.ElemCfg{},
	}))

	app.AddModuleToGroup("example_group_2", "module_2", modules.NewModuleRunSd(modules.ModuleRunSdCfg{
		Name: "mock_2",
		Run: modules.ElemCfg{
			TotalDur:         time.Millisecond * 4000,
			DelayOnStopByCtx: time.Millisecond * 1000,
			WantFail:         false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 600,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("example_group_3", "module_3", modules.NewModuleRunSd(modules.ModuleRunSdCfg{
		Name: "mock_3",
		Run: modules.ElemCfg{
			TotalDur:         time.Millisecond * 1000,
			DelayOnStopByCtx: time.Millisecond * 1000,
			WantFail:         true,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 600,
			WantFail: false,
		},
	}))

	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 2)
	// 		ctxTo, cancel := context.WithTimeout(context.Background(), time.Millisecond*1000)
	// 		errs := app.Healthcheck(ctxTo)
	// 		cancel()
	// 		if len(errs) != 0 {
	// 			log.Printf("*** healthcheck errors reporting: %+v", errs)
	// 		}
	// 	}
	// }()
	ctx := context.Background()
	app.Start(ctx)
}
