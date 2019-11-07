package retrymanager

import (
	"time"
)

type InmemManager struct {
	data             map[string]int
	retryBackOffTime int
	maximumRetyCount int
}

func NewInmemManager(retryBackOffTime, maximumRetryCount int) *InmemManager {
	return &InmemManager{
		data:             map[string]int{},
		retryBackOffTime: retryBackOffTime,
		maximumRetyCount: maximumRetryCount,
	}
}

func (rm *InmemManager) AddRetryCount(key string) int {
	value, isExist := rm.data[key]
	if !isExist {
		value = 0
	}

	rm.data[key] = value + 1
	return 0
}

func (rm *InmemManager) GetRetryCount(key string) int {
	if value, isExist := rm.data[key]; isExist {
		return value
	}

	return 0
}

func (rm *InmemManager) ClearRetryCount(key string) {
	delete(rm.data, key)
}

func (rm *InmemManager) DelayProcessFollowBackOffTime(key string) {
	retryCount := rm.GetRetryCount(key)
	time.Sleep(time.Duration(rm.retryBackOffTime*retryCount) * time.Second)
}

func (rm *InmemManager) IsMaximumRetry(key string) bool {
	return (rm.maximumRetyCount + 1) == rm.GetRetryCount(key)
}
