package ctsarama

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/ctifaces"
	"github.com/surkovvs/ct/internal/tools"
)

const fetchBytes int32 = 64 << 10 // 64 * 1024

type BatchConsumer[T any] struct {
	handler       batchHandler[T]
	addresses     []string
	groupID       string
	consumerGroup sarama.ConsumerGroup
	Config        *sarama.Config
	enabled       bool
}

type (
	DecodeFunc[T any]  func(msg Message) (T, error)
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

func NewBatchConsumer[T any](params BatchConsumerParameters[T], opts ...CfgOpt) (*BatchConsumer[T], error) {
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

	dropOffsets := false
	initialOffset := consCfg.GetInitialOffset()
	if initialOffset != nil && *initialOffset == -3 {
		dropOffsets = true
		*initialOffset = sarama.OffsetOldest
	}
	tools.SetIfNotNil(&sConfig.Consumer.Offsets.Initial, initialOffset)

	tools.SetIfNotNil(&sConfig.Consumer.MaxProcessingTime, consCfg.GetMaxProcessingTime())

	logger := params.Config.GetLogger()
	if logger == nil {
		logger = loggerStub{}
	}

	for _, opt := range opts {
		opt(sConfig)
	}

	return &BatchConsumer[T]{
		Config:    sConfig,
		enabled:   consCfg.ConsumerEnabled(),
		addresses: clientCfg.GetBrokerAddresses(),
		groupID:   consCfg.GetGroupID(),
		handler: batchHandler[T]{
			fp: flushParameters[T]{
				decode:           params.DecodeFunc,
				decodeErrHandle:  params.DecodeErrHandle,
				process:          params.ProcessFunc,
				processErrHandle: params.ProcessErrHandle,
				size:             consCfg.GetBatchSize(),
			},
			flushInterval: consCfg.GetBatchInterval(),
			topic:         consCfg.GetTopic(),
			logger:        logger,
			dropOffsets:   dropOffsets,
		},
	}, nil
}

func (cons *BatchConsumer[T]) Init(ctx context.Context) error {
	if !cons.enabled {
		cons.handler.logger.Info("consumer_disabled",
			"topic", cons.handler.topic,
		)
		return nil
	}

	client, err := sarama.NewClient(cons.addresses, cons.Config)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	reportOffsets(client, cons.handler.topic, cons.handler.logger)

	cons.consumerGroup, err = sarama.NewConsumerGroupFromClient(cons.groupID, client)
	if err != nil {
		return fmt.Errorf("new consumer group: %w", err)
	}

	cons.setDefaultErrHanldeFunc()

	return nil
}

func reportOffsets(client sarama.Client, topic string, logger ctifaces.Logger) error {
	parts, err := client.Partitions(topic)
	if err != nil {
		return fmt.Errorf("partitions: %w", err)
	}
	byPartition := make(map[int32]int64, len(parts))
	for _, p := range parts {
		offset, err := client.GetOffset(topic, p, sarama.OffsetOldest)
		if err != nil {
			return fmt.Errorf("get offset: %w", err)
		}
		byPartition[p] = offset
	}
	logger.Debug("earliest_available_offsets",
		"topic", topic,
		"by_partition", byPartition)

	for _, p := range parts {
		offset, err := client.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("get offset: %w", err)
		}
		byPartition[p] = offset
	}
	logger.Debug("latest_offsets",
		"topic", topic,
		"by_partition", byPartition)
	return nil
}

func (cons *BatchConsumer[T]) setDefaultErrHanldeFunc() {
	if cons.handler.fp.decodeErrHandle == nil {
		cons.handler.fp.decodeErrHandle = func(msg *sarama.ConsumerMessage, err error) error {
			cons.handler.logger.Error("decode_message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"error", err,
			)
			return err
		}
	}
	if cons.handler.fp.processErrHandle == nil {
		cons.handler.fp.processErrHandle = func(msgs []*sarama.ConsumerMessage, err error) error {
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

func (cons *BatchConsumer[T]) Run(ctx context.Context) error {
	err := cons.consumerGroup.Consume(ctx, []string{cons.handler.topic}, cons.handler)
	if err != nil {
		return fmt.Errorf("consumer group consume: %w", err)
	}
	return nil
}

func (cons *BatchConsumer[T]) Shutdown(_ context.Context) error {
	if cons.consumerGroup != nil {
		err := cons.consumerGroup.Close()
		if err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("consumer group close: %w", err)
		}
	}
	return nil
}

func (cons *BatchConsumer[T]) GetModuleNamePrefix() string {
	return "kafka_batch_consumer"
}
