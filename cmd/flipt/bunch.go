package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/oci"
)

type bunchCommand struct {
	rootDir string
}

func newBunchCommand() *cobra.Command {
	bunch := &bunchCommand{}

	cmd := &cobra.Command{
		Use:   "bunch",
		Short: "Manage flipt bunches",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list all bunches",
		RunE:  bunch.list,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "create a bunch",
		RunE:  bunch.create,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "push",
		Short: "push a bunch",
		RunE:  bunch.push,
		Args:  cobra.ExactArgs(1),
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pull",
		Short: "pull a bunch",
		RunE:  bunch.pull,
		Args:  cobra.ExactArgs(1),
	})

	return cmd
}

func (c *bunchCommand) list(cmd *cobra.Command, args []string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	bunches, err := store.List()
	if err != nil {
		return err
	}

	wr := writer()

	fmt.Fprintf(wr, "DIGEST\tREPO\tTAG\tCREATED\t\n")
	for _, bunch := range bunches {
		fmt.Fprintf(wr, "%s\t%s\t%s\t%s\t\n", bunch.Digest.Hex()[:7], bunch.Repository, bunch.Tag, bunch.CreatedAt)
	}

	return wr.Flush()
}

func (c *bunchCommand) create(cmd *cobra.Command, args []string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	var opts []containers.Option[oci.CreateOption]
	if len(args) > 0 {
		opts = append(opts, oci.WithRef(args[0]))
	}

	bunch, err := store.Create(cmd.Context(), os.DirFS("."), opts...)
	if err != nil {
		return err
	}

	fmt.Println(bunch.Digest)

	return nil
}

func (c *bunchCommand) push(cmd *cobra.Command, args []string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	err = store.Push(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return nil
}

func (c *bunchCommand) pull(cmd *cobra.Command, args []string) error {
	store, err := getStore()
	if err != nil {
		return err
	}

	err = store.Pull(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return nil
}

func getStore() (*oci.Store, error) {
	dir, err := ensureBunchesDir()
	if err != nil {
		return nil, err
	}

	return oci.NewStore(dir)
}

func ensureBunchesDir() (string, error) {
	stateDir, err := defaultUserStateDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(stateDir, "bunches")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating bunches directory: %w", err)
	}

	return dir, nil
}

func writer() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
}
