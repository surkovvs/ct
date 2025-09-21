package ctdb

import (
	"crypto/tls"
	"fmt"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.SQLConfigurator = SQLConfig{}

type SQLConfig struct {
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
func (cfg SQLConfig) GetDSN() string {
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

func (cfg SQLConfig) GetLogger() ctifaces.Logger {
	return cfg.Logger
}

func (cfg SQLConfig) WithLogger(log ctifaces.Logger) SQLConfig {
	cfg.Logger = log
	return cfg
}

func (cfg SQLConfig) LogQueriesEnabled() bool {
	return cfg.LogQueries
}

func (cfg SQLConfig) GetPGPoolConfig() ctifaces.PGPoolConfigurator {
	return cfg.PGPoolConfig
}

func (cfg SQLConfig) GetTLS() *tls.Config {
	return cfg.TLS
}
