package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/util"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

// Provider represents a Git service provider type
type Provider int

const (
	ProviderUnknown Provider = iota
	ProviderGitHub
	ProviderGitLab
	ProviderBitBucket
	ProviderGitea
	ProviderAzure
	ProviderGit
)

var providerNames = map[Provider]string{
	ProviderGitHub:    "GitHub",
	ProviderGitLab:    "GitLab",
	ProviderBitBucket: "BitBucket",
	ProviderGitea:     "Gitea",
	ProviderAzure:     "Azure",
	ProviderGit:       "Git",
}

func (p Provider) String() string {
	if name, ok := providerNames[p]; ok {
		return name
	}
	return "Unknown"
}

func (p Provider) IsHosted() bool {
	return p != ProviderGit && p != ProviderGitea
}

type quickstart struct {
	// Configuration
	configFile string
	cfg        *config.Config

	// Repository information
	repo struct {
		url       string
		owner     string
		name      string
		branch    string
		directory string
	}

	// Provider information
	provider struct {
		typ      Provider
		name     string
		isCustom bool
		apiURL   string
		token    string
	}

	// Internal state
	pendingCredentials map[string]map[string]any
}

const (
	// Default values
	DefaultBranch    = "main"
	DefaultDirectory = "flipt"
	DefaultStorage   = "default"
	DefaultEnv       = "default"

	// File permissions
	ConfigDirPerm  = 0700
	ConfigFilePerm = 0600

	// URLs and endpoints
	GitHubTokenURL    = "https://github.com/settings/tokens/new?description=Flipt%20Access&scopes=repo"
	GitLabTokenURL    = "https://gitlab.com/-/user_settings/personal_access_tokens"
	AzureTokenURL     = "https://dev.azure.com/_usersSettings/tokens"
	BitBucketTokenURL = "https://bitbucket.org/account/settings/app-passwords/"

	// YAML schema comment
	yamlSchemaComment = "# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/v2/config/flipt.schema.json\n\n"
)

var (
	// Enhanced color scheme with better visual hierarchy
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D"))

	// New styles for improved visual hierarchy
	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4FF")).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B9D")).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B8BCC8"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			MarginTop(1)

	configItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)
)

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// isInterruptError checks if the error is a user interrupt (Ctrl+C) for bubbletea/huh
func isInterruptError(err error) bool {
	return errors.Is(err, tea.ErrInterrupted) || errors.Is(err, huh.ErrUserAborted)
}

// validateRepositoryURL validates a repository URL
func (c *quickstart) validateRepositoryURL(repoURL string) error {
	if repoURL == "" {
		return ValidationError{Field: "repository_url", Message: "repository URL is required"}
	}
	if _, err := url.Parse(repoURL); err != nil {
		return ValidationError{Field: "repository_url", Message: "invalid URL format"}
	}
	return nil
}

// validateAPIURL validates an API URL
func (c *quickstart) validateAPIURL(apiURL string) error {
	if apiURL == "" {
		return ValidationError{Field: "api_url", Message: "API URL is required"}
	}
	if _, err := url.Parse(apiURL); err != nil {
		return ValidationError{Field: "api_url", Message: "invalid URL format"}
	}
	return nil
}

// validateToken validates an access token
func (c *quickstart) validateToken(token string) error {
	if token == "" {
		return ValidationError{Field: "access_token", Message: "access token is required"}
	}
	return nil
}

func (c *quickstart) run() error {
	defaultFile := providedConfigFile
	if defaultFile == "" {
		defaultFile = userConfigFile
	}
	c.configFile = defaultFile

	// Initialize config
	c.cfg = config.Default()
	c.pendingCredentials = make(map[string]map[string]any)

	fmt.Println()
	fmt.Println(titleStyle.Render("✨ Flipt v2 Quickstart"))
	fmt.Println(subtitleStyle.Render("Configure Git storage syncing with a remote repository"))
	fmt.Println()

	// Check for existing config
	if _, err := os.Stat(c.configFile); err == nil {
		fmt.Println(warningStyle.Render("⚠  Warning: This will overwrite your existing config file"))
		fmt.Println()
	}

	// Step 1: Confirmation
	if err := c.runConfirmationStep(); err != nil {
		return err
	}

	// Step 2: Repository Configuration
	if err := c.runRepositoryStep(); err != nil {
		return err
	}

	// Step 3: Provider Configuration
	if err := c.runProviderStep(); err != nil {
		return err
	}

	// Step 4: Branch and Directory
	if err := c.runBranchDirectoryStep(); err != nil {
		return err
	}

	// Step 5: Authentication (if needed)
	if err := c.runAuthenticationStep(); err != nil {
		return err
	}

	// Step 6: Review and Confirm
	if err := c.runReviewStep(); err != nil {
		return err
	}

	// Write configuration
	return c.writeConfig()
}

