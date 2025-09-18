package data

import (
	"github.com/google/wire"
	"github.com/orzkratos/demokratos/demo2kratos/internal/conf"
	"github.com/orzkratos/zapkratos"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	// TODO wrapped database client
}

// NewData .
func NewData(c *conf.Data, zapKratos *zapkratos.ZapKratos) (*Data, func(), error) {
	zapLog := zapKratos.SubZap()
	zapLog.SUG.Info("creating data resources")
	cleanup := func() {
		zapLog.SUG.Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
