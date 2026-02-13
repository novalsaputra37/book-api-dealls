package usecase

import (
	"context"

	"github.com/adf-code/beta-book-api/config"
	"github.com/adf-code/beta-book-api/internal/delivery/request"
	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/adf-code/beta-book-api/internal/pkg/mail"
	"github.com/adf-code/beta-book-api/internal/pkg/messages"
	"github.com/adf-code/beta-book-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type BookMongoUseCase interface {
	GetAll(ctx context.Context, params request.BookListQueryParams) ([]entity.Book, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error)
	Create(ctx context.Context, book entity.Book) (*entity.Book, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type bookMongoUseCase struct {
	bookRepo    repository.BookMongoRepository
	logger      zerolog.Logger
	emailClient mail.EmailClient
	kafka       messages.KafkaClient
	cfg         *config.AppConfig
}

func NewBookMongoUseCase(bookRepo repository.BookMongoRepository, logger zerolog.Logger, emailClient mail.EmailClient, kafka messages.KafkaClient, cfg *config.AppConfig) BookMongoUseCase {
	return &bookMongoUseCase{
		bookRepo:    bookRepo,
		logger:      logger,
		emailClient: emailClient,
		kafka:       kafka,
		cfg:         cfg,
	}
}

func (uc *bookMongoUseCase) GetAll(ctx context.Context, params request.BookListQueryParams) ([]entity.Book, error) {
	uc.logger.Info().Str("usecase", "GetAll").Msg("⚙️ [v2-mongo] Fetching all books")
	return uc.bookRepo.FetchWithQueryParams(ctx, params)
}

func (uc *bookMongoUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error) {
	uc.logger.Info().Str("usecase", "GetByID").Msg("⚙️ [v2-mongo] Fetching book by ID")
	return uc.bookRepo.FetchByID(ctx, id)
}

func (uc *bookMongoUseCase) Create(ctx context.Context, book entity.Book) (*entity.Book, error) {
	uc.logger.Info().Str("usecase", "Create").Msg("⚙️ [v2-mongo] Store book")

	err := uc.bookRepo.Store(ctx, &book)
	if err != nil {
		uc.logger.Error().Err(err).Msg("❌ [v2-mongo] Failed to store book")
		return nil, err
	}

	err = uc.emailClient.SendBookCreatedEmail(book)
	if err != nil {
		uc.logger.Error().Err(err).Msg("❌ [v2-mongo] Failed to send email")
		return nil, err
	}

	uc.logger.Info().Str("book_id", book.ID.String()).Msg("✅ [v2-mongo] Book created and email sent successfully")

	// Publish event to Kafka
	if err := uc.kafka.Publish(uc.cfg.KafkaTopicBookCreated, book.ID.String(), book); err != nil {
		uc.logger.Error().Err(err).Msg("⚠️ [v2-mongo] Failed to publish Kafka event (non-blocking)")
	}

	return &book, nil
}

func (uc *bookMongoUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	uc.logger.Info().Str("usecase", "Delete").Msg("⚙️ [v2-mongo] Remove book")
	return uc.bookRepo.Remove(ctx, id)
}
