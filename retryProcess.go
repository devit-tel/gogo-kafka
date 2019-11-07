package gogo_kafka

//go:generate mockery -name=RetryProcess
type RetryProcess interface {
	AddRetryCount(key string) int
	GetRetryCount(key string) int
	DelayProcessFollowBackOffTime(string string)
	ClearRetryCount(string string)
	IsMaximumRetry(key string) bool
}
