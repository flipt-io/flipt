package license

import (
	"context"
	"sync"
	"time"

	"github.com/keygen-sh/keygen-go"
	"go.uber.org/zap"
)

// LicenseManager handles enterprise license validation and periodic revalidation.
type LicenseManager struct {
	logger          *zap.Logger
	accountID       string
	productID       string
	licenseKey      string
	isEnterprise    bool
	mu              sync.RWMutex
	cancel          context.CancelFunc
	forceEnterprise bool
}

const revalidateInterval = 12 * time.Hour

type LicenseManagerOption func(*LicenseManager)

func WithForceEnterprise() LicenseManagerOption {
	return func(lm *LicenseManager) {
		lm.forceEnterprise = true
		lm.isEnterprise = true
	}
}

// NewLicenseManager creates a new LicenseManager and starts periodic revalidation.
func NewLicenseManager(ctx context.Context, logger *zap.Logger, accountID, productID, licenseKey string, opts ...LicenseManagerOption) *LicenseManager {
	ctx, cancel := context.WithCancel(ctx)
	lm := &LicenseManager{
		logger:          logger,
		accountID:       accountID,
		productID:       productID,
		licenseKey:      licenseKey,
		cancel:          cancel,
		forceEnterprise: false,
	}
	for _, opt := range opts {
		opt(lm)
	}

	if lm.forceEnterprise {
		lm.logger.Warn("enterprise features are enabled for development purposes only. It is in violation of the Flipt Fair Core License (FCL) if you are using this in production.", zap.String("url", "https://github.com/flipt-io/flipt/blob/v2/LICENSE"))
	} else {
		lm.validateAndSet()
		go lm.periodicRevalidate(ctx)
	}

	return lm
}

// IsEnterprise returns true if the license is valid for enterprise features.
func (lm *LicenseManager) IsEnterprise() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.isEnterprise
}

// Close stops the background revalidation goroutine.
func (lm *LicenseManager) Close() {
	lm.cancel()
}

func (lm *LicenseManager) periodicRevalidate(ctx context.Context) {
	ticker := time.NewTicker(revalidateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lm.validateAndSet()
		case <-ctx.Done():
			return
		}
	}
}

func (lm *LicenseManager) validateAndSet() {
	if lm.licenseKey == "" {
		lm.setEnterprise(false)
		lm.logger.Warn("No enterprise license key provided; enterprise features are disabled.")
		return
	}

	keygen.Account = lm.accountID
	keygen.Product = lm.productID
	keygen.LicenseKey = lm.licenseKey

	license, err := keygen.Validate()
	if err != nil || license == nil {
		lm.setEnterprise(false)
		lm.logger.Warn("Failed to validate enterprise license key; enterprise features are disabled.", zap.Error(err))
		return
	}

	lm.setEnterprise(true)
	lm.logger.Info("Enterprise license validated; enterprise features enabled.")
}

func (lm *LicenseManager) setEnterprise(val bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.isEnterprise = val
}
