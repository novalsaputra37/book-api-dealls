package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/adf-code/beta-book-api/internal/pkg/object_storage"
	"github.com/adf-code/beta-book-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type BookCoverMongoUseCase interface {
	Upload(ctx context.Context, bookID uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (*entity.BookCover, error)
	GetByBookID(ctx context.Context, bookID uuid.UUID) ([]entity.BookCover, error)
}

type bookCoverMongoUseCase struct {
	bookCoverRepo repository.BookCoverMongoRepository
	logger        zerolog.Logger
	objectStorage object_storage.ObjectStorageClient
}

func NewBookCoverMongoUseCase(bookCoverRepo repository.BookCoverMongoRepository, logger zerolog.Logger, storage object_storage.ObjectStorageClient) BookCoverMongoUseCase {
	return &bookCoverMongoUseCase{
		bookCoverRepo: bookCoverRepo,
		logger:        logger,
		objectStorage: storage,
	}
}

func (uc *bookCoverMongoUseCase) Upload(ctx context.Context, bookID uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader) (*entity.BookCover, error) {
	uc.logger.Info().Str("usecase", "UploadCover").Msg("⚙️ [v2-mongo] Upload book cover")
	timestamp := time.Now().Format(time.RFC3339)
	sanitizedTimestamp := strings.ReplaceAll(timestamp, ":", "-")
	objectName := fmt.Sprintf("covers/book_%s_%s_%s", bookID, sanitizedTimestamp, fileHeader.Filename)

	url, err := uc.objectStorage.UploadFile(ctx, file, objectName, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		uc.logger.Error().Err(err).Msg("❌ [v2-mongo] Failed upload file to object storage")
		return nil, err
	}

	cover := entity.BookCover{
		BookID:   bookID,
		FileName: fileHeader.Filename,
		FileURL:  url,
	}

	err = uc.bookCoverRepo.Store(ctx, &cover)
	if err != nil {
		uc.logger.Error().Err(err).Msg("❌ [v2-mongo] Failed to store book cover")
		return nil, err
	}

	uc.logger.Info().Str("book_cover_id", cover.ID.String()).Msg("✅ [v2-mongo] Book Cover uploaded successfully")
	return &cover, nil
}

func (uc *bookCoverMongoUseCase) GetByBookID(ctx context.Context, bookID uuid.UUID) ([]entity.BookCover, error) {
	uc.logger.Info().Str("usecase", "GetByBookID").Msg("⚙️ [v2-mongo] Fetching books cover by book id")
	return uc.bookCoverRepo.FetchByBookID(ctx, bookID)
}
