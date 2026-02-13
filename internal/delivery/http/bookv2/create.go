package bookv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adf-code/beta-book-api/internal/delivery/response"
	"github.com/adf-code/beta-book-api/internal/entity"
)

func (h *BookV2Handler) Create(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Incoming Create request")
	var book entity.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to store book, invalid data")
		response.Failed(w, 422, "books", "createBook", "Invalid Data, Create Book")
		return
	}

	newBook, err := h.BookUC.Create(r.Context(), book)
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to store book, general")
		response.Failed(w, 500, "books", "createBook", "Error Create Book")
		return
	}
	newBook.BookCover = make([]entity.BookCover, 0)
	h.Logger.Info().Str("data", fmt.Sprint(newBook)).Msg("✅ [v2] Successfully stored book")
	response.Success(w, 201, "books", "createBook", "Success Create Book", newBook)
}
