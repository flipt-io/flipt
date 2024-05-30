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
	"go.flipt.io/flipt/internal/config"
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

type cloudError struct {
	Error string `json:"error"`
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

	serveCmd := &cobra.Command{
		Use:   "serve [flags]",
		Short: "Serve your local instance via Flipt Cloud",
		RunE:  cloud.serve,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(serveCmd)

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

func (c *cloudCommand) serve(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	logger, cfg, err := buildConfig(ctx)
	if err != nil {
		return err
	}

	if !cfg.Experimental.Cloud.Enabled {
		return errors.New("cloud feature is not enabled")
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parsing cloud URL: %w", err)
	}

	srvConfig := func(cfg *config.Config, t cloudTunnel) *config.Config {
		cfg.Cloud.Host = u.Hostname()
		cfg.Cloud.Gateway = t.Gateway
		cfg.Cloud.Organization = t.Organization
		cfg.Cloud.Authentication.ApiKey = "" // clear API key if present to use JWT
		cfg.Server.Cloud.Enabled = true
		cfg.Authentication.Session.Domain = u.Host
		cfg.Authentication.Methods.Cloud.Enabled = true
		return cfg
	}

	// first check for existing of auth token/cloud.json
	// if not found, prompt user to login
AUTH:
	cloudAuthFile := filepath.Join(userConfigDir, "cloud.json")
	f, err := os.ReadFile(cloudAuthFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("\n✗ No cloud authentication token found.")

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

			// otherwise, try reading the file again
			goto AUTH
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
		if errors.Is(err, jwt.ErrTokenExpired) {
			fmt.Println("✗ Existing cloud authentication token expired.")

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

			// otherwise, try reading the file again
			goto AUTH
		}

		return fmt.Errorf("parsing JWT: %w", err)
	}

	if !parsed.Valid {
		return errors.New("invalid JWT")
	}

	fmt.Println("✓ Validated Flipt Cloud authentication.")

	if auth.Tunnel != nil {
		// check if gateway actually exists
		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/gateways/%s/status", c.url, auth.Tunnel.ID), nil)

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

		_ = resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusNotFound:
			fmt.Println("Existing linked Flipt Cloud gateway not found.")
			ok, err := util.PromptConfirm("Continue with new gateway?", false)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

		case http.StatusOK:
			// check if gateway has not expired
			if time.Now().Unix() <= auth.Tunnel.ExpiresAt {
				fmt.Println("✓ Found existing linked Flipt Cloud gateway.")
				// prompt user to see if they want to use existing gateway
				ok, err := util.PromptConfirm("Use existing gateway?", false)
				if err != nil {
					return err
				}

				if ok {
					logger, cfg, err := buildConfig(ctx)
					if err != nil {
						return err
					}

					fmt.Println("✓ Starting local instance linked with Flipt Cloud.")
					return run(ctx, logger, srvConfig(cfg, *auth.Tunnel))
				}
			} else {
				fmt.Println("Existing linked Flipt Cloud gateway has expired.")
				ok, err := util.PromptConfirm("Continue with new gateway?", false)
				if err != nil {
					return err
				}

				if !ok {
					return nil
				}
			}
		default:
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/api/gateways", c.url), nil)
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

	fmt.Println("✓ Created temporary gateway in Flipt Cloud.")
	var t cloudTunnel
	if err := json.Unmarshal(body, &t); err != nil {
		return fmt.Errorf("unmarshalling response body: %w", err)
	}

	auth.Tunnel = &t
	cloudAuthBytes, err := json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("marshalling cloud auth token: %w", err)
	}

	if err := os.WriteFile(cloudAuthFile, cloudAuthBytes, 0600); err != nil {
		return fmt.Errorf("writing cloud auth token: %w", err)
	}

	fmt.Println("✓ Starting local instance linked with Flipt Cloud.")
	return run(ctx, logger, srvConfig(cfg, t))
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
