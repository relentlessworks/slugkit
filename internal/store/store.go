package store

import (
	"sync/atomic"
)

// Stats tracks in-memory usage counters.
type Stats struct {
	SlugsGenerated int64
	SlugsValidated  int64
}

// Store is an in-memory usage tracker. No persistence needed for a stateless service.
type Store struct {
	slugs     int64
	validated int64
}

// New creates a new in-memory store.
func New() *Store {
	return &Store{}
}

func (s *Store) IncrSlug()    { atomic.AddInt64(&s.slugs, 1) }
func (s *Store) IncrValid()   { atomic.AddInt64(&s.validated, 1) }

// GetStats returns a snapshot of current usage counters.
func (s *Store) GetStats() Stats {
	return Stats{
		SlugsGenerated: atomic.LoadInt64(&s.slugs),
		SlugsValidated: atomic.LoadInt64(&s.validated),
	}
}
