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

const fetchBytes int32 = 64 << 10 // 64 * 1024

type BatchConsumer[T any] struct {
	handler       batchHandler[T]
	addresses     []string
	groupID       string
	consumerGroup sarama.ConsumerGroup
	Config        *sarama.Config
	enabled       bool
}

type batchHandler[T any] struct {
	logger           ctifaces.Logger
	topic            string
	batchInterval    time.Duration
	batchSize        int
	decodeFunc       DecodeFunc[T]
	decodeErrHandle  DecodeErrHandle
	processFunc      ProcessFunc[T]
	processErrHandle ProcessErrHandle
	dropOffsets      bool
}

func (handler batchHandler[T]) Setup(cgs sarama.ConsumerGroupSession) error {
	handler.logger.Debug("starting new session",
		"topic", handler.topic,
		"generation_id", cgs.GenerationID(),
		"member_id", cgs.MemberID())

	if handler.dropOffsets {
		partitions, ok := cgs.Claims()[handler.topic]
		if !ok {
			return errors.New("")
		}
		for _, p := range partitions {
			cgs.ResetOffset(handler.topic, p, 0, "")
		}

		handler.logger.Info("offsets dropped",
			"topic", handler.topic)
	}

	return nil
}

func (_ batchHandler[T]) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler batchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	t := time.NewTicker(handler.batchInterval)
	defer t.Stop()
	rawBatch := make([]*sarama.ConsumerMessage, 0, handler.batchSize)
	var msg *sarama.ConsumerMessage
	var ok bool
	for {
		select {
		case msg, ok = <-claim.Messages():
			if !ok {
				handler.logger.Info("claim message channel closed",
					"topic", claim.Topic(),
					"ctx_error", session.Context().Err(),
				)

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
			err := session.Context().Err()
			if err != nil {
				return fmt.Errorf("session context: %w", err)
			}
			return nil
		}
	}
}

func (handler *batchHandler[T]) flushBatch(ctx context.Context, rawBatch []*sarama.ConsumerMessage) error {
	batch := make([]T, 0, len(rawBatch))
	for _, msg := range rawBatch {
		decoded, err := handler.decodeFunc(fromSarama(msg))
		if err != nil {
			handledErr := handler.decodeErrHandle(msg, err)
			if handledErr != nil {
				return handledErr
			}
		}
		batch = append(batch, decoded)
	}

	if err := handler.processFunc(ctx, batch); err != nil {
		handledErr := handler.processErrHandle(rawBatch, err)
		if handledErr != nil {
			return handledErr
		}
	}
	return nil
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
		initialOffset = nil
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
			batchInterval:    consCfg.GetBatchInterval(),
			batchSize:        consCfg.GetBatchSize(),
			topic:            consCfg.GetTopic(),
			decodeFunc:       params.DecodeFunc,
			decodeErrHandle:  params.DecodeErrHandle,
			processFunc:      params.ProcessFunc,
			processErrHandle: params.ProcessErrHandle,
			logger:           logger,
			dropOffsets:      dropOffsets,
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
		cons.handler.logger.Info("consumer_disabled",
			"topic", cons.handler.topic,
		)

		return nil
	}

	cons.setDefaultErrHanldeFunc()

	// TODO: inspect
	go func() {
		for err := range cons.consumerGroup.Errors() {
			cons.handler.logger.Error("consumer_group",
				"broker_addrs", cons.addresses,
				"group_id", cons.groupID,
				"topic", cons.handler.topic,
				"error", err,
			)
		}
	}()

	err = cons.consumerGroup.Consume(ctx, []string{cons.handler.topic}, cons.handler)
	if err != nil {
		return fmt.Errorf("consumer group consume: %w", err)
	}
	return nil
}

func (cons *BatchConsumer[T]) setDefaultErrHanldeFunc() {
	if cons.handler.decodeErrHandle == nil {
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
	if cons.handler.processErrHandle == nil {
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

func (cons *BatchConsumer[T]) Shutdown(_ context.Context) error {
	if cons.consumerGroup != nil {
		err := cons.consumerGroup.Close()
		if err != nil {
			return fmt.Errorf("consumer group close: %w", err)
		}
	}
	return nil
}

func (cons *BatchConsumer[T]) GetModuleNamePrefix() string {
	return "kafka_batch_consumer"
}
