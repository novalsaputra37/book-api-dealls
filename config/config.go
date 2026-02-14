package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port                string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	Env                 string
	TelemetryAPIKey     string
	TelemetryEndpoint   string
	SendGridAPIKey      string
	SendGridSenderEmail string
	SMTPHost            string
	SMTPPort            int

	MiniEndpoint    string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string

	MongoURI    string
	MongoDBName string

	KafkaBrokers          string
	KafkaClientID         string
	KafkaAPIKey           string
	KafkaAPISecret        string
	KafkaSecurityProto    string
	KafkaTopicBookCreated string
	KafkaTopicBookPending string
	KafkaConsumerGroupID  string
}

func LoadConfig() *AppConfig {
	// Load from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	return &AppConfig{
		Env:                   getEnv("ENV", "development"),
		Port:                  getEnv("APP_PORT", "8080"),
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBUser:                getEnv("DB_USER", "postgres"),
		DBPassword:            getEnv("DB_PASSWORD", ""),
		DBName:                getEnv("DB_NAME", "bookdb"),
		DBSSLMode:             getEnv("DB_SSLMODE", "disable"),
		TelemetryAPIKey:       getEnv("TELEMETRY_API_KEY", "not_set"),
		TelemetryEndpoint:     getEnv("TELEMETRY_ENDPOINT", "not_set"),
		SendGridAPIKey:        getEnv("SENDGRID_API_KEY", "not_set"),
		SendGridSenderEmail:   getEnv("SENDGRID_SENDER_EMAIL", "not_set"),
		SMTPHost:              getEnv("SMTP_HOST", "localhost"),
		SMTPPort:              getEnvInt("SMTP_PORT", 1025),
		MiniEndpoint:          getEnv("MINIO_ENDPOINT", "not_set"),
		MinioAccessKey:        getEnv("MINIO_ACCESS_KEY", "not_set"),
		MinioSecretKey:        getEnv("MINIO_SECRET_KEY", "not_set"),
		MinioBucketName:       getEnv("MINIO_BUCKET_NAME", "not_set"),
		MongoURI:              getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:           getEnv("MONGO_DB_NAME", "beta_book_api"),
		KafkaBrokers:          getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaClientID:         getEnv("KAFKA_CLIENT_ID", "beta-book-api"),
		KafkaAPIKey:           getEnv("KAFKA_API_KEY", ""),
		KafkaAPISecret:        getEnv("KAFKA_API_SECRET", ""),
		KafkaSecurityProto:    getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		KafkaTopicBookCreated: getEnv("KAFKA_TOPIC_BOOK_CREATED", "topic_1"),
		KafkaTopicBookPending: getEnv("KAFKA_TOPIC_BOOK_PENDING", "topic_book_pending"),
		KafkaConsumerGroupID:  getEnv("KAFKA_CONSUMER_GROUP_ID", "beta-book-consumer-group"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
