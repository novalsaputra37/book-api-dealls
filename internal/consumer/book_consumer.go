package consumer

import (
	"context"
	"encoding/json"

	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/adf-code/beta-book-api/internal/pkg/messages"
	"github.com/adf-code/beta-book-api/internal/repository"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
)

// BookConsumer reads book messages from Kafka.
// Fibonacci books are saved to MongoDB books collection;
// non-fibonacci books are saved to the book_queue collection.
type BookConsumer struct {
	consumer  messages.KafkaConsumer
	bookRepo  repository.BookMongoRepository
	queueRepo repository.BookQueueMongoRepository
	logger    zerolog.Logger
}

func NewBookConsumer(
	consumer messages.KafkaConsumer,
	bookRepo repository.BookMongoRepository,
	queueRepo repository.BookQueueMongoRepository,
	logger zerolog.Logger,
) *BookConsumer {
	return &BookConsumer{
		consumer:  consumer,
		bookRepo:  bookRepo,
		queueRepo: queueRepo,
		logger:    logger,
	}
}

// Start begins consuming messages from the given Kafka topic.
// It blocks until the context is cancelled.
func (c *BookConsumer) Start(ctx context.Context, topic string) {
	if err := c.consumer.Subscribe(topic); err != nil {
		c.logger.Fatal().Err(err).Str("topic", topic).Msg("❌ Failed to subscribe to topic")
	}

	c.logger.Info().Str("topic", topic).Msg("🚀 Book consumer started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("🛑 Book consumer stopping...")
			return
		default:
			msg, err := c.consumer.Poll(1000)
			if err != nil {
				c.logger.Error().Err(err).Msg("❌ Consumer poll error")
				continue
			}
			if msg == nil {
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *BookConsumer) processMessage(ctx context.Context, msg *kafka.Message) {
	var bookMsg entity.BookMessage
	if err := json.Unmarshal(msg.Value, &bookMsg); err != nil {
		c.logger.Error().Err(err).
			Str("raw", string(msg.Value)).
			Msg("❌ Failed to unmarshal book message")
		// Commit to skip malformed messages
		_ = c.consumer.CommitMessage(msg)
		return
	}

	c.logger.Info().
		Str("book_id", bookMsg.Book.ID.String()).
		Str("title", bookMsg.Book.Title).
		Int64("position", bookMsg.Position).
		Bool("is_fibonacci", bookMsg.IsFibonacci).
		Msg("📥 Received book message from Kafka")

	if bookMsg.IsFibonacci {
		// Fibonacci → save to MongoDB books collection
		book := bookMsg.Book
		if err := c.bookRepo.Store(ctx, &book); err != nil {
			c.logger.Error().Err(err).
				Str("book_id", bookMsg.Book.ID.String()).
				Msg("❌ Failed to store fibonacci book in MongoDB")
			// Don't commit → will be retried on next poll
			return
		}

		c.logger.Info().
			Str("book_id", bookMsg.Book.ID.String()).
			Str("title", bookMsg.Book.Title).
			Int64("position", bookMsg.Position).
			Msg("✅ Fibonacci book saved to MongoDB")
	} else {
		// Non-fibonacci → save to queue collection in MongoDB
		book := bookMsg.Book
		if err := c.queueRepo.Store(ctx, &book); err != nil {
			c.logger.Error().Err(err).
				Str("book_id", bookMsg.Book.ID.String()).
				Msg("❌ Failed to store non-fibonacci book in queue")
			return
		}

		c.logger.Info().
			Str("book_id", bookMsg.Book.ID.String()).
			Str("title", bookMsg.Book.Title).
			Int64("position", bookMsg.Position).
			Msg("📦 Non-fibonacci book saved to queue collection")
	}

	// Commit the message offset
	if err := c.consumer.CommitMessage(msg); err != nil {
		c.logger.Error().Err(err).Msg("❌ Failed to commit Kafka message")
	}
}
