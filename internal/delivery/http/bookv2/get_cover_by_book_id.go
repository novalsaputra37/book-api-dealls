package bookv2

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/adf-code/beta-book-api/internal/delivery/http/router"
	"github.com/adf-code/beta-book-api/internal/delivery/response"
	"github.com/adf-code/beta-book-api/internal/entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (h *BookV2Handler) GetCoverByBookID(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Incoming GetCoverByBookID request")
	idStr := router.GetParam(r, "id")
	if idStr == "" {
		h.Logger.Error().Msg("❌ [v2] Failed to get book by ID, missing ID parameter")
		response.Failed(w, 422, "books", "getBookByID", "Missing ID Parameter, Get Book by ID")
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to get book by ID, invalid UUID parameter")
		response.Failed(w, 422, "books", "getBookByID", "Invalid UUID, Get Book by ID")
		return
	}
	book, err := h.BookUC.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			h.Logger.Info().Msg("✅ [v2] Successfully get book by id, data not found")
			response.Success(w, 404, "books", "getBookByID", "Book not Found", nil)
			return
		}
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to get book by ID, general")
		response.Failed(w, 500, "books", "getBookByID", "Error Get Book by ID")
		return
	}
	booksCover, err := h.BookCoverUC.GetByBookID(r.Context(), id)
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to fetch books cover, general")
		response.FailedWithMeta(w, 500, "books", "getAllBooks", "Error Get Book Cover by Book ID", nil)
		return
	}

	if len(booksCover) == 0 {
		book.BookCover = make([]entity.BookCover, 0)
	} else {
		book.BookCover = booksCover
	}
	h.Logger.Info().Str("data", fmt.Sprint(book.ID)).Msg("✅ [v2] Successfully get book by id")
	response.Success(w, 200, "books", "getBookByID", "Success Get Book by ID", book)
}
