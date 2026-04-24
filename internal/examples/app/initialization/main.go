//nolint:mnd,gochecknoglobals // example
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	ctapp "github.com/surkovvs/ct/ctapp"
	"github.com/surkovvs/ct/internal/examples/app/modules"
)

var (
	appName         = "example_init"
	initTimeout     = time.Millisecond * 1000
	shutdownTimeout = time.Millisecond * 1200

	runDuration      = time.Millisecond * 1500
	shutdownDuration = time.Millisecond * 250
)

func main() {
	// log.SetOutput(io.Discard)
	log.SetFlags(log.Lmicroseconds)

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

	app.AddNamedEgressModule("module_bg_1_initRun", modules.NewModuleInitRun(modules.ModuleInitRunCfg{
		Name: "mock_bg_1_initRun",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 150,
			WantFail: false,
		},
		Run: modules.ElemCfg{TotalDur: runDuration},
	}))

	app.AddNamedEgressModule("module_bg_2_initSd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_bg_2_initSd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 300,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{TotalDur: shutdownDuration},
	}))

	app.AddNamedIngressModule("module_bgs_1_initRun", modules.NewModuleInitRun(modules.ModuleInitRunCfg{
		Name: "mock_bgs_1_initRun",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 300,
			WantFail: false,
		},
		Run: modules.ElemCfg{TotalDur: runDuration},
	}))

	app.AddNamedIngressModule("module_bgs_2_initSd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_bgs_2_initSd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 300,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{TotalDur: shutdownDuration},
	}))

	app.AddModuleToGroup("example_group_1", "module_1:1_initSd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_1:1_initSd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 150,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{TotalDur: shutdownDuration},
	}))

	app.AddModuleToGroup("example_group_1", "module_1:2_initRun", modules.NewModuleInitRun(modules.ModuleInitRunCfg{
		Name: "mock_1:2_initRun",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 150,
			WantFail: false,
		},
		Run: modules.ElemCfg{TotalDur: runDuration},
	}))

	app.AddModuleToGroup("example_group_2", "module_2:1_initRun", modules.NewModuleInitRun(modules.ModuleInitRunCfg{
		Name: "mock_2:1_initRun",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 150,
			WantFail: false,
		},
		Run: modules.ElemCfg{TotalDur: runDuration},
	}))

	app.AddModuleToGroup("example_group_2", "module_2:2_initSd", modules.NewModuleInitSd(modules.ModuleInitSdCfg{
		Name: "mock_2:2_initSd",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 150,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{TotalDur: shutdownDuration},
	}))

	ctx := context.Background()
	app.Start(ctx)
}
