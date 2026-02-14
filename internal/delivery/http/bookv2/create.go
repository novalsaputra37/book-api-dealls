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

	result, err := h.BookUC.Create(r.Context(), book)
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to store book, general")
		response.Failed(w, 500, "books", "createBook", "Error Create Book")
		return
	}

	if result.Queued {
		h.Logger.Info().
			Int64("position", result.Position).
			Msg("📦 [v2] Book queued (non-fibonacci)")
		msg := fmt.Sprintf("Book queued at position #%d (next fibonacci: waiting)", result.Position)
		response.Success(w, 202, "books", "createBook", msg, result)
		return
	}

	h.Logger.Info().
		Int64("position", result.Position).
		Int("processed", result.ProcessedCount).
		Msg("✅ [v2] Book created (fibonacci hit)")
	msg := fmt.Sprintf("Book created at fibonacci position #%d, processed %d pending books", result.Position, result.ProcessedCount)
	response.Success(w, 201, "books", "createBook", msg, result)
}
