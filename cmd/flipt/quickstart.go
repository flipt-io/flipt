package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/util"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

type provider string

const (
	ProviderGitHub    provider = "GitHub"
	ProviderGitLab    provider = "GitLab"
	ProviderBitBucket provider = "BitBucket"
	ProviderGitea     provider = "Gitea"
	ProviderAzure     provider = "Azure"
	ProviderGit       provider = "Git"
)

func (p provider) String() string {
	return string(p)
}

func (p provider) Hosted() bool {
	return p != ProviderGit && p != ProviderGitea
}

type quickstart struct {
	configFile         string
	cfg                *config.Config
	pendingCredentials map[string]map[string]any
	repoOwner          string
	repoName           string
	provider           provider
}

// isInterruptError checks if the error is a user interrupt (Ctrl+C)
func isInterruptError(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}

func (c *quickstart) run(cmd *cobra.Command, args []string) error {
	defaultFile := providedConfigFile

	if defaultFile == "" {
		defaultFile = userConfigFile
	}

	c.configFile = defaultFile

	fmt.Println("🚀 Welcome to Flipt v2 Quickstart!")
	fmt.Println("\nThis wizard will help you configure Git storage syncing with a remote repository.")
	fmt.Println()

	// Only show overwrite warning if config file exists
	if _, err := os.Stat(c.configFile); err == nil {
		fmt.Println("⚠️  This will overwrite your existing config file.")
		fmt.Println()
	}

	// prompt if they want to continue, if not return
	ok, err := util.PromptConfirm("Would you like to continue?", true)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	return c.runGitSetup()
}

func (c *quickstart) runGitSetup() error {
	c.cfg = config.Default()

	// Ask for repository URL
	repoURL, err := util.PromptPlaintext("Git repository URL (e.g., https://github.com/owner/repo.git):", "")
	if err != nil {
		return err
	}

	// Parse repository URL to detect prvder
	prvder, repoOwner, repoName, err := parseRepositoryURL(repoURL)
	if err != nil {
		return fmt.Errorf("parsing repository URL: %w", err)
	}

	c.repoOwner = repoOwner
	c.repoName = repoName

	// prompt them if the provider is correct, if not allow them to choose a different provider
	correctProvider, err := util.PromptConfirm(fmt.Sprintf("Is %s the correct provider?", prvder), true)
	if err != nil {
		return err
	}

	var providerString string
	if !correctProvider {
		if err := survey.AskOne(&survey.Select{
			Message: "Which SCM provider would you like to integrate with?",
			Options: []string{"GitHub", "GitLab", "BitBucket", "Azure", "Gitea"},
		}, &providerString); err != nil {
			return err
		}

		prvder = provider(providerString)
	}

	c.provider = prvder

	fmt.Printf("✅ Using %s repository: %s/%s\n\n", prvder, repoOwner, repoName)

	// Configure environment with directory
	if c.cfg.Environments == nil {
		c.cfg.Environments = make(config.EnvironmentsConfig)
	}

	if c.cfg.Environments["default"] == nil {
		c.cfg.Environments["default"] = &config.EnvironmentConfig{
			Name:    "default",
			Storage: "default",
			Default: true,
			SCM: &config.SCMConfig{
				Type: config.SCMType(strings.ToLower(string(prvder))),
			},
		}
	}

	// Ask for branch
	branch, err := util.PromptPlaintext("Default branch name:", "main")
	if err != nil {
		return err
	}

	directory, err := util.PromptPlaintext("Directory to store data in remote repository:", "flipt")
	if err != nil {
		return err
	}

	if directory != "" && directory != "." {
		c.cfg.Environments["default"].Directory = directory
	}

	var credentialsName string

	switch prvder {
	case ProviderGitHub, ProviderGitLab, ProviderBitBucket, ProviderAzure:
		promptToOpenBrowser := true
		c.cfg.Environments["default"].SCM = &config.SCMConfig{
			Type: config.SCMType(strings.ToLower(string(prvder))),
		}

		customAPI, err := util.PromptConfirm(fmt.Sprintf("Are you using a self-hosted/enterprise %s instance?", string(prvder)), false)
		if err != nil {
			return err
		}

		if customAPI {
			apiURL, err := util.PromptPlaintext(fmt.Sprintf("%s API URL:", prvder), "")
			if err != nil {
				return err
			}
			c.cfg.Environments["default"].SCM.ApiURL = apiURL
			promptToOpenBrowser = false
		}
		// Setup credentials for SCM API access
		credentialsName, err = c.setupSCMCredentials(promptToOpenBrowser)
		if err != nil {
			return err
		}
	case ProviderGitea:
		c.cfg.Environments["default"].SCM = &config.SCMConfig{
			Type: config.SCMType(strings.ToLower(string(prvder))),
		}

		apiURL, err := util.PromptPlaintext(fmt.Sprintf("%s API URL:", prvder), "")
		if err != nil {
			return err
		}
		c.cfg.Environments["default"].SCM.ApiURL = apiURL

		// Setup credentials for SCM API access
		credentialsName, err = c.setupSCMCredentials(false)
		if err != nil {
			return err
		}
	}

	// Configure storage
	c.cfg.Storage = config.StoragesConfig{
		"default": &config.StorageConfig{
			Remote: repoURL,
			Branch: branch,
			Backend: config.StorageBackendConfig{
				Type: config.MemoryStorageBackendType,
			},
			PollInterval: 30 * time.Second,
		},
	}

	if credentialsName != "" {
		c.cfg.Storage["default"].Credentials = credentialsName

		if prvder != ProviderGit {
			c.cfg.Environments["default"].SCM.Credentials = &credentialsName
		}
	}

	return c.writeConfig()
}

