package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/oci"
)

type bundleCommand struct{}

func newBundleCommand() *cobra.Command {
	bundle := &bundleCommand{}

	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Manage Flipt bundles",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Build a bundle",
		RunE:  bundle.build,
		Args:  cobra.ExactArgs(1),
	})

	return cmd
}

func (c *bundleCommand) build(cmd *cobra.Command, args []string) error {
	logger, cfg, err := buildConfig()
	if err != nil {
		return err
	}

	var opts []containers.Option[oci.StoreOptions]
	if cfg := cfg.Storage.OCI; cfg != nil {
		if cfg.BundleDirectory != "" {
			opts = append(opts, oci.WithBundleDir(cfg.BundleDirectory))
		}

		if cfg.Authentication != nil {
			opts = append(opts, oci.WithCredentials(
				cfg.Authentication.Username,
				cfg.Authentication.Password,
			))
		}
	}

	store, err := oci.NewStore(logger, fmt.Sprintf("flipt://local/%s", args[0]), opts...)
	if err != nil {
		return err
	}

	bundle, err := store.Build(cmd.Context(), os.DirFS("."))
	if err != nil {
		return err
	}

	fmt.Println(bundle.Digest)

	return nil
}
