package cthttp

import (
	"crypto/tls"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.HTTPServerConfigurator = (*ServerConfig)(nil)

type ServerConfig struct {
	Network string
	Address string `validate:"required"`
	TLS     *tls.Config
}

func (cfg ServerConfig) GetNetwork() string {
	if cfg.Network == "" {
		return "tcp"
	}
	return cfg.Network
}

func (cfg ServerConfig) GetAddress() string {
	return cfg.Address
}

func (cfg ServerConfig) GetTLS() *tls.Config {
	return cfg.TLS
}
