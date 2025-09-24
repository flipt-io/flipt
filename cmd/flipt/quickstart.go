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
	"github.com/mattn/go-isatty"
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
	ProviderBitBucket: "Bitbucket",
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

// WizardStep represents the current step in the wizard
type WizardStep int

const (
	StepConfirmation WizardStep = iota
	StepRepository
	StepProvider
	StepBranchDirectory
	StepAuthentication
	StepReview
	StepComplete
)

var stepNames = map[WizardStep]string{
	StepConfirmation:    "Welcome",
	StepRepository:      "Repository",
	StepProvider:        "Provider",
	StepBranchDirectory: "Storage",
	StepAuthentication:  "Authentication",
	StepReview:          "Review",
	StepComplete:        "Complete",
}

func (s WizardStep) String() string {
	if name, ok := stepNames[s]; ok {
		return name
	}
	return "Unknown"
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
	currentStep        WizardStep
	totalSteps         int
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

// Note: Common styles are imported from styles.go to maintain consistency across commands

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

// renderProgressIndicator creates a visual progress bar for the wizard
func (c *quickstart) renderProgressIndicator() string {
	steps := []string{"Welcome", "Repository", "Provider", "Storage", "Auth", "Review", "Complete"}
	var indicators []string

	for i, step := range steps {
		var style lipgloss.Style
		var symbol string

		if WizardStep(i) < c.currentStep {
			style = ProgressCompleteStyle
			symbol = "✓"
		} else if WizardStep(i) == c.currentStep {
			style = ProgressActiveStyle
			symbol = "●"
		} else {
			style = ProgressInactiveStyle
			symbol = "○"
		}

		// Add the indicator with proper spacing
		indicator := style.Render(fmt.Sprintf("%s %s", symbol, step))
		indicators = append(indicators, indicator)

		// Add spacing between items (except after the last one)
		if i < len(steps)-1 {
			indicators = append(indicators, " ")
		}
	}

	progressLine := lipgloss.JoinHorizontal(lipgloss.Left, indicators...)

	return CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			SectionHeaderStyle.Render("Setup Progress"),
			progressLine,
		),
	)
}

// renderHeader creates the header with title and progress
func (c *quickstart) renderHeader() string {
	title := TitleStyle.Render("Flipt v2 Quickstart")
	subtitle := SubtitleStyle.Render("Configure Git storage syncing with a remote repository")

	headerContent := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)

	if c.currentStep != StepComplete {
		headerContent = lipgloss.JoinVertical(lipgloss.Center,
			headerContent,
			lipgloss.NewStyle().MarginTop(1).Render(c.renderProgressIndicator()),
		)
	}

	return headerContent
}

func (c *quickstart) run() error {
	// Check if we're in a TTY session
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("quickstart requires an interactive terminal (TTY) session\n" +
			"Please run this command in an interactive terminal or use 'flipt config init' for non-interactive setup")
	}

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	defaultFile := providedConfigFile
	if defaultFile == "" {
		defaultFile = userConfigFile
	}
	c.configFile = defaultFile

	// Initialize config
	c.cfg = config.Default()
	c.pendingCredentials = make(map[string]map[string]any)
	c.currentStep = StepConfirmation
	c.totalSteps = 7

	fmt.Println()
	fmt.Println(c.renderHeader())
	fmt.Println()

	// Check for existing config
	if _, err := os.Stat(c.configFile); err == nil {
		warningCard := CardStyle.Copy().
			BorderForeground(Amber).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					WarningStyle.Render("⚠  Existing Configuration Detected"),
					HelperTextStyle.Render("\nThis will overwrite your existing config file at:"),
					ValueStyle.Render(c.configFile),
				),
			)
		fmt.Println(warningCard)
		fmt.Println()
	}

	// Step 1: Confirmation
	c.currentStep = StepConfirmation
	if err := c.runConfirmationStep(); err != nil {
		return err
	}

	// Step 2: Repository Configuration
	c.currentStep = StepRepository
	fmt.Print("\033[H\033[2J")
	fmt.Println(c.renderHeader())
	if err := c.runRepositoryStep(); err != nil {
		return err
	}

	// Step 3: Provider Configuration
	c.currentStep = StepProvider
	fmt.Print("\033[H\033[2J")
	fmt.Println(c.renderHeader())
	if err := c.runProviderStep(); err != nil {
		return err
	}

	// Step 4: Branch and Directory
	c.currentStep = StepBranchDirectory
	fmt.Print("\033[H\033[2J")
	fmt.Println(c.renderHeader())
	if err := c.runBranchDirectoryStep(); err != nil {
		return err
	}

	// Step 5: Authentication (if needed)
	c.currentStep = StepAuthentication
	fmt.Print("\033[H\033[2J")
	fmt.Println(c.renderHeader())
	if err := c.runAuthenticationStep(); err != nil {
		return err
	}

	// Step 6: Review and Confirm
	c.currentStep = StepReview
	fmt.Print("\033[H\033[2J")
	fmt.Println(c.renderHeader())
	if err := c.runReviewStep(); err != nil {
		return err
	}

	// Step 7: Complete
	c.currentStep = StepComplete
	fmt.Print("\033[H\033[2J")

	// Write configuration
	return c.writeConfig()
}

