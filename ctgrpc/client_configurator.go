package ctgrpc

import (
	"crypto/tls"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.GRPCClientConfigurator = (*ClientConfig)(nil)

type ClientConfig struct {
	Address string `validate:"required"`
	TLS     *tls.Config
}

func (cfg ClientConfig) GetAddress() string {
	return cfg.Address
}

func (cfg ClientConfig) GetTLS() *tls.Config {
	return cfg.TLS
}