func (c *quickstart) runConfirmationStep() error {
	var proceed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(highlightStyle.Render("Setup Confirmation")).
				Description("Would you like to continue with the Git storage setup?").
				Value(&proceed).
				Affirmative("Yes, continue").
				Negative("No, cancel"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return fmt.Errorf("running confirmation form: %w", err)
	}

	if !proceed {
		return tea.ErrInterrupted
	}

	return nil
}

func (c *quickstart) runRepositoryStep() error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Git repository URL").
				Description("Enter the URL of your Git repository").
				Placeholder("https://github.com/owner/repo.git").
				Value(&c.repo.url).
				Validate(c.validateRepositoryURL),
		).Title(highlightStyle.Render("Repository Configuration")),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return fmt.Errorf("running repository configuration form: %w", err)
	}

	// Parse repository URL
	providerType, owner, name, err := parseRepositoryURL(c.repo.url)
	if err != nil {
		return fmt.Errorf("parsing repository URL: %w", err)
	}

	c.provider.typ = providerType
	c.provider.name = providerType.String()
	c.repo.owner = owner
	c.repo.name = name

	return nil
}

func (c *quickstart) runProviderStep() error {
	var correctProvider bool

	// Confirm detected provider
	providerDisplay := accentStyle.Render(c.provider.name)
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(highlightStyle.Render("Provider Configuration")).
				Description(fmt.Sprintf("%s %s\n\nIs this the correct provider?",
					labelStyle.Render("Detected provider:"),
					providerDisplay)).
				Value(&correctProvider).
				Affirmative("Yes").
				Negative("No, let me choose"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("running provider confirmation form: %w", err)
	}

	if !correctProvider {
		// Let user select provider
		providerOptions := []huh.Option[string]{
			huh.NewOption("GitHub", "GitHub"),
			huh.NewOption("GitLab", "GitLab"),
			huh.NewOption("BitBucket", "BitBucket"),
			huh.NewOption("Azure DevOps", "Azure"),
			huh.NewOption("Gitea", "Gitea"),
			huh.NewOption("Generic Git", "Git"),
		}

		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(highlightStyle.Render("Select SCM Provider")).
					Description("Choose your source control management provider").
					Options(providerOptions...).
					Value(&c.provider.name),
			),
		).WithTheme(huh.ThemeCharm())

		if err := selectForm.Run(); err != nil {
			return fmt.Errorf("running provider selection form: %w", err)
		}

		// Find provider type by name
		for provType, name := range providerNames {
			if name == c.provider.name {
				c.provider.typ = provType
				break
			}
		}
	}

	// Handle custom API URL for hosted providers
	if c.provider.typ.IsHosted() {
		customForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Are you using a self-hosted/enterprise %s instance?", accentStyle.Render(c.provider.name))).
					Value(&c.provider.isCustom).
					Affirmative("Yes, self-hosted").
					Negative("No, cloud version"),
			),
		).WithTheme(huh.ThemeCharm())

		if err := customForm.Run(); err != nil {
			return fmt.Errorf("running custom API configuration form: %w", err)
		}

		if c.provider.isCustom {
			apiForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(fmt.Sprintf("%s API URL", accentStyle.Render(c.provider.name))).
						Description("Enter the API URL for your instance").
						Placeholder("https://git.example.com/api/v4").
						Value(&c.provider.apiURL).
						Validate(c.validateAPIURL),
				),
			).WithTheme(huh.ThemeCharm())

			if err := apiForm.Run(); err != nil {
				return fmt.Errorf("running API URL configuration form: %w", err)
			}
		}
	} else if c.provider.typ == ProviderGitea {
		// Gitea always needs API URL
		apiForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(accentStyle.Render("Gitea") + " API URL").
					Description("Enter the API URL for your Gitea instance").
					Placeholder("https://gitea.example.com/api/v1").
					Value(&c.provider.apiURL).
					Validate(func(s string) error {
						if err := c.validateAPIURL(s); err != nil {
							if s == "" {
								return ValidationError{Field: "api_url", Message: "API URL is required for Gitea"}
							}
							return err
						}
						return nil
					}),
			),
		).WithTheme(huh.ThemeCharm())

		if err := apiForm.Run(); err != nil {
			return fmt.Errorf("running Gitea API URL configuration form: %w", err)
		}
	}

	return nil
}

