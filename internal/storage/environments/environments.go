package environments

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/credentials"
	serverconfig "go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage/environments/evaluation"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	configcoreflipt "go.flipt.io/flipt/internal/storage/environments/fs/flipt"
	environmentsgit "go.flipt.io/flipt/internal/storage/environments/git"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	"go.uber.org/zap"
)

// NewStore is a constructor that handles all the known declarative backend storage types
// Given the provided storage type is know, the relevant backend is configured and returned
func NewStore(ctx context.Context, logger *zap.Logger, cfg *config.Config) (
	_ *serverconfig.EnvironmentStore,
	err error,
) {
	if len(cfg.Environments) == 0 {
		return nil, errors.New("no environments configured")
	}

	var (
		builder = sourceBuilder{
			logger: logger,
			cfg:    cfg,
			repos:  map[string]*storagegit.Repository{},
		}
		envs        []serverconfig.Environment
		credentials = credentials.New(logger, cfg.Credentials)
	)

	for name, envConf := range cfg.Environments {
		env, err := builder.forEnvironment(ctx, logger, envConf, credentials)
		if err != nil {
			return nil, fmt.Errorf("environment %q: %w", name, err)
		}

		envs = append(envs, env)
	}

	// perform an extra fetch before proceeding to ensure any
	// branched environments have been added
	for _, repo := range builder.repos {
		if err := repo.Fetch(ctx); err != nil {
			if !errors.Is(err, transport.ErrEmptyRemoteRepository) &&
				!errors.Is(err, git.NoMatchingRefSpecError{}) {
				return nil, err
			}
		}
	}

	envStore, err := serverconfig.NewEnvironmentStore(logger, envs...)
	if err != nil {
		return nil, err
	}

	// subscribe to all git environments to be notified of new branches
	for _, env := range envs {
		gitEnv, ok := env.(*environmentsgit.Environment)
		if !ok {
			continue
		}

		gitEnv.Repository().Subscribe(&environmentSubscriber{
			Environment: gitEnv,
			envs:        envStore,
		})
	}

	return envStore, nil
}

type sourceBuilder struct {
	logger *zap.Logger
	cfg    *config.Config
	repos  map[string]*storagegit.Repository
}

func (s *sourceBuilder) forEnvironment(
	ctx context.Context,
	logger *zap.Logger,
	envConf *config.EnvironmentConfig,
	credentials *credentials.CredentialSource,
) (serverconfig.Environment, error) {
	var (
		srcs        = s.cfg.Storage
		storage, ok = srcs[envConf.Storage]
	)

	if !ok {
		return nil, fmt.Errorf("missing storage for name %q", envConf.Storage)
	}

	fileStorage := fs.NewStorage(
		logger,
		configcoreflipt.NewFlagStorage(logger),
		configcoreflipt.NewSegmentStorage(logger),
	)

	repo, err := s.getOrCreateGitRepo(ctx, envConf, storage, credentials)
	if err != nil {
		return nil, err
	}

	// build new git backend environment over repository
	env, err := environmentsgit.NewEnvironmentFromRepo(
		ctx,
		logger,
		envConf,
		repo,
		fileStorage,
		evaluation.NewSnapshotPublisher(logger),
	)
	if err != nil {
		return nil, err
	}

	logger.Debug("configured environment",
		zap.String("environment", envConf.Name),
		zap.String("storage", envConf.Storage))

	return env, nil
}

func (s *sourceBuilder) getOrCreateGitRepo(ctx context.Context, envConf *config.EnvironmentConfig, storage *config.StorageConfig, credentials *credentials.CredentialSource) (_ *storagegit.Repository, err error) {
	logger := s.logger

	gitRepo, ok := s.repos[envConf.Storage]
	if !ok {
		// fallback to using first environments branch as the default
		// in the case where it is not configured on the source
		defaultBranch := storage.Branch

		opts := []containers.Option[storagegit.Repository]{
			storagegit.WithInsecureTLS(storage.InsecureSkipTLS),
			storagegit.WithSignature(storage.Signature.Name, storage.Signature.Email),
			storagegit.WithDefaultBranch(defaultBranch),
		}

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

		gitRepo, err = storagegit.NewRepository(ctx, logger, opts...)
		if err != nil {
			return nil, err
		}

		s.repos[envConf.Storage] = gitRepo
	}

	return gitRepo, nil
}

type environmentSubscriber struct {
	*environmentsgit.Environment
	envs *serverconfig.EnvironmentStore
}

func (s *environmentSubscriber) Notify(ctx context.Context, refs map[string]string) error {
	// this returns the set of newly observed environments based on this one to be added to the store
	envs, err := s.RefreshEnvironment(ctx, refs)
	if err != nil {
		return err
	}

	for _, env := range envs {
		// we ignore the error as we're only interested in new environments
		// add only errors when attempting to add an existing env
		_ = s.envs.Add(env)
	}

	return err
}
