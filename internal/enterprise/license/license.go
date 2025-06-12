package license

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/keygen-sh/keygen-go/v3"
	"github.com/keygen-sh/machineid"
	"go.uber.org/zap"
)

var (
	managerOnce sync.Once
	manager     *Manager
	managerFunc func(context.Context) error = func(context.Context) error { return nil }
	managerErr  error
)

// Manager handles enterprise license validation and periodic revalidation.
type Manager struct {
	logger          *zap.Logger
	accountID       string
	productID       string
	licenseKey      string
	fingerprint     string
	verifyKey       string // base64 encoded for validation
	license         *keygen.License
	isEnterprise    bool
	mu              sync.RWMutex
	done            chan struct{}
	cancel          context.CancelFunc
	forceEnterprise bool
}

const (
	revalidateInterval = 12 * time.Hour
)

type LicenseManagerOption func(*Manager)

func WithForceEnterprise() LicenseManagerOption {
	return func(lm *Manager) {
		lm.forceEnterprise = true
		lm.isEnterprise = true
	}
}

func WithVerificationKey(verifyKey string) LicenseManagerOption {
	return func(lm *Manager) {
		lm.verifyKey = verifyKey
	}
}

// NewManager creates a new Manager and starts periodic revalidation.
func NewManager(ctx context.Context, logger *zap.Logger, accountID, productID, licenseKey string, opts ...LicenseManagerOption) (*Manager, func(context.Context) error, error) {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		lm := &Manager{
			logger:          logger,
			accountID:       accountID,
			productID:       productID,
			licenseKey:      licenseKey,
			cancel:          cancel,
			forceEnterprise: false,
			done:            make(chan struct{}),
		}

		logger.Debug("creating license manager")

		for _, opt := range opts {
			opt(lm)
		}

		if lm.forceEnterprise {
			lm.logger.Warn("enterprise features are enabled for development purposes only. It is in violation of the Flipt Fair Core License (FCL) if you are using this in production.", zap.String("url", "https://github.com/flipt-io/flipt/blob/v2/LICENSE"))
		} else {
			c := retryablehttp.NewClient()
			c.Backoff = retryablehttp.LinearJitterBackoff
			c.RetryMax = 5

			keygen.HTTPClient = c.StandardClient()
			keygen.Account = lm.accountID
			keygen.Product = lm.productID
			keygen.LicenseKey = lm.licenseKey
			keygen.Logger = keygen.NewNilLogger()

			if lm.verifyKey != "" {
				keygen.PublicKey = lm.verifyKey
			}

			fingerprint, err := machineid.ProtectedID(keygen.Product)
			if err != nil {
				managerErr = fmt.Errorf("failed to get machine fingerprint; enterprise features are disabled.: %w", err)
				return
			}

			lm.fingerprint = fingerprint
			lm.validateAndSet(ctx)
			go lm.periodicRevalidate(ctx)
		}

		manager = lm
		managerFunc = func(ctx context.Context) error {
			return lm.Shutdown(ctx)
		}
	})

	return manager, managerFunc, managerErr
}

// IsEnterprise returns true if the license is valid for enterprise features.
func (lm *Manager) IsEnterprise() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.isEnterprise
}

// Close stops the background revalidation goroutine.
func (lm *Manager) Shutdown(ctx context.Context) error {
	lm.cancel()
	<-lm.done
	if lm.license != nil {
		// deactivate the license for this machine so it can be used on another machine
		if err := lm.license.Deactivate(ctx, lm.fingerprint); err != nil {
			lm.logger.Warn("failed to deactivate license", zap.Error(err))
		}
	}
	return nil
}

func (lm *Manager) periodicRevalidate(ctx context.Context) {
	ticker := time.NewTicker(revalidateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lm.validateAndSet(ctx)
		case <-ctx.Done():
			close(lm.done)
			return
		}
	}
}

func (lm *Manager) validateAndSet(ctx context.Context) {
	if lm.licenseKey == "" {
		lm.mu.Lock()
		lm.isEnterprise = false
		lm.mu.Unlock()
		lm.logger.Warn("no enterprise license key provided; enterprise features are disabled.")
		return
	}

	license, err := keygen.Validate(ctx, lm.fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, keygen.ErrLicenseNotActivated):
			// Activate the current fingerprint
			_, err := license.Activate(ctx, lm.fingerprint)
			if err != nil {
				if errors.Is(err, keygen.ErrMachineLimitExceeded) {
					lm.logger.Warn("machine limit has been exceeded; enterprise features are disabled.")
					return
				}
				lm.logger.Warn("machine activation failed; enterprise features are disabled.")
				return
			}
		case errors.Is(err, keygen.ErrLicenseExpired):
			lm.logger.Warn("license is expired; enterprise features are disabled.")
			return
		default:
			lm.logger.Warn("license is invalid; enterprise features are disabled.", zap.Error(err))
			return
		}
	}

	if license == nil {
		lm.logger.Error("license is nil; enterprise features are disabled.")
		return
	}

	if license.Expiry == nil {
		lm.logger.Error("license has no expiry date; enterprise features are disabled.")
		return
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.isEnterprise = true
	lm.license = license

	lm.logger.Info("enterprise license validated; enterprise features enabled.",
		zap.Time("expires_at", *license.Expiry))

}