func (c *quickstart) runBranchDirectoryStep() error {
	// Set defaults
	c.repo.branch = DefaultBranch
	c.repo.directory = DefaultDirectory

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Default branch name").
				Description("The Git branch to use for configuration").
				Value(&c.repo.branch).
				Placeholder(DefaultBranch),
			huh.NewInput().
				Title("Directory to store data").
				Description("Directory in the repository to store Flipt data").
				Value(&c.repo.directory).
				Placeholder(DefaultDirectory),
		).Title(highlightStyle.Render("Storage Configuration")),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return fmt.Errorf("running branch and directory configuration form: %w", err)
	}
	return nil
}

func (c *quickstart) runAuthenticationStep() error {
	// Skip auth for plain Git provider
	if c.provider.typ == ProviderGit {
		return nil
	}

	// Offer to open browser for PAT creation (if not custom API)
	if !c.provider.isCustom && c.provider.typ != ProviderGitea {
		var openBrowser bool

		browserForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Would you like to open your browser to create an access token?").
					Description("We'll open the correct page for creating a Personal Access Token").
					Value(&openBrowser).
					Affirmative("Yes, open browser").
					Negative("No, I'll enter it manually"),
			).Title(highlightStyle.Render("Authentication Setup")),
		).WithTheme(huh.ThemeCharm())

		if err := browserForm.Run(); err != nil {
			return fmt.Errorf("running browser confirmation form: %w", err)
		}

		if openBrowser {
			patURL := c.getPATCreationURL()
			if patURL != "" {
				fmt.Println(labelStyle.Render("Opening browser: ") + accentStyle.Render(patURL))
				if err := util.OpenBrowser(patURL); err != nil {
					fmt.Println(warningStyle.Render("⚠  Couldn't open browser automatically."))
					fmt.Println(labelStyle.Render("Please visit: ") + accentStyle.Render(patURL))
				}
			}
		}
	}

	// Get token
	tokenForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Personal Access Token").
				Description("Enter your access token (will be hidden)").
				Value(&c.provider.token).
				EchoMode(huh.EchoModePassword).
				Validate(c.validateToken),
		),
	).WithTheme(huh.ThemeCharm())

	if err := tokenForm.Run(); err != nil {
		return fmt.Errorf("running token input form: %w", err)
	}
	return nil
}

func (c *quickstart) runReviewStep() error {
	// Create configuration summary with improved formatting
	var configLines []string
	configLines = append(configLines, fmt.Sprintf("%s     %s",
		labelStyle.Render("Repository:"),
		valueStyle.Render(c.repo.url)))
	configLines = append(configLines, fmt.Sprintf("%s       %s",
		labelStyle.Render("Provider:"),
		accentStyle.Render(c.provider.name)))
	configLines = append(configLines, fmt.Sprintf("%s         %s",
		labelStyle.Render("Branch:"),
		valueStyle.Render(c.repo.branch)))
	configLines = append(configLines, fmt.Sprintf("%s      %s",
		labelStyle.Render("Directory:"),
		valueStyle.Render(c.repo.directory)))

	if c.provider.apiURL != "" {
		configLines = append(configLines, fmt.Sprintf("%s        %s",
			labelStyle.Render("API URL:"),
			valueStyle.Render(c.provider.apiURL)))
	}

	if c.provider.token != "" {
		configLines = append(configLines, fmt.Sprintf("%s %s",
			labelStyle.Render("Authentication:"),
			successStyle.Render("✓ Configured")))
	}

	configSummary := strings.Join(configLines, "\n")

	var confirm bool
	configPath := highlightStyle.Render(c.configFile)
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(highlightStyle.Render("Review Configuration")).
				Description(fmt.Sprintf("%s\n\n%s\n\n%s %s",
					labelStyle.Render("Configuration:"),
					configSummary,
					labelStyle.Render("Config will be saved to:"),
					configPath)).
				Value(&confirm).
				Affirmative("Yes, save configuration").
				Negative("No, cancel setup"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("running configuration review form: %w", err)
	}

	if !confirm {
		return tea.ErrInterrupted
	}

	// Build configuration
	c.buildConfiguration()

	return nil
}

