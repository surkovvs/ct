package ctdb

import (
	"crypto/tls"
	"fmt"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.SQLConfigurator = (*Config)(nil)

type Config struct {
	ctifaces.Logger `mapstructure:"-"`
	// prefer DSN for coonection
	DSN        string
	Host       string
	Port       uint16
	Name       string
	User       string
	Pass       string
	LogQueries bool

	PGPoolConfig `mapstructure:"Pool"`

	TLS *tls.Config
}

// TODO: add TLS
func (cfg Config) GetDSN() string {
	if cfg.DSN != "" {
		return cfg.DSN
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.User,
		cfg.Pass,
	)
}

func (cfg Config) GetLogger() ctifaces.Logger {
	return cfg.Logger
}

func (cfg Config) WithLogger(log ctifaces.Logger) Config {
	cfg.Logger = log
	return cfg
}

func (cfg Config) LogQueriesEnabled() bool {
	return cfg.LogQueries
}

func (cfg Config) GetPGPoolConfig() ctifaces.PGPoolConfigurator {
	return cfg.PGPoolConfig
}

func (cfg Config) GetTLS() *tls.Config {
	return cfg.TLS
}
