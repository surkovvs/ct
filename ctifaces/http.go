package ctifaces

import "crypto/tls"

type HTTPServerConfigurator interface {
	GetNetwork() string
	GetAddress() string
	GetTLS() *tls.Config
}