func (c *quickstart) buildConfiguration() {
	// Configure environment
	if c.cfg.Environments == nil {
		c.cfg.Environments = make(config.EnvironmentsConfig)
	}

	c.cfg.Environments[DefaultEnv] = &config.EnvironmentConfig{
		Name:    DefaultEnv,
		Storage: DefaultStorage,
		Default: true,
	}

	if c.repo.directory != "" && c.repo.directory != "." {
		c.cfg.Environments[DefaultEnv].Directory = c.repo.directory
	}

	// Configure SCM if needed
	if c.provider.typ != ProviderGit {
		c.cfg.Environments[DefaultEnv].SCM = &config.SCMConfig{
			Type: config.SCMType(strings.ToLower(c.provider.name)),
		}

		if c.provider.apiURL != "" {
			c.cfg.Environments[DefaultEnv].SCM.ApiURL = c.provider.apiURL
		}

		if c.provider.token != "" {
			credentialsName := fmt.Sprintf("%s-api", strings.ToLower(c.provider.name))
			c.cfg.Environments[DefaultEnv].SCM.Credentials = &credentialsName

			// Add to pending credentials
			c.pendingCredentials[credentialsName] = map[string]any{
				"type":         "access_token",
				"access_token": c.provider.token,
			}
		}
	}

	// Configure storage
	c.cfg.Storage = config.StoragesConfig{
		DefaultStorage: &config.StorageConfig{
			Remote: c.repo.url,
			Branch: c.repo.branch,
			Backend: config.StorageBackendConfig{
				Type: config.MemoryStorageBackendType,
			},
			PollInterval: 30 * time.Second,
		},
	}

	// Add credentials to storage if set
	if c.provider.token != "" && c.provider.typ != ProviderGit {
		credentialsName := fmt.Sprintf("%s-api", strings.ToLower(c.provider.name))
		c.cfg.Storage[DefaultStorage].Credentials = credentialsName
	}
}

func parseRepositoryURL(repoURL string) (providerType Provider, repoOwner, repoName string, err error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return ProviderUnknown, "", "", fmt.Errorf("parsing repository URL: %w", err)
	}

	// Detect provider from hostname
	switch {
	case strings.Contains(u.Host, "github.com"):
		providerType = ProviderGitHub
	case strings.Contains(u.Host, "gitlab.com"):
		providerType = ProviderGitLab
	case strings.Contains(u.Host, "bitbucket.org"):
		providerType = ProviderBitBucket
	case strings.Contains(u.Host, "dev.azure.com") || strings.Contains(u.Host, "visualstudio.com"):
		providerType = ProviderAzure
	case strings.Contains(u.Host, "gitea.com"):
		providerType = ProviderGitea
	default:
		providerType = ProviderGit // generic git
	}

	// Parse owner/repo from path
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		repoOwner = parts[0]
		repoName = parts[1]
	}

	return providerType, repoOwner, repoName, nil
}

func (c *quickstart) getPATCreationURL() string {
	switch c.provider.typ {
	case ProviderGitHub:
		return GitHubTokenURL
	case ProviderGitLab:
		return GitLabTokenURL
	case ProviderBitBucket:
		if c.repo.owner != "" && c.repo.name != "" {
			return fmt.Sprintf("https://bitbucket.org/%s/%s/admin/access-tokens", c.repo.owner, c.repo.name)
		}
		return BitBucketTokenURL
	case ProviderAzure:
		return AzureTokenURL
	default:
		return ""
	}
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
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.configFile), ConfigDirPerm); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	yamlConfig := c.convertConfigToYAML()
	if c.pendingCredentials != nil && len(c.pendingCredentials) > 0 {
		yamlConfig["credentials"] = c.pendingCredentials
	}

	out, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("marshaling configuration to YAML: %w", err)
	}

	// Add schema comment
	content := yamlSchemaComment + string(out)

	if err := os.WriteFile(c.configFile, []byte(content), ConfigFilePerm); err != nil {
		return fmt.Errorf("writing configuration file: %w", err)
	}

	fmt.Println(successStyle.Render("✅ Configuration successfully written!"))
	fmt.Println(configItemStyle.Render(labelStyle.Render("Location: ") + valueStyle.Render(c.configFile)))
	fmt.Println()
	fmt.Println(titleStyle.Render("🎉 Setup Complete!"))
	fmt.Println()
	fmt.Println(subtitleStyle.Render("Next steps:"))
	fmt.Println(configItemStyle.Render("1. Start Flipt:         " + accentStyle.Render("flipt server")))
	fmt.Println(configItemStyle.Render("2. Open Flipt UI:       " + accentStyle.Render("http://localhost:8080")))
	fmt.Println(configItemStyle.Render("3. Create your first feature flag"))
	fmt.Println()

	return nil
}

func newQuickstartCommand() *cobra.Command {
	quickstartCmd := &quickstart{}

	cmd := &cobra.Command{
		Use:   "quickstart",
		Short: "Interactive setup wizard for Flipt Git storage",
		Long: `Setup wizard helps you configure Flipt v2 with Git storage.

The wizard will guide you through:
  • Configuring Git repository storage with SCM integration
  • Setting up authentication (Personal Access Token)

Examples:
  flipt quickstart              # Interactive setup wizard
  flipt quickstart --config /path/to/config.yml # Path to write to config file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := quickstartCmd.run()

			// Handle user cancellation and interrupts silently
			if isInterruptError(err) {
				return nil
			}

			return err
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")

	return cmd
}
