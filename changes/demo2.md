# Changes

Code differences compared to source project demokratos.

## Makefile (+6 -1)

```diff
@@ -1,6 +1,11 @@
 GOHOSTOS:=$(shell go env GOHOSTOS)
 GOPATH:=$(shell go env GOPATH)
-VERSION=$(shell git describe --tags --always)
+#这是官方推荐的
+#VERSION=$(shell git describe --tags --always)
+#因为在开发阶段都是不打标签的，在很长的时间里可能都没有标签，这里使用较长的提交哈希
+#VERSION=$(shell git describe --tags 2>/dev/null || git rev-parse HEAD)
+#这样就能涵盖需要的
+VERSION=$(shell git describe --tags --always --abbrev=40 --dirty=+code)
 
 ifeq ($(GOHOSTOS), windows)
 	#the `find.exe` is different from `find` in bash/shell.
```

## cmd/demo2kratos/main.go (+42 -14)

```diff
@@ -3,18 +3,23 @@
 import (
 	"flag"
 	"os"
+	"path/filepath"
 
 	"github.com/go-kratos/kratos/v2"
 	"github.com/go-kratos/kratos/v2/config"
 	"github.com/go-kratos/kratos/v2/config/file"
-	"github.com/go-kratos/kratos/v2/log"
-	"github.com/go-kratos/kratos/v2/middleware/tracing"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
+	"github.com/orzkratos/demokratos/demo2kratos"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/zapkratos"
 	"github.com/yyle88/done"
 	"github.com/yyle88/must"
+	"github.com/yyle88/osexistpath/osmustexist"
 	"github.com/yyle88/rese"
+	"github.com/yyle88/tern/zerotern"
+	"github.com/yyle88/zaplog"
+	"go.uber.org/zap"
 )
 
 // go build -ldflags "-X main.Version=x.y.z"
@@ -31,13 +36,13 @@
 	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
 }
 
-func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
+func newApp(gs *grpc.Server, hs *http.Server, zapKratos *zapkratos.ZapKratos) *kratos.App {
 	return kratos.New(
 		kratos.ID(done.VCE(os.Hostname()).Omit()),
 		kratos.Name(Name),
 		kratos.Version(Version),
 		kratos.Metadata(map[string]string{}),
-		kratos.Logger(logger),
+		kratos.Logger(zapKratos.NewLogger("network-service")),
 		kratos.Server(
 			gs,
 			hs,
@@ -47,15 +52,38 @@
 
 func main() {
 	flag.Parse()
-	logger := log.With(log.NewStdLogger(os.Stdout),
-		"ts", log.DefaultTimestamp,
-		"caller", log.DefaultCaller,
-		"service.id", kratos.ID(done.VCE(os.Hostname()).Omit()),
-		"service.name", Name,
-		"service.version", Version,
-		"trace.id", tracing.TraceID(),
-		"span.id", tracing.SpanID(),
-	)
+
+	{
+		rootBin := osmustexist.ROOT(filepath.Join(demo2kratos.SourceRoot(), "bin"))
+		path1 := filepath.Join(rootBin, "log-newest.log")
+		path2 := filepath.Join(rootBin, "log-oldest.log")
+
+		// Clean session log on startup
+		// 启动时清空会话日志
+		if osmustexist.IsFile(path1) {
+			must.Done(os.Truncate(path1, 0))
+		}
+
+		// Set default zap log to stdout and disk-file
+		// 设置默认 zap 日志输出到标准输出和日志文件
+		zaplog.SetLog(rese.P1(zaplog.NewZapLog(zaplog.NewConfig().
+			AddOutputPaths(
+				path1, path2, // Also log to file // 也输出到文件
+			))).With(
+			zap.String("service", zerotern.VF(Name, func() string {
+				return filepath.Base(demo2kratos.SourceRoot())
+			})),
+			zap.String("version", zerotern.VV(Version, "v0.0.0")),
+		))
+	}
+
+	// Create zapkratos logger with default zaplog
+	// 使用默认的 zaplog 创建 zapkratos 日志
+	zapKratos := zapkratos.NewZapKratos(zaplog.LOGGER, zapkratos.NewOptions())
+	zapLog := zapKratos.SubZap()
+	zapLog.LOG.Info("application starting...")
+	zapLog.LOG.Info("reading-config-from-path", zap.String("config", flagconf))
+
 	c := config.New(
 		config.WithSource(
 			file.NewSource(flagconf),
@@ -68,7 +96,7 @@
 	var cfg conf.Bootstrap
 	must.Done(c.Scan(&cfg))
 
-	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
+	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, zapKratos))
 	defer cleanup()
 
 	// start and wait for stop signal
```

