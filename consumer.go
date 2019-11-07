package gogo_kafka

import (
	"github.com/Shopify/sarama"
)

type consumerSarama struct {
	ready        chan bool
	handlers     map[string]WorkerHandler
	retryManager RetryProcess
	panicHandler func(err interface{})
}

func (consumer *consumerSarama) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *consumerSarama) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *consumerSarama) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var key string

	// Testing will be not panic. please ignore comment this defer func before testing
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// check is panic handler exist
			if consumer.panicHandler != nil {
				// call panic handler for manage panic
				consumer.panicHandler(panicErr)
			}

			if key != "" {
				consumer.processRetryAndDelay(key)
			}
		}
	}()

	for message := range claim.Messages() {
		key = string(message.Key)

		handler, isExist := consumer.handlers[message.Topic]
		if !isExist {
			session.MarkMessage(message, "")
			continue
		}

		if consumer.retryManager.IsMaximumRetry(key) {
			session.MarkMessage(message, "")
			continue
		}

		if err := handler(message.Value); err != nil {
			consumer.processRetryAndDelay(key)
			return err
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func (consumer *consumerSarama) processRetryAndDelay(key string) {
	consumer.retryManager.AddRetryCount(key)
	if !consumer.retryManager.IsMaximumRetry(key) {
		consumer.retryManager.DelayProcessFollowBackOffTime(key)
	}
}
