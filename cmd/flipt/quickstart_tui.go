package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/util"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

type quickstartTUI struct {
	configFile string
	cfg        *config.Config

	// Form data
	repoURL      string
	provider     string
	providerType provider
	branch       string
	directory    string
	isCustomAPI  bool
	apiURL       string
	token        string

	// Internal state
	pendingCredentials map[string]map[string]any
	repoOwner          string
	repoName           string
}

var (
	// Simple, clean styling
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D")).
			Bold(true)
)

var (
	ErrUserCancelled = errors.New("user cancelled setup")
)

func (c *quickstartTUI) run() error {
	// Clear screen for clean start
	fmt.Print("\033[H\033[2J")
	
	defaultFile := providedConfigFile
	if defaultFile == "" {
		defaultFile = userConfigFile
	}
	c.configFile = defaultFile

	// Initialize config
	c.cfg = config.Default()
	c.pendingCredentials = make(map[string]map[string]any)

	fmt.Println(titleStyle.Render("✨ Flipt v2 Quickstart"))
	fmt.Println(subtitleStyle.Render("Configure Git storage syncing with a remote repository"))
	fmt.Println()

	// Check for existing config
	if _, err := os.Stat(c.configFile); err == nil {
		fmt.Println(warningStyle.Render("⚠ This will overwrite your existing config file"))
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

func (c *quickstartTUI) runConfirmationStep() error {
	var proceed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like to continue with the setup?").
				Value(&proceed).
				Affirmative("Yes").
				Negative("No"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		if isInterruptError(err) {
			return err
		}
		return err
	}

	if !proceed {
		fmt.Println("Setup cancelled.")
		return ErrUserCancelled
	}

	return nil
}

func (c *quickstartTUI) runRepositoryStep() error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Git repository URL").
				Description("Enter the URL of your Git repository").
				Placeholder("https://github.com/owner/repo.git").
				Value(&c.repoURL).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("repository URL is required")
					}
					_, err := url.Parse(s)
					if err != nil {
						return fmt.Errorf("invalid URL format")
					}
					return nil
				}),
		).Title("Repository Configuration"),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		if isInterruptError(err) {
			return err
		}
		return err
	}

	// Parse repository URL
	prvder, owner, name, err := parseRepositoryURL(c.repoURL)
	if err != nil {
		return fmt.Errorf("parsing repository URL: %w", err)
	}

	c.providerType = prvder
	c.provider = string(prvder)
	c.repoOwner = owner
	c.repoName = name

	return nil
}

func (c *quickstartTUI) runProviderStep() error {
	var correctProvider bool

	// Confirm detected provider
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Provider Configuration").
				Description(fmt.Sprintf("Detected provider: %s\n\nIs this the correct provider?", c.provider)).
				Value(&correctProvider).
				Affirmative("Yes").
				Negative("No, let me choose"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := confirmForm.Run(); err != nil {
		if isInterruptError(err) {
			return err
		}
		return err
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
					Title("Select your SCM provider").
					Options(providerOptions...).
					Value(&c.provider),
			),
		).WithTheme(huh.ThemeCharm())

		if err := selectForm.Run(); err != nil {
			if isInterruptError(err) {
				return err
			}
			return err
		}

		c.providerType = provider(c.provider)
	}

	// Handle custom API URL for hosted providers
	if c.providerType.Hosted() {
		customForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Are you using a self-hosted/enterprise %s instance?", c.provider)).
					Value(&c.isCustomAPI).
					Affirmative("Yes").
					Negative("No"),
			),
		).WithTheme(huh.ThemeCharm())

		if err := customForm.Run(); err != nil {
			if isInterruptError(err) {
				return err
			}
			return err
		}

		if c.isCustomAPI {
			apiForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(fmt.Sprintf("%s API URL", c.provider)).
						Description("Enter the API URL for your instance").
						Placeholder("https://git.example.com/api/v4").
						Value(&c.apiURL).
						Validate(func(s string) error {
							if s == "" {
								return fmt.Errorf("API URL is required")
							}
							_, err := url.Parse(s)
							if err != nil {
								return fmt.Errorf("invalid URL format")
							}
							return nil
						}),
				),
			).WithTheme(huh.ThemeCharm())

			if err := apiForm.Run(); err != nil {
				if isInterruptError(err) {
					return err
				}
				return err
			}
		}
	} else if c.providerType == ProviderGitea {
		// Gitea always needs API URL
		apiForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Gitea API URL").
					Description("Enter the API URL for your Gitea instance").
					Placeholder("https://gitea.example.com/api/v1").
					Value(&c.apiURL).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("API URL is required for Gitea")
						}
						_, err := url.Parse(s)
						if err != nil {
							return fmt.Errorf("invalid URL format")
						}
						return nil
					}),
			),
		).WithTheme(huh.ThemeCharm())

		if err := apiForm.Run(); err != nil {
			if isInterruptError(err) {
				return err
			}
			return err
		}
	}

	return nil
}

