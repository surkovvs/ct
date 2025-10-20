package ctsarama

import (
	"context"

	"github.com/IBM/sarama"
)

type saramaBatch[T any] []*sarama.ConsumerMessage

type flushParameters[T any] struct {
	decode           DecodeFunc[T]
	decodeErrHandle  DecodeErrHandle
	process          ProcessFunc[T]
	processErrHandle ProcessErrHandle
	size             int
}

// Auto flush on reaching the limit.
func (b saramaBatch[T]) append(
	ctx context.Context,
	msg *sarama.ConsumerMessage,
	fp flushParameters[T],
) (flushed bool, batch saramaBatch[T], err error) {
	b = append(b, msg)
	if len(b) == fp.size {
		b, err = b.flush(ctx, fp)
		return true, b, err
	}
	return false, b, nil
}

func (b saramaBatch[T]) flush(
	ctx context.Context,
	fp flushParameters[T],
) (saramaBatch[T], error) {
	if len(b) == 0 {
		return b, nil
	}

	decodedBatch := make([]T, len(b))
	for i, msg := range b {
		decoded, err := fp.decode(fromSarama(msg))
		if err != nil {
			handledErr := fp.decodeErrHandle(msg, err)
			if handledErr != nil {
				return b[:0], handledErr
			}
		}
		decodedBatch[i] = decoded
	}

	if err := fp.process(ctx, decodedBatch); err != nil {
		handledErr := fp.processErrHandle(b, err)
		if handledErr != nil {
			return b[:0], handledErr
		}
	}

	return b[:0], nil
}