## cmd/demo2kratos/wire.go (+2 -2)

```diff
@@ -6,16 +6,16 @@
 
 import (
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/data"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/server"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+	"github.com/orzkratos/zapkratos"
 )
 
 // wireApp init kratos application.
-func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
+func wireApp(*conf.Server, *conf.Data, *zapkratos.ZapKratos) (*kratos.App, func(), error) {
 	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
 }
```

## cmd/demo2kratos/wire_gen.go (+9 -9)

```diff
@@ -7,28 +7,28 @@
 
 import (
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/data"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/server"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+	"github.com/orzkratos/zapkratos"
 )
 
 // Injectors from wire.go:
 
 // wireApp init kratos application.
-func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
-	dataData, cleanup, err := data.NewData(confData, logger)
+func wireApp(confServer *conf.Server, confData *conf.Data, zapKratos *zapkratos.ZapKratos) (*kratos.App, func(), error) {
+	dataData, cleanup, err := data.NewData(confData, zapKratos)
 	if err != nil {
 		return nil, nil, err
 	}
-	greeterRepo := data.NewGreeterRepo(dataData, logger)
-	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
-	greeterService := service.NewGreeterService(greeterUsecase)
-	grpcServer := server.NewGRPCServer(confServer, greeterService, logger)
-	httpServer := server.NewHTTPServer(confServer, greeterService, logger)
-	app := newApp(logger, grpcServer, httpServer)
+	greeterRepo := data.NewGreeterRepo(dataData, zapKratos)
+	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, zapKratos)
+	greeterService := service.NewGreeterService(greeterUsecase, zapKratos)
+	grpcServer := server.NewGRPCServer(confServer, greeterService, zapKratos)
+	httpServer := server.NewHTTPServer(confServer, greeterService, zapKratos)
+	app := newApp(grpcServer, httpServer, zapKratos)
 	return app, func() {
 		cleanup()
 	}, nil
```

## internal/biz/greeter.go (+10 -6)

```diff
@@ -4,8 +4,9 @@
 	"context"
 
 	"github.com/go-kratos/kratos/v2/errors"
-	"github.com/go-kratos/kratos/v2/log"
 	v1 "github.com/orzkratos/demokratos/demo2kratos/api/helloworld/v1"
+	"github.com/orzkratos/zapkratos"
+	"github.com/yyle88/zaplog"
 )
 
 var (
@@ -29,17 +30,20 @@
 
 // GreeterUsecase is a Greeter usecase.
 type GreeterUsecase struct {
-	repo GreeterRepo
-	log  *log.Helper
+	repo   GreeterRepo
+	zapLog *zaplog.Zap
 }
 
 // NewGreeterUsecase new a Greeter usecase.
-func NewGreeterUsecase(repo GreeterRepo, logger log.Logger) *GreeterUsecase {
-	return &GreeterUsecase{repo: repo, log: log.NewHelper(logger)}
+func NewGreeterUsecase(repo GreeterRepo, zapKratos *zapkratos.ZapKratos) *GreeterUsecase {
+	return &GreeterUsecase{
+		repo:   repo,
+		zapLog: zapKratos.SubZap(),
+	}
 }
 
 // CreateGreeter creates a Greeter, and returns the new Greeter.
 func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *Greeter) (*Greeter, error) {
-	uc.log.WithContext(ctx).Infof("CreateGreeter: %v", g.Hello)
+	uc.zapLog.SUG.Infof("CreateGreeter: %v", g.Hello)
 	return uc.repo.Save(ctx, g)
 }
```

## internal/data/data.go (+5 -3)

```diff
@@ -1,9 +1,9 @@
 package data
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
+	"github.com/orzkratos/zapkratos"
 )
 
 // ProviderSet is data providers.
@@ -15,9 +15,11 @@
 }
 
 // NewData .
-func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
+func NewData(c *conf.Data, zapKratos *zapkratos.ZapKratos) (*Data, func(), error) {
+	zapLog := zapKratos.SubZap()
+	zapLog.SUG.Info("creating data resources")
 	cleanup := func() {
-		log.NewHelper(logger).Info("closing the data resources")
+		zapLog.SUG.Info("closing the data resources")
 	}
 	return &Data{}, cleanup, nil
 }
```

