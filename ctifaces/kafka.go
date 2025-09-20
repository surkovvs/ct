package ctifaces

import "time"

type KafkaClientConfigurator interface {
	GetBrokerAddresses() []string
	GetCLientID() string
	GetChannelBufferSize() *int
	GetCertPath() *string
	AuthNeeded() bool
	GetUserName() string
	GetPassword() string
}

type KafkaProducerConfigurator interface {
	ProduerEnabled() bool
	GetTopic() string
	GetTxnID() *string
}

type KafkaConsumerConfigurator interface {
	ConsumerEnabled() bool
	GetGroupID() string
	GetTopic() string
	GetBatchSize() int
	GetBatchInterval() time.Duration
	GetInitialOffset() *int64
	GetMaxProcessingTime() *time.Duration
}

type KafkaConfigurator interface {
	GetLogger() Logger
	LogKafkaEvents() bool
	GetEventsLogTitle() string
	GetKafkaCLients() map[string]KafkaClientConfigurator
	GetKafkaProducers() map[string]KafkaProducerConfigurator
	GetKafkaConsumers() map[string]KafkaConsumerConfigurator
}
