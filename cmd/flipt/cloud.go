package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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

func newCloudCommand() *cobra.Command {
	cloud := &cloudCommand{}

	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Interact with Flipt Cloud",
	}

	cmd.PersistentFlags().StringVarP(&cloud.url, "url", "u", "https://flipt.cloud", "Flipt Cloud URL")

	authCmd := &cobra.Command{
		Use:   "auth [flags]",
		Short: "Authenticate with Flipt Cloud",
		RunE:  cloud.auth,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(authCmd)
	return cmd
}

func (c *cloudCommand) auth(cmd *cobra.Command, args []string) error {
	done := make(chan struct{})
	const callbackURL = "http://localhost:8080/cloud/auth/callback"

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
		if err := flow.StartServer(nil); !errors.Is(err, http.ErrServerClosed) {
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
			if err := flow.Close(); !errors.Is(err, context.Canceled) {
				return err
			}
		}
		return nil
	})

	browserParams := cloud.BrowserParams{
		RedirectURL: callbackURL,
	}

	g.Go(func() error {
		url, err := flow.BrowserURL(fmt.Sprintf("%s/login/device", c.url), browserParams)
		if err != nil {
			return fmt.Errorf("creating browser URL: %w", err)
		}
		return util.OpenBrowser(url)
	})

	return g.Wait()
}
