package store

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/oci"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/internal/storage/fs/git"
	"go.flipt.io/flipt/internal/storage/fs/local"
	"go.flipt.io/flipt/internal/storage/fs/object/azblob"
	"go.flipt.io/flipt/internal/storage/fs/object/s3"
	storageoci "go.flipt.io/flipt/internal/storage/fs/oci"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// NewStore is a constructor that handles all the known declarative backend storage types
// Given the provided storage type is know, the relevant backend is configured and returned
func NewStore(ctx context.Context, logger *zap.Logger, cfg *config.Config) (_ storage.Store, err error) {
	switch cfg.Storage.Type {
	case config.GitStorageType:
		opts := []containers.Option[git.SnapshotStore]{
			git.WithRef(cfg.Storage.Git.Ref),
			git.WithPollOptions(
				storagefs.WithInterval(cfg.Storage.Git.PollInterval),
			),
			git.WithInsecureTLS(cfg.Storage.Git.InsecureSkipTLS),
		}

		if cfg.Storage.Git.CaCertBytes != "" {
			opts = append(opts, git.WithCABundle([]byte(cfg.Storage.Git.CaCertBytes)))
		} else if cfg.Storage.Git.CaCertPath != "" {
			if bytes, err := os.ReadFile(cfg.Storage.Git.CaCertPath); err == nil {
				opts = append(opts, git.WithCABundle(bytes))
			} else {
				return nil, err
			}
		}

		auth := cfg.Storage.Git.Authentication
		switch {
		case auth.BasicAuth != nil:
			opts = append(opts, git.WithAuth(&http.BasicAuth{
				Username: auth.BasicAuth.Username,
				Password: auth.BasicAuth.Password,
			}))
		case auth.TokenAuth != nil:
			opts = append(opts, git.WithAuth(&http.TokenAuth{
				Token: auth.TokenAuth.AccessToken,
			}))
		case auth.SSHAuth != nil:
			var method *gitssh.PublicKeys
			if auth.SSHAuth.PrivateKeyBytes != "" {
				method, err = gitssh.NewPublicKeys(
					auth.SSHAuth.User,
					[]byte(auth.SSHAuth.PrivateKeyBytes),
					auth.SSHAuth.Password,
				)
			} else {
				method, err = gitssh.NewPublicKeysFromFile(
					auth.SSHAuth.User,
					auth.SSHAuth.PrivateKeyPath,
					auth.SSHAuth.Password,
				)
			}
			if err != nil {
				return nil, err
			}

			// we're protecting against this explicitly so we can disable
			// the gosec linting rule
			if auth.SSHAuth.InsecureIgnoreHostKey {
				// nolint:gosec
				method.HostKeyCallback = ssh.InsecureIgnoreHostKey()
			}

			opts = append(opts, git.WithAuth(method))
		}

		snapStore, err := git.NewSnapshotStore(ctx, logger, cfg.Storage.Git.Repository, opts...)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(snapStore), nil
	case config.LocalStorageType:
		snapStore, err := local.NewSnapshotStore(ctx, logger, cfg.Storage.Local.Path)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(snapStore), nil
	case config.ObjectStorageType:
		return newObjectStore(ctx, cfg, logger)
	case config.OCIStorageType:
		var opts []containers.Option[oci.StoreOptions]
		if auth := cfg.Storage.OCI.Authentication; auth != nil {
			opts = append(opts, oci.WithCredentials(
				auth.Username,
				auth.Password,
			))
		}

		ocistore, err := oci.NewStore(logger, cfg.Storage.OCI.BundlesDirectory, opts...)
		if err != nil {
			return nil, err
		}

		ref, err := oci.ParseReference(cfg.Storage.OCI.Repository)
		if err != nil {
			return nil, err
		}

		snapStore, err := storageoci.NewSnapshotStore(ctx, logger, ocistore, ref,
			storageoci.WithPollOptions(
				storagefs.WithInterval(cfg.Storage.OCI.PollInterval),
			),
		)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(snapStore), nil
	}

	return nil, fmt.Errorf("unexpected storage type: %q", cfg.Storage.Type)
}

// newObjectStore create a new storate.Store from the object config
func newObjectStore(ctx context.Context, cfg *config.Config, logger *zap.Logger) (store storage.Store, err error) {
	objectCfg := cfg.Storage.Object
	// keep this as a case statement in anticipation of
	// more object types in the future
	// nolint:gocritic
	switch objectCfg.Type {
	case config.S3ObjectSubStorageType:
		opts := []containers.Option[s3.SnapshotStore]{
			s3.WithPollOptions(
				storagefs.WithInterval(objectCfg.S3.PollInterval),
			),
		}
		if objectCfg.S3.Endpoint != "" {
			opts = append(opts, s3.WithEndpoint(objectCfg.S3.Endpoint))
		}
		if objectCfg.S3.Region != "" {
			opts = append(opts, s3.WithRegion(objectCfg.S3.Region))
		}

		snapStore, err := s3.NewSnapshotStore(ctx, logger, objectCfg.S3.Bucket, opts...)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(snapStore), nil
	case config.AZBlobObjectSubStorageType:
		opts := []containers.Option[azblob.SnapshotStore]{
			azblob.WithEndpoint(objectCfg.AZBlob.Endpoint),
			azblob.WithPollOptions(
				storagefs.WithInterval(objectCfg.AZBlob.PollInterval),
			),
		}

		snapStore, err := azblob.NewSnapshotStore(ctx, logger, objectCfg.AZBlob.Container, opts...)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(snapStore), nil
	}

	return nil, fmt.Errorf("unexpected object storage subtype: %q", objectCfg.Type)
}
