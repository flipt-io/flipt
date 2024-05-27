package store

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"oras.land/oras-go/v2"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/oci"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.flipt.io/flipt/internal/storage/fs/git"
	"go.flipt.io/flipt/internal/storage/fs/local"
	"go.flipt.io/flipt/internal/storage/fs/object"
	storageoci "go.flipt.io/flipt/internal/storage/fs/oci"
	"go.uber.org/zap"
	"gocloud.dev/blob/azureblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"golang.org/x/crypto/ssh"
)

// NewStore is a constructor that handles all the known declarative backend storage types
// Given the provided storage type is know, the relevant backend is configured and returned
func NewStore(ctx context.Context, logger *zap.Logger, cfg *config.Config) (_ storage.Store, err error) {
	switch cfg.Storage.Type {
	case config.GitStorageType:
		storage := cfg.Storage.Git
		opts := []containers.Option[git.SnapshotStore]{
			git.WithRef(cfg.Storage.Git.Ref),
			git.WithPollOptions(
				storagefs.WithInterval(cfg.Storage.Git.PollInterval),
			),
			git.WithInsecureTLS(cfg.Storage.Git.InsecureSkipTLS),
			git.WithDirectory(cfg.Storage.Git.Directory),
		}

		if storage.RefType == config.GitRefTypeSemver {
			opts = append(opts, git.WithSemverResolver())
		}

		switch storage.Backend.Type {
		case config.GitBackendLocal:
			path := storage.Backend.Path
			if path == "" {
				path, err = os.MkdirTemp(os.TempDir(), "flipt-git-*")
				if err != nil {
					return nil, fmt.Errorf("making tempory directory for git storage: %w", err)
				}
			}

			opts = append(opts, git.WithFilesystemStorage(path))
			logger = logger.With(zap.String("git_storage_type", "filesystem"), zap.String("git_storage_path", path))
		case config.GitBackendMemory:
			logger = logger.With(zap.String("git_storage_type", "memory"))
		}

		if storage.CaCertBytes != "" {
			opts = append(opts, git.WithCABundle([]byte(storage.CaCertBytes)))
		} else if storage.CaCertPath != "" {
			if bytes, err := os.ReadFile(storage.CaCertPath); err == nil {
				opts = append(opts, git.WithCABundle(bytes))
			} else {
				return nil, err
			}
		}

		auth := storage.Authentication
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

		snapStore, err := git.NewSnapshotStore(ctx, logger, storage.Repository, opts...)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(snapStore), nil
	case config.LocalStorageType:
		snapStore, err := local.NewSnapshotStore(ctx, logger, cfg.Storage.Local.Path)
		if err != nil {
			return nil, err
		}

		return storagefs.NewStore(storagefs.NewSingleReferenceStore(logger, snapStore)), nil
	case config.ObjectStorageType:
		return newObjectStore(ctx, cfg, logger)
	case config.OCIStorageType:
		var opts []containers.Option[oci.StoreOptions]
		if auth := cfg.Storage.OCI.Authentication; auth != nil {
			opt, err := oci.WithCredentials(
				auth.Type,
				auth.Username,
				auth.Password,
			)
			if err != nil {
				return nil, err
			}
			opts = append(opts, opt)
		}

		// The default is the 1.1 version, this is why we don't need to check it in here.
		if cfg.Storage.OCI.ManifestVersion == config.OCIManifestVersion10 {
			opts = append(opts, oci.WithManifestVersion(oras.PackManifestVersion1_0))
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

		return storagefs.NewStore(storagefs.NewSingleReferenceStore(logger, snapStore)), nil
	}

	return nil, fmt.Errorf("unexpected storage type: %q", cfg.Storage.Type)
}

// newObjectStore create a new storate.Store from the object config
func newObjectStore(ctx context.Context, cfg *config.Config, logger *zap.Logger) (store storage.Store, err error) {
	var (
		ocfg       = cfg.Storage.Object
		opts       []containers.Option[object.SnapshotStore]
		scheme     string
		bucketName string
		values     = url.Values{}
	)
	// keep this as a case statement in anticipation of
	// more object types in the future
	// nolint:gocritic
	switch ocfg.Type {
	case config.S3ObjectSubStorageType:
		scheme = s3blob.Scheme
		bucketName = ocfg.S3.Bucket
		if ocfg.S3.Endpoint != "" {
			values.Set("endpoint", ocfg.S3.Endpoint)
		}

		if ocfg.S3.Region != "" {
			values.Set("region", ocfg.S3.Region)
		}

		if ocfg.S3.Prefix != "" {
			values.Set("prefix", ocfg.S3.Prefix)
		}

		opts = append(opts,
			object.WithPollOptions(
				storagefs.WithInterval(ocfg.S3.PollInterval),
			),
			object.WithPrefix(ocfg.S3.Prefix),
		)
	case config.AZBlobObjectSubStorageType:
		scheme = azureblob.Scheme
		bucketName = ocfg.AZBlob.Container

		url, err := url.Parse(ocfg.AZBlob.Endpoint)
		if err != nil {
			return nil, err
		}

		os.Setenv("AZURE_STORAGE_PROTOCOL", url.Scheme)
		os.Setenv("AZURE_STORAGE_IS_LOCAL_EMULATOR", strconv.FormatBool(url.Scheme == "http"))
		os.Setenv("AZURE_STORAGE_DOMAIN", url.Host)

		opts = append(opts,
			object.WithPollOptions(
				storagefs.WithInterval(ocfg.AZBlob.PollInterval),
			),
		)
	case config.GSBlobObjectSubStorageType:
		scheme = gcsblob.Scheme
		bucketName = ocfg.GS.Bucket
		if ocfg.GS.Prefix != "" {
			values.Set("prefix", ocfg.GS.Prefix)
		}

		opts = append(opts,
			object.WithPollOptions(
				storagefs.WithInterval(ocfg.GS.PollInterval),
			),
			object.WithPrefix(ocfg.GS.Prefix),
		)
	default:
		return nil, fmt.Errorf("unexpected object storage subtype: %q", ocfg.Type)
	}

	u := &url.URL{
		Scheme:   scheme,
		Host:     bucketName,
		RawQuery: values.Encode(),
	}

	bucket, err := object.OpenBucket(ctx, u)
	if err != nil {
		return nil, err
	}

	snap, err := object.NewSnapshotStore(ctx, logger, scheme, bucket, opts...)
	if err != nil {
		return nil, err
	}

	return storagefs.NewStore(storagefs.NewSingleReferenceStore(logger, snap)), nil
}