func (c *quickstartTUI) runBranchDirectoryStep() error {
	// Set defaults
	c.branch = "main"
	c.directory = "flipt"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Default branch name").
				Description("The Git branch to use for configuration").
				Value(&c.branch).
				Placeholder("main"),
			huh.NewInput().
				Title("Directory to store data").
				Description("Directory in the repository to store Flipt data").
				Value(&c.directory).
				Placeholder("flipt"),
		).Title("Storage Configuration"),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		if isInterruptError(err) {
			return err
		}
		return err
	}
	return nil
}

func (c *quickstartTUI) runAuthenticationStep() error {
	// Skip auth for plain Git provider
	if c.providerType == ProviderGit {
		return nil
	}

	// Offer to open browser for PAT creation (if not custom API)
	if !c.isCustomAPI && c.providerType != ProviderGitea {
		var openBrowser bool

		browserForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Would you like to open your browser to create an access token?").
					Description("We'll open the correct page for creating a Personal Access Token").
					Value(&openBrowser).
					Affirmative("Yes, open browser").
					Negative("No, I'll enter it manually"),
			).Title("Authentication Setup"),
		).WithTheme(huh.ThemeCharm())

		if err := browserForm.Run(); err != nil {
			if isInterruptError(err) {
				return err
			}
			return err
		}

		if openBrowser {
			patURL := c.getPATCreationURL()
			if patURL != "" {
				fmt.Printf("Opening browser: %s\n", patURL)
				if err := util.OpenBrowser(patURL); err != nil {
					fmt.Printf("⚠ Couldn't open browser automatically. Please visit:\n%s\n", patURL)
				}
				fmt.Println()
			}
		}
	}

	// Get token
	tokenForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Personal Access Token").
				Description("Enter your access token").
				Value(&c.token).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("access token is required")
					}
					return nil
				}),
		).Title("Authentication Setup"),
	).WithTheme(huh.ThemeCharm())

	if err := tokenForm.Run(); err != nil {
		if isInterruptError(err) {
			return err
		}
		return err
	}
	return nil
}

func (c *quickstartTUI) runReviewStep() error {
	// Create configuration summary for form description
	var configLines []string
	configLines = append(configLines, fmt.Sprintf("Repository:     %s", c.repoURL))
	configLines = append(configLines, fmt.Sprintf("Provider:       %s", c.provider))
	configLines = append(configLines, fmt.Sprintf("Branch:         %s", c.branch))
	configLines = append(configLines, fmt.Sprintf("Directory:      %s", c.directory))

	if c.apiURL != "" {
		configLines = append(configLines, fmt.Sprintf("API URL:        %s", c.apiURL))
	}

	if c.token != "" {
		configLines = append(configLines, "Authentication: ✓ Configured")
	}

	configSummary := strings.Join(configLines, "\n")

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Write this configuration to file?").
				Description(fmt.Sprintf("Configuration:\n\n%s\n\nConfig will be saved to: %s", configSummary, c.configFile)).
				Value(&confirm).
				Affirmative("Yes, save configuration").
				Negative("No, cancel setup"),
		).Title("Review Configuration"),
	).WithTheme(huh.ThemeCharm())

	if err := confirmForm.Run(); err != nil {
		if isInterruptError(err) {
			return err
		}
		return err
	}

	if !confirm {
		return ErrUserCancelled
	}

	// Build configuration
	c.buildConfiguration()

	return nil
}

