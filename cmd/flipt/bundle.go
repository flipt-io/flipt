package main

import (
	"fmt"
	"os"
	"text/tabwriter"

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

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all bundles",
		RunE:  bundle.list,
	})

	return cmd
}

func (c *bundleCommand) build(cmd *cobra.Command, args []string) error {
	store, err := c.getStore()
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
	store, err := c.getStore()
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

func (c *bundleCommand) getStore() (*oci.Store, error) {
	logger, cfg, err := buildConfig()
	if err != nil {
		return nil, err
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

	return oci.NewStore(logger, opts...)
}

func writer() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
}
