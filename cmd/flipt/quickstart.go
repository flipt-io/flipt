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
	"golang.org/x/term"
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
	minContentWidth  = 48

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

func (c *quickstart) availableWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		return contentWidth
	}

	usable := width - 4
	if usable < minContentWidth {
		if usable < 1 {
			return 1
		}
		return usable
	}
	if usable > contentWidth {
		return contentWidth
	}
	return usable
}

func (c *quickstart) heroHeader(title, subtitle string) string {
	width := c.availableWidth()

	lines := []string{
		TitleStyle.Copy().Width(width).Render(title),
		DividerStyle.Render(strings.Repeat("─", width)),
	}

	if subtitle != "" {
		lines = append(lines, SubtitleStyle.Copy().Width(width).Render(subtitle))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type quickstartLayout struct {
	quickstart *quickstart
	base       huh.Layout
}

func (l *quickstartLayout) View(f *huh.Form) string {
	body := l.base.View(f)
	return stack(l.quickstart.renderHeader(), body)
}

func (l *quickstartLayout) GroupWidth(f *huh.Form, g *huh.Group, w int) int {
	return l.base.GroupWidth(f, g, w)
}

func (c *quickstart) newForm(groups ...*huh.Group) *huh.Form {
	form := huh.NewForm(groups...).WithTheme(huh.ThemeCharm())
	form = form.WithLayout(&quickstartLayout{
		quickstart: c,
		base:       huh.LayoutDefault,
	})
	form = form.WithProgramOptions(
		tea.WithOutput(os.Stdout),
		tea.WithAltScreen(),
		tea.WithReportFocus(),
	)
	return form
}

func (c *quickstart) noteFor(content string) *huh.Note {
	return huh.NewNote().
		Description(content).
		Height(lipgloss.Height(content))
}

func (c *quickstart) cardStyle() lipgloss.Style {
	return CardStyle.Copy().Width(c.availableWidth())
}

// renderProgressIndicator creates a visual progress bar for the wizard
func (c *quickstart) renderProgressIndicator() string {
	steps := []string{"Welcome", "Repository", "Provider", "Storage", "Auth", "Review", "Complete"}
	var chips []string
	for i, step := range steps {
		var chipStyle lipgloss.Style
		var prefix string

		switch {
		case WizardStep(i) < c.currentStep:
			chipStyle = StepChipCompleteStyle
			prefix = "✓"
		case WizardStep(i) == c.currentStep:
			chipStyle = StepChipActiveStyle
			prefix = fmt.Sprintf("%d", i+1)
		default:
			chipStyle = StepChipInactiveStyle
			prefix = fmt.Sprintf("%d", i+1)
		}

		chips = append(chips, chipStyle.Render(fmt.Sprintf("%s %s", prefix, step)))
	}

	availableWidth := c.availableWidth()
	if availableWidth <= 0 {
		availableWidth = contentWidth
	}

	var (
		rows     [][]string
		current  []string
		rowWidth int
		flushRow = func() {
			if len(current) == 0 {
				return
			}
			rows = append(rows, current)
			current = nil
			rowWidth = 0
		}
	)

	for _, chip := range chips {
		chipWidth := lipgloss.Width(chip)
		if len(current) > 0 && rowWidth+chipWidth > availableWidth {
			flushRow()
		}
		current = append(current, chip)
		rowWidth += chipWidth
	}

	flushRow()

	if len(rows) == 0 {
		return ""
	}

	renderedRows := make([]string, 0, len(rows))
	for _, row := range rows {
		renderedRows = append(renderedRows, lipgloss.JoinHorizontal(lipgloss.Left, row...))
	}

	indicator := lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	return lipgloss.NewStyle().PaddingBottom(0).Render(indicator)
}

// renderHeader creates the header with title and progress
func (c *quickstart) renderHeader() string {
	hero := c.heroHeader("Flipt v2 Quickstart", "Configure Git storage syncing with a remote repository")

	if c.currentStep == StepComplete {
		return hero
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		hero,
		lipgloss.NewStyle().MarginTop(1).Render(c.renderProgressIndicator()),
		"",
	)
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

	// Check for existing config
	if _, err := os.Stat(c.configFile); err == nil {
		warningCard := applySectionSpacing(c.cardStyle().
			BorderForeground(Amber).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					WarningStyle.Render("⚠  Existing Configuration Detected"),
					HelperTextStyle.Render("\nThis will overwrite your existing config file at:"),
					ValueStyle.Render(c.configFile),
				),
			))
		fmt.Println(warningCard)
	}

	// Step 1: Confirmation
	c.currentStep = StepConfirmation
	if err := c.runConfirmationStep(); err != nil {
		return err
	}

	// Step 2: Repository Configuration
	c.currentStep = StepRepository
	fmt.Print("\033[H\033[2J")
	if err := c.runRepositoryStep(); err != nil {
		return err
	}

	// Step 3: Provider Configuration
	c.currentStep = StepProvider
	fmt.Print("\033[H\033[2J")
	if err := c.runProviderStep(); err != nil {
		return err
	}

	// Step 4: Branch and Directory
	c.currentStep = StepBranchDirectory
	fmt.Print("\033[H\033[2J")
	if err := c.runBranchDirectoryStep(); err != nil {
		return err
	}

	// Step 5: Authentication (if needed)
	c.currentStep = StepAuthentication
	fmt.Print("\033[H\033[2J")
	if err := c.runAuthenticationStep(); err != nil {
		return err
	}

	// Step 6: Review and Confirm
	c.currentStep = StepReview
	fmt.Print("\033[H\033[2J")
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

	group := huh.NewGroup(
		huh.NewConfirm().
			Title("Ready to begin?").
			Description("This setup will take about 2 minutes").
			Value(&proceed).
			Affirmative("Yes, let's start").
			Negative("No, maybe later"),
	)

	form := c.newForm(group)

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
	guidanceCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "REPOSITORY", "Connect your Git source"),
			HelperTextStyle.Render("Flipt stores feature flags in Git. Provide an HTTPS or SSH URL."),
			ConfigItemStyle.Render(renderBulletList([]string{
				"https://github.com/your-org/feature-flags.git",
				"https://gitlab.com/team/config-repo.git",
				"git@github.com:company/flipt-config.git",
			})),
		),
	))

	note := c.noteFor(guidanceCard)

	group := huh.NewGroup(
		note,
		huh.NewInput().
			Title(InputLabelStyle.Render("Git Repository URL")).
			Description("Enter the URL of your Git repository (HTTPS or SSH)").
			Placeholder("https://github.com/owner/repo.git").
			Value(&c.repo.url).
			Validate(c.validateRepositoryURL),
	)

	form := c.newForm(group)

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
	detectedCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeSuccessStyle, "DETECTED", "Repository Details"),
			ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				renderKeyValue("Provider", ValueStyle.Render(c.provider.name)),
				renderKeyValue("Repository", ValueStyle.Render(fmt.Sprintf("%s/%s", owner, name))),
			)),
		),
	))
	fmt.Println(detectedCard)

	return nil
}

