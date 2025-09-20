package ctsarama

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/ctifaces"
	"github.com/surkovvs/ct/internal/tools"
)

var fetchBytes int32 = 64 << 10 // 64 * 1024

func DecodeStub(_, v []byte) ([]byte, error) {
	return v, nil
}

type BatchConsumer[T any] struct {
	Config        *sarama.Config
	enabled       bool
	addresses     []string
	groupID       string
	topic         string
	consumerGroup sarama.ConsumerGroup

	handler batchHandler[T]
}

func (handler batchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (handler batchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (handler batchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	t := time.NewTicker(handler.batchInterval)
	rawBatch := make([]*sarama.ConsumerMessage, 0, handler.batchSize)
	var msg *sarama.ConsumerMessage
	var ok bool
	for {
		select {
		case msg, ok = <-claim.Messages():
			if !ok {
				if handler.logger != nil {
					handler.logger.Info("claim message channel closed",
						"topic", claim.Topic(),
						"ctx_error", session.Context().Err(),
					)
				}
				return nil
			}
			rawBatch = append(rawBatch, msg)
			if len(rawBatch) == handler.batchSize {
				if err := handler.flushBatch(session.Context(), rawBatch); err != nil {
					return err
				}

				t.Reset(handler.batchInterval)
				rawBatch = rawBatch[:0]
				session.MarkMessage(msg, "")
				session.Commit()
			}

		case <-t.C:
			if len(rawBatch) == 0 {
				continue
			}

			if err := handler.flushBatch(session.Context(), rawBatch); err != nil {
				return err
			}

			t.Reset(handler.batchInterval)
			rawBatch = rawBatch[:0]
			session.MarkMessage(msg, "")
			session.Commit()

		case <-session.Context().Done():
			return session.Context().Err()
		}
	}
}

func (handler *batchHandler[T]) flushBatch(ctx context.Context, rawBatch []*sarama.ConsumerMessage) error {
	batch := make([]T, 0, len(rawBatch))
	for _, msg := range rawBatch {
		decoded, err := handler.decodeFunc(msg.Key, msg.Value)
		if err != nil {
			if err := handler.decodeErrHandle(msg, err); err != nil {
				return err
			}
		}
		batch = append(batch, decoded)
	}

	if err := handler.processFunc(ctx, batch); err != nil {
		if err := handler.processErrHandle(rawBatch, err); err != nil {
			return err
		}
	}
	return nil
}

type batchHandler[T any] struct {
	batchInterval    time.Duration
	batchSize        int
	decodeFunc       DecodeFunc[T]
	decodeErrHandle  DecodeErrHandle
	processFunc      ProcessFunc[T]
	processErrHandle ProcessErrHandle
	logger           ctifaces.Logger
}

type (
	DecodeFunc[T any]  func(k, v []byte) (T, error)
	DecodeErrHandle    func(msg *sarama.ConsumerMessage, err error) error
	ProcessFunc[T any] func(ctx context.Context, batch []T) error
	ProcessErrHandle   func(msgs []*sarama.ConsumerMessage, err error) error
)

type BatchConsumerParameters[T any] struct {
	Config           ctifaces.KafkaConfigurator
	ClientName       string
	ConsumerName     string
	DecodeFunc       DecodeFunc[T]
	DecodeErrHandle  DecodeErrHandle
	ProcessFunc      ProcessFunc[T]
	ProcessErrHandle ProcessErrHandle
}

func NewBatchConsumer[T any](params BatchConsumerParameters[T], opts ...cfgOpt) (*BatchConsumer[T], error) {
	if params.Config.LogKafkaEvents() {
		InitSaramaLogger(params.Config.GetEventsLogTitle(), params.Config.GetLogger())
	}

	clientCfg, ok := params.Config.GetKafkaCLients()[params.ClientName]
	if !ok {
		return nil, errors.New("no config for client named: " + params.ClientName)
	}
	sConfig, err := newClientConfigDefault(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("new config: %w", err)
	}
	sConfig.Consumer.Return.Errors = true
	sConfig.Consumer.Fetch.Min = 1
	sConfig.Consumer.Fetch.Max = fetchBytes
	sConfig.Consumer.Fetch.Default = fetchBytes
	sConfig.Consumer.Offsets.AutoCommit.Enable = false
	sConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRange(),
	}

	consCfg, ok := params.Config.GetKafkaConsumers()[params.ConsumerName]
	if !ok {
		return nil, errors.New("no config for consumer named: " + params.ClientName)
	}
	tools.SetIfNotNil(&sConfig.Consumer.MaxProcessingTime, consCfg.GetMaxProcessingTime())
	tools.SetIfNotNil(&sConfig.Consumer.Offsets.Initial, consCfg.GetInitialOffset())

	for _, opt := range opts {
		opt(sConfig)
	}

	return &BatchConsumer[T]{
		Config:    sConfig,
		enabled:   consCfg.ConsumerEnabled(),
		addresses: clientCfg.GetBrokerAddresses(),
		groupID:   consCfg.GetGroupID(),
		topic:     consCfg.GetTopic(),
		handler: batchHandler[T]{
			batchInterval:    consCfg.GetBatchInterval(),
			batchSize:        consCfg.GetBatchSize(),
			decodeFunc:       params.DecodeFunc,
			decodeErrHandle:  params.DecodeErrHandle,
			processFunc:      params.ProcessFunc,
			processErrHandle: params.ProcessErrHandle,
			logger:           params.Config.GetLogger(),
		},
	}, nil
}

func (cons *BatchConsumer[T]) Run(ctx context.Context) error {
	var err error
	cons.consumerGroup, err = sarama.NewConsumerGroup(cons.addresses, cons.groupID, cons.Config)
	if err != nil {
		return fmt.Errorf("new consumer group: %w", err)
	}
	if !cons.enabled {
		if cons.handler.logger != nil {
			cons.handler.logger.Info("consumer_disabled",
				"topic", cons.topic,
			)
		}
		return nil
	}

	cons.setDefaultErrHanldeFunc()

	// TODO: inspect
	go func() {
		for err := range cons.consumerGroup.Errors() {
			cons.handler.logger.Error("consumer_group",
				"broker_addrs", cons.addresses,
				"group_id", cons.groupID,
				"topic", cons.topic,
				"error", err,
			)
		}
	}()

	return cons.consumerGroup.Consume(ctx, []string{cons.topic}, cons.handler)
}

func (cons *BatchConsumer[T]) setDefaultErrHanldeFunc() {
	if cons.handler.decodeErrHandle == nil && cons.handler.logger != nil {
		cons.handler.decodeErrHandle = func(msg *sarama.ConsumerMessage, err error) error {
			cons.handler.logger.Error("decode_message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"error", err,
			)
			return err
		}
	}
	if cons.handler.processErrHandle == nil && cons.handler.logger != nil {
		cons.handler.processErrHandle = func(msgs []*sarama.ConsumerMessage, err error) error {
			cons.handler.logger.Error("process_message",
				"topic", msgs[0].Topic,
				"partition", msgs[0].Partition,
				"offset_from", msgs[0].Offset,
				"offset_to", msgs[len(msgs)-1].Offset,
				"error", err,
			)

			return err
		}
	}
}

func (cons *BatchConsumer[T]) Shutdown(ctx context.Context) error {
	if cons.consumerGroup != nil {
		return cons.consumerGroup.Close()
	}
	return nil
}

func (p *BatchConsumer[T]) GetModuleNamePrefix() string {
	return "kafka_batch_consumer"
}
