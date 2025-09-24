package ctkafka

import "github.com/surkovvs/ct/ctifaces"

var _ ctifaces.KafkaConfigurator = (*Config)(nil)

type Config struct {
	LogEnable bool
	LogTitle  string
	Logger    ctifaces.Logger

	Clients   map[string]ClientConfig   `mapstructure:"Clients"`
	Producers map[string]ProducerConfig `mapstructure:"Producers"`
	Consumers map[string]ConsumerConfig `mapstructure:"Consumers"`
}

func (cfg Config) WithLogger(log ctifaces.Logger) Config {
	cfg.Logger = log
	return cfg
}

func (cfg Config) GetLogger() ctifaces.Logger {
	return cfg.Logger
}

func (cfg Config) LogKafkaEvents() bool {
	return cfg.LogEnable
}

func (cfg Config) GetEventsLogTitle() string {
	return cfg.LogTitle
}

func (cfg Config) GetKafkaCLients() map[string]ctifaces.KafkaClientConfigurator {
	res := make(map[string]ctifaces.KafkaClientConfigurator, len(cfg.Clients))
	for k, v := range cfg.Clients {
		res[k] = v
	}
	return res
}

func (cfg Config) GetKafkaProducers() map[string]ctifaces.KafkaProducerConfigurator {
	res := make(map[string]ctifaces.KafkaProducerConfigurator, len(cfg.Clients))
	for k, v := range cfg.Producers {
		res[k] = v
	}
	return res
}

func (cfg Config) GetKafkaConsumers() map[string]ctifaces.KafkaConsumerConfigurator {
	res := make(map[string]ctifaces.KafkaConsumerConfigurator, len(cfg.Clients))
	for k, v := range cfg.Consumers {
		res[k] = v
	}
	return res
}