## internal/data/greeter.go (+9 -6)

```diff
@@ -3,24 +3,27 @@
 import (
 	"context"
 
-	"github.com/go-kratos/kratos/v2/log"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/zapkratos"
+	"github.com/yyle88/zaplog"
+	"go.uber.org/zap"
 )
 
 type greeterRepo struct {
-	data *Data
-	log  *log.Helper
+	data   *Data
+	zapLog *zaplog.Zap
 }
 
 // NewGreeterRepo .
-func NewGreeterRepo(data *Data, logger log.Logger) biz.GreeterRepo {
+func NewGreeterRepo(data *Data, zapKratos *zapkratos.ZapKratos) biz.GreeterRepo {
 	return &greeterRepo{
-		data: data,
-		log:  log.NewHelper(logger),
+		data:   data,
+		zapLog: zapKratos.SubZap(),
 	}
 }
 
 func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
+	r.zapLog.LOG.Info("save-greeter-message", zap.String("hello", g.Hello))
 	return g, nil
 }
 
```

## internal/server/grpc.go (+4 -2)

```diff
@@ -1,19 +1,21 @@
 package server
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
+	"github.com/go-kratos/kratos/v2/middleware/logging"
 	"github.com/go-kratos/kratos/v2/middleware/recovery"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	v1 "github.com/orzkratos/demokratos/demo2kratos/api/helloworld/v1"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+	"github.com/orzkratos/zapkratos"
 )
 
 // NewGRPCServer new a gRPC server.
-func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *grpc.Server {
+func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, zapKratos *zapkratos.ZapKratos) *grpc.Server {
 	var opts = []grpc.ServerOption{
 		grpc.Middleware(
 			recovery.Recovery(),
+			logging.Server(zapKratos.GetLogger("grpc-request")),
 		),
 	}
 	if c.Grpc.Network != "" {
```

## internal/server/http.go (+4 -2)

```diff
@@ -1,19 +1,21 @@
 package server
 
 import (
-	"github.com/go-kratos/kratos/v2/log"
+	"github.com/go-kratos/kratos/v2/middleware/logging"
 	"github.com/go-kratos/kratos/v2/middleware/recovery"
 	"github.com/go-kratos/kratos/v2/transport/http"
 	v1 "github.com/orzkratos/demokratos/demo2kratos/api/helloworld/v1"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/service"
+	"github.com/orzkratos/zapkratos"
 )
 
 // NewHTTPServer new an HTTP server.
-func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *http.Server {
+func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, zapKratos *zapkratos.ZapKratos) *http.Server {
 	var opts = []http.ServerOption{
 		http.Middleware(
 			recovery.Recovery(),
+			logging.Server(zapKratos.GetLogger("http-request")),
 		),
 	}
 	if c.Http.Network != "" {
```

## internal/service/greeter.go (+12 -3)

```diff
@@ -5,25 +5,34 @@
 
 	v1 "github.com/orzkratos/demokratos/demo2kratos/api/helloworld/v1"
 	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
+	"github.com/orzkratos/zapkratos"
+	"github.com/yyle88/zaplog"
+	"go.uber.org/zap"
 )
 
 // GreeterService is a greeter service.
 type GreeterService struct {
 	v1.UnimplementedGreeterServer
 
-	uc *biz.GreeterUsecase
+	uc     *biz.GreeterUsecase
+	zapLog *zaplog.Zap
 }
 
 // NewGreeterService new a greeter service.
-func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
-	return &GreeterService{uc: uc}
+func NewGreeterService(uc *biz.GreeterUsecase, zapKratos *zapkratos.ZapKratos) *GreeterService {
+	return &GreeterService{
+		uc:     uc,
+		zapLog: zapKratos.SubZap(),
+	}
 }
 
 // SayHello implements helloworld.GreeterServer.
 func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
+	s.zapLog.LOG.Info("receive-hello-message", zap.String("name", in.Name))
 	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
 	if err != nil {
 		return nil, err
 	}
+	s.zapLog.LOG.Info("reply-a-hello-message", zap.String("name", in.Name))
 	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
 }
```

