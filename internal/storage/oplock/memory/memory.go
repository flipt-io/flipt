package memory

import (
	"context"
	"sync"
	"time"

	"go.flipt.io/flipt/internal/storage/oplock"
)

// Service is an in-memory implementation of the oplock.Service.
// It is only safe for single instance / in-process use.
type Service struct {
	mu sync.Mutex

	ops map[oplock.Operation]oplock.LockEntry
}

// New constructs and configures a new service instance.
func New() *Service {
	return &Service{ops: map[oplock.Operation]oplock.LockEntry{}}
}

// TryAcquire will attempt to obtain a lock for the supplied operation name for the specified duration.
// If it succeeds then the returned boolean (acquired) will be true, else false.
// The lock entry associated with the last successful acquisition is also returned.
// Given the lock was acquired successfully this will be the entry just created.
func (s *Service) TryAcquire(ctx context.Context, operation oplock.Operation, duration time.Duration) (acquired bool, entry oplock.LockEntry, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	entry, ok := s.ops[operation]
	if !ok {
		entry.Operation = operation
		entry.Version = 1
		entry.LastAcquired = now
		entry.AcquiredUntil = now.Add(duration)
		s.ops[operation] = entry
		return true, entry, nil
	}

	if entry.AcquiredUntil.Before(now) {
		entry.Version++
		entry.LastAcquired = now
		entry.AcquiredUntil = now.Add(duration)
		s.ops[operation] = entry
		return true, entry, nil
	}

	return false, entry, nil
}