func (c *quickstart) setupSCMCredentials(promptToOpenBrowser bool) (string, error) {
	credentialsName := fmt.Sprintf("%s-api", strings.ToLower(string(c.provider)))

	if promptToOpenBrowser {
		// Offer to open browser to create PAT
		openBrowser, err := util.PromptConfirm("Would you like to open your browser to create an access token?", true)
		if err != nil {
			return "", err
		}

		if openBrowser {
			patURL := c.getPATCreationURL()

			if patURL != "" {
				fmt.Printf("🌐 Opening %q to create an access token...\n", patURL)
				if err := util.OpenBrowser(patURL); err != nil {
					fmt.Printf("⚠️  Couldn't open browser automatically. Please visit: %q\n", patURL)
				}
				fmt.Println()
				fmt.Println("📝 Required permissions for SCM integration:")
				c.printSCMPermissions()
				fmt.Println()
			}
		}
	}

	token, err := util.PromptPassword("Enter your access token:")
	if err != nil {
		return "", err
	}

	if c.pendingCredentials == nil {
		c.pendingCredentials = make(map[string]map[string]any)
	}

	// Add credentials to pending credentials
	c.pendingCredentials[credentialsName] = map[string]any{
		"type":         "access_token",
		"access_token": token,
	}

	return credentialsName, nil
}

func (c *quickstart) convertConfigToYAML() map[string]any {
	result := make(map[string]any)

	if c.cfg.Storage != nil && len(c.cfg.Storage) > 0 {
		storage := make(map[string]any)
		for name, s := range c.cfg.Storage {
			storageMap := map[string]any{
				"backend": map[string]any{
					"type": string(s.Backend.Type),
				},
				"branch":        s.Branch,
				"poll_interval": s.PollInterval.String(),
			}
			if s.Remote != "" {
				storageMap["remote"] = s.Remote
			}
			if s.Credentials != "" {
				storageMap["credentials"] = s.Credentials
			}
			if s.Backend.Path != "" {
				storageMap["backend"].(map[string]any)["path"] = s.Backend.Path
			}
			storage[name] = storageMap
		}
		result["storage"] = storage
	}

	if c.cfg.Environments != nil && len(c.cfg.Environments) > 0 {
		environments := make(map[string]any)
		for name, env := range c.cfg.Environments {
			envMap := map[string]any{
				"name":    env.Name,
				"storage": env.Storage,
				"default": env.Default,
			}
			if env.Directory != "" {
				envMap["directory"] = env.Directory
			}
			if env.SCM != nil {
				scmMap := map[string]any{
					"type": string(env.SCM.Type),
				}
				if env.SCM.Credentials != nil {
					scmMap["credentials"] = *env.SCM.Credentials
				}
				if env.SCM.ApiURL != "" {
					scmMap["api_url"] = env.SCM.ApiURL
				}
				envMap["scm"] = scmMap
			}
			environments[name] = envMap
		}
		result["environments"] = environments
	}

	return result
}

