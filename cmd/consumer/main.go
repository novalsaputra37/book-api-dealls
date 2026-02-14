package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/adf-code/beta-book-api/config"
	"github.com/adf-code/beta-book-api/internal/consumer"
	pkgDatabase "github.com/adf-code/beta-book-api/internal/pkg/database"
	pkgLogger "github.com/adf-code/beta-book-api/internal/pkg/logger"
	pkgMessages "github.com/adf-code/beta-book-api/internal/pkg/messages"
	"github.com/adf-code/beta-book-api/internal/repository"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.LoadConfig()
	logger := pkgLogger.InitLoggerWithTelemetry(cfg)

	logger.Info().Msg("🚀 Starting Book Consumer...")

	// MongoDB
	mongoClient := pkgDatabase.NewMongoClient(cfg, logger)
	mongoDB := mongoClient.InitMongoDB()

	// Kafka Consumer
	kafkaConsumer := pkgMessages.NewConfluentConsumer(cfg, logger)

	// Repositories
	bookRepo := repository.NewBookMongoRepo(mongoDB)
	queueRepo := repository.NewBookQueueMongoRepo(mongoDB)

	// Book Consumer
	bookConsumer := consumer.NewBookConsumer(kafkaConsumer, bookRepo, queueRepo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in goroutine
	go func() {
		bookConsumer.Start(ctx, cfg.KafkaTopicBookPending)
	}()

	logger.Info().
		Str("topic", cfg.KafkaTopicBookPending).
		Msg("✅ Book consumer is running")

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("🛑 Shutting down consumer...")
	cancel()
	kafkaConsumer.Close()
	pkgDatabase.CloseMongoDB(mongoDB.Client(), logger)
	logger.Info().Msg("✅ Consumer shutdown completed.")
}
