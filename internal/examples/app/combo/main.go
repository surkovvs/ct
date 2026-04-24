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
	initTimeout     = time.Millisecond * 1500
	shutdownTimeout = time.Millisecond * 1500
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

	app.AddNamedEgressModule("module_eg_1", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_eg_1",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
	}))

	app.AddNamedEgressModule("module_eg_2", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_eg_2",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
	}))

	app.AddNamedIngressModule("module_ing_1", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_ing_1",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
	}))
	app.AddNamedIngressModule("module_ing_2", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_ing_2",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 450,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("group_1", "module_seq_1:group_1", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_seq_1:group_1",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 200,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 200,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("group_1", "module_seq_2:group_1", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_seq_2:group_1",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 200,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 200,
			WantFail: false,
		},
	}))

	app.AddModuleToGroup("group_2", "module_seq:group_2", modules.NewModuleInitRunSd(modules.ModuleInitRunSdCfg{
		Name: "module_seq:group_2",
		Init: modules.ElemCfg{
			TotalDur: time.Millisecond * 200,
			WantFail: false,
		},
		Run: modules.ElemCfg{
			TotalDur: time.Millisecond * 2000,
			WantFail: false,
		},
		Shutdown: modules.ElemCfg{
			TotalDur: time.Millisecond * 200,
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
