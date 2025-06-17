package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/util"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

type provider string

const (
	ProviderGitHub provider = "GitHub"
	ProviderGitLab provider = "GitLab"
	ProviderGitea  provider = "Gitea"
	ProviderGit    provider = "Git"
)

func (p provider) String() string {
	return string(p)
}

type quickstart struct {
	configFile         string
	cfg                *config.Config
	pendingCredentials map[string]map[string]any
}

func (c *quickstart) run(cmd *cobra.Command, args []string) error {
	defaultFile := providedConfigFile

	if defaultFile == "" {
		defaultFile = userConfigFile
	}

	c.configFile = defaultFile

	fmt.Println("🚀 Welcome to Flipt v2 Quickstart!")
	fmt.Println("This wizard will help you configure Git storage syncing with a remote repository.")
	fmt.Println()

	fmt.Println("This will overwrite your existing config file if it exists.")
	fmt.Println()

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

	// Parse repository URL to detect provider
	provider, repoOwner, repoName, err := parseRepositoryURL(repoURL)
	if err != nil {
		return fmt.Errorf("parsing repository URL: %w", err)
	}

	// prompt them if the provider is correct, if not allow them to choose a different provider
	correctProvider, err := util.PromptConfirm(fmt.Sprintf("Is %s the correct provider?", provider), true)
	if err != nil {
		return err
	}

	if !correctProvider {
		if err := survey.AskOne(&survey.Select{
			Message: "Which SCM provider would you like to integrate with?",
			Options: []string{"GitHub", "GitLab", "Gitea"},
		}, &provider); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Using %s repository: %s/%s\n\n", provider, repoOwner, repoName)

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
				Type: config.SCMType(strings.ToLower(string(provider))),
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

	var promptToOpenBrowser = true
	if provider != ProviderGit {
		c.cfg.Environments["default"].SCM = &config.SCMConfig{
			Type: config.SCMType(strings.ToLower(string(provider))),
		}

		customAPI, err := util.PromptConfirm(fmt.Sprintf("Are you using a self-hosted/enterprise %s instance?", string(provider)), false)
		if err != nil {
			return err
		}

		if customAPI {
			apiURL, err := util.PromptPlaintext(fmt.Sprintf("%s API URL:", provider), "")
			if err != nil {
				return err
			}
			c.cfg.Environments["default"].SCM.ApiURL = apiURL
			promptToOpenBrowser = false
		}
	}

	// Setup credentials for SCM API access
	credentialsName, err := c.setupSCMCredentials(provider, promptToOpenBrowser)
	if err != nil {
		return err
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

		if provider != ProviderGit {
			c.cfg.Environments["default"].SCM.Credentials = &credentialsName
		}
	}

	return c.writeConfig()
}

func (c *quickstart) setupSCMCredentials(provider provider, promptToOpenBrowser bool) (string, error) {
	credentialsName := fmt.Sprintf("%s-api", strings.ToLower(string(provider)))

	if promptToOpenBrowser {
		// Offer to open browser to create PAT
		openBrowser, err := util.PromptConfirm("Would you like to open your browser to create an API token?", true)
		if err != nil {
			return "", err
		}

		if openBrowser {
			patURL := getPATCreationURL(provider)
			if patURL != "" {
				fmt.Printf("🌐 Opening %s to create an API token...\n", patURL)
				if err := util.OpenBrowser(patURL); err != nil {
					fmt.Printf("⚠️  Couldn't open browser automatically. Please visit: %s\n", patURL)
				}
				fmt.Println()
				fmt.Println("📝 Required permissions for SCM integration:")
				printSCMPermissions(provider)
				fmt.Println()
			}
		}
	}

	token, err := util.PromptPassword("Enter your API token:")
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
	// This is a simplified conversion - in practice you'd want to use
	// proper struct-to-map conversion or mapstructure
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

func getPATCreationURL(provider provider) string {
	switch provider {
	case ProviderGitHub:
		return "https://github.com/settings/tokens"
	case ProviderGitLab:
		return "https://gitlab.com/-/user_settings/personal_access_tokens"
	default:
		return ""
	}
}

func printSCMPermissions(provider provider) {
	switch provider {
	case ProviderGitHub:
		fmt.Println("  • repo (Full control of private repositories)")
		fmt.Println("  • pull_requests (Create and manage pull requests)")
	case ProviderGitLab:
		fmt.Println("  • read_repository (Read repository)")
		fmt.Println("  • write_repository (Write repository)")
	case ProviderGitea:
		fmt.Println("  • repository (Repository access)")
		fmt.Println("  • issue (Issue and pull request access)")
	}
}

func newQuickstartCommand() *cobra.Command {
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
		RunE: quickstartCmd.run,
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")

	return cmd
}
