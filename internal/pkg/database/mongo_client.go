package database

import (
	"context"
	"time"

	"github.com/adf-code/beta-book-api/config"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoClient struct {
	uri    string
	dbName string
	logger zerolog.Logger
}

func NewMongoClient(cfg *config.AppConfig, logger zerolog.Logger) *MongoClient {
	return &MongoClient{
		uri:    cfg.MongoURI,
		dbName: cfg.MongoDBName,
		logger: logger,
	}
}

func (m *MongoClient) InitMongoDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(m.uri))
	if err != nil {
		m.logger.Fatal().Err(err).Msg("❌ Failed to connect to MongoDB")
	}

	if err := client.Ping(ctx, nil); err != nil {
		m.logger.Fatal().Err(err).Msg("❌ Failed to ping MongoDB")
	}

	m.logger.Info().Msgf("✅ Connected to MongoDB successfully (db: %s)", m.dbName)

	return client.Database(m.dbName)
}

func CloseMongoDB(client *mongo.Client, logger zerolog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		logger.Info().Msgf("⚠️ Failed to close MongoDB connection: %v", err)
	} else {
		logger.Info().Msg("🔒 MongoDB connection closed.")
	}
}
