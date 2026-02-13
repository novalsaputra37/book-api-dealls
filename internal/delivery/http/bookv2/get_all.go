package bookv2

import (
	"net/http"

	"github.com/adf-code/beta-book-api/internal/delivery/request"
	"github.com/adf-code/beta-book-api/internal/delivery/response"
)

func (h *BookV2Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Incoming GetAll request")
	params := request.ParseBookQueryParams(r)
	books, err := h.BookUC.GetAll(r.Context(), params)
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to fetch books, general")
		response.FailedWithMeta(w, 500, "books", "getAllBooks", "Error Get All Books", nil)
		return
	}
	h.Logger.Info().Int("count", len(books)).Msg("✅ [v2] Successfully fetched books")
	response.SuccessWithMeta(w, 200, "books", "getAllBooks", "Success Get All Books", &params, books)
}
