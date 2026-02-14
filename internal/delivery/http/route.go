package http

import (
	"github.com/adf-code/beta-book-api/internal/delivery/http/book"
	bookv2 "github.com/adf-code/beta-book-api/internal/delivery/http/bookv2"
	"github.com/adf-code/beta-book-api/internal/delivery/http/health"
	"github.com/adf-code/beta-book-api/internal/delivery/http/middleware"
	"github.com/adf-code/beta-book-api/internal/delivery/http/router"
	"github.com/adf-code/beta-book-api/internal/usecase"
	"github.com/rs/zerolog"

	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupHandler(
	bookUC usecase.BookUseCase,
	bookCoverUC usecase.BookCoverUseCase,
	bookMongoUC usecase.BookMongoUseCase,
	bookCoverMongoUC usecase.BookCoverMongoUseCase,
	logger zerolog.Logger,
) http.Handler {
	bookHandler := book.NewBookHandler(bookUC, bookCoverUC, logger)
	bookV2Handler := bookv2.NewBookV2Handler(bookMongoUC, bookCoverMongoUC, logger)
	healthHandler := health.NewHealthHandler(logger)
	auth := middleware.AuthMiddleware(logger)
	log := middleware.LoggingMiddleware(logger)

	r := router.NewRouter()

	r.HandlePrefix(http.MethodGet, "/swagger/", httpSwagger.WrapHandler)

	r.Handle("GET", "/healthz", middleware.Chain(log)(healthHandler.Check))

	// v1 routes (PostgreSQL)
	r.Handle("GET", "/api/v1/books/cover/{id}", middleware.Chain(log, auth)(bookHandler.GetCoverByBookID))
	r.Handle("GET", "/api/v1/books/{id}", middleware.Chain(log, auth)(bookHandler.GetByID))
	r.Handle("GET", "/api/v1/books", middleware.Chain(log, auth)(bookHandler.GetAll))
	r.Handle("POST", "/api/v1/books/upload-cover", middleware.Chain(log, auth)(bookHandler.UploadCover))
	r.Handle("POST", "/api/v1/books", middleware.Chain(log, auth)(bookHandler.Create))
	r.Handle("DELETE", "/api/v1/books/{id}", middleware.Chain(log, auth)(bookHandler.Delete))

	// v2 routes (MongoDB)
	r.Handle("GET", "/api/v2/books/cover/{id}", middleware.Chain(log, auth)(bookV2Handler.GetCoverByBookID))
	r.Handle("GET", "/api/v2/books/fibonacci/status", middleware.Chain(log, auth)(bookV2Handler.GetFibonacciStatus))
	r.Handle("GET", "/api/v2/books/fibonacci/queue", middleware.Chain(log, auth)(bookV2Handler.GetFibonacciQueue))
	r.Handle("POST", "/api/v2/books/fibonacci/reset", middleware.Chain(log, auth)(bookV2Handler.ResetFibonacci))
	r.Handle("GET", "/api/v2/books/{id}", middleware.Chain(log, auth)(bookV2Handler.GetByID))
	r.Handle("GET", "/api/v2/books", middleware.Chain(log, auth)(bookV2Handler.GetAll))
	r.Handle("POST", "/api/v2/books/upload-cover", middleware.Chain(log, auth)(bookV2Handler.UploadCover))
	r.Handle("POST", "/api/v2/books", middleware.Chain(log, auth)(bookV2Handler.Create))
	r.Handle("DELETE", "/api/v2/books/{id}", middleware.Chain(log, auth)(bookV2Handler.Delete))

	return r
}
