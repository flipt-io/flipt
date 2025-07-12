package environments

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/coss/license"
	cosssigning "go.flipt.io/flipt/internal/coss/signing"
	cossgit "go.flipt.io/flipt/internal/coss/storage/environments/git"
	"go.flipt.io/flipt/internal/coss/storage/environments/git/gitea"
	"go.flipt.io/flipt/internal/coss/storage/environments/git/github"
	"go.flipt.io/flipt/internal/coss/storage/environments/git/gitlab"
	"go.flipt.io/flipt/internal/credentials"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/secrets"
	serverconfig "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage/environments/evaluation"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	configcoreflipt "go.flipt.io/flipt/internal/storage/environments/fs/flipt"
	environmentsgit "go.flipt.io/flipt/internal/storage/environments/git"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	"go.uber.org/zap"
)

// RepositoryManager manages git repositories.
type RepositoryManager struct {
	logger         *zap.Logger
	cfg            *config.Config
	repos          map[string]*storagegit.Repository
	secretsManager secrets.Manager
	licenseManager license.Manager
}

func NewRepositoryManager(logger *zap.Logger, cfg *config.Config, secretsManager secrets.Manager, licenseManager license.Manager) *RepositoryManager {
	return &RepositoryManager{
		logger:         logger,
		cfg:            cfg,
		repos:          map[string]*storagegit.Repository{},
		secretsManager: secretsManager,
		licenseManager: licenseManager,
	}
}

func (rm *RepositoryManager) GetOrCreate(ctx context.Context, envConf *config.EnvironmentConfig, storage *config.StorageConfig, credentials *credentials.CredentialSource) (*storagegit.Repository, error) {
	var (
		logger = rm.logger
		_, ok  = rm.repos[envConf.Storage]
	)

	if !ok {
		// fallback to using first environments branch as the default
		var (
			defaultBranch = storage.Branch
			opts          = []containers.Option[storagegit.Repository]{
				storagegit.WithInsecureTLS(storage.InsecureSkipTLS),
				storagegit.WithSignature(storage.Signature.Name, storage.Signature.Email),
				storagegit.WithDefaultBranch(defaultBranch),
			}
		)

		if storage.Remote != "" {
			opts = append(opts, storagegit.WithRemote("origin", storage.Remote))
		}

		pollInterval := storage.PollInterval
		if pollInterval == 0 {
			pollInterval = 30 * time.Second
		}

		opts = append(opts, storagegit.WithInterval(pollInterval))

		switch storage.Backend.Type {
		case config.LocalStorageBackendType:
			path := storage.Backend.Path
			if path == "" {
				var err error
				path, err = os.MkdirTemp(os.TempDir(), "flipt-git-*")
				if err != nil {
					return nil, fmt.Errorf("making tempory directory for git storage: %w", err)
				}
			}
			opts = append(opts, storagegit.WithFilesystemStorage(path))
			logger = logger.With(zap.String("git_storage_type", "filesystem"), zap.String("git_storage_path", path))

		case config.MemoryStorageBackendType:
			logger = logger.With(zap.String("git_storage_type", "memory"))
		}

		if storage.CaCertBytes != "" {
			opts = append(opts, storagegit.WithCABundle([]byte(storage.CaCertBytes)))
		} else if storage.CaCertPath != "" {
			if bytes, err := os.ReadFile(storage.CaCertPath); err == nil {
				opts = append(opts, storagegit.WithCABundle(bytes))
			} else {
				return nil, err
			}
		}

		if storage.Credentials != "" {
			creds, err := credentials.Get(storage.Credentials)
			if err != nil {
				return nil, err
			}
			auth, err := creds.GitAuthentication()
			if err != nil {
				return nil, err
			}
			opts = append(opts, storagegit.WithAuth(auth))
		}

		// Configure commit signing if enabled
		if storage.Signature.Enabled {
			// Check license for commit signing (COSS feature)
			if rm.licenseManager.Product() == product.OSS {
				logger.Warn("commit signing requires a paid license")
			} else {

				if storage.Signature.KeyRef == nil {
					return nil, errors.New("signing key reference is required when commit signing is enabled")
				}

				// Create GPG signer
				signer, err := cosssigning.NewGPGSigner(
					*storage.Signature.KeyRef,
					storage.Signature.KeyID,
					rm.secretsManager,
					logger.With(zap.String("component", "gpg-signer")),
				)
				if err != nil {
					return nil, fmt.Errorf("creating GPG signer: %w", err)
				}

				opts = append(opts, storagegit.WithSigner(signer))
				logger.Info("commit signing enabled",
					zap.String("provider", storage.Signature.KeyRef.Provider),
					zap.String("key_id", storage.Signature.KeyID))
			}
		}

		newRepo, err := storagegit.NewRepository(ctx, logger, opts...)
		if err != nil {
			return nil, err
		}

		rm.repos[envConf.Storage] = newRepo
	}

	return rm.repos[envConf.Storage], nil
}

