package gogo_kafka

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
)

type WorkerHandler func(data []byte) error

type Worker struct {
	topicNames []string
	consumer   *consumerSarama
	client     sarama.ConsumerGroup
	ctx        context.Context
	cancel     context.CancelFunc
	config     *Config
}

func New(config *Config, consumerGroup sarama.ConsumerGroup, retryManager RetryProcess) (*Worker, error) {
	consumer := &consumerSarama{
		ready:        make(chan bool),
		retryManager: retryManager,
		handlers:     map[string]WorkerHandler{},
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		topicNames: []string{},
		consumer:   consumer,
		ctx:        ctx,
		client:     consumerGroup,
		cancel:     cancel,
		config:     config,
	}, nil
}

func (worker *Worker) SetLogger(logger *log.Logger) {
	sarama.Logger = logger
}

func (worker *Worker) SetPanicHandler(recover func(err interface{})) {
	worker.consumer.panicHandler = recover
}

func (worker *Worker) RegisterHandler(topicName string, handler WorkerHandler) {
	worker.topicNames = append(worker.topicNames, topicName)
	worker.consumer.handlers[topicName] = handler
}

func (worker *Worker) Start() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := worker.client.Consume(worker.ctx, worker.topicNames, worker.consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}

			if worker.ctx.Err() != nil {
				return
			}
			worker.consumer.ready = make(chan bool)
		}
	}()

	<-worker.consumer.ready // Await till the consumer has been set up
	log.Printf("[Worker: %s]: start subscribe topics %v", worker.config.GroupName, worker.topicNames)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-worker.ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}

	worker.cancel()
	wg.Wait()
	if err := worker.client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Create sarama consumer by custom config
func NewSaramaConsumerWithConfig(config *Config) (sarama.ConsumerGroup, error) {
	if config.Debug {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	saramaVersion, err := sarama.ParseKafkaVersion(config.KafkaVersion)
	if err != nil {
		return nil, err
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = saramaVersion
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	client, err := sarama.NewConsumerGroup(config.BrokerEndpoints, config.GroupName, saramaConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
