package ctkafka

import (
	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.KafkaClientConfigurator = (*ClientConfig)(nil)

type (
	ClientConfig struct {
		Addrs             []string `validate:"required,min=1"`
		ClientID          string   `validate:"required"`
		ChannelBufferSize *int
		Auth              *Auth
		CertPath          *string
	}
	Auth struct {
		User string
		Pass string
	}
)

func (c ClientConfig) GetBrokerAddresses() []string {
	return c.Addrs
}

func (c ClientConfig) GetCLientID() string {
	return c.ClientID
}

func (c ClientConfig) GetChannelBufferSize() *int {
	return c.ChannelBufferSize
}

func (c ClientConfig) GetCertPath() *string {
	return c.CertPath
}

func (c ClientConfig) AuthNeeded() bool {
	if c.Auth != nil {
		return true
	}
	return false
}

func (c ClientConfig) GetUserName() string {
	return c.Auth.User
}

func (c ClientConfig) GetPassword() string {
	return c.Auth.Pass
}