func (c *quickstartTUI) buildConfiguration() {
	// Configure environment
	if c.cfg.Environments == nil {
		c.cfg.Environments = make(config.EnvironmentsConfig)
	}

	c.cfg.Environments["default"] = &config.EnvironmentConfig{
		Name:    "default",
		Storage: "default",
		Default: true,
	}

	if c.directory != "" && c.directory != "." {
		c.cfg.Environments["default"].Directory = c.directory
	}

	// Configure SCM if needed
	if c.providerType != ProviderGit {
		c.cfg.Environments["default"].SCM = &config.SCMConfig{
			Type: config.SCMType(strings.ToLower(string(c.providerType))),
		}

		if c.apiURL != "" {
			c.cfg.Environments["default"].SCM.ApiURL = c.apiURL
		}

		if c.token != "" {
			credentialsName := fmt.Sprintf("%s-api", strings.ToLower(string(c.providerType)))
			c.cfg.Environments["default"].SCM.Credentials = &credentialsName

			// Add to pending credentials
			c.pendingCredentials[credentialsName] = map[string]any{
				"type":         "access_token",
				"access_token": c.token,
			}
		}
	}

	// Configure storage
	c.cfg.Storage = config.StoragesConfig{
		"default": &config.StorageConfig{
			Remote: c.repoURL,
			Branch: c.branch,
			Backend: config.StorageBackendConfig{
				Type: config.MemoryStorageBackendType,
			},
			PollInterval: 30 * time.Second,
		},
	}

	// Add credentials to storage if set
	if c.token != "" && c.providerType != ProviderGit {
		credentialsName := fmt.Sprintf("%s-api", strings.ToLower(string(c.providerType)))
		c.cfg.Storage["default"].Credentials = credentialsName
	}
}

func (c *quickstartTUI) getPATCreationURL() string {
	switch c.providerType {
	case ProviderGitHub:
		return "https://github.com/settings/tokens/new?description=Flipt%20Access&scopes=repo"
	case ProviderGitLab:
		return "https://gitlab.com/-/user_settings/personal_access_tokens"
	case ProviderBitBucket:
		if c.repoOwner != "" && c.repoName != "" {
			return fmt.Sprintf("https://bitbucket.org/%s/%s/admin/access-tokens", c.repoOwner, c.repoName)
		}
		return "https://bitbucket.org/account/settings/app-passwords/"
	case ProviderAzure:
		return "https://dev.azure.com/_usersSettings/tokens"
	default:
		return ""
	}
}

func (c *quickstartTUI) convertConfigToYAML() map[string]any {
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

func (c *quickstartTUI) writeConfig() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.configFile), 0700); err != nil {
		return err
	}

	yamlConfig := c.convertConfigToYAML()
	if c.pendingCredentials != nil && len(c.pendingCredentials) > 0 {
		yamlConfig["credentials"] = c.pendingCredentials
	}

	out, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return err
	}

	// Add schema comment
	content := "# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/v2/config/flipt.schema.json\n\n" + string(out)

	if err := os.WriteFile(c.configFile, []byte(content), 0600); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(successStyle.Render("✅ Configuration written to " + c.configFile))
	fmt.Println()
	fmt.Println("🎉 Setup complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Start Flipt: flipt server")
	fmt.Println("  2. Open Flipt UI: http://localhost:8080")
	fmt.Println("  3. Create your first feature flag")

	return nil
}

func newQuickstartCommand() *cobra.Command {
	quickstartCmd := &quickstartTUI{}

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
			if errors.Is(err, ErrUserCancelled) || isInterruptError(err) {
				return nil
			}
			
			return err
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")

	return cmd
}
