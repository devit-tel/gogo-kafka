// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"

import mock "github.com/stretchr/testify/mock"
import sarama "github.com/Shopify/sarama"

// ConsumerGroupSession is an autogenerated mock type for the ConsumerGroupSession type
type ConsumerGroupSession struct {
	mock.Mock
}

// Claims provides a mock function with given fields:
func (_m *ConsumerGroupSession) Claims() map[string][]int32 {
	ret := _m.Called()

	var r0 map[string][]int32
	if rf, ok := ret.Get(0).(func() map[string][]int32); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]int32)
		}
	}

	return r0
}

// Context provides a mock function with given fields:
func (_m *ConsumerGroupSession) Context() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// GenerationID provides a mock function with given fields:
func (_m *ConsumerGroupSession) GenerationID() int32 {
	ret := _m.Called()

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}

// MarkMessage provides a mock function with given fields: msg, metadata
func (_m *ConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	_m.Called(msg, metadata)
}

// MarkOffset provides a mock function with given fields: topic, partition, offset, metadata
func (_m *ConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	_m.Called(topic, partition, offset, metadata)
}

// MemberID provides a mock function with given fields:
func (_m *ConsumerGroupSession) MemberID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ResetOffset provides a mock function with given fields: topic, partition, offset, metadata
func (_m *ConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	_m.Called(topic, partition, offset, metadata)
}
