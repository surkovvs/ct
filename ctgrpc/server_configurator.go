package ctgrpc

import (
	"crypto/tls"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.GRPCServerConfigurator = (*ServerConfig)(nil)

type ServerConfig struct {
	Network *string // "tcp" by default
	Address string  `validate:"required"`
	TLS     *tls.Config
}

// GetNetwork - returns "tcp" if network is nil.
func (cfg ServerConfig) GetNetwork() string {
	if cfg.Network == nil {
		return "tcp"
	}
	return *cfg.Network
}

func (cfg ServerConfig) GetAddress() string {
	return cfg.Address
}

func (cfg ServerConfig) GetTLS() *tls.Config {
	return cfg.TLS
}