func (c *quickstart) runConfirmationStep() error {
	var proceed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Ready to begin?").
				Description("This setup will take about 2 minutes").
				Value(&proceed).
				Affirmative("Yes, let's start").
				Negative("No, maybe later"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return fmt.Errorf("running confirmation form: %w", err)
	}

	if !proceed {
		fmt.Println(HelperTextStyle.Render("Setup cancelled. You can run 'flipt quickstart' anytime to continue."))
		return tea.ErrInterrupted
	}

	return nil
}

func (c *quickstart) runRepositoryStep() error {
	// Show examples without card
	fmt.Println(HelperTextStyle.Render("We'll store your feature flags in a Git repository."))
	fmt.Println()
	fmt.Println(InputLabelStyle.Render("Examples:"))
	fmt.Println(lipgloss.NewStyle().Foreground(MutedGray).Render("  • https://github.com/your-org/feature-flags.git"))
	fmt.Println(lipgloss.NewStyle().Foreground(MutedGray).Render("  • https://gitlab.com/team/config-repo.git"))
	fmt.Println(lipgloss.NewStyle().Foreground(MutedGray).Render("  • git@github.com:company/flipt-config.git"))
	fmt.Println()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(InputLabelStyle.Render("Git Repository URL")).
				Description("Enter the URL of your Git repository (HTTPS or SSH)").
				Placeholder("https://github.com/owner/repo.git").
				Value(&c.repo.url).
				Validate(c.validateRepositoryURL),
		),
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

	// Show detected information
	fmt.Println()
	fmt.Println(SuccessStyle.Render("✓ Repository detected!"))
	fmt.Println(fmt.Sprintf("%s %s", LabelStyle.Render("Provider:"), ValueStyle.Render(c.provider.name)))
	fmt.Println(fmt.Sprintf("%s %s/%s", LabelStyle.Render("Repository:"), ValueStyle.Render(owner), ValueStyle.Render(name)))
	fmt.Println()

	return nil
}

func (c *quickstart) runProviderStep() error {
	var correctProvider bool

	// Show provider detection result without card
	fmt.Println(fmt.Sprintf("%s %s", LabelStyle.Render("Detected:"), ValueStyle.Render(c.provider.name)))
	fmt.Println(fmt.Sprintf("%s %s", LabelStyle.Render("Repository:"), ValueStyle.Render(fmt.Sprintf("%s/%s", c.repo.owner, c.repo.name))))
	fmt.Println()

	// Confirm detected provider
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Is this correct?").
				Description(fmt.Sprintf("We detected %s as your Git provider", c.provider.name)).
				Value(&correctProvider).
				Affirmative("Yes, that's correct").
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
			huh.NewOption("Bitbucket", "Bitbucket"),
			huh.NewOption("Azure DevOps", "Azure"),
			huh.NewOption("Gitea", "Gitea"),
			huh.NewOption("Generic Git", "Git"),
		}

		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select SCM Provider").
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
					Title(fmt.Sprintf("Are you using a self-hosted/enterprise %s instance?", c.provider.name)).
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
						Title(fmt.Sprintf("%s API URL", AccentStyle.Render(c.provider.name))).
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
					Title(AccentStyle.Render("Gitea") + " API URL").
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

	// Show organization tips without card
	fmt.Println(HelperTextStyle.Render("Configure how your feature flags will be organized in the repository."))
	fmt.Println()
	fmt.Println(InputLabelStyle.Render("Organization patterns:"))
	fmt.Println(lipgloss.NewStyle().Foreground(MutedGray).Render("  • By environment: /dev, /staging, /production"))
	fmt.Println(lipgloss.NewStyle().Foreground(MutedGray).Render("  • By service: /auth-service, /payment-service"))
	fmt.Println(lipgloss.NewStyle().Foreground(MutedGray).Render("  • By team: /team-platform, /team-product"))
	fmt.Println()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(InputLabelStyle.Render("Branch Name")).
				Description("The Git branch to sync configuration from").
				Value(&c.repo.branch).
				Placeholder(DefaultBranch),
			huh.NewInput().
				Title(InputLabelStyle.Render("Storage Directory")).
				Description("Directory path in the repository (e.g., 'flipt' or 'config/features')").
				Value(&c.repo.directory).
				Placeholder(DefaultDirectory),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return fmt.Errorf("running branch and directory configuration form: %w", err)
	}

	// Ensure we have values
	if c.repo.branch == "" {
		c.repo.branch = DefaultBranch
	}
	if c.repo.directory == "" {
		c.repo.directory = DefaultDirectory
	}

	return nil
}

