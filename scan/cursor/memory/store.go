package memory

import (
	"context"
	"sync"
)

// Store is a generic, thread-safe in-memory cursor store.
// Each Store instance is scoped to a single scanner.
type Store[Cur any] struct {
	mu  sync.RWMutex
	cur Cur
	ok  bool
}

// NewStore returns a new in-memory cursor store.
func NewStore[Cur any]() *Store[Cur] {
	return &Store[Cur]{}
}

// Get returns the stored cursor and whether it has been set.
func (s *Store[Cur]) Get(_ context.Context) (Cur, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cur, s.ok, nil
}

// Set stores the cursor.
func (s *Store[Cur]) Set(_ context.Context, cur Cur) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cur = cur
	s.ok = true
	return nil
}