func (c *quickstart) runProviderStep() error {
	var correctProvider bool

	providerOverview := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "PROVIDER", "Detected settings"),
			ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				renderKeyValue("Provider", ValueStyle.Render(c.provider.name)),
				renderKeyValue("Repository", ValueStyle.Render(fmt.Sprintf("%s/%s", c.repo.owner, c.repo.name))),
			)),
		),
	))

	note := c.noteFor(providerOverview)

	// Confirm detected provider
	confirmGroup := huh.NewGroup(
		note,
		huh.NewConfirm().
			Title("Is this correct?").
			Description(fmt.Sprintf("We detected %s as your Git provider", c.provider.name)).
			Value(&correctProvider).
			Affirmative("Yes, that's correct").
			Negative("No, let me choose"),
	)

	confirmForm := c.newForm(confirmGroup)

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

		selectGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select SCM Provider").
				Description("Choose your source control management provider").
				Options(providerOptions...).
				Value(&c.provider.name),
		)

		selectForm := c.newForm(selectGroup)

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
		customGroup := huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Are you using a self-hosted/enterprise %s instance?", c.provider.name)).
				Value(&c.provider.isCustom).
				Affirmative("Yes, self-hosted").
				Negative("No, cloud version"),
		)

		customForm := c.newForm(customGroup)

		if err := customForm.Run(); err != nil {
			return fmt.Errorf("running custom API configuration form: %w", err)
		}

		if c.provider.isCustom {
			apiGroup := huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("%s API URL", AccentStyle.Render(c.provider.name))).
					Description("Enter the API URL for your instance").
					Placeholder("https://git.example.com/api/v4").
					Value(&c.provider.apiURL).
					Validate(c.validateAPIURL),
			)

			apiForm := c.newForm(apiGroup)

			if err := apiForm.Run(); err != nil {
				return fmt.Errorf("running API URL configuration form: %w", err)
			}
		}
	} else if c.provider.typ == ProviderGitea {
		// Gitea always needs API URL
		apiGroup := huh.NewGroup(
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
		)

		apiForm := c.newForm(apiGroup)

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

	organizationCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "STORAGE", "Organize your configuration"),
			HelperTextStyle.Render("Choose the branch and directory structure Flipt should sync."),
			ConfigItemStyle.Render(renderBulletList([]string{
				"By environment: /dev, /staging, /production",
				"By service: /auth-service, /payment-service",
				"By team: /team-platform, /team-product",
			})),
		),
	))

	note := c.noteFor(organizationCard)

	group := huh.NewGroup(
		note,
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
	)

	form := c.newForm(group)

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
		fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "SKIP", "No authentication needed for generic Git repositories")))
		return nil
	}

	permissionsBlock := lipgloss.JoinVertical(lipgloss.Left,
		LabelStyle.Render("Required permissions:"),
		ConfigItemStyle.Render(renderBulletList(c.getRequiredPermissions())),
	)

	authCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "AUTH", fmt.Sprintf("%s access token", c.provider.name)),
			HelperTextStyle.Render(fmt.Sprintf("Generate a Personal Access Token so Flipt can sync with %s.", c.provider.name)),
			permissionsBlock,
		),
	))
	contextNote := c.noteFor(authCard)

	// Offer to open browser for PAT creation (if not custom API)
	if !c.provider.isCustom && c.provider.typ != ProviderGitea {
		var openBrowser bool

		browserGroup := huh.NewGroup(
			contextNote,
			huh.NewConfirm().
				Title("Open browser to create token?").
				Description("We'll open the correct page for creating a Personal Access Token").
				Value(&openBrowser).
				Affirmative("Yes, open browser").
				Negative("No, I have a token"),
		)

		browserForm := c.newForm(browserGroup)

		if err := browserForm.Run(); err != nil {
			return fmt.Errorf("running browser confirmation form: %w", err)
		}

		if openBrowser {
			patURL := c.getPATCreationURL()
			if patURL != "" {
				if err := util.OpenBrowser(patURL); err != nil {
					failureMessage := stack(
						renderInlineStatus(BadgeWarnStyle, "BROWSER", "Couldn't open the browser automatically"),
						lipgloss.JoinHorizontal(lipgloss.Left,
							LabelStyle.Render("Visit:"),
							lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render(patURL)),
						),
					)
					fmt.Println(applySectionSpacing(failureMessage))
					fmt.Println()
				}
			}
		}
	}

	// Get token
	tokenGroup := huh.NewGroup(
		c.noteFor(authCard),
		huh.NewInput().
			Title(InputLabelStyle.Render("Personal Access Token")).
			Description("Paste your access token here (it will be hidden)").
			Value(&c.provider.token).
			EchoMode(huh.EchoModePassword).
			Validate(c.validateToken),
	)

	tokenForm := c.newForm(tokenGroup)

	if err := tokenForm.Run(); err != nil {
		return fmt.Errorf("running token input form: %w", err)
	}

	fmt.Println(applySectionSpacing(renderInlineStatus(BadgeSuccessStyle, "READY", "Authentication configured")))

	return nil
}

