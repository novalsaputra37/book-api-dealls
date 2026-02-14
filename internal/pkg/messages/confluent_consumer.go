package messages

import (
	"github.com/adf-code/beta-book-api/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

// KafkaConsumer defines the interface for consuming Kafka messages.
type KafkaConsumer interface {
	Subscribe(topic string) error
	Poll(timeoutMs int) (*kafka.Message, error)
	CommitMessage(msg *kafka.Message) error
	Close()
}

type confluentConsumer struct {
	consumer *kafka.Consumer
	logger   zerolog.Logger
}

// NewConfluentConsumer creates a new Kafka consumer using confluent-kafka-go.
func NewConfluentConsumer(cfg *config.AppConfig, logger zerolog.Logger) KafkaConsumer {
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":  cfg.KafkaBrokers,
		"group.id":           cfg.KafkaConsumerGroupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	// Confluent Cloud (SASL_SSL)
	if cfg.KafkaAPIKey != "" && cfg.KafkaAPISecret != "" {
		kafkaConfig.SetKey("security.protocol", cfg.KafkaSecurityProto)
		kafkaConfig.SetKey("sasl.mechanisms", "PLAIN")
		kafkaConfig.SetKey("sasl.username", cfg.KafkaAPIKey)
		kafkaConfig.SetKey("sasl.password", cfg.KafkaAPISecret)
	}

	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("❌ Failed to create Kafka consumer")
	}

	logger.Info().
		Str("brokers", cfg.KafkaBrokers).
		Str("group_id", cfg.KafkaConsumerGroupID).
		Msg("✅ Kafka consumer connected")

	return &confluentConsumer{
		consumer: c,
		logger:   logger,
	}
}

func (c *confluentConsumer) Subscribe(topic string) error {
	err := c.consumer.Subscribe(topic, nil)
	if err != nil {
		c.logger.Error().Err(err).Str("topic", topic).Msg("❌ Failed to subscribe to topic")
		return err
	}
	c.logger.Info().Str("topic", topic).Msg("✅ Subscribed to Kafka topic")
	return nil
}

func (c *confluentConsumer) Poll(timeoutMs int) (*kafka.Message, error) {
	ev := c.consumer.Poll(timeoutMs)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		return e, nil
	case kafka.Error:
		c.logger.Error().Err(e).Msg("❌ Kafka consumer error")
		return nil, e
	default:
		return nil, nil
	}
}

func (c *confluentConsumer) CommitMessage(msg *kafka.Message) error {
	_, err := c.consumer.CommitMessage(msg)
	if err != nil {
		c.logger.Error().Err(err).Msg("❌ Failed to commit Kafka message")
	}
	return err
}

func (c *confluentConsumer) Close() {
	c.consumer.Close()
	c.logger.Info().Msg("🔒 Kafka consumer closed.")
}
