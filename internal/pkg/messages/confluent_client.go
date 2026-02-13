package messages

import (
	"encoding/json"
	"fmt"

	"github.com/adf-code/beta-book-api/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

type KafkaClient interface {
	Publish(topic string, key string, value interface{}) error
	Close()
}

type confluentClient struct {
	producer *kafka.Producer
	logger   zerolog.Logger
}

func NewConfluentClient(cfg *config.AppConfig, logger zerolog.Logger) KafkaClient {
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
		"client.id":         cfg.KafkaClientID,
		"acks":              "all",
	}

	// Confluent Cloud (SASL_SSL)
	if cfg.KafkaAPIKey != "" && cfg.KafkaAPISecret != "" {
		kafkaConfig.SetKey("security.protocol", cfg.KafkaSecurityProto)
		kafkaConfig.SetKey("sasl.mechanisms", "PLAIN")
		kafkaConfig.SetKey("sasl.username", cfg.KafkaAPIKey)
		kafkaConfig.SetKey("sasl.password", cfg.KafkaAPISecret)
	}

	p, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("❌ Failed to create Kafka producer")
	}

	// Handle delivery reports in background
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logger.Error().Err(ev.TopicPartition.Error).
						Str("topic", *ev.TopicPartition.Topic).
						Msg("❌ Kafka delivery failed")
				} else {
					logger.Debug().
						Str("topic", *ev.TopicPartition.Topic).
						Int32("partition", ev.TopicPartition.Partition).
						Msg("✅ Kafka message delivered")
				}
			}
		}
	}()

	logger.Info().
		Str("brokers", cfg.KafkaBrokers).
		Str("client_id", cfg.KafkaClientID).
		Msg("✅ Kafka producer connected")

	return &confluentClient{
		producer: p,
		logger:   logger,
	}
}

func (c *confluentClient) Publish(topic string, key string, value interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		c.logger.Error().Err(err).Msg("❌ Failed to marshal Kafka message")
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          payload,
	}

	err = c.producer.Produce(msg, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("topic", topic).Msg("❌ Failed to produce Kafka message")
		return fmt.Errorf("failed to produce message: %w", err)
	}

	c.logger.Info().Str("topic", topic).Str("key", key).Msg("📤 Kafka message produced")
	return nil
}

func (c *confluentClient) Close() {
	c.producer.Flush(5000)
	c.producer.Close()
	c.logger.Info().Msg("🔒 Kafka producer closed.")
}
