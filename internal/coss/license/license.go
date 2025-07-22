package license

import (
	"context"
	"errors"
	"io"
	"log"
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
	manager     *ManagerImpl
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
	if w.License == nil {
		return nil, errors.New("license is nil")
	}
	_, err := w.License.Activate(ctx, fingerprint)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *keygenLicenseWrapper) Deactivate(ctx context.Context, fingerprint string) error {
	if w.License == nil {
		return errors.New("license is nil")
	}
	return w.License.Deactivate(ctx, fingerprint)
}

func (w *keygenLicenseWrapper) GetExpiry() *time.Time {
	if w.License == nil {
		return nil
	}
	return w.License.Expiry
}

// validateLicense directly calls keygen.Validate and wraps the result
func (lm *ManagerImpl) validateLicense(ctx context.Context, fingerprint string) (License, error) {
	license, err := keygen.Validate(ctx, fingerprint)
	if err != nil {
		// For ErrLicenseNotActivated, we still need to return the license object
		// so it can be activated. The keygen library guarantees license is non-nil
		// when returning ErrLicenseNotActivated.
		if errors.Is(err, keygen.ErrLicenseNotActivated) {
			return &keygenLicenseWrapper{License: license}, err
		}
		return nil, err
	}
	if license == nil {
		return nil, errors.New("license validation returned nil license without error")
	}
	return &keygenLicenseWrapper{License: license}, nil
}

type Manager interface {
	Product() product.Product
	Shutdown(ctx context.Context) error
}

var _ Manager = (*ManagerImpl)(nil)

// ManagerImpl handles commercial license validation and periodic revalidation.
type ManagerImpl struct {
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

type LicenseManagerOption func(*ManagerImpl)

func WithProduct(product product.Product) LicenseManagerOption {
	return func(lm *ManagerImpl) {
		lm.force = true
		lm.product = product
	}
}

func WithVerificationKey(verifyKey string) LicenseManagerOption {
	return func(lm *ManagerImpl) {
		lm.verifyKey = verifyKey
	}
}

// NewManager creates a new Manager and starts periodic revalidation.
func NewManager(ctx context.Context, logger *zap.Logger, accountID, productID, licenseKey string, opts ...LicenseManagerOption) (*ManagerImpl, func(context.Context) error) {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		lm := &ManagerImpl{
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
			manager = lm
			managerFunc = func(ctx context.Context) error { return nil }
			return
		}

		c := retryablehttp.NewClient()
		c.Backoff = retryablehttp.LinearJitterBackoff
		c.RetryMax = 5
		c.Logger = log.New(io.Discard, "", log.LstdFlags)

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
		manager = lm
		managerFunc = func(ctx context.Context) error {
			return lm.Shutdown(ctx)
		}
	})

	return manager, managerFunc
}

// Product returns the product that the license is valid for.
func (lm *ManagerImpl) Product() product.Product {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.product
}

// Close stops the background revalidation goroutine.
func (lm *ManagerImpl) Shutdown(ctx context.Context) error {
	lm.cancel()
	<-lm.done
	lm.mu.Lock()
	defer lm.mu.Unlock()
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

func (lm *ManagerImpl) periodicRevalidate(ctx context.Context) {
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

// setOSSProduct is a helper method to set the product to OSS and avoid code duplication
func (lm *ManagerImpl) setOSSProduct() {
	lm.mu.Lock()
	lm.product = product.OSS
	lm.mu.Unlock()
}

func (lm *ManagerImpl) validateAndSet(ctx context.Context) {
	if lm.licenseKey == "" {
		lm.setOSSProduct()
		lm.logger.Warn("no license key provided; additional features are disabled.")
		return
	}

	fingerprint, err := lm.fingerprinter(lm.productID)
	if err != nil {
		lm.setOSSProduct()
		lm.logger.Warn("failed to get machine fingerprint; additional features are disabled.", zap.Error(err))
		return
	}

	license, err := lm.validateLicense(ctx, fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, keygen.ErrLicenseNotActivated):
			// Activate the current fingerprint
			activatedLicense, activateErr := license.Activate(ctx, fingerprint)
			if activateErr != nil {
				lm.setOSSProduct()
				lm.logger.Warn("failed to activate license; additional features are disabled.", zap.Error(activateErr))
				return
			}
			// Use the activated license directly, no need to re-validate
			license = activatedLicense
			lm.logger.Debug("license activated successfully", zap.String("fingerprint", fingerprint))
		case errors.Is(err, keygen.ErrLicenseExpired):
			lm.setOSSProduct()
			lm.logger.Warn("license is expired; additional features are disabled.")
			return
		default:
			lm.setOSSProduct()
			lm.logger.Warn("license is invalid; additional features are disabled.", zap.Error(err))
			return
		}
	}

	if license == nil {
		lm.setOSSProduct()
		lm.logger.Error("license is nil after validation; additional features are disabled.")
		return
	}

	expiry := license.GetExpiry()
	if expiry == nil {
		lm.setOSSProduct()
		lm.logger.Error("license has no expiry date; additional features are disabled.")
		return
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.product = product.Pro
	lm.license = license

	lm.logger.Info("license validated; additional features enabled.",
		zap.Time("expires_at", *expiry))
}
