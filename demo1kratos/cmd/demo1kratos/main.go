package main

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/orzkratos/demokratos/demo1kratos/internal/conf"
	"github.com/orzkratos/zapkratos"
	"github.com/yyle88/done"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func newApp(gs *grpc.Server, hs *http.Server, zapKratos *zapkratos.ZapKratos) *kratos.App {
	return kratos.New(
		kratos.ID(done.VCE(os.Hostname()).Omit()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(zapKratos.NewLogger("network-service")),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()

	zapKratos := zapkratos.NewZapKratos(zaplog.LOGGER, zapkratos.NewOptions())
	zapLog := zapKratos.SubZap()
	zapLog.LOG.Info("version", zap.String("version", Version))
	zapLog.LOG.Info("reading-config-from-path", zap.String("config", flagconf))

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer rese.F0(c.Close)

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, zapKratos)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
