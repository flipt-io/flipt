package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"oras.land/oras-go/v2"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
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
		Use:   "build [flags] <name>",
		Short: "Build a bundle",
		RunE:  bundle.build,
		Args:  cobra.ExactArgs(1),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list [flags]",
		Short: "List all bundles",
		RunE:  bundle.list,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "push [flags] <from> <to>",
		Short: "Push local bundle to remote",
		RunE:  bundle.push,
		Args:  cobra.ExactArgs(2),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pull [flags] <remote>",
		Short: "Pull a remote bundle",
		RunE:  bundle.pull,
		Args:  cobra.ExactArgs(1),
	})

	return cmd
}

func (c *bundleCommand) build(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	store, err := c.getStore(ctx)
	if err != nil {
		return err
	}

	ref, err := oci.ParseReference(args[0])
	if err != nil {
		return err
	}

	bundle, err := store.Build(cmd.Context(), os.DirFS("."), ref)
	if err != nil {
		return err
	}

	fmt.Println(bundle.Digest)

	return nil
}

func (c *bundleCommand) list(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	store, err := c.getStore(ctx)
	if err != nil {
		return err
	}

	bundles, err := store.List(cmd.Context())
	if err != nil {
		return err
	}

	wr := writer()

	fmt.Fprintf(wr, "DIGEST\tREPO\tTAG\tCREATED\t\n")
	for _, bundle := range bundles {
		fmt.Fprintf(wr, "%s\t%s\t%s\t%s\t\n", bundle.Digest.Hex()[:7], bundle.Repository, bundle.Tag, bundle.CreatedAt)
	}

	return wr.Flush()
}

func (c *bundleCommand) push(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	store, err := c.getStore(ctx)
	if err != nil {
		return err
	}

	src, err := oci.ParseReference(args[0])
	if err != nil {
		return err
	}

	dst, err := oci.ParseReference(args[1])
	if err != nil {
		return err
	}

	bundle, err := store.Copy(cmd.Context(), src, dst)
	if err != nil {
		return err
	}

	fmt.Println(bundle.Digest)

	return nil
}

func (c *bundleCommand) pull(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	store, err := c.getStore(ctx)
	if err != nil {
		return err
	}

	src, err := oci.ParseReference(args[0])
	if err != nil {
		return err
	}

	// copy source into destination and rewrite
	// to reference the local equivalent name
	dst := src
	dst.Registry = "local"
	dst.Scheme = "flipt"

	bundle, err := store.Copy(cmd.Context(), src, dst)
	if err != nil {
		return err
	}

	fmt.Println(bundle.Digest)

	return nil
}

func (c *bundleCommand) getStore(ctx context.Context) (*oci.Store, error) {
	logger, cfg, err := buildConfig(ctx)
	if err != nil {
		return nil, err
	}

	dir, err := config.DefaultBundleDir()
	if err != nil {
		return nil, err
	}

	var opts []containers.Option[oci.StoreOptions]
	if cfg := cfg.Storage.OCI; cfg != nil {
		if cfg.Authentication != nil {
			if !cfg.Authentication.Type.IsValid() {
				cfg.Authentication.Type = oci.AuthenticationTypeStatic
			}
			opt, err := oci.WithCredentials(
				cfg.Authentication.Type,
				cfg.Authentication.Username,
				cfg.Authentication.Password,
			)
			if err != nil {
				return nil, err
			}
			opts = append(opts, opt)
		}

		// The default is the 1.1 version, this is why we don't need to check it in here.
		if cfg.ManifestVersion == config.OCIManifestVersion10 {
			opts = append(opts, oci.WithManifestVersion(oras.PackManifestVersion1_0))
		}

		if cfg.BundlesDirectory != "" {
			dir = cfg.BundlesDirectory
		}
	}

	return oci.NewStore(logger, dir, opts...)
}

func writer() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
}
