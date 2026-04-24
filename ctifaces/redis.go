package ctifaces

import "crypto/tls"

type RedisClientConfigurator interface {
	GetAddresses() []string
	GetClientName() string
	GetDB() int
	GetPassword() string
	GetUsername() string
	GetTLS() *tls.Config
}