// EnvironmentFactory creates environments and wraps them as needed.
type EnvironmentFactory struct {
	logger         *zap.Logger
	cfg            *config.Config
	credentials    *credentials.CredentialSource
	repoManager    *RepositoryManager
	licenseManager interface{ Product() product.Product } // Accepts LicenseManager or a mock for tests
}

func NewEnvironmentFactory(logger *zap.Logger, cfg *config.Config, credentials *credentials.CredentialSource, repoManager *RepositoryManager, licenseManager license.Manager) *EnvironmentFactory {
	return &EnvironmentFactory{
		logger:         logger,
		cfg:            cfg,
		credentials:    credentials,
		repoManager:    repoManager,
		licenseManager: licenseManager,
	}
}

func (f *EnvironmentFactory) Create(ctx context.Context, name string, envConf *config.EnvironmentConfig) (serverconfig.Environment, error) {
	var (
		srcs        = f.cfg.Storage
		storage, ok = srcs[envConf.Storage]
	)
	if !ok {
		return nil, fmt.Errorf("missing storage for name %q", envConf.Storage)
	}

	fileStorage := fs.NewStorage(
		f.logger,
		configcoreflipt.NewFlagStorage(f.logger),
		configcoreflipt.NewSegmentStorage(f.logger),
	)

	repo, err := f.repoManager.GetOrCreate(ctx, envConf, storage, f.credentials)
	if err != nil {
		return nil, err
	}

	// build new git backend environment over repository
	env, err := environmentsgit.NewEnvironmentFromRepo(
		ctx,
		f.logger,
		envConf,
		repo,
		fileStorage,
		evaluation.NewSnapshotPublisher(f.logger),
	)
	if err != nil {
		return nil, err
	}

	// wrap the environment with an SCM if configured
	if envConf.SCM != nil {
		// License check: only allow SCM if pro is enabled
		if f.licenseManager == nil || f.licenseManager.Product() != product.Pro {
			f.logger.Warn("paid license required for SCM integration; using noop SCM.", zap.String("environment", envConf.Name))
			wrapped := cossgit.NewEnvironment(f.logger, env, &cossgit.SCMNotImplemented{})
			return wrapped, nil
		}

		var scm cossgit.SCM = &cossgit.SCMNotImplemented{}

		repoURL, err := cossgit.ParseGitURL(repo.GetRemote())
		if err != nil {
			return nil, fmt.Errorf("failed to parse git url: %w", err)
		}

		var (
			repoOwner = repoURL.Owner
			repoName  = repoURL.Repo
		)

		//nolint
		switch envConf.SCM.Type {
		case config.GitHubSCMType:
			opts := []github.ClientOption{}

			// To support GitHub Enterprise
			if envConf.SCM.ApiURL != "" {
				apiURL, err := url.Parse(envConf.SCM.ApiURL)
				if err != nil {
					return nil, fmt.Errorf("failed to parse api url: %w", err)
				}
				opts = append(opts, github.WithApiURL(apiURL))
			}

			if envConf.SCM.Credentials != nil {
				creds, err := f.credentials.Get(*envConf.SCM.Credentials)
				if err != nil {
					return nil, err
				}

				opts = append(opts, github.WithApiAuth(creds.APIAuthentication()))
			}

			scm, err = github.NewSCM(f.logger, repoOwner, repoName, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to setup github scm: %w", err)
			}
		case config.GitLabSCMType:
			opts := []gitlab.ClientOption{}

			// To support GitLab Enterprise
			if envConf.SCM.ApiURL != "" {
				apiURL, err := url.Parse(envConf.SCM.ApiURL)
				if err != nil {
					return nil, fmt.Errorf("failed to parse api url: %w", err)
				}
				opts = append(opts, gitlab.WithApiURL(apiURL))
			}

			if envConf.SCM.Credentials != nil {
				creds, err := f.credentials.Get(*envConf.SCM.Credentials)
				if err != nil {
					return nil, err
				}

				opts = append(opts, gitlab.WithApiAuth(creds.APIAuthentication()))
			}

			scm, err = gitlab.NewSCM(f.logger, repoOwner, repoName, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to setup gitlab scm: %w", err)
			}
		case config.GiteaSCMType:
			opts := []gitea.ClientOption{}

			if envConf.SCM.Credentials != nil {
				creds, err := f.credentials.Get(*envConf.SCM.Credentials)
				if err != nil {
					return nil, err
				}

				opts = append(opts, gitea.WithApiAuth(creds.APIAuthentication()))
			}

			scm, err = gitea.NewSCM(f.logger, envConf.SCM.ApiURL, repoOwner, repoName, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to setup gitea scm: %w", err)
			}
		}

		wrapped := cossgit.NewEnvironment(f.logger, env, scm)
		return wrapped, nil
	}

	f.logger.Debug("configured environment",
		zap.String("environment", envConf.Name),
		zap.String("storage", envConf.Storage))

	return env, nil
}

