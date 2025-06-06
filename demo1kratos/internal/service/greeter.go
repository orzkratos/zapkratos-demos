package service

import (
	"context"

	v1 "github.com/orzkratos/demokratos/demo1kratos/api/helloworld/v1"
	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
	"github.com/orzkratos/zapkratos"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc     *biz.GreeterUsecase
	zapLog *zaplog.Zap
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase, zapKratos *zapkratos.ZapKratos) *GreeterService {
	return &GreeterService{
		uc:     uc,
		zapLog: zapKratos.SubZap(),
	}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	s.zapLog.LOG.Info("receive-hello-message", zap.String("name", in.Name))
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	s.zapLog.LOG.Info("reply-a-hello-message", zap.String("name", in.Name))
	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
}
