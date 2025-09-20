package ctkafka

import (
	"strings"
	"time"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.KafkaConsumerConfigurator = (*ConsumerConfig)(nil)

type ConsumerConfig struct {
	Enabled           bool          `validate:"required"`
	GroupID           string        `validate:"required"`
	Topic             string        `validate:"required"`
	BatchSize         int           `validate:"min=1"`
	BatchInterval     time.Duration `validate:"min=1"`
	InitialOffset     any           // from sarama -2: OffsetOldest -1: OffsetNewest
	MaxProcessingTime *time.Duration
}

func (c ConsumerConfig) ConsumerEnabled() bool {
	return c.Enabled
}

func (c ConsumerConfig) GetGroupID() string {
	return c.GroupID
}

func (c ConsumerConfig) GetTopic() string {
	return c.Topic
}

func (c ConsumerConfig) GetBatchSize() int {
	return c.BatchSize
}

func (c ConsumerConfig) GetBatchInterval() time.Duration {
	return c.BatchInterval
}

func (c ConsumerConfig) GetInitialOffset() *int64 {
	switch offset := c.InitialOffset.(type) {
	case int:
		r := int64(offset)
		return &r
	case string:
		switch {
		case strings.EqualFold("oldest", offset):
			r := int64(-2)
			return &r
		case strings.EqualFold("newest", offset):
			r := int64(-1)
			return &r
		}
	}
	return nil
}

func (c ConsumerConfig) GetMaxProcessingTime() *time.Duration {
	return c.MaxProcessingTime
}
