package memory

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage/oplock"
	storagesql "go.flipt.io/flipt/internal/storage/sql"
	"go.uber.org/zap"
)

// Service is an in-memory implementation of the oplock.Service.
// It is only safe for single instance / in-process use.
type Service struct {
	logger  *zap.Logger
	driver  storagesql.Driver
	builder sq.StatementBuilderType
}

// New constructs and configures a new service instance.
func New(logger *zap.Logger, driver storagesql.Driver, builder sq.StatementBuilderType) *Service {
	return &Service{
		logger:  logger,
		driver:  driver,
		builder: builder,
	}
}

// TryAcquire will attempt to obtain a lock for the supplied operation name for the specified duration.
// If it succeeds then the returned boolean (acquired) will be true, else false.
// The lock entry associated with the last successful acquisition is also returned.
// Given the lock was acquired successfully this will be the entry just created.
func (s *Service) TryAcquire(ctx context.Context, operation oplock.Operation, duration time.Duration) (acquired bool, entry oplock.LockEntry, err error) {
	entry, err = s.readEntry(ctx, operation)
	if err != nil {
		if _, match := errors.As[errors.ErrNotFound](err); match {
			// entry does not exist so we try and create one
			entry, err := s.insertEntry(ctx, operation, duration)
			if err != nil {
				if _, match := errors.As[errors.ErrInvalid](err); match {
					// check if the entry is invalid due to
					// uniqueness constraint violation
					// if so re-read the current entry and return that
					entry, err := s.readEntry(ctx, operation)
					return false, entry, err
				}

				return false, entry, err
			}

			return true, entry, nil
		}

		// something went wrong
		return false, entry, err
	}

	// entry exists so first check the acquired until has elapsed
	if time.Now().UTC().Before(entry.AcquiredUntil) {
		// return early as the lock is still acquired
		return false, entry, nil
	}

	acquired, err = s.acquireEntry(ctx, &entry, duration)

	return acquired, entry, err
}

func (s *Service) acquireEntry(ctx context.Context, entry *oplock.LockEntry, dur time.Duration) (acquired bool, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("updating existing entry: %w", s.driver.AdaptError(err))
		}
	}()

	now := time.Now().UTC()
	query := s.builder.Update("operation_lock").
		Set("version", entry.Version+1).
		Set("last_acquired_at", now).
		Set("acquired_until", now.Add(dur)).
		Where(sq.Eq{
			"operation": string(entry.Operation),
			// ensure current entry has not been updated
			"version": entry.Version,
		})

	res, err := query.ExecContext(ctx)
	if err != nil {
		return false, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if count < 1 {
		// current entry version does not match
		// therefore we can assume it was updated
		// by concurrent lock acquirer
		return false, nil
	}

	entry.Version++
	entry.LastAcquired = now
	entry.AcquiredUntil = now.Add(dur)
	return true, nil
}

func (s *Service) insertEntry(ctx context.Context, op oplock.Operation, dur time.Duration) (entry oplock.LockEntry, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("inserting new entry: %w", err)
		}
	}()

	entry.Operation = op
	entry.Version = 1
	entry.LastAcquired = time.Now().UTC()
	entry.AcquiredUntil = entry.LastAcquired.Add(dur)

	_, err = s.builder.Insert("operation_lock").
		Columns(
			"operation",
			"version",
			"last_acquired_at",
			"acquired_until",
		).Values(
		&entry.Operation,
		&entry.Version,
		&entry.LastAcquired,
		&entry.AcquiredUntil,
	).ExecContext(ctx)

	return entry, s.driver.AdaptError(err)
}

func (s *Service) readEntry(ctx context.Context, operation oplock.Operation) (entry oplock.LockEntry, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("reading entry: %w", err)
		}
	}()

	err = s.builder.Select(
		"operation",
		"version",
		"last_acquired_at",
		"acquired_until",
	).From("operation_lock").
		Where(sq.Eq{"operation": string(operation)}).
		QueryRowContext(ctx).
		Scan(
			&entry.Operation,
			&entry.Version,
			&entry.LastAcquired,
			&entry.AcquiredUntil,
		)

	return entry, s.driver.AdaptError(err)
}
