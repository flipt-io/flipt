package license

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/keygen-sh/keygen-go/v3"
	"github.com/keygen-sh/machineid"
	"go.flipt.io/flipt/internal/product"
	"go.uber.org/zap"
)

var (
	managerOnce sync.Once
	manager     *Manager
	managerFunc func(context.Context) error = func(context.Context) error { return nil }
)

type fingerprintFunc func(string) (string, error)

// License represents a license that can be activated, deactivated, and has an expiry date.
type License interface {
	Activate(ctx context.Context, fingerprint string) (License, error)
	Deactivate(ctx context.Context, fingerprint string) error
	GetExpiry() *time.Time
}

// keygenLicenseWrapper wraps the keygen.License to implement our License interface.
type keygenLicenseWrapper struct {
	*keygen.License
}

func (w *keygenLicenseWrapper) Activate(ctx context.Context, fingerprint string) (License, error) {
	_, err := w.License.Activate(ctx, fingerprint)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *keygenLicenseWrapper) Deactivate(ctx context.Context, fingerprint string) error {
	return w.License.Deactivate(ctx, fingerprint)
}

func (w *keygenLicenseWrapper) GetExpiry() *time.Time {
	return w.License.Expiry
}

type validator func(ctx context.Context, fingerprints ...string) (License, error)

// keygenValidator wraps the keygen.Validate function to return our License interface.
func keygenValidator(ctx context.Context, fingerprints ...string) (License, error) {
	license, err := keygen.Validate(ctx, fingerprints...)
	if err != nil {
		return nil, err
	}
	if license == nil {
		return nil, nil
	}
	return &keygenLicenseWrapper{License: license}, nil
}

var licenseValidator validator = keygenValidator

// Manager handles commercial license validation and periodic revalidation.
type Manager struct {
	logger        *zap.Logger
	accountID     string
	productID     string
	licenseKey    string
	fingerprinter fingerprintFunc
	verifyKey     string // base64 encoded for validation
	license       License
	product       product.Product
	mu            sync.RWMutex
	done          chan struct{}
	doneOnce      sync.Once
	cancel        context.CancelFunc
	force         bool
}

const revalidateInterval = 12 * time.Hour

type LicenseManagerOption func(*Manager)

func WithProduct(product product.Product) LicenseManagerOption {
	return func(lm *Manager) {
		lm.force = true
		lm.product = product
	}
}

func WithVerificationKey(verifyKey string) LicenseManagerOption {
	return func(lm *Manager) {
		lm.verifyKey = verifyKey
	}
}

// NewManager creates a new Manager and starts periodic revalidation.
func NewManager(ctx context.Context, logger *zap.Logger, accountID, productID, licenseKey string, opts ...LicenseManagerOption) (*Manager, func(context.Context) error) {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		lm := &Manager{
			logger:        logger,
			accountID:     accountID,
			productID:     productID,
			licenseKey:    licenseKey,
			fingerprinter: machineid.ProtectedID,
			cancel:        cancel,
			force:         false,
			done:          make(chan struct{}),
		}

		logger.Debug("creating license manager")

		for _, opt := range opts {
			opt(lm)
		}

		if lm.force {
			lm.logger.Warn(string(lm.product)+" features are enabled for Flipt development purposes only. It is in violation of the Flipt Fair Core License (FCL) if you are using this software in any other context.", zap.String("url", "https://github.com/flipt-io/flipt/blob/v2/LICENSE"))
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

			lm.validateAndSet(ctx)
			go lm.periodicRevalidate(ctx)
		}

		manager = lm
		managerFunc = func(ctx context.Context) error {
			return lm.Shutdown(ctx)
		}
	})

	return manager, managerFunc
}

// Product returns the product that the license is valid for.
func (lm *Manager) Product() product.Product {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.product
}

// Close stops the background revalidation goroutine.
func (lm *Manager) Shutdown(ctx context.Context) error {
	lm.cancel()
	lm.doneOnce.Do(func() { close(lm.done) })
	<-lm.done
	if lm.license != nil {
		fingerprint, err := lm.fingerprinter(lm.productID)
		if err != nil {
			lm.logger.Warn("failed to get machine fingerprint for license deactivation.", zap.Error(err))
			return err
		}
		// deactivate the license for this machine so it can be used on another machine
		if err := lm.license.Deactivate(ctx, fingerprint); err != nil {
			lm.logger.Warn("failed to deactivate license", zap.Error(err))
		}
	}
	return nil
}

func (lm *Manager) periodicRevalidate(ctx context.Context) {
	for {
		select {
		case <-time.After(revalidateInterval):
			lm.validateAndSet(ctx)
		case <-ctx.Done():
			lm.doneOnce.Do(func() { close(lm.done) })
			return
		}
	}
}

func (lm *Manager) validateAndSet(ctx context.Context) {
	if lm.licenseKey == "" {
		lm.mu.Lock()
		lm.product = product.OSS
		lm.mu.Unlock()
		lm.logger.Warn("no license key provided; additional features are disabled.")
		return
	}

	fingerprint, err := lm.fingerprinter(lm.productID)
	if err != nil {
		lm.logger.Warn("failed to get machine fingerprint; additional features are disabled.", zap.Error(err))
		return
	}

	license, err := licenseValidator(ctx, fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, keygen.ErrLicenseNotActivated):
			// Activate the current fingerprint
			_, err := license.Activate(ctx, fingerprint)
			if err != nil {
				lm.logger.Warn("failed to activate license; additional features are disabled.", zap.Error(err))
				return
			}
		case errors.Is(err, keygen.ErrLicenseExpired):
			lm.logger.Warn("license is expired; additional features are disabled.")
			return
		default:
			lm.logger.Warn("license is invalid; additional features are disabled.", zap.Error(err))
			return
		}
	}

	if license == nil {
		lm.logger.Error("license is nil; additional features are disabled.")
		return
	}

	if license.GetExpiry() == nil {
		lm.logger.Error("license has no expiry date; additional features are disabled.")
		return
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.product = product.Pro
	lm.license = license

	lm.logger.Info("license validated; additional features enabled.",
		zap.Time("expires_at", *license.GetExpiry()))

}
