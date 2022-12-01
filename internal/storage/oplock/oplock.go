package oplock

import (
	"context"
	"time"
)

// Operation is a string which identifies a particular unique operation name.
type Operation string

type LockEntry struct {
	Operation     Operation
	Version       int64
	LastAcquired  time.Time
	AcquiredUntil time.Time
}

// Service is an operation lock service which provides the ability to lock access
// to perform a named operation up until an ellapsed duration.
// Implementations of this type can be used to ensure an operation occurs once per
// the provided elapsed duration between a set of Flipt instances.
// If coordinating a distributed set of Flipt instances then a remote backend (e.g. SQL)
// will be required. In memory implementations will only work for single instance deployments.
type Service interface {
	// TryAcquire will attempt to obtain a lock for the supplied operation name for the specified duration.
	// If it succeeds then the returned boolean (acquired) will be true, else false.
	// The lock entry associated with the last successful acquisition is also returned.
	// Given the lock was acquired successfully this will be the entry just created.
	TryAcquire(ctx context.Context, operation Operation, duration time.Duration) (acquired bool, entry LockEntry, err error)
}
