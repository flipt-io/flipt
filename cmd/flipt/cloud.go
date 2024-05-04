package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/cloud"
	"go.flipt.io/flipt/internal/cmd/util"
	"golang.org/x/sync/errgroup"
)

type cloudCommand struct {
	url string
}

type cloudAuth struct {
	Token    string         `json:"token"`
	Instance *cloudInstance `json:"instance,omitempty"`
}

type cloudInstance struct {
	ID           string `json:"id"`
	Instance     string `json:"instance"`
	Organization string `json:"organization"`
	Status       string `json:"status"`
	ExpiresAt    int64  `json:"expiresAt"`
}

type cloudError struct {
	Error string `json:"error"`
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

		fmt.Println("\n✓ Authenticated with Flipt Cloud!\nYou can now run commands that require cloud authentication.")

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
	ctx := cmd.Context()
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parsing cloud URL: %w", err)
	}

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

	fmt.Println("\n✓ Found Flipt Cloud authentication.")

	// validate JWT using our JWKS endpoint
	jwksURL := fmt.Sprintf("%s%s", c.url, "/api/auth/jwks")

	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return fmt.Errorf("creating keyfunc: %w", err)
	}

	parsed, err := jwt.Parse(auth.Token, k.Keyfunc, jwt.WithExpirationRequired())
	if err != nil {
		// TODO: handle case where token is expired
		// Error: parsing JWT: token has invalid claims: token is expired
		return fmt.Errorf("parsing JWT: %w", err)
	}

	if !parsed.Valid {
		return errors.New("invalid JWT")
	}

	fmt.Println("✓ Validated Flipt Cloud authentication.")

	if auth.Instance != nil {
		// check if instance has not expired
		if time.Now().Unix() <= auth.Instance.ExpiresAt {
			fmt.Println("✓ Found existing linked Flipt Cloud instance.")
			// prompt user to see if they want to use existing instance
			ok, err := util.PromptConfirm("Use existing instance?", false)
			if err != nil {
				return err
			}

			if ok {
				logger, cfg, err := buildConfig()
				if err != nil {
					return err
				}

				cfg.Server.Cloud.Enabled = true
				cfg.Server.Cloud.Address = u.Host
				cfg.Server.Cloud.Instance = auth.Instance.Instance
				cfg.Server.Cloud.Organization = auth.Instance.Organization

				fmt.Println("✓ Starting local instance linked with Flipt Cloud.")
				return run(ctx, logger, cfg)
			}
		} else {
			fmt.Println("Existing linked Flipt Cloud instance has expired.")
			ok, err := util.PromptConfirm("Continue with new instance?", false)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/api/instances", c.url), nil)
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		var cloudErr cloudError
		if err := json.Unmarshal(body, &cloudErr); err != nil {
			return fmt.Errorf("unmarshalling response body: %w", err)
		}

		return errors.New(cloudErr.Error)
	}

	fmt.Println("✓ Created temporary instance in Flipt Cloud.")
	var instance cloudInstance
	if err := json.Unmarshal(body, &instance); err != nil {
		return fmt.Errorf("unmarshalling response body: %w", err)
	}

	// save instance to auth file
	auth.Instance = &instance
	cloudAuthBytes, err := json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("marshalling cloud auth token: %w", err)
	}

	if err := os.WriteFile(cloudAuthFile, cloudAuthBytes, 0600); err != nil {
		return fmt.Errorf("writing cloud auth token: %w", err)
	}

	logger, cfg, err := buildConfig()
	if err != nil {
		return err
	}

	cfg.Server.Cloud.Enabled = true
	cfg.Server.Cloud.Address = u.Host
	cfg.Server.Cloud.Instance = instance.Instance
	cfg.Server.Cloud.Organization = instance.Organization

	fmt.Println("✓ Starting local instance linked with Flipt Cloud.")
	return run(ctx, logger, cfg)
}