func (c *quickstart) writeConfig() error {
	configFile := c.configFile
	if configFile == "" {
		configFile = userConfigFile
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return err
	}

	yamlConfig := c.convertConfigToYAML()
	if c.pendingCredentials != nil {
		yamlConfig["credentials"] = c.pendingCredentials
	}

	out, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return err
	}

	// Add schema comment
	content := "# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/v2/config/flipt.schema.json\n\n" + string(out)

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return err
	}

	fmt.Printf("✅ Configuration written to %s\n", configFile)
	fmt.Println()
	fmt.Println("🎉 Setup complete! Next steps:")
	fmt.Println("   1. Start Flipt: flipt server")
	fmt.Println("   2. Open Flipt UI: http://localhost:8080")
	fmt.Println("   3. Create your first feature flag!")

	// Show what was configured
	fmt.Println()
	fmt.Println("📋 What was configured:")
	if c.cfg.Storage != nil && c.cfg.Storage["default"] != nil {
		s := c.cfg.Storage["default"]
		fmt.Printf("   • Git repository: %s\n", s.Remote)
		fmt.Printf("   • Branch: %s\n", s.Branch)
		if s.Credentials != "" {
			fmt.Printf("   • Authentication: configured\n")
		}
	}

	if c.cfg.Environments != nil && c.cfg.Environments["default"] != nil && c.cfg.Environments["default"].SCM != nil {
		fmt.Printf("   • SCM integration: %s\n", c.cfg.Environments["default"].SCM.Type)
	}

	return nil
}

// Helper functions
func parseRepositoryURL(repoURL string) (provider provider, owner, repo string, err error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", "", fmt.Errorf("parsing repository URL: %w", err)
	}

	// Detect provider from hostname
	switch {
	case strings.Contains(u.Host, "github.com"):
		provider = ProviderGitHub
	case strings.Contains(u.Host, "gitlab.com"):
		provider = ProviderGitLab
	case strings.Contains(u.Host, "bitbucket.org"):
		provider = ProviderBitBucket
	case strings.Contains(u.Host, "dev.azure.com") || strings.Contains(u.Host, "visualstudio.com"):
		provider = ProviderAzure
	case strings.Contains(u.Host, "gitea.com"):
		provider = ProviderGitea
	default:
		provider = ProviderGit // generic git
	}

	// Parse owner/repo from path
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		owner = parts[0]
		repo = parts[1]
	}

	return provider, owner, repo, nil
}

func (c *quickstart) getPATCreationURL() string {
	switch c.provider {
	case ProviderGitHub:
		return "https://github.com/settings/tokens"
	case ProviderGitLab:
		return "https://gitlab.com/-/user_settings/personal_access_tokens"
	case ProviderBitBucket:
		if c.repoOwner != "" && c.repoName != "" {
			return fmt.Sprintf("https://bitbucket.org/%s/%s/admin/access-tokens", c.repoOwner, c.repoName)
		}
		// Fallback URL if we don't have repository details
		return "https://bitbucket.org/account/settings/app-passwords/"
	case ProviderAzure:
		return "https://dev.azure.com/_usersSettings/tokens"
	default:
		return ""
	}
}

func (c *quickstart) printSCMPermissions() {
	switch c.provider {
	case ProviderGitHub:
		fmt.Println("  • repo (Full control of private repositories)")
		fmt.Println("  • pull_requests (Create and manage pull requests)")
	case ProviderGitLab:
		fmt.Println("  • read_repository (Read repository)")
		fmt.Println("  • write_repository (Write repository)")
	case ProviderBitBucket:
		fmt.Println("  • Repositories: Read, Write")
		fmt.Println("  • Pull requests: Read, Write")
	case ProviderAzure:
		fmt.Println("  • Code (read & write) - Access to source code and metadata")
		fmt.Println("  • Pull Requests (read & write) - Create and manage pull requests")
	case ProviderGitea:
		fmt.Println("  • repository (Repository access)")
		fmt.Println("  • issue (Issue and pull request access)")
	}
}

func newQuickstartCommandOld() *cobra.Command {
	quickstartCmd := &quickstart{}

	cmd := &cobra.Command{
		Use:   "quickstart ",
		Short: "Interactive setup wizard for Flipt Git storage",
		Long: `Setup wizard helps you configure Flipt v2 with Git storage.

The wizard will guide you through:
  • Configuring Git repository storage with SCM integration
  • Setting up authentication (Personal Access Token)

Examples:
  flipt quickstart              # Interactive setup wizard
  flipt quickstart --config /path/to/config.yml # Path to write to config file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := quickstartCmd.run(cmd, args); err != nil {
				if isInterruptError(err) {
					fmt.Println("\nQuickstart cancelled.")
					return nil
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")

	return cmd
}
