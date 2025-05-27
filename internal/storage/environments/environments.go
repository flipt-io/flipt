package environments

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	giturl "github.com/kubescape/go-git-url"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/credentials"
	enterprisegit "go.flipt.io/flipt/internal/enterprise/storage/environments/git"
	"go.flipt.io/flipt/internal/enterprise/storage/environments/git/github"
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
	logger *zap.Logger
	cfg    *config.Config
	repos  map[string]*storagegit.Repository
}

func NewRepositoryManager(logger *zap.Logger, cfg *config.Config) *RepositoryManager {
	return &RepositoryManager{
		logger: logger,
		cfg:    cfg,
		repos:  map[string]*storagegit.Repository{},
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
	logger      *zap.Logger
	cfg         *config.Config
	credentials *credentials.CredentialSource
	repoManager *RepositoryManager
}

func NewEnvironmentFactory(logger *zap.Logger, cfg *config.Config, credentials *credentials.CredentialSource, repoManager *RepositoryManager) *EnvironmentFactory {
	return &EnvironmentFactory{
		logger:      logger,
		cfg:         cfg,
		credentials: credentials,
		repoManager: repoManager,
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
	// TODO: check for enterprise license
	if envConf.SCM != nil {
		var scm enterprisegit.SCM = &enterprisegit.SCMNotImplemented{}

		gitURL, err := giturl.NewGitURL(repo.GetRemote())
		if err != nil {
			return nil, fmt.Errorf("failed to parse git url: %w", err)
		}
		var (
			repoOwner = gitURL.GetOwnerName()
			repoName  = gitURL.GetRepoName()
		)

		switch envConf.SCM.Type {
		case config.GitHubSCMType:
			if envConf.SCM.Credentials != nil {
				creds, err := f.credentials.Get(*envConf.SCM.Credentials)
				if err != nil {
					return nil, err
				}

				client, err := creds.HTTPClient(ctx)
				if err != nil {
					return nil, err
				}

				scm = github.NewSCM(f.logger, repoOwner, repoName, client)
			}
		}

		wrapped := enterprisegit.NewEnvironment(f.logger, env, scm)
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
func NewStore(ctx context.Context, logger *zap.Logger, cfg *config.Config) (
	_ *serverconfig.EnvironmentStore,
	err error,
) {
	if len(cfg.Environments) == 0 {
		return nil, errors.New("no environments configured")
	}

	var (
		credentials = credentials.New(logger, cfg.Credentials)
		repoManager = NewRepositoryManager(logger, cfg)
		factory     = NewEnvironmentFactory(logger, cfg, credentials, repoManager)
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
