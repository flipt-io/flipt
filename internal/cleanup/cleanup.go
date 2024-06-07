package cleanup

import (
	"context"
	"fmt"
	"time"

	"go.flipt.io/flipt/internal/config"
	authstorage "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/oplock"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const minCleanupInterval = 5 * time.Minute

// AuthenticationService is configured to run background goroutines which
// will clear out expired authentication tokens.
type AuthenticationService struct {
	logger *zap.Logger
	lock   oplock.Service
	store  authstorage.Store
	config config.AuthenticationConfig

	errgroup errgroup.Group
	cancel   func()
}

// NewAuthenticationService constructs and configures a new instance of authentication service.
func NewAuthenticationService(logger *zap.Logger, lock oplock.Service, store authstorage.Store, config config.AuthenticationConfig) *AuthenticationService {
	return &AuthenticationService{
		logger: logger.With(zap.String("service", "authentication cleanup service")),
		lock:   lock,
		store:  store,
		config: config,
		cancel: func() {},
	}
}

// Run starts up a background goroutine per configure authentication method schedule.
func (s *AuthenticationService) Run(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	for _, info := range s.config.Methods.AllMethods(ctx) {
		logger := s.logger.With(zap.Stringer("method", info.Method))
		if info.Cleanup == nil {
			if info.Enabled {
				logger.Debug("cleanup for auth method not defined (skipping)")
			}

			continue
		}

		if !info.RequiresDatabase {
			if info.Enabled {
				logger.Debug("cleanup for auth method not required (skipping)")
			}

			continue
		}

		var (
			method    = info.Method
			schedule  = info.Cleanup
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

					logger.Warn("attempting to acquire lock", zap.Error(err))
					continue
				}

				// update the next sleep target to current entries acquired until
				acquiredUntil = entry.AcquiredUntil

				if !acquired {
					logger.Debug("cleanup process not acquired", zap.Time("next_attempt", entry.AcquiredUntil))
					continue
				}

				expiredBefore := time.Now().UTC().Add(-schedule.GracePeriod)
				logger.Info("cleanup process deleting authentications", zap.Time("expired_before", expiredBefore))
				if err := s.store.DeleteAuthentications(ctx, authstorage.Delete(
					authstorage.WithMethod(method),
					authstorage.WithExpiredBefore(expiredBefore),
				)); err != nil {
					logger.Error("attempting to delete expired authentications", zap.Error(err))
				}
			}
		})
	}
}

// Stop signals for the cleanup goroutines to cancel and waits for them to finish.
func (s *AuthenticationService) Shutdown(ctx context.Context) error {
	s.logger.Debug("shutting down...")
	defer func() {
		s.logger.Debug("shutdown complete")
	}()

	s.cancel()

	return s.errgroup.Wait()
}
