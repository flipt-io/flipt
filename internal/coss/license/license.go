package license

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/keygen-sh/keygen-go/v3"
	"github.com/keygen-sh/machineid"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/product"
	"go.uber.org/zap"
)

var (
	managerOnce sync.Once
	manager     *ManagerImpl
	managerFunc func(context.Context) error = func(context.Context) error { return nil }
)

type fingerprintFunc func(string) (string, error)

// validateOnline directly calls the keygen API to validate the license and handles activation
func (lm *ManagerImpl) validateOnline(ctx context.Context) (*keygen.License, error) {
	fingerprint, err := lm.fingerprinter(lm.productID)
	if err != nil {
		return nil, err
	}

	license, err := keygen.Validate(ctx, fingerprint)
	if err != nil {
		// For ErrLicenseNotActivated, we still need to return the license object
		// so it can be activated. The keygen library guarantees license is non-nil
		// when returning ErrLicenseNotActivated.
		if errors.Is(err, keygen.ErrLicenseNotActivated) {
			// Activate the current fingerprint
			if _, err := license.Activate(ctx, fingerprint); err != nil {
				return nil, err
			}
			lm.logger.Debug("license activated successfully", zap.String("fingerprint", fingerprint))
			return license, nil
		}
		return nil, err
	}

	if license == nil {
		return nil, errors.New("license validation returned nil license without error")
	}

	return license, nil
}

// validateOffline validates and decrypts the license file and returns the license object
func (lm *ManagerImpl) validateOffline(_ context.Context) (*keygen.License, error) {
	cert, err := os.ReadFile(lm.config.File)
	if err != nil {
		return nil, err
	}

	license := &keygen.LicenseFile{Certificate: string(cert)}
	if err := license.Verify(); err != nil {
		return nil, err
	}

	// Decrypt and validate
	dataset, err := license.Decrypt(lm.config.Key)
	if err != nil {
		return nil, err
	}

	if dataset == nil {
		return nil, errors.New("license file is invalid")
	}

	return &dataset.License, nil
}

type Manager interface {
	Product() product.Product
	Shutdown(ctx context.Context) error
}

var _ Manager = (*ManagerImpl)(nil)

type LicenseType string

const (
	LicenseTypeOnline  LicenseType = "online"
	LicenseTypeOffline LicenseType = "offline"
)

// ManagerImpl handles commercial license validation and periodic revalidation.
type ManagerImpl struct {
	logger        *zap.Logger
	accountID     string
	productID     string
	licenseType   LicenseType
	config        *config.LicenseConfig
	fingerprinter fingerprintFunc
	verifyKey     string // base64 encoded for validation
	license       *keygen.License
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
func NewManager(ctx context.Context, logger *zap.Logger, accountID, productID string, config *config.LicenseConfig, opts ...LicenseManagerOption) (*ManagerImpl, func(context.Context) error) {
	managerOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		lm := &ManagerImpl{
			logger:        logger,
			accountID:     accountID,
			productID:     productID,
			config:        config,
			licenseType:   LicenseTypeOnline,
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
		c.HTTPClient.Timeout = 10 * time.Second
		c.Backoff = retryablehttp.LinearJitterBackoff
		c.RetryMax = 3
		c.Logger = log.New(io.Discard, "", log.LstdFlags)

		keygen.HTTPClient = c.StandardClient()
		keygen.Account = lm.accountID
		keygen.Product = lm.productID
		keygen.LicenseKey = lm.config.Key
		keygen.Logger = keygen.NewNilLogger()

		if lm.verifyKey != "" {
			keygen.PublicKey = lm.verifyKey
		}

		// If a license file is provided, we need to validate it offline
		if lm.config.File != "" {
			lm.licenseType = LicenseTypeOffline
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
	// wait for existing revalidation goroutine to finish
	<-lm.done

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// If the license is offline, we don't need to deactivate it
	if lm.licenseType == LicenseTypeOffline {
		return nil
	}

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
	if lm.config.Key == "" {
		lm.setOSSProduct()
		lm.logger.Warn("no license key provided; additional features are disabled.")
		return
	}

	// Add random startup delay (0-30s) to prevent thundering herd during mass pod restarts
	// Only delay when we have a license key that will make API calls
	delay := time.Duration(rand.Intn(30)) * time.Second
	lm.logger.Debug("adding startup delay to prevent rate limits", zap.Duration("delay", delay))
	time.Sleep(delay)

	var (
		license *keygen.License
		err     error
	)

	switch lm.licenseType {
	case LicenseTypeOnline:
		license, err = lm.validateOnline(ctx)
		if err != nil {
			lm.setOSSProduct()
			lm.logger.Warn("license is invalid; additional features are disabled.", zap.Error(err))
			return
		}

	case LicenseTypeOffline:
		license, err = lm.validateOffline(ctx)
		if err != nil {
			lm.setOSSProduct()
			lm.logger.Warn("license is invalid; additional features are disabled.", zap.Error(err))
			return
		}
	}

	if license.Expiry == nil {
		lm.logger.Warn("license has no expiry date; additional features are disabled.")
		return
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.product = product.Pro
	lm.license = license

	lm.logger.Info("license validated; additional features enabled.",
		zap.Time("expires", *license.Expiry))
}
