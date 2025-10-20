package ctsarama

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/surkovvs/ct/ctifaces"
)

type batchHandler[T any] struct {
	fp            flushParameters[T]
	logger        ctifaces.Logger
	topic         string
	flushInterval time.Duration
	dropOffsets   bool
}

func (handler batchHandler[T]) Setup(cgs sarama.ConsumerGroupSession) error {
	handler.logger.Debug("starting new session",
		"topic", handler.topic,
		"generation_id", cgs.GenerationID(),
		"member_id", cgs.MemberID())

	if handler.dropOffsets {
		partitions, ok := cgs.Claims()[handler.topic]
		if !ok {
			return errors.New("topic not found in session claims")
		}
		for _, part := range partitions {
			cgs.ResetOffset(handler.topic, part, 0, "")
		}

		handler.logger.Warn("offsets dropped",
			"topic", handler.topic)
	}

	return nil
}

func (_ batchHandler[T]) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Individualy running by partition.
//
//nolint:gocognit // STUB
func (handler batchHandler[T]) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	bf, err := handler.reportInitialOffsets(session, claim)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	} else if err != nil {
		return nil
	}

	t := time.NewTicker(handler.flushInterval)
	defer t.Stop()
	var msg *sarama.ConsumerMessage
	var ok, flushed bool
	fp := handler.fp
	for {
		select {
		case msg, ok = <-claim.Messages():
			if !ok {
				handler.logger.Info("claim message channel closed",
					"topic", claim.Topic(),
					"ctx_error", session.Context().Err(),
				)

				err := session.Context().Err()
				if err != nil && !errors.Is(err, context.Canceled) {
					return fmt.Errorf("session context: %w", err)
				}
				return nil
			}
			flushed, bf, err = bf.append(session.Context(), msg, fp)
			if err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("batch flushing: %w", err)
			} else if flushed {
				t.Reset(handler.flushInterval)
				session.MarkMessage(msg, "")
				session.Commit()
			}

		case <-t.C:
			bf, err = bf.flush(session.Context(), fp)
			if err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("batch flushing: %w", err)
			}

			t.Reset(handler.flushInterval)
			session.MarkMessage(msg, "")
			session.Commit()

		case <-session.Context().Done():
			err := session.Context().Err()
			if err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("session context: %w", err)
			}
			return nil
		}
	}
}

func (handler batchHandler[T]) reportInitialOffsets(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) (saramaBatch[T], error) {
	bf := make(saramaBatch[T], 0, handler.fp.size)
	select {
	case msg, ok := <-claim.Messages():
		if !ok {
			handler.logger.Info("claim message channel closed",
				"topic", claim.Topic(),
				"ctx_error", session.Context().Err(),
			)

			err := session.Context().Err()
			if err != nil && !errors.Is(err, context.Canceled) {
				return nil, fmt.Errorf("session context: %w", err)
			}
			return nil, context.Canceled
		}

		handler.logger.Debug("initial_offset",
			"topic", handler.topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)
		var flushed bool
		var err error
		flushed, bf, err = bf.append(session.Context(), msg, handler.fp)
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("batch flushing: %w", err)
		}
		if flushed {
			session.MarkMessage(msg, "")
			session.Commit()
		}
	case <-session.Context().Done():
		err := session.Context().Err()
		if err != nil && !errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("session context: %w", err)
		}
		return nil, context.Canceled
	}
	return bf, nil
}
