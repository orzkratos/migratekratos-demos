package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/cfgpath"
	"github.com/orzkratos/demokratos/demo2kratos/cmd/demo2kratos/subcmds"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/orzkratos/demokratos/demo2kratos/internal/pkg/appcfg"
	"github.com/spf13/cobra"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/must/mustslice"
	"github.com/yyle88/rese"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
)

func init() {
	fmt.Println("service-name:", Name)
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(done.VCE(os.Hostname()).Omit()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", kratos.ID(done.VCE(os.Hostname()).Omit()),
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	var rootCmd = &cobra.Command{
		Use:   "demo2kratos",
		Short: "A Kratos microservice with database migration",
		Run: func(cmd *cobra.Command, args []string) {
			mustslice.None(args)
			if cfg := appcfg.ParseConfig(cfgpath.ConfigPath); cfg.Server.AutoRun {
				runApp(cfg, logger)
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "conf", "c", "./configs", "config path, eg: --conf=config.yaml")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Start the application",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
			runApp(cfg, logger)
		},
	})

	rootCmd.AddCommand(subcmds.NewVersionCmd(Name, Version, logger))
	rootCmd.AddCommand(subcmds.NewMigrateCmd(logger))

	must.Done(rootCmd.Execute())
}

func runApp(cfg *conf.Bootstrap, logger log.Logger) {
	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
