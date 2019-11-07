package gogo_kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

//go:generate mockery -name=ConsumerGroupSession
type ConsumerGroupSession interface {
	Claims() map[string][]int32
	MemberID() string
	GenerationID() int32
	MarkOffset(topic string, partition int32, offset int64, metadata string)
	ResetOffset(topic string, partition int32, offset int64, metadata string)
	MarkMessage(msg *sarama.ConsumerMessage, metadata string)
	Context() context.Context
}

//go:generate mockery -name=ConsumerGroupClaim
type ConsumerGroupClaim interface {
	Topic() string
	Partition() int32
	InitialOffset() int64
	HighWaterMarkOffset() int64
	Messages() <-chan *sarama.ConsumerMessage
}
