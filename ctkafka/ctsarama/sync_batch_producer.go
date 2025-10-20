package ctsarama

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/ctifaces"
)

var (
	ErrProducerIsClosed = errors.New("producer is closed")
	ErrProducerDisabled = errors.New("producer disabled")
)

type SyncBatchProducer[T any] struct {
	enabled      bool
	addrs        []string
	topic        string
	Config       *sarama.Config
	syncProducer sarama.SyncProducer
	encodeFunc   EncodeFunc[T]
	logger       ctifaces.Logger

	running chan struct{}
	closed  chan struct{}

	sendChan chan []*sarama.ProducerMessage
	errChan  chan error
}

type EncodeFunc[T any] func(obj T) (msg Message, err error)

type BatchProducerParameters[T any] struct {
	Config       ctifaces.KafkaConfigurator
	ClientName   string
	ProducerName string
	EncodeFunc   EncodeFunc[T]
}

func NewSyncBatchProducer[T any](params BatchProducerParameters[T],
	opts ...CfgOpt,
) (*SyncBatchProducer[T], error) {
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
	defaultProducerCfgApply()(sConfig)

	prodCfg, ok := params.Config.GetKafkaProducers()[params.ProducerName]
	if !ok {
		return nil, errors.New("no config for producer named: " + params.ProducerName)
	}

	logger := params.Config.GetLogger()
	if logger == nil {
		logger = loggerStub{}
	}

	for _, opt := range opts {
		opt(sConfig)
	}

	return &SyncBatchProducer[T]{
		enabled:    prodCfg.ProduerEnabled(),
		topic:      prodCfg.GetTopic(),
		encodeFunc: params.EncodeFunc,
		addrs:      clientCfg.GetBrokerAddresses(),
		Config:     sConfig,
		logger:     logger,

		running: make(chan struct{}),
		closed:  make(chan struct{}),

		sendChan: make(chan []*sarama.ProducerMessage),
		errChan:  make(chan error),
	}, nil
}

func defaultProducerCfgApply() CfgOpt {
	return func(c *sarama.Config) {
		c.Producer.Idempotent = true
		c.Producer.Return.Errors = true
		c.Producer.Return.Successes = true
		c.Producer.RequiredAcks = sarama.WaitForAll
		c.Producer.Transaction.Retry.Backoff = 10
		c.Producer.Retry.Max = 10
		c.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}
}

func (p *SyncBatchProducer[T]) Run(ctx context.Context) error {
	if !p.enabled {
		p.logger.Info("producer_disabled",
			"topic", p.topic,
		)

		return nil
	}
	var err error
	p.syncProducer, err = sarama.NewSyncProducer(p.addrs, p.Config)
	if err != nil {
		return fmt.Errorf("new sync producer: %w", err)
	}

	close(p.running)

LoopLable:
	for {
		select {
		case <-ctx.Done():
			p.Close()
			close(p.sendChan)
			break LoopLable
		case <-p.closed:
			close(p.sendChan)
			break LoopLable
		case msgs := <-p.sendChan:
			var errs []error
			for _, msg := range msgs {
				if part, offset, err := p.syncProducer.SendMessage(msg); err != nil {
					errs = append(errs, fmt.Errorf("topic: %s,patrition: %d,offset: %d,desc: %w", p.topic, part, offset, err))
				}
			}
			if err := errors.Join(errs...); err != nil {
				p.errChan <- fmt.Errorf("sarama send message:\n%w", err)
			} else {
				p.errChan <- nil
			}
		}
	}

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("context error: %w", err)
	}

	return nil
}

func (p *SyncBatchProducer[T]) Shutdown(_ context.Context) error {
	if err := p.syncProducer.Close(); err != nil {
		return fmt.Errorf("close sync producer: %w", err)
	}
	return nil
}

func (p *SyncBatchProducer[T]) batchEncode(objs []T) ([]*sarama.ProducerMessage, error) {
	batch := make([]*sarama.ProducerMessage, 0, len(objs))
	for _, obj := range objs {
		msg, err := p.encodeFunc(obj)
		if err != nil {
			return nil, err
		}
		msg.topic = p.topic
		batch = append(batch, msg.toSarama())
	}
	return batch, nil
}

func (p *SyncBatchProducer[T]) SendBatch(objs []T) error {
	if !p.enabled {
		return ErrProducerDisabled
	}

	batch, err := p.batchEncode(objs)
	if err != nil {
		return fmt.Errorf("batch encode: %w", err)
	}

	<-p.running
	select {
	case <-p.closed:
		return ErrProducerIsClosed
	default:
		p.sendChan <- batch
		if err := <-p.errChan; err != nil {
			return fmt.Errorf("send chan: %w", err)
		}
	}

	return nil
}

func (p *SyncBatchProducer[T]) Close() error {
	select {
	case <-p.closed:
		return ErrProducerIsClosed
	default:
		close(p.closed)
	}
	return nil
}

func (p *SyncBatchProducer[T]) GetModuleNamePrefix() string {
	return "kafka_batch_producer"
}

func (p *SyncBatchProducer[T]) PreidentifyModuleGroup() string {
	return "background"
}
