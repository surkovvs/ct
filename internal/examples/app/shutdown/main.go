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
	appName         = "example_shutdown"
	shutdownTimeout = time.Millisecond * 1000
	owertime        = shutdownTimeout + time.Millisecond*100
)

func main() {
	// log.SetOutput(io.Discard)

	app := ctapp.New(
		ctapp.WithConfig(ctapp.ConfigApp{
			Silient: false,
			// TolerantMode:    false,
			TolerantMode:    true,
			Name:            &appName,
			InitTimeout:     nil,
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

	app.AddBackgroundModule("module_bg_1_init_sd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_bg_1_init_sd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	app.AddBackgroundModule("module_bg_2_run_sd", modules.NewModuleRunSd(modules.ModuleRunSdCfg{
		Name: "mock_bg_2_run_sd",
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: true,
		},
	}))

	app.AddBackgroundSyncModule("module_bgs_1_init_sd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_bgs_1_init_sd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			// TotalDur: owertime,
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	app.AddBackgroundSyncModule("module_bgs_2_run_sd", modules.NewModuleRunSd(modules.ModuleRunSdCfg{
		Name: "mock_bgs_2_run_sd",
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("example_group_1", "module_1_1_init_sd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_1_1_init_sd",
		Init: modules.ElemCfg{
			TotalDur: 0,
			// DelayOnStopByCtx: owertime,
			// WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("example_group_1", "module_1_2_run_sd", modules.NewModuleRunSd(modules.ModuleRunSdCfg{
		Name: "mock_1_2_run_sd",
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: true,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("example_group_2", "module_2_1_init_sd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_2_1_init_sd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: true,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("example_group_2", "module_2_2_run_sd", modules.NewModuleRunSd(modules.ModuleRunSdCfg{
		Name: "mock_2_2_run_sd",
		Run: modules.ElemCfg{
			TotalDur: 0,
			// DelayOnStopByCtx: owertime,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 250,
			WantFail: false,
		},
	}))

	ctx := context.Background()
	app.Start(ctx)
}
