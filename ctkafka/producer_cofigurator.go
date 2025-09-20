package ctkafka

import "github.com/surkovvs/ct/ctifaces"

var _ ctifaces.KafkaProducerConfigurator = (*ProducerConfig)(nil)

type ProducerConfig struct {
	Enabled bool   `validate:"required"`
	Topic   string `validate:"required"`
	TxnID   *string
}

func (c ProducerConfig) ProduerEnabled() bool {
	return c.Enabled
}

func (c ProducerConfig) GetTopic() string {
	return c.Topic
}

func (c ProducerConfig) GetTxnID() *string {
	return c.TxnID
}
