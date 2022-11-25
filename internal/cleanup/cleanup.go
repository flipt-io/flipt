package cleanup

import (
	"context"
	"fmt"
	"time"

	"go.flipt.io/flipt/internal/config"
	authstorage "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/internal/storage/oplock"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const minCleanupInterval = 5 * time.Minute

// AuthenticationService is configured to run background goroutines which
// will clear out expired authentication tokens.
type AuthenticationService struct {
	logger    *zap.Logger
	lock      oplock.Service
	store     authstorage.Store
	schedules config.AuthenticationCleanupSchedules

	errgroup errgroup.Group
	cancel   func()
}

// NewAuthenticationService constructs and configures a new instance of authentication service.
func NewAuthenticationService(logger *zap.Logger, lock oplock.Service, store authstorage.Store, schedules config.AuthenticationCleanupSchedules) *AuthenticationService {
	return &AuthenticationService{
		logger:    logger,
		lock:      lock,
		store:     store,
		schedules: schedules,
		cancel:    func() {},
	}
}

// Run starts up a background goroutine per configure authentication method schedule.
func (s *AuthenticationService) Run(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)
	for method, schedule := range s.schedules {
		var (
			method    = method
			schedule  = schedule
			operation = oplock.Operation(fmt.Sprintf("cleanup_auth_%s", method))
		)

		s.errgroup.Go(func() error {
			// on the first attempt to run the cleanup authentication service
			// we attempt to obtain the lock immediately. If the lock is already
			// held the service should return false and return the current acquired
			// current timestamp
			acquiredUntil := time.Now().UTC()
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Until(acquiredUntil)):
				}

				acquired, entry, err := s.lock.TryAcquire(ctx, operation, schedule.Interval)
				if err != nil {
					// ensure we dont go into hot loop when the operation lock service
					// enters an error state by ensuring we sleep for at-least the minimum
					// interval.
					now := time.Now().UTC()
					if acquiredUntil.Before(now) {
						acquiredUntil = now.Add(minCleanupInterval)
					}

					s.logger.Warn("attempting to acquire lock", zap.Error(err))
					continue
				}

				// update the next sleep target to current entries acquired until
				acquiredUntil = entry.AcquiredUntil

				if !acquired {
					s.logger.Info("cleanup process not acquired", zap.Time("next_attempt", entry.AcquiredUntil))
					continue
				}

				expiredBefore := time.Now().UTC().Add(-schedule.GracePeriod)
				s.logger.Info("cleanup process deleting authentications", zap.Time("expired_before", expiredBefore))
				if err := s.store.DeleteAuthentications(ctx, authstorage.Delete(
					authstorage.WithMethod(method),
					authstorage.WithExpiredBefore(expiredBefore),
				)); err != nil {
					s.logger.Error("attempting to delete expired authentications", zap.Error(err))
				}
			}
		})
	}
}

// Stop signals for the cleanup goroutines to cancel and waits for them to finish.
func (s *AuthenticationService) Stop() error {
	s.cancel()

	return s.errgroup.Wait()
}
