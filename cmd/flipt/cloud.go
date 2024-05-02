package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/cloud"
	"go.flipt.io/flipt/internal/cmd/util"
	"golang.org/x/sync/errgroup"
)

type cloudCommand struct {
	url string
}

type cloudAuth struct {
	Token string `json:"token"`
}

type cloudInstance struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	OrganizationID string `json:"organizationID"`
	Status         string `json:"status"`
}

func newCloudCommand() *cobra.Command {
	cloud := &cloudCommand{}

	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Interact with Flipt Cloud",
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

	serveCmd := &cobra.Command{
		Use:   "serve [flags]",
		Short: "Serve Flipt Cloud locally",
		RunE:  cloud.serve,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(serveCmd)

	return cmd
}

func (c *cloudCommand) login(cmd *cobra.Command, args []string) error {
	done := make(chan struct{})

	ok, err := util.PromptConfirm("Open browser to authenticate with Flipt Cloud?", false)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	flow, err := cloud.InitFlow()
	if err != nil {
		return fmt.Errorf("initializing flow: %w", err)
	}

	var (
		g           errgroup.Group
		ctx, cancel = context.WithCancel(cmd.Context())
	)

	defer cancel()

	cloudAuthFile := filepath.Join(userConfigDir, "cloud.json")

	g.Go(func() error {
		tok, err := flow.Wait(ctx)
		if err != nil {
			return fmt.Errorf("waiting for token: %w", err)
		}

		cloudAuth := cloudAuth{
			Token: tok,
		}

		cloudAuthBytes, err := json.Marshal(cloudAuth)
		if err != nil {
			return fmt.Errorf("marshalling cloud auth token: %w", err)
		}

		if err := os.WriteFile(cloudAuthFile, cloudAuthBytes, 0600); err != nil {
			return fmt.Errorf("writing cloud auth token: %w", err)
		}

		fmt.Println("\n✅ Authenticated with Flipt Cloud!\nYou can now run commands that require cloud authentication.")

		return nil
	})

	g.Go(func() error {
		if err := flow.StartServer(nil); err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("starting server: %w", err)
		}
		close(done)
		return nil
	})

	g.Go(func() error {
		select {
		case <-done:
			cancel()
		case <-ctx.Done():
			if err := flow.Close(); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
		}
		return nil
	})

	g.Go(func() error {
		url, err := flow.BrowserURL(fmt.Sprintf("%s/login/device", c.url))
		if err != nil {
			return fmt.Errorf("creating browser URL: %w", err)
		}
		return util.OpenBrowser(url)
	})

	return g.Wait()
}

func (c *cloudCommand) serve(cmd *cobra.Command, args []string) error {
	// first check for existing of auth token/cloud.json
	// if not found, prompt user to login
	cloudAuthFile := filepath.Join(userConfigDir, "cloud.json")
	f, err := os.ReadFile(cloudAuthFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No cloud authentication token found. Please run 'flipt cloud login' to authenticate with Flipt Cloud.")
			return nil
		}

		return fmt.Errorf("reading cloud auth payload %w", err)
	}

	var auth cloudAuth

	if err := json.Unmarshal(f, &auth); err != nil {
		return fmt.Errorf("unmarshalling cloud auth payload: %w", err)
	}

	fmt.Println("\n✅ Found Flipt Cloud authentication.")

	// TODO: check for expiration of token
	// TODO: check for existing instance

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/instances", c.url), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", auth.Token))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return fmt.Errorf("reading response body: %w", err)
	}

	_ = resp.Body.Close()

	fmt.Println("✅ Created temporary instance in Flipt Cloud.")
	var instance cloudInstance
	if err := json.Unmarshal(body, &instance); err != nil {
		return fmt.Errorf("unmarshalling response body: %w", err)
	}

	// download config file from cloud
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/api/instances/%s/config", c.url, instance.ID), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", auth.Token))
	req.Header.Set("Accept", "text/yaml")

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return fmt.Errorf("reading response body: %w", err)
	}

	_ = resp.Body.Close()

	fmt.Println("✅ Downloaded configuration file from Flipt Cloud.")

	// write to stdout for now

	fmt.Println(string(body))
	return nil
}
