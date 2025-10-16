package ctsarama

import (
	"github.com/IBM/sarama"
)

type (
	Header struct {
		Key   []byte
		Value []byte
	}
	Message struct {
		Key     []byte
		Value   []byte
		Headers []Header

		topic     string
		offset    int64
		partition int32
	}
)

func (m Message) WithPartition(partition int32) Message {
	m.partition = partition
	return m
}

func (m Message) GetPartition() int32 {
	return m.partition
}

func (m Message) WithOffset(offset int64) Message {
	m.offset = offset
	return m
}

func (m Message) GetOffset() int64 {
	return m.offset
}

func (m Message) toSarama() *sarama.ProducerMessage {
	headers := make([]sarama.RecordHeader, 0, len(m.Headers))
	for _, h := range m.Headers {
		headers = append(headers, sarama.RecordHeader{
			Key:   h.Key,
			Value: h.Value,
		})
	}
	return &sarama.ProducerMessage{
		Topic:     m.topic,
		Key:       sarama.ByteEncoder(m.Key),
		Value:     sarama.ByteEncoder(m.Value),
		Headers:   headers,
		Offset:    m.offset,
		Partition: m.partition,
	}
}

func fromSarama(msg *sarama.ConsumerMessage) Message {
	headers := make([]Header, 0, len(msg.Headers))
	for _, h := range msg.Headers {
		headers = append(headers, Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}
	return Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   headers,
		offset:    msg.Offset,
		partition: msg.Partition,
	}
}
