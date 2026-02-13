package mocks

import (
	"github.com/stretchr/testify/mock"
)

type KafkaClient struct {
	mock.Mock
}

func (m *KafkaClient) Publish(topic string, key string, value interface{}) error {
	args := m.Called(topic, key, value)
	return args.Error(0)
}

func (m *KafkaClient) Close() {
	m.Called()
}
