package ctcfg

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

func ParseFile[T any](filepath *string) (*T, error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	switch {
	case filepath != nil:
		viper.SetConfigFile(*filepath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf(`read config by file path: %w`, err)
		}
	default:
		viper.SetConfigFile("config.yml")
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf(`trying to read config from ./config.yml: %w`, err)
		}
	}

	viper.SetOptions(viper.KeyDelimiter("|"))

	var cfg T
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf(`unmarshal config: %w`, err)
	}

	return &cfg, nil
}

func MustParseFile[T any](filepath *string) *T {
	cfg, err := ParseFile[T](filepath)
	if err != nil {
		panic(fmt.Errorf("parse file: %w", err))
	}
	return cfg
}
