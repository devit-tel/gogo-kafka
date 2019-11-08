package gogo_kafka

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/devit-tel/gogo-kafka/mocks"
	"github.com/stretchr/testify/suite"
)

type ConsumerTestSuite struct {
	suite.Suite

	consumer             *consumerSarama
	consumerGroupClaim   *mocks.ConsumerGroupClaim
	consumerGroupSession *mocks.ConsumerGroupSession
	retryManager         *mocks.RetryProcess
}

func (suite *ConsumerTestSuite) SetupTest() {
	suite.consumerGroupClaim = &mocks.ConsumerGroupClaim{}
	suite.consumerGroupSession = &mocks.ConsumerGroupSession{}
	suite.retryManager = &mocks.RetryProcess{}

	suite.consumer = &consumerSarama{
		retryManager: suite.retryManager,
		ready:        make(chan bool),
		handlers:     map[string]WorkerHandler{},
	}
}

func TestConsumerSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}

func makeTestChannelConsumer(messages []*sarama.ConsumerMessage) func() <-chan *sarama.ConsumerMessage {
	return func() <-chan *sarama.ConsumerMessage {
		out := make(chan *sarama.ConsumerMessage)
		go func() {
			for _, n := range messages {
				out <- n
			}
			close(out)
		}()
		return out
	}
}

func (suite *ConsumerTestSuite) TestConsumerSarama_Success() {
	var result string
	testFunc := func(key string, data []byte) error {
		result = string(data)
		return nil
	}

	// Map topic to handler
	suite.consumer.handlers["testTopic"] = testFunc

	// Make message data for test
	messages := []*sarama.ConsumerMessage{
		{Key: []byte("key_1"), Value: []byte("data_1"), Topic: "testTopic"},
		{Key: []byte("key_2"), Value: []byte("data_2"), Topic: "testTopic"},
		{Key: []byte("key_3"), Value: []byte("data_3"), Topic: "testTopic"},
	}

	suite.consumerGroupClaim.On("Messages").Once().Return(makeTestChannelConsumer(messages))
	suite.retryManager.On("IsMaximumRetry", "key_1").Once().Return(false)
	suite.retryManager.On("IsMaximumRetry", "key_2").Once().Return(false)
	suite.retryManager.On("IsMaximumRetry", "key_3").Once().Return(false)
	suite.consumerGroupSession.On("MarkMessage", messages[0], "").Once()
	suite.consumerGroupSession.On("MarkMessage", messages[1], "").Once()
	suite.consumerGroupSession.On("MarkMessage", messages[2], "").Once()

	err := suite.consumer.ConsumeClaim(suite.consumerGroupSession, suite.consumerGroupClaim)
	suite.NoError(err)
	suite.Equal("data_3", result)

	suite.retryManager.AssertExpectations(suite.T())
	suite.consumerGroupClaim.AssertExpectations(suite.T())
	suite.consumerGroupSession.AssertExpectations(suite.T())
}

func (suite *ConsumerTestSuite) TestConsumerSarama_CustomRecoveryFunc() {
	var resultData, resultTopic, resultKey, resultPanic string
	testFunc := func(key string, data []byte) error {
		if string(data) == "data_2" {
			panic("data_from_panic_handler")
		}
		return nil
	}

	panicFunc := func(err interface{}, topic, key string, data []byte) error {
		resultData = string(data)
		resultTopic = topic
		resultKey = key
		resultPanic = fmt.Sprintf("%v", err)
		return nil
	}

	// Map topic to handler
	suite.consumer.handlers["testTopic"] = testFunc
	suite.consumer.panicHandler = panicFunc

	// Make message data for test
	messages := []*sarama.ConsumerMessage{
		{Key: []byte("key_1"), Value: []byte("data_1"), Topic: "testTopic"},
		{Key: []byte("key_2"), Value: []byte("data_2"), Topic: "testTopic"},
		{Key: []byte("key_3"), Value: []byte("data_3"), Topic: "testTopic"},
	}

	suite.consumerGroupClaim.On("Messages").Once().Return(makeTestChannelConsumer(messages))

	// Lap: 1
	suite.retryManager.On("IsMaximumRetry", "key_1").Once().Return(false)
	suite.consumerGroupSession.On("MarkMessage", messages[0], "").Once()

	// Lap: 2
	suite.retryManager.On("IsMaximumRetry", "key_2").Once().Return(false)
	suite.retryManager.On("AddRetryCount", "key_2").Once().Return(3)
	suite.retryManager.On("IsMaximumRetry", "key_2").Once().Return(true)

	err := suite.consumer.ConsumeClaim(suite.consumerGroupSession, suite.consumerGroupClaim)
	suite.NoError(err)
	suite.Equal("data_2", resultData)
	suite.Equal("testTopic", resultTopic)
	suite.Equal("key_2", resultKey)
	suite.Equal("data_from_panic_handler", resultPanic)

	suite.retryManager.AssertExpectations(suite.T())
	suite.consumerGroupClaim.AssertExpectations(suite.T())
	suite.consumerGroupSession.AssertExpectations(suite.T())
}

