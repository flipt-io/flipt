package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/cloud"
	"go.flipt.io/flipt/internal/cmd/util"
	"golang.org/x/sync/errgroup"
)

const cloudAuthVersion = "0.1.0"

type cloudCommand struct {
	url string
}

type cloudAuth struct {
	Version string       `json:"version,omitempty"`
	Token   string       `json:"token"`
	Tunnel  *cloudTunnel `json:"tunnel,omitempty"`
}

type cloudTunnel struct {
	ID           string `json:"id"`
	Gateway      string `json:"gateway"`
	Organization string `json:"organization"`
	Status       string `json:"status"`
	ExpiresAt    int64  `json:"expiresAt,omitempty"`
}

func newCloudCommand() *cobra.Command {
	cloud := &cloudCommand{}

	cmd := &cobra.Command{
		Use:    "cloud",
		Short:  "Interact with Flipt Cloud",
		Hidden: true,
	}

	cmd.PersistentFlags().StringVarP(&cloud.url, "url", "u", "https://flipt.cloud", "Flipt Cloud URL")

	loginCmd := &cobra.Command{
		Use:   "login [flags]",
		Short: "Authenticate with Flipt Cloud",
		RunE:  cloud.login,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(loginCmd)

	logoutCmd := &cobra.Command{
		Use:   "logout [flags]",
		Short: "Logout from Flipt Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.Remove(filepath.Join(userConfigDir, "cloud.json")); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("removing cloud auth token: %w", err)
			}

			fmt.Println("Logged out from Flipt Cloud.")
			return nil
		},
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(logoutCmd)

	return cmd
}

func (c *cloudCommand) login(cmd *cobra.Command, args []string) error {
	var (
		ctx, cancel = context.WithCancel(cmd.Context())
		_, cfg, err = buildConfig(ctx)
	)
	defer cancel()

	if err != nil {
		return err
	}

	if !cfg.Experimental.Cloud.Enabled {
		return errors.New("cloud feature is not enabled")
	}

	_, err = url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parsing cloud URL: %w", err)
	}

	ok, err := util.PromptConfirm("Open browser to authenticate with Flipt Cloud?", false)
	if err != nil {
		return err
	}

	// if they didn't attempt login, exit
	if !ok {
		return nil
	}

	if err := c.loginFlow(ctx); err != nil {
		return err
	}

	fmt.Println("\n✓ Authenticated with Flipt Cloud!\nYou can now run commands that require cloud authentication.")
	return nil
}

func (c *cloudCommand) loginFlow(ctx context.Context) error {
	flow, err := cloud.InitFlow()
	if err != nil {
		return fmt.Errorf("initializing flow: %w", err)
	}

	defer flow.Close()

	var g errgroup.Group

	g.Go(func() error {
		if err := flow.StartServer(nil); err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("starting server: %w", err)
		}
		return nil
	})

	url, err := flow.BrowserURL(fmt.Sprintf("%s/login/device", c.url))
	if err != nil {
		return fmt.Errorf("creating browser URL: %w", err)
	}

	if err := util.OpenBrowser(url); err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}

	cloudAuthFile := filepath.Join(userConfigDir, "cloud.json")

	tok, err := flow.Wait(ctx)
	if err != nil {
		return fmt.Errorf("waiting for token: %w", err)
	}

	if err := flow.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("closing flow: %w", err)
	}

	cloudAuth := cloudAuth{
		Version: cloudAuthVersion,
		Token:   tok,
	}

	cloudAuthBytes, err := json.Marshal(cloudAuth)
	if err != nil {
		return fmt.Errorf("marshalling cloud auth token: %w", err)
	}

	if err := os.WriteFile(cloudAuthFile, cloudAuthBytes, 0600); err != nil {
		return fmt.Errorf("writing cloud auth token: %w", err)
	}

	return g.Wait()
}
