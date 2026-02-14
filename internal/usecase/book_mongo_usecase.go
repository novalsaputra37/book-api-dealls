package usecase

import (
	"context"
	"time"

	"github.com/adf-code/beta-book-api/config"
	"github.com/adf-code/beta-book-api/internal/delivery/request"
	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/adf-code/beta-book-api/internal/pkg/fibonacci"
	"github.com/adf-code/beta-book-api/internal/pkg/messages"
	"github.com/adf-code/beta-book-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CreateBookResult holds the result of a v2 create book operation with fibonacci metadata.
type CreateBookResult struct {
	Book           *entity.Book  `json:"book"`
	Position       int64         `json:"position"`
	IsFibonacci    bool          `json:"is_fibonacci"`
	Queued         bool          `json:"queued"`
	ProcessedCount int           `json:"processed_count"`
	ProcessedBooks []entity.Book `json:"processed_books,omitempty"`
}

type BookMongoUseCase interface {
	GetAll(ctx context.Context, params request.BookListQueryParams) ([]entity.Book, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error)
	Create(ctx context.Context, book entity.Book) (*CreateBookResult, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetFibonacciStatus() fibonacci.Status
	GetPendingQueue(ctx context.Context) ([]entity.Book, error)
	ResetFibonacci(ctx context.Context) error
}

type bookMongoUseCase struct {
	bookRepo      repository.BookMongoRepository
	bookQueueRepo repository.BookQueueMongoRepository
	logger        zerolog.Logger
	kafka         messages.KafkaClient
	cfg           *config.AppConfig
	fibTracker    *fibonacci.Tracker
}

func NewBookMongoUseCase(
	bookRepo repository.BookMongoRepository,
	bookQueueRepo repository.BookQueueMongoRepository,
	logger zerolog.Logger,
	kafka messages.KafkaClient,
	cfg *config.AppConfig,
	fibTracker *fibonacci.Tracker,
) BookMongoUseCase {
	return &bookMongoUseCase{
		bookRepo:      bookRepo,
		bookQueueRepo: bookQueueRepo,
		logger:        logger,
		kafka:         kafka,
		cfg:           cfg,
		fibTracker:    fibTracker,
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

// Create publishes the book to Kafka with fibonacci metadata.
// The Kafka consumer will process the message and save to MongoDB.
//   - If the current request position is a fibonacci number: marked as is_fibonacci=true
//   - If the position is NOT a fibonacci number: marked as is_fibonacci=false (queued)
func (uc *bookMongoUseCase) Create(ctx context.Context, book entity.Book) (*CreateBookResult, error) {
	position, isFib := uc.fibTracker.NextHit()

	uc.logger.Info().
		Int64("position", position).
		Bool("is_fibonacci", isFib).
		Msg("⚙️ [v2-mongo] Create book — fibonacci check")

	// Prepare book data
	now := time.Now()
	book.ID = uuid.New()
	book.CreatedAt = &now
	book.UpdatedAt = &now
	book.BookCover = make([]entity.BookCover, 0)

	// Build Kafka message with fibonacci metadata
	bookMsg := entity.BookMessage{
		Book:        book,
		Position:    position,
		IsFibonacci: isFib,
	}

	// Publish to Kafka pending topic
	if err := uc.kafka.Publish(uc.cfg.KafkaTopicBookPending, book.ID.String(), bookMsg); err != nil {
		uc.logger.Error().Err(err).Msg("❌ [v2-mongo] Failed to publish book to Kafka")
		return nil, err
	}

	result := &CreateBookResult{
		Book:        &book,
		Position:    position,
		IsFibonacci: isFib,
	}

	if isFib {
		uc.logger.Info().
			Str("book_id", book.ID.String()).
			Int64("position", position).
			Msg("✅ [v2-mongo] Fibonacci book published to Kafka (will be saved by consumer)")
		result.Queued = false
	} else {
		uc.logger.Info().
			Str("book_id", book.ID.String()).
			Int64("position", position).
			Msg("📦 [v2-mongo] Non-fibonacci book published to Kafka (queued)")
		result.Queued = true
	}

	return result, nil
}

func (uc *bookMongoUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	uc.logger.Info().Str("usecase", "Delete").Msg("⚙️ [v2-mongo] Remove book")
	return uc.bookRepo.Remove(ctx, id)
}

// GetFibonacciStatus returns current fibonacci tracker status.
func (uc *bookMongoUseCase) GetFibonacciStatus() fibonacci.Status {
	return uc.fibTracker.CurrentStatus()
}

// GetPendingQueue returns non-fibonacci books from the MongoDB queue collection.
func (uc *bookMongoUseCase) GetPendingQueue(ctx context.Context) ([]entity.Book, error) {
	return uc.bookQueueRepo.FetchAll(ctx)
}

// ResetFibonacci resets the fibonacci counter and clears the queue collection.
func (uc *bookMongoUseCase) ResetFibonacci(ctx context.Context) error {
	uc.fibTracker.Reset()
	if err := uc.bookQueueRepo.RemoveAll(ctx); err != nil {
		uc.logger.Error().Err(err).Msg("❌ [v2-mongo] Failed to clear book queue")
		return err
	}
	uc.logger.Info().Msg("🔄 [v2-mongo] Fibonacci tracker reset and queue cleared")
	return nil
}