func (c *quickstart) runAuthenticationStep() error {
	// Skip auth for plain Git provider
	if c.provider.typ == ProviderGit {
		fmt.Println(HelperTextStyle.Render("No authentication needed for generic Git repositories"))
		return nil
	}

	// Authentication info without card
	fmt.Println(HelperTextStyle.Render(fmt.Sprintf("To access your %s repository, you'll need a Personal Access Token.", c.provider.name)))
	fmt.Println()
	fmt.Println(InputLabelStyle.Render("Required permissions:"))
	fmt.Println(c.getRequiredPermissions())
	fmt.Println()

	// Offer to open browser for PAT creation (if not custom API)
	if !c.provider.isCustom && c.provider.typ != ProviderGitea {
		var openBrowser bool

		browserForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Open browser to create token?").
					Description("We'll open the correct page for creating a Personal Access Token").
					Value(&openBrowser).
					Affirmative("Yes, open browser").
					Negative("No, I have a token"),
			),
		).WithTheme(huh.ThemeCharm())

		if err := browserForm.Run(); err != nil {
			return fmt.Errorf("running browser confirmation form: %w", err)
		}

		if openBrowser {
			patURL := c.getPATCreationURL()
			if patURL != "" {
				if err := util.OpenBrowser(patURL); err != nil {
					fmt.Println(WarningStyle.Render("⚠  Couldn't open browser automatically."))
					fmt.Println()
					fmt.Println(InputLabelStyle.Render("Please visit this URL:"))
					fmt.Println(AccentStyle.Render(patURL))
				}
				fmt.Println()
			}
		}
	}

	// Get token
	tokenForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(InputLabelStyle.Render("Personal Access Token")).
				Description("Paste your access token here (it will be hidden)").
				Value(&c.provider.token).
				EchoMode(huh.EchoModePassword).
				Validate(c.validateToken),
		),
	).WithTheme(huh.ThemeCharm())

	if err := tokenForm.Run(); err != nil {
		return fmt.Errorf("running token input form: %w", err)
	}

	fmt.Println(SuccessStyle.Render("✓ Authentication configured successfully!"))
	fmt.Println()

	return nil
}

// getRequiredPermissions returns the required permissions for each provider
func (c *quickstart) getRequiredPermissions() string {
	var permissions []string

	switch c.provider.typ {
	case ProviderGitHub:
		permissions = []string{
			"  • repo (Full control of private repositories)",
			"  • read:org (Read org and team membership)",
		}
	case ProviderGitLab:
		permissions = []string{
			"  • api (Complete read/write access)",
			"  • read_repository (Read access to repositories)",
		}
	case ProviderBitBucket:
		permissions = []string{
			"  • repository:read (Read repository)",
			"  • repository:write (Write to repository)",
		}
	case ProviderAzure:
		permissions = []string{
			"  • Code (Read & Write)",
			"  • Project and Team (Read)",
		}
	default:
		permissions = []string{
			"  • Read access to repository",
			"  • Write access to repository",
		}
	}

	return lipgloss.NewStyle().Foreground(MutedGray).Render(strings.Join(permissions, "\n"))
}

