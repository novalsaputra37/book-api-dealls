package bookv2

import (
	"net/http"

	"github.com/adf-code/beta-book-api/internal/delivery/http/router"
	"github.com/adf-code/beta-book-api/internal/delivery/response"
	"github.com/google/uuid"
)

func (h *BookV2Handler) Delete(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Incoming Delete request")
	idStr := router.GetParam(r, "id")
	if idStr == "" {
		h.Logger.Error().Msg("❌ [v2] Failed to remove book, missing ID parameter")
		response.Failed(w, 422, "books", "deleteBookByID", "Missing ID Parameter, Delete Book by ID")
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to remove book, invalid UUID parameter")
		response.Failed(w, 422, "books", "deleteBookByID", "Invalid UUID, Delete Book by ID")
		return
	}
	if err := h.BookUC.Delete(r.Context(), id); err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to remove book, general")
		response.Failed(w, 500, "books", "deleteBookByID", "Error Delete Book")
		return
	}
	h.Logger.Info().Msg("✅ [v2] Successfully removed book")
	response.Success(w, 202, "books", "deleteBookByID", "Success Delete Book", nil)
}