func (suite *ConsumerTestSuite) TestConsumerSarama_ProcessError() {
	var retryCounter int
	testFunc := func(key string, data []byte) error {
		if string(data) == "data_1" && retryCounter == 0 {
			retryCounter += 1
			return errors.New("test_error")
		}

		return nil
	}

	// Map topic to handler
	suite.consumer.handlers["testTopic"] = testFunc

	// Make message data for test
	messages := []*sarama.ConsumerMessage{
		{Key: []byte("key_1"), Value: []byte("data_1"), Topic: "testTopic"},
		{Key: []byte("key_2"), Value: []byte("data_2"), Topic: "testTopic"},
		{Key: []byte("key_3"), Value: []byte("data_3"), Topic: "testTopic"},
	}

	suite.consumerGroupClaim.On("Messages").Once().Return(makeTestChannelConsumer(messages))
	suite.retryManager.On("IsMaximumRetry", "key_1").Once().Return(false)
	suite.retryManager.On("AddRetryCount", "key_1").Once().Return(1)
	suite.retryManager.On("IsMaximumRetry", "key_1").Once().Return(false)
	suite.retryManager.On("DelayProcessFollowBackOffTime", "key_1").Once()

	err := suite.consumer.ConsumeClaim(suite.consumerGroupSession, suite.consumerGroupClaim)
	suite.Equal(errors.New("test_error"), err)
	suite.retryManager.AssertExpectations(suite.T())
	suite.consumerGroupClaim.AssertExpectations(suite.T())
	suite.consumerGroupSession.AssertExpectations(suite.T())
}

func (suite *ConsumerTestSuite) TestConsumerSarama_MaximumRetryAndSkip() {
	var retryCounter int
	testFunc := func(key string, data []byte) error {
		if string(data) == "data_1" && retryCounter == 0 {
			retryCounter += 1
			return errors.New("test_error")
		}

		return nil
	}

	// Map topic to handler
	suite.consumer.handlers["testTopic"] = testFunc

	// Make message data for test
	messages := []*sarama.ConsumerMessage{
		{Key: []byte("key_1"), Value: []byte("data_1"), Topic: "testTopic"},
		{Key: []byte("key_2"), Value: []byte("data_2"), Topic: "testTopic"},
		{Key: []byte("key_3"), Value: []byte("data_3"), Topic: "testTopic"},
	}

	suite.consumerGroupClaim.On("Messages").Once().Return(makeTestChannelConsumer(messages))
	suite.retryManager.On("IsMaximumRetry", "key_1").Once().Return(true)
	suite.retryManager.On("IsMaximumRetry", "key_2").Once().Return(false)
	suite.retryManager.On("IsMaximumRetry", "key_3").Once().Return(false)

	suite.consumerGroupSession.On("MarkMessage", messages[0], "").Once()
	suite.consumerGroupSession.On("MarkMessage", messages[1], "").Once()
	suite.consumerGroupSession.On("MarkMessage", messages[2], "").Once()

	err := suite.consumer.ConsumeClaim(suite.consumerGroupSession, suite.consumerGroupClaim)
	suite.NoError(err)
	suite.retryManager.AssertExpectations(suite.T())
	suite.consumerGroupClaim.AssertExpectations(suite.T())
	suite.consumerGroupSession.AssertExpectations(suite.T())
}