// getRequiredPermissions returns the required permissions for each provider
func (c *quickstart) getRequiredPermissions() []string {
	switch c.provider.typ {
	case ProviderGitHub:
		return []string{
			"repo: full control of private repositories",
			"read:org: read org and team membership",
		}
	case ProviderGitLab:
		return []string{
			"api: complete read/write access",
			"read_repository: read access to repositories",
		}
	case ProviderBitBucket:
		return []string{
			"repository:read",
			"repository:write",
		}
	case ProviderAzure:
		return []string{
			"Code (Read & Write)",
			"Project and Team (Read)",
		}
	default:
		return []string{
			"Read access to repository",
			"Write access to repository",
		}
	}
}

func (c *quickstart) runReviewStep() error {
	// Authentication + save location summary
	var outputLines []string
	if c.provider.token != "" {
		outputLines = append(outputLines, HelperTextStyle.Copy().Foreground(White).Render("Personal access token stored securely."))
	}
	outputLines = append(outputLines, renderKeyValue("Config", AccentStyle.Render(c.configFile)))

	outputSection := applySectionSpacing(c.cardStyle().Render(
		lipgloss.JoinVertical(lipgloss.Left,
			renderSectionBadge(BadgeInfoStyle, "Output", "What we'll create"),
			lipgloss.JoinVertical(lipgloss.Left, outputLines...),
		),
	))

	summaryNote := c.noteFor(outputSection)

	// Confirmation prompt
	var confirm bool
	confirmGroup := huh.NewGroup(
		summaryNote,
		huh.NewConfirm().
			Title("Ready to save configuration?").
			Description("Your Flipt configuration will be created with the settings above").
			Value(&confirm).
			Affirmative("Yes, create configuration").
			Negative("No, cancel setup"),
	)

	confirmForm := c.newForm(confirmGroup)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("running configuration review form: %w", err)
	}

	if !confirm {
		fmt.Println(applySectionSpacing(renderInlineStatus(BadgeWarnStyle, "CANCELLED", "Setup cancelled. No changes were made.")))
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
	fmt.Println(c.heroHeader("Setup complete!", "Your Flipt configuration is ready to sync."))

	summaryRows := []string{
		renderKeyValue("Config file", AccentStyle.Render(c.configFile)),
		renderKeyValue("Repository", AccentStyle.Render(c.repo.url)),
		renderKeyValue("Branch", ValueStyle.Render(c.repo.branch)),
	}

	configSummary := applySectionSpacing(c.cardStyle().
		BorderForeground(Green).
		Background(SurfaceMuted).
		Render(
			stack(
				renderSectionBadge(BadgeSuccessStyle, "CONFIG", "Configuration created"),
				ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, summaryRows...)),
			),
		))
	fmt.Println(configSummary)

	command := fmt.Sprintf("flipt server --config %q", c.configFile)
	commandLine := lipgloss.NewStyle().
		Background(CodeBlockBg).
		Foreground(Green).
		Padding(0, 1).
		Render(command)

	commandsCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "RUN", "Start the server"),
			ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				commandLine,
				HelperTextStyle.Render("Start the Flipt server with your new configuration"),
			)),
		),
	))
	fmt.Println(commandsCard)

	nextSteps := renderBulletList([]string{
		"Access the Flipt UI at " + AccentStyle.Render("http://localhost:8080"),
		"Create your first feature flag",
		"Integrate with your application using our SDKs",
		"Set up your team's workflow with pull requests",
	})

	nextStepsCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "NEXT", "What to do next"),
			ConfigItemStyle.Render(nextSteps),
		),
	))
	fmt.Println(nextStepsCard)

	resourcesRows := []string{
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("Documentation:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("https://docs.flipt.io/v2")),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("Discord:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("https://flipt.io/discord")),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("GitHub:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("https://github.com/flipt-io/flipt")),
		),
	}

	resourcesCard := applySectionSpacing(c.cardStyle().Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "RESOURCES", "Keep exploring"),
			ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, resourcesRows...)),
		),
	))
	fmt.Println(resourcesCard)

	fmt.Println(applySectionSpacing(lipgloss.NewStyle().
		Foreground(Purple).
		Bold(true).
		Align(lipgloss.Center).
		Width(c.availableWidth()).
		Render("Thank you for choosing Flipt!")))
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
