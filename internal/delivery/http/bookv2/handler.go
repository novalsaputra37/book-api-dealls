package bookv2

import (
	"github.com/adf-code/beta-book-api/internal/usecase"
	"github.com/rs/zerolog"
)

type BookV2Handler struct {
	BookUC      usecase.BookMongoUseCase
	BookCoverUC usecase.BookCoverMongoUseCase
	Logger      zerolog.Logger
}

func NewBookV2Handler(bookUC usecase.BookMongoUseCase, bookCoverUC usecase.BookCoverMongoUseCase, logger zerolog.Logger) *BookV2Handler {
	return &BookV2Handler{BookUC: bookUC, BookCoverUC: bookCoverUC, Logger: logger}
}
