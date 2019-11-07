package gogo_kafka

type Config struct {
	BackOffTime     int
	MaximumRetry    int
	KafkaVersion    string
	GroupName       string
	BrokerEndpoints []string
	Debug           bool
}

func NewConfig(brokerEndpoints []string, groupName string) *Config {
	return &Config{
		KafkaVersion:    "2.1.1",
		BackOffTime:     3,
		MaximumRetry:    3,
		GroupName:       groupName,
		BrokerEndpoints: brokerEndpoints,
	}
}

func (c *Config) EnableDebug() *Config {
	c.Debug = true

	return c
}

func (c *Config) SetBackOffTime(backOffTime int) {
	c.BackOffTime = backOffTime
}

func (c *Config) SetMaximumRetry(maximumRetry int) {
	c.MaximumRetry = maximumRetry
}
