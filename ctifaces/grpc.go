package ctifaces

import "crypto/tls"

type GRPCServerConfigurator interface {
	GetNetwork() string
	GetAddress() string
	GetTLS() *tls.Config
}

type GRPCClientConfigurator interface {
	GetAddress() string
	GetTLS() *tls.Config
}
