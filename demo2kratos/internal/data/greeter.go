package data

import (
	"context"

	"github.com/orzkratos/demokratos/demo2kratos/internal/biz"
	"github.com/orzkratos/zapkratos"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

type greeterRepo struct {
	data   *Data
	zapLog *zaplog.Zap
}

// NewGreeterRepo .
func NewGreeterRepo(data *Data, zapKratos *zapkratos.ZapKratos) biz.GreeterRepo {
	return &greeterRepo{
		data:   data,
		zapLog: zapKratos.SubZap(),
	}
}

func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	r.zapLog.LOG.Info("save-greeter-message", zap.String("hello", g.Hello))
	return g, nil
}

func (r *greeterRepo) Update(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) FindByID(context.Context, int64) (*biz.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListByHello(context.Context, string) ([]*biz.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListAll(context.Context) ([]*biz.Greeter, error) {
	return nil, nil
}