// environmentSubscriber is a subscriber for storagegit.Repository
type environmentSubscriber struct {
	branchesFn func() []string
	notifyFn   func(ctx context.Context, refs map[string]string) error
}

func (s *environmentSubscriber) Branches() []string {
	return s.branchesFn()
}

func (s *environmentSubscriber) Notify(ctx context.Context, refs map[string]string) error {
	return s.notifyFn(ctx, refs)
}

type repoEnv interface {
	Repository() *storagegit.Repository
	RefreshEnvironment(context.Context, map[string]string) ([]serverconfig.Environment, error)
	Branches() []string
}

// NewStore is a constructor that handles all the environment storage types
// Given the provided storage type is know, the relevant backend is configured and returned
func NewStore(ctx context.Context, logger *zap.Logger, cfg *config.Config, secretsManager secrets.Manager, licenseManager license.Manager) (
	_ *serverconfig.EnvironmentStore,
	err error,
) {
	if len(cfg.Environments) == 0 {
		return nil, errors.New("no environments configured")
	}

	var (
		credentials = credentials.New(logger, cfg.Credentials)
		repoManager = NewRepositoryManager(logger, cfg, secretsManager, licenseManager)
		factory     = NewEnvironmentFactory(logger, cfg, credentials, repoManager, licenseManager)
	)

	var envs []serverconfig.Environment
	for name, envConf := range cfg.Environments {
		env, err := factory.Create(ctx, name, envConf)
		if err != nil {
			return nil, fmt.Errorf("environment %q: %w", name, err)
		}
		envs = append(envs, env)
	}

	envStore, err := serverconfig.NewEnvironmentStore(logger, envs...)
	if err != nil {
		return nil, err
	}

	// subscribe all git environments to their repositories using a custom subscriber
	for _, env := range envs {
		gitEnv, ok := env.(repoEnv)
		if !ok {
			continue
		}

		var (
			repo = gitEnv.Repository()
			sub  = &environmentSubscriber{
				branchesFn: func() []string {
					return gitEnv.Branches()
				},
				notifyFn: func(ctx context.Context, refs map[string]string) error {
					newEnvs, err := gitEnv.RefreshEnvironment(ctx, refs)
					if err != nil {
						return err
					}
					for _, e := range newEnvs {
						envStore.Add(e)
					}
					return nil
				},
			}
		)

		repo.Subscribe(sub)
	}

	// perform an extra fetch before proceeding to ensure any
	// branched environments have been added
	for _, repo := range repoManager.repos {
		if err := repo.Fetch(ctx); err != nil {
			if !errors.Is(err, transport.ErrEmptyRemoteRepository) &&
				!errors.Is(err, git.NoMatchingRefSpecError{}) {
				return nil, err
			}
		}
	}

	return envStore, nil
}
