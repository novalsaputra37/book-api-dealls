package bookv2

import (
	"net/http"

	"github.com/adf-code/beta-book-api/internal/delivery/response"
)

// GetFibonacciStatus returns the current fibonacci tracker state:
// counter position, whether it's a fibonacci number, and pending queue count.
func (h *BookV2Handler) GetFibonacciStatus(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Get fibonacci status")
	status := h.BookUC.GetFibonacciStatus()
	response.Success(w, 200, "fibonacci", "getStatus", "Current fibonacci status", status)
}

// GetFibonacciQueue returns all books currently waiting in the non-fibonacci queue.
func (h *BookV2Handler) GetFibonacciQueue(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Get fibonacci pending queue")
	pending, err := h.BookUC.GetPendingQueue(r.Context())
	if err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to fetch pending queue")
		response.Failed(w, 500, "fibonacci", "getQueue", "Error fetching pending queue")
		return
	}
	response.Success(w, 200, "fibonacci", "getQueue", "Pending queue data", pending)
}

// ResetFibonacci resets the fibonacci counter and clears the pending queue.
func (h *BookV2Handler) ResetFibonacci(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("📥 [v2] Reset fibonacci tracker")
	if err := h.BookUC.ResetFibonacci(r.Context()); err != nil {
		h.Logger.Error().Err(err).Msg("❌ [v2] Failed to reset fibonacci")
		response.Failed(w, 500, "fibonacci", "reset", "Error resetting fibonacci tracker")
		return
	}
	response.Success(w, 200, "fibonacci", "reset", "Fibonacci tracker reset successfully", nil)
}