func (c *quickstart) runReviewStep() error {
	// Create comprehensive review sections
	var sections []string

	// Repository section with single card
	repoSection := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("Repository Configuration"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", LabelStyle.Render("URL:"), ValueStyle.Render(c.repo.url)),
					fmt.Sprintf("%s %s", LabelStyle.Render("Provider:"), ValueStyle.Render(c.provider.name)),
					fmt.Sprintf("%s %s", LabelStyle.Render("Branch:"), ValueStyle.Render(c.repo.branch)),
					fmt.Sprintf("%s %s", LabelStyle.Render("Directory:"), ValueStyle.Render(c.repo.directory)),
				),
			),
		),
	)
	sections = append(sections, repoSection)

	// Add API URL to main section if custom
	if c.provider.isCustom && c.provider.apiURL != "" {
		// Include in the main repository section instead of separate card
		sections[0] = CardStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				SectionHeaderStyle.Render("Repository Configuration"),
				ConfigItemStyle.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						fmt.Sprintf("%s %s", LabelStyle.Render("URL:"), ValueStyle.Render(c.repo.url)),
						fmt.Sprintf("%s %s", LabelStyle.Render("Provider:"), ValueStyle.Render(c.provider.name)),
						fmt.Sprintf("%s %s", LabelStyle.Render("API URL:"), ValueStyle.Render(c.provider.apiURL)),
						fmt.Sprintf("%s %s", LabelStyle.Render("Branch:"), ValueStyle.Render(c.repo.branch)),
						fmt.Sprintf("%s %s", LabelStyle.Render("Directory:"), ValueStyle.Render(c.repo.directory)),
					),
				),
			),
		)
	}

	// Authentication section combined with save location
	var authAndSaveContent []string
	if c.provider.token != "" {
		authAndSaveContent = append(authAndSaveContent,
			SectionHeaderStyle.Render("Authentication"),
			ConfigItemStyle.Render(SuccessStyle.Render("✓ Personal Access Token configured")),
			"",
		)
	}
	authAndSaveContent = append(authAndSaveContent,
		SectionHeaderStyle.Render("Save Location"),
		ConfigItemStyle.Render(ValueStyle.Render(c.configFile)),
	)

	configSection := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, authAndSaveContent...),
	)
	sections = append(sections, configSection)

	// Display all sections
	fmt.Println(lipgloss.JoinVertical(lipgloss.Left, sections...))
	fmt.Println()

	// Confirmation prompt
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Ready to save configuration?").
				Description("Your Flipt configuration will be created with the settings above").
				Value(&confirm).
				Affirmative("Yes, create configuration").
				Negative("No, cancel setup"),
		),
	).WithTheme(huh.ThemeCharm())

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("running configuration review form: %w", err)
	}

	if !confirm {
		fmt.Println(WarningStyle.Render("⚠ Setup cancelled. No changes were made."))
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

// getPATCreationURL returns the URL for creating Personal Access Tokens based on the provider type.
// Returns an empty string for providers that don't have a standard token creation URL.
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

// renderSuccessScreen creates a celebratory success screen
func (c *quickstart) renderSuccessScreen() {
	// Success header
	successHeader := lipgloss.NewStyle().
		Foreground(Green).
		Bold(true).
		Padding(1, 3).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(Green).
		Align(lipgloss.Center).
		Render("Setup Complete!")

	fmt.Println(lipgloss.NewStyle().Width(80).Align(lipgloss.Center).Render(successHeader))
	fmt.Println()

	// Configuration summary
	configSummary := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("Configuration Created"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", LabelStyle.Render("File:"), ValueStyle.Render(c.configFile)),
					fmt.Sprintf("%s %s", LabelStyle.Render("Repository:"), ValueStyle.Render(c.repo.url)),
					fmt.Sprintf("%s %s", LabelStyle.Render("Branch:"), ValueStyle.Render(c.repo.branch)),
				),
			),
		),
	)
	fmt.Println(configSummary)

	// Quick start commands
	commandsCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("Start Server"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					lipgloss.NewStyle().
						Background(CodeBlockBg).
						Foreground(Green).
						Padding(0, 1).
						Render("flipt server"),
					HelperTextStyle.Render("Start the Flipt server with your new configuration"),
				),
			),
		),
	)
	fmt.Println(commandsCard)

	// Next steps
	nextStepsCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("Next Steps"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", SuccessStyle.Render("1."), "Access the Flipt UI at "+AccentStyle.Render("http://localhost:8080")),
					fmt.Sprintf("%s %s", SuccessStyle.Render("2."), "Create your first feature flag"),
					fmt.Sprintf("%s %s", SuccessStyle.Render("3."), "Integrate with your application using our SDKs"),
					fmt.Sprintf("%s %s", SuccessStyle.Render("4."), "Set up your team's workflow with pull requests"),
				),
			),
		),
	)
	fmt.Println(nextStepsCard)

	// Resources
	resourcesCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("Resources"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", LabelStyle.Render("Documentation:"), AccentStyle.Render("https://docs.flipt.io/v2")),
					fmt.Sprintf("%s %s", LabelStyle.Render("Discord:"), AccentStyle.Render("https://flipt.io/discord")),
					fmt.Sprintf("%s %s", LabelStyle.Render("GitHub:"), AccentStyle.Render("https://github.com/flipt-io/flipt")),
				),
			),
		),
	)
	fmt.Println(resourcesCard)

	// Footer
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().
		Foreground(Purple).
		Bold(true).
		Align(lipgloss.Center).
		Width(80).
		Render("Thank you for choosing Flipt!"))
	fmt.Println()
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

	// Show success screen
	c.renderSuccessScreen()

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
