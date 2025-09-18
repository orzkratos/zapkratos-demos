package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/orzkratos/demokratos/demo2kratos"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/orzkratos/zapkratos"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/osexistpath/osmustexist"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
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

	rootBin := osmustexist.ROOT(filepath.Join(demo2kratos.SourceRoot(), "bin"))
	path1 := filepath.Join(rootBin, "log-newest.log")
	path2 := filepath.Join(rootBin, "log-oldest.log")

	if osmustexist.IsFile(path1) {
		must.Done(os.Truncate(path1, 0))
	}

	// Set default zap log to stdout and disk-file
	// 设置默认 zap 日志输出到标准输出和日志文件
	zaplog.SetLog(rese.P1(zaplog.NewZapLog(zaplog.NewConfig().
		AddOutputPaths(
			path1, path2, // Also log to file // 也输出到文件
		))))

	// Create zapkratos logger with default zaplog
	// 使用默认的 zaplog 创建 zapkratos 日志
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

	must.Done(c.Load())

	var cfg conf.Bootstrap
	must.Done(c.Scan(&cfg))

	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, zapKratos))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
