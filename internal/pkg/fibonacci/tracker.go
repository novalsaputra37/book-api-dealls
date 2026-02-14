package fibonacci

import (
	"math"
	"sync"

	"github.com/adf-code/beta-book-api/internal/entity"
)

// Status represents the current state of the fibonacci tracker.
type Status struct {
	Counter       int64 `json:"counter"`
	IsFibonacci   bool  `json:"is_fibonacci"`
	NextFibonacci int64 `json:"next_fibonacci"`
	PendingCount  int   `json:"pending_count"`
}

// Tracker manages a request counter and determines fibonacci-based processing.
// Requests hitting a fibonacci position are processed immediately;
// non-fibonacci requests are queued until the next fibonacci trigger.
type Tracker struct {
	mu      sync.Mutex
	counter int64
	pending []entity.Book
}

// NewTracker creates a new fibonacci tracker with an empty queue.
func NewTracker() *Tracker {
	return &Tracker{
		pending: make([]entity.Book, 0),
	}
}

// NextHit increments the counter and returns the new position and whether it is a fibonacci number.
func (t *Tracker) NextHit() (int64, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.counter++
	return t.counter, IsFibonacci(t.counter)
}

// CurrentStatus returns the current fibonacci status snapshot.
func (t *Tracker) CurrentStatus() Status {
	t.mu.Lock()
	defer t.mu.Unlock()
	return Status{
		Counter:       t.counter,
		IsFibonacci:   IsFibonacci(t.counter),
		NextFibonacci: nextFibonacciAfter(t.counter),
		PendingCount:  len(t.pending),
	}
}

// AddPending adds a book to the pending queue.
func (t *Tracker) AddPending(book entity.Book) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pending = append(t.pending, book)
}

// DrainPending removes and returns all pending books from the queue.
func (t *Tracker) DrainPending() []entity.Book {
	t.mu.Lock()
	defer t.mu.Unlock()
	drained := make([]entity.Book, len(t.pending))
	copy(drained, t.pending)
	t.pending = make([]entity.Book, 0)
	return drained
}

// GetPending returns a copy of the current pending books without draining.
func (t *Tracker) GetPending() []entity.Book {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]entity.Book, len(t.pending))
	copy(result, t.pending)
	return result
}

// Reset clears the counter and pending queue.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.counter = 0
	t.pending = make([]entity.Book, 0)
}

// isPerfectSquare checks if n is a perfect square.
func isPerfectSquare(n int64) bool {
	if n < 0 {
		return false
	}
	s := int64(math.Sqrt(float64(n)))
	return s*s == n
}

// IsFibonacci checks if n is a fibonacci number.
// A number is fibonacci if and only if (5*n*n + 4) or (5*n*n - 4) is a perfect square.
func IsFibonacci(n int64) bool {
	if n <= 0 {
		return false
	}
	return isPerfectSquare(5*n*n+4) || isPerfectSquare(5*n*n-4)
}

// nextFibonacciAfter returns the smallest fibonacci number greater than n.
func nextFibonacciAfter(n int64) int64 {
	if n <= 0 {
		return 1
	}
	a, b := int64(1), int64(1)
	for b <= n {
		a, b = b, a+b
	}
	return b
}
