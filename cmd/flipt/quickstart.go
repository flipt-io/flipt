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
	"golang.org/x/term"
	"gopkg.in/yaml.v2"

	"go.flipt.io/flipt/internal/cmd/util"
	"go.flipt.io/flipt/internal/config"
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
	// Default configuration values
	DefaultBranch    = "main"
	DefaultDirectory = "flipt"
	DefaultStorage   = "default"
	DefaultEnv       = "default"
	minContentWidth  = 48

	// File system permissions
	ConfigDirPerm  = 0o700
	ConfigFilePerm = 0o600

	// Token creation URLs for supported providers
	GitHubTokenURL    = "https://github.com/settings/tokens/new?description=Flipt%20Access&scopes=repo"
	GitLabTokenURL    = "https://gitlab.com/-/user_settings/personal_access_tokens"
	AzureTokenURL     = "https://dev.azure.com/_usersSettings/tokens"
	BitBucketTokenURL = "https://bitbucket.org/account/settings/app-passwords/"

	// Configuration file metadata
	yamlSchemaComment = "# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/v2/config/flipt.schema.json\n\n"

	// Progress bar display constants
	totalDisplaySteps = 5
	compactBarWidth   = 14
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

// contentSection represents a structured content section with consistent formatting
type contentSection struct {
	badge       lipgloss.Style
	badgeText   string
	heading     string
	helperText  string
	configItems []string
	bulletItems []string
}

// render creates the formatted content section
func (cs *contentSection) render() string {
	var elements []string

	// Add section badge and heading
	elements = append(elements, renderSectionBadge(cs.badge, cs.badgeText, cs.heading))
	elements = append(elements, "")

	// Add helper text if provided
	if cs.helperText != "" {
		elements = append(elements, HelperTextStyle.Render(cs.helperText))
		elements = append(elements, "")
	}

	// Add config items or bullet list
	if len(cs.configItems) > 0 {
		elements = append(elements, ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, cs.configItems...)))
	} else if len(cs.bulletItems) > 0 {
		elements = append(elements, ConfigItemStyle.Render(renderBulletList(cs.bulletItems)))
	}

	elements = append(elements, "")

	return stack(elements...)
}

// isInterruptError checks if the error is a user interrupt (Ctrl+C) for bubbletea/huh
func isInterruptError(err error) bool {
	return errors.Is(err, tea.ErrInterrupted) || errors.Is(err, huh.ErrUserAborted)
}

// validationFunc represents a validation function type
type validationFunc func(string) error

// createURLValidator creates a validation function for URLs with a specific field name
func createURLValidator(fieldName, displayName string) validationFunc {
	return func(value string) error {
		if value == "" {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", displayName)}
		}
		if _, err := url.Parse(value); err != nil {
			return ValidationError{Field: fieldName, Message: "invalid URL format"}
		}
		return nil
	}
}

// createRequiredValidator creates a validation function for required fields
func createRequiredValidator(fieldName, displayName string) validationFunc {
	return func(value string) error {
		if value == "" {
			return ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", displayName)}
		}
		return nil
	}
}

// Validation functions using the factory pattern
func (c *quickstart) validateRepositoryURL(repoURL string) error {
	return createURLValidator("repository_url", "repository URL")(repoURL)
}

func (c *quickstart) validateAPIURL(apiURL string) error {
	return createURLValidator("api_url", "API URL")(apiURL)
}

func (c *quickstart) validateToken(token string) error {
	return createRequiredValidator("access_token", "access token")(token)
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

	// Create the main title line with step counter if not on first or last screen
	var titleLine string
	var progressLine string

	if c.currentStep != StepConfirmation && c.currentStep != StepComplete {
		// Calculate step info
		currentStepNum := int(c.currentStep)
		displayStepNum := currentStepNum
		if currentStepNum == int(StepReview) {
			displayStepNum = 5
		}
		progressPercent := float64(displayStepNum) / 5.0 * 100

		// Simple title without step counter
		titleLine = TitleStyle.Copy().Width(width).Align(lipgloss.Center).Render(title)

		// Create a thin progress bar using thinner Unicode characters
		progressFilled := int(float64(width) * progressPercent / 100.0)
		if progressFilled > width {
			progressFilled = width
		}

		// Build progress bar with filled and remaining sections using thin blocks
		filledSection := strings.Repeat("▬", progressFilled)
		remainingSection := strings.Repeat("▬", width-progressFilled)

		// Combine filled (current color) and remaining (lighter) sections
		progressBar := lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(PurpleAccent).Render(filledSection),
			lipgloss.NewStyle().Foreground(PurpleDark).Render(remainingSection),
		)

		progressLine = progressBar
	} else {
		titleLine = TitleStyle.Copy().Width(width).Align(lipgloss.Center).Render(title)

		// For welcome/complete screens, use a decorative border
		decorativeBorder := strings.Repeat("━", width)
		progressLine = lipgloss.NewStyle().
			Foreground(PurpleAccent).
			Render(decorativeBorder)
	}

	// Build the header sections
	lines := []string{
		"", // Top spacing
		titleLine,
	}

	// Add progress line with some spacing
	if progressLine != "" {
		lines = append(lines,
			lipgloss.NewStyle().MarginTop(1).Render(progressLine),
		)
	}

	// Add subtitle if present
	if subtitle != "" {
		lines = append(lines,
			"", // Spacing
			SubtitleStyle.Copy().Width(width).Align(lipgloss.Center).Render(subtitle),
		)
	}

	// Add bottom spacing
	lines = append(lines, "")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type quickstartLayout struct {
	quickstart *quickstart
	base       huh.Layout
}

func (l *quickstartLayout) View(f *huh.Form) string {
	body := l.base.View(f)
	header := l.quickstart.renderHeader()
	return stack(header, body)
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

// renderProgressIndicator creates a stylish progress bar for the wizard
// progressInfo holds calculated progress information
type progressInfo struct {
	currentStep   int
	displayStep   int
	totalSteps    int
	progressPct   float64
	percentageInt int
}

// calculateProgress computes progress information for the current step
func (c *quickstart) calculateProgress() progressInfo {
	currentStepNum := int(c.currentStep)
	displayStepNum := currentStepNum

	// Review is actually step 5 (index), should show as 5/5
	if currentStepNum == int(StepReview) {
		displayStepNum = totalDisplaySteps
	}

	progressPercent := float64(displayStepNum) / float64(totalDisplaySteps)
	if progressPercent > 1.0 {
		progressPercent = 1.0
	}

	return progressInfo{
		currentStep:   currentStepNum,
		displayStep:   displayStepNum,
		totalSteps:    totalDisplaySteps,
		progressPct:   progressPercent,
		percentageInt: int(progressPercent * 100),
	}
}

// renderProgressIndicator creates a stylish progress bar for the wizard
func (c *quickstart) renderProgressIndicator() string {
	availableWidth := c.availableWidth()
	progress := c.calculateProgress()

	// For narrow terminals, use compact version
	if availableWidth < 50 {
		return c.renderCompactProgressBar(progress)
	}

	return c.renderFullProgressBar(progress, availableWidth)
}

// renderFullProgressBar creates a segmented progress bar for wider terminals
func (c *quickstart) renderFullProgressBar(progress progressInfo, availableWidth int) string {
	// Create a segmented progress bar where each segment represents a step
	barWidth := availableWidth - 20 // Reserve space for labels and counter
	if barWidth < compactBarWidth {
		return c.renderCompactProgressBar(progress)
	}

	// Each step gets equal width in the bar
	stepWidth := barWidth / progress.totalSteps
	remainderWidth := barWidth % progress.totalSteps

	var barElements []string

	// Draw segments for each step
	for i := 1; i <= progress.totalSteps; i++ {
		segmentWidth := stepWidth
		// Distribute remainder pixels across first segments
		if i-1 < remainderWidth {
			segmentWidth++
		}

		barElements = append(barElements, c.createProgressSegment(i, progress.displayStep, segmentWidth))
	}

	progressBar := lipgloss.JoinHorizontal(lipgloss.Left, barElements...)
	counter := ProgressCounterStyle.Render(fmt.Sprintf(" %d/%d ", progress.displayStep, progress.totalSteps))
	percentage := ProgressLabelStyle.Render(fmt.Sprintf("(%d%%)", progress.percentageInt))

	// Combine all elements
	progressLine := lipgloss.JoinHorizontal(lipgloss.Left, progressBar, counter, percentage)

	// Center the progress bar
	return lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(progressLine)
}

// createProgressSegment creates a single segment of the progress bar
func (c *quickstart) createProgressSegment(segmentIndex, currentStep, width int) string {
	var segment string
	switch {
	case segmentIndex < currentStep:
		// Completed: solid fill with completion character
		segment = ProgressBarCompleteStyle.Render(strings.Repeat("█", width))
	case segmentIndex == currentStep:
		// Active: highlighted fill with different character
		segment = ProgressBarActiveStyle.Render(strings.Repeat("█", width))
	default:
		// Pending: light track
		segment = ProgressBarTrackStyle.Render(strings.Repeat("░", width))
	}
	return segment
}

// renderCompactProgressBar creates a compact progress bar for narrow terminals
func (c *quickstart) renderCompactProgressBar(progress progressInfo) string {
	// Create a simple progress bar with blocks
	filledBlocks := int(progress.progressPct * float64(compactBarWidth))

	var barElements []string
	for i := 0; i < compactBarWidth; i++ {
		var element string
		switch {
		case i < filledBlocks:
			element = ProgressBarCompleteStyle.Render("█")
		case i == filledBlocks && progress.currentStep < int(StepReview):
			element = ProgressBarActiveStyle.Render("█")
		default:
			element = ProgressBarTrackStyle.Render("░")
		}
		barElements = append(barElements, element)
	}

	progressBar := lipgloss.JoinHorizontal(lipgloss.Left, barElements...)
	counter := ProgressCounterStyle.Render(fmt.Sprintf(" %d/%d", progress.displayStep, progress.totalSteps))
	percentage := ProgressLabelStyle.Render(fmt.Sprintf(" (%d%%)", progress.percentageInt))

	compactLine := lipgloss.JoinHorizontal(lipgloss.Left, progressBar, counter, percentage)
	return ProgressBarStyle.Render(compactLine)
}

// renderHeader creates the header with title and step counter
func (c *quickstart) renderHeader() string {
	// Only show subtitle on first screen
	var subtitle string
	if c.currentStep == StepConfirmation {
		subtitle = "Configure Git storage syncing with a remote repository"
	}
	return c.heroHeader("Flipt v2 Quickstart", subtitle)
}

// wizardStep represents a step in the wizard with its execution function
type wizardStep struct {
	step       WizardStep
	runFunc    func() error
	needsClear bool
}

// initializeQuickstart sets up the initial quickstart configuration
func (c *quickstart) initializeQuickstart() error {
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

	return nil
}

// showExistingConfigWarning displays a warning if a config file already exists
func (c *quickstart) showExistingConfigWarning() {
	if _, err := os.Stat(c.configFile); err == nil {
		warningContent := stack(
			renderSectionBadge(BadgeWarnStyle, "WARNING", "Existing configuration detected"),
			"",
			HelperTextStyle.Render("This setup will overwrite your existing configuration file:"),
			"",
			ValueStyle.Render(c.configFile),
			"",
			HintStyle.Render("💡 Consider backing up your current configuration before proceeding."),
			"",
		)
		fmt.Println(applySectionSpacing(warningContent))
	}
}

// getWizardSteps returns the sequence of wizard steps to execute
func (c *quickstart) getWizardSteps() []wizardStep {
	return []wizardStep{
		{StepConfirmation, c.runConfirmationStep, false},
		{StepRepository, c.runRepositoryStep, true},
		{StepProvider, c.runProviderStep, true},
		{StepBranchDirectory, c.runBranchDirectoryStep, true},
		{StepAuthentication, c.runAuthenticationStep, true},
		{StepReview, c.runReviewStep, true},
	}
}

// executeWizardSteps runs through all the wizard steps
func (c *quickstart) executeWizardSteps() error {
	steps := c.getWizardSteps()

	for _, step := range steps {
		c.currentStep = step.step
		if step.needsClear {
			fmt.Print("\033[H\033[2J")
		}
		if err := step.runFunc(); err != nil {
			return err
		}
	}

	return nil
}

// run executes the complete quickstart wizard
func (c *quickstart) run() error {
	if err := c.initializeQuickstart(); err != nil {
		return err
	}

	c.showExistingConfigWarning()

	if err := c.executeWizardSteps(); err != nil {
		return err
	}

	// Final step: Complete and write configuration
	c.currentStep = StepComplete
	fmt.Print("\033[H\033[2J")

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

// createRepositoryGuidanceContent creates the guidance content for repository step
func (c *quickstart) createRepositoryGuidanceContent() string {
	section := &contentSection{
		badge:      BadgeInfoStyle,
		badgeText:  "REPOSITORY",
		heading:    "Connect your Git source",
		helperText: "Flipt stores feature flags in Git repositories. Supported formats:",
		bulletItems: []string{
			"HTTPS: https://github.com/your-org/feature-flags.git",
			"SSH: git@github.com:company/flipt-config.git",
			"GitLab: https://gitlab.com/team/config-repo.git",
		},
	}
	return section.render()
}

// createRepositoryDetectedContent creates content showing detected repository information
func (c *quickstart) createRepositoryDetectedContent(owner, name string) string {
	section := &contentSection{
		badge:     BadgeSuccessStyle,
		badgeText: "DETECTED",
		heading:   "Repository Details",
		configItems: []string{
			renderKeyValue("Provider", ValueStyle.Render(c.provider.name)),
			renderKeyValue("Repository", ValueStyle.Render(fmt.Sprintf("%s/%s", owner, name))),
			"",
			HelperTextStyle.Render("✓ Repository information parsed successfully"),
		},
	}
	return section.render()
}

// runRepositoryStep handles the repository configuration step
func (c *quickstart) runRepositoryStep() error {
	guidanceContent := c.createRepositoryGuidanceContent()
	note := c.noteFor(guidanceContent)

	group := huh.NewGroup(
		note,
		huh.NewInput().
			Title(InputLabelStyle.Render("Git Repository URL")).
			Description("Enter the complete URL of your Git repository").
			Placeholder("https://github.com/your-org/flipt-config.git").
			Value(&c.repo.url).
			Validate(c.validateRepositoryURL),
	)

	form := c.newForm(group)
	if err := form.Run(); err != nil {
		return fmt.Errorf("running repository configuration form: %w", err)
	}

	// Parse repository URL and update configuration
	if err := c.parseAndSetRepositoryInfo(); err != nil {
		return fmt.Errorf("parsing repository URL: %w", err)
	}

	// Show detected information
	detectedInfo := c.createRepositoryDetectedContent(c.repo.owner, c.repo.name)
	fmt.Println(applySectionSpacing(detectedInfo))

	return nil
}

// parseAndSetRepositoryInfo parses the repository URL and sets provider/repo information
func (c *quickstart) parseAndSetRepositoryInfo() error {
	providerType, owner, name, err := parseRepositoryURL(c.repo.url)
	if err != nil {
		return err
	}

	c.provider.typ = providerType
	c.provider.name = providerType.String()
	c.repo.owner = owner
	c.repo.name = name

	return nil
}

// createProviderOverviewContent creates content showing detected provider information
func (c *quickstart) createProviderOverviewContent() string {
	section := &contentSection{
		badge:     BadgeInfoStyle,
		badgeText: "PROVIDER",
		heading:   "Detected settings",
		configItems: []string{
			renderKeyValue("Provider", ValueStyle.Render(c.provider.name)),
			renderKeyValue("Repository", ValueStyle.Render(fmt.Sprintf("%s/%s", c.repo.owner, c.repo.name))),
		},
	}
	return section.render()
}

// getProviderOptions returns the available provider options for selection
func (c *quickstart) getProviderOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("GitHub", "GitHub"),
		huh.NewOption("GitLab", "GitLab"),
		huh.NewOption("Bitbucket", "Bitbucket"),
		huh.NewOption("Azure DevOps", "Azure"),
		huh.NewOption("Gitea", "Gitea"),
		huh.NewOption("Generic Git", "Git"),
	}
}

// updateProviderTypeFromName updates the provider type based on the selected name
func (c *quickstart) updateProviderTypeFromName() {
	for provType, name := range providerNames {
		if name == c.provider.name {
			c.provider.typ = provType
			break
		}
	}
}

// runProviderStep handles the provider configuration step
func (c *quickstart) runProviderStep() error {
	if err := c.confirmDetectedProvider(); err != nil {
		return err
	}

	if err := c.handleProviderAPIConfiguration(); err != nil {
		return err
	}

	return nil
}

// confirmDetectedProvider handles provider confirmation or manual selection
func (c *quickstart) confirmDetectedProvider() error {
	var correctProvider bool

	providerOverview := c.createProviderOverviewContent()
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
		return c.selectProviderManually()
	}

	return nil
}

// selectProviderManually allows manual provider selection
func (c *quickstart) selectProviderManually() error {
	providerOptions := c.getProviderOptions()

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

	c.updateProviderTypeFromName()
	return nil
}

// handleProviderAPIConfiguration handles API URL configuration for providers that need it
func (c *quickstart) handleProviderAPIConfiguration() error {
	switch {
	case c.provider.typ.IsHosted():
		return c.handleHostedProviderAPI()
	case c.provider.typ == ProviderGitea:
		return c.handleGiteaAPI()
	default:
		return nil
	}
}

// handleHostedProviderAPI handles API configuration for hosted providers
func (c *quickstart) handleHostedProviderAPI() error {
	var instanceType string = "Cloud version" // Default to cloud

	options := []huh.Option[string]{
		huh.NewOption("Cloud version", "Cloud version"),
		huh.NewOption("Self-hosted/Enterprise", "Self-hosted/Enterprise"),
	}

	customGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Title(fmt.Sprintf("Which %s instance are you using?", c.provider.name)).
			Description("Choose the type of instance you're connecting to").
			Options(options...).
			Value(&instanceType),
	)

	customForm := c.newForm(customGroup)
	if err := customForm.Run(); err != nil {
		return fmt.Errorf("running instance type selection form: %w", err)
	}

	c.provider.isCustom = (instanceType == "Self-hosted/Enterprise")

	if c.provider.isCustom {
		return c.collectAPIURL(fmt.Sprintf("%s API URL", c.provider.name), "https://git.example.com/api/v4")
	}

	return nil
}

// handleGiteaAPI handles API configuration for Gitea (always required)
func (c *quickstart) handleGiteaAPI() error {
	validator := func(s string) error {
		if s == "" {
			return ValidationError{Field: "api_url", Message: "API URL is required for Gitea"}
		}
		return c.validateAPIURL(s)
	}

	return c.collectAPIURLWithValidator("Gitea API URL", "https://gitea.example.com/api/v1", validator)
}

// collectAPIURL collects API URL with standard validation
func (c *quickstart) collectAPIURL(title, placeholder string) error {
	return c.collectAPIURLWithValidator(title, placeholder, c.validateAPIURL)
}

// collectAPIURLWithValidator collects API URL with custom validation
func (c *quickstart) collectAPIURLWithValidator(title, placeholder string, validator validationFunc) error {
	apiGroup := huh.NewGroup(
		huh.NewInput().
			Title(InputLabelStyle.Render(title)).
			Description("Enter the API URL for your instance").
			Placeholder(placeholder).
			Value(&c.provider.apiURL).
			Validate(validator),
	)

	apiForm := c.newForm(apiGroup)
	if err := apiForm.Run(); err != nil {
		return fmt.Errorf("running API URL configuration form: %w", err)
	}

	return nil
}

// createOrganizationGuidanceContent creates guidance content for the branch/directory step
func (c *quickstart) createOrganizationGuidanceContent() string {
	organizationPatterns := []string{
		"By environment: config/dev, config/staging, config/production",
		"By service: auth-service/flags, payment-service/flags",
		"By team: platform-team/features, product-team/features",
		"Simple setup: flipt/ (recommended for getting started)",
	}

	configItems := []string{
		LabelStyle.Render("Common Organization Patterns:"),
		"",
		renderBulletList(organizationPatterns),
	}

	section := &contentSection{
		badge:       BadgeInfoStyle,
		badgeText:   "STORAGE",
		heading:     "Organize your configuration",
		helperText:  "Choose how to organize your feature flag configurations:",
		configItems: configItems,
	}
	return section.render()
}

// runBranchDirectoryStep handles the branch and directory configuration step
func (c *quickstart) runBranchDirectoryStep() error {
	// Set defaults
	c.repo.branch = DefaultBranch
	c.repo.directory = DefaultDirectory

	organizationContent := c.createOrganizationGuidanceContent()
	note := c.noteFor(organizationContent)

	group := huh.NewGroup(
		note,
		huh.NewInput().
			Title(InputLabelStyle.Render("Branch Name")).
			Description("Git branch containing your feature flag configurations").
			Value(&c.repo.branch).
			Placeholder(DefaultBranch),
		huh.NewInput().
			Title(InputLabelStyle.Render("Configuration Directory")).
			Description("Directory path where Flipt will look for configuration files").
			Value(&c.repo.directory).
			Placeholder("flipt (recommended for new setups)"),
	)

	form := c.newForm(group)
	if err := form.Run(); err != nil {
		return fmt.Errorf("running branch and directory configuration form: %w", err)
	}

	// Ensure we have values
	c.ensureRepositoryDefaults()

	return nil
}

// ensureRepositoryDefaults ensures branch and directory have default values if empty
func (c *quickstart) ensureRepositoryDefaults() {
	if c.repo.branch == "" {
		c.repo.branch = DefaultBranch
	}
	if c.repo.directory == "" {
		c.repo.directory = DefaultDirectory
	}
}

// createAuthenticationContent creates the authentication guidance content
func (c *quickstart) createAuthenticationContent() string {
	permissionsBlock := lipgloss.JoinVertical(lipgloss.Left,
		LabelStyle.Render("Required permissions:"),
		ConfigItemStyle.Render(renderBulletList(c.getRequiredPermissions())),
	)

	section := &contentSection{
		badge:       BadgeInfoStyle,
		badgeText:   "AUTH",
		heading:     fmt.Sprintf("%s access token", c.provider.name),
		helperText:  fmt.Sprintf("Generate a Personal Access Token so Flipt can sync with %s.", c.provider.name),
		configItems: []string{permissionsBlock},
	}
	return section.render()
}

// shouldOfferBrowserOpen determines if we should offer to open browser for token creation
func (c *quickstart) shouldOfferBrowserOpen() bool {
	return !c.provider.isCustom && c.provider.typ != ProviderGitea
}

// handleBrowserTokenCreation handles opening browser for token creation if requested
func (c *quickstart) handleBrowserTokenCreation(authContent string) error {
	var openBrowser bool
	contextNote := c.noteFor(authContent)

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
		return c.openTokenCreationURL()
	}

	return nil
}

// openTokenCreationURL opens the browser to the token creation page
func (c *quickstart) openTokenCreationURL() error {
	patURL := c.getPATCreationURL()
	if patURL == "" {
		return nil
	}

	if err := util.OpenBrowser(patURL); err != nil {
		failureMessage := stack(
			renderInlineStatus(BadgeWarnStyle, "BROWSER", "Couldn't open the browser automatically"),
			"",
			lipgloss.JoinHorizontal(lipgloss.Left,
				LabelStyle.Render("Visit:"),
				lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render(patURL)),
			),
		)
		fmt.Println(applySectionSpacing(failureMessage))
	}

	return nil
}

// collectAccessToken collects the access token from the user
func (c *quickstart) collectAccessToken(authContent string) error {
	tokenGroup := huh.NewGroup(
		c.noteFor(authContent),
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

	return nil
}

// runAuthenticationStep handles the authentication configuration step
func (c *quickstart) runAuthenticationStep() error {
	// Skip auth for plain Git provider
	if c.provider.typ == ProviderGit {
		fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "SKIP", "No authentication needed for generic Git repositories")))
		return nil
	}

	authContent := c.createAuthenticationContent()

	// Offer to open browser for PAT creation (if not custom API)
	if c.shouldOfferBrowserOpen() {
		if err := c.handleBrowserTokenCreation(authContent); err != nil {
			return err
		}
	}

	// Collect the access token
	if err := c.collectAccessToken(authContent); err != nil {
		return err
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

// createReviewContent creates the review step content showing what will be created
func (c *quickstart) createReviewContent() string {
	outputLines := []string{
		renderKeyValue("Config", AccentStyle.Render(c.configFile)),
	}

	section := &contentSection{
		badge:       BadgeInfoStyle,
		badgeText:   "OUTPUT",
		heading:     "What we'll create",
		configItems: outputLines,
	}
	return section.render()
}

// runReviewStep handles the review and confirmation step
func (c *quickstart) runReviewStep() error {
	outputSection := c.createReviewContent()
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

// buildEnvironmentConfig configures the default environment
func (c *quickstart) buildEnvironmentConfig() {
	if c.cfg.Environments == nil {
		c.cfg.Environments = make(config.EnvironmentsConfig)
	}

	envConfig := &config.EnvironmentConfig{
		Name:    DefaultEnv,
		Storage: DefaultStorage,
		Default: true,
	}

	if c.repo.directory != "" && c.repo.directory != "." {
		envConfig.Directory = c.repo.directory
	}

	c.cfg.Environments[DefaultEnv] = envConfig
}

// buildSCMConfig configures SCM integration if needed
func (c *quickstart) buildSCMConfig() {
	if c.provider.typ == ProviderGit {
		return
	}

	scmConfig := &config.SCMConfig{
		Type: config.SCMType(strings.ToLower(c.provider.name)),
	}

	if c.provider.apiURL != "" {
		scmConfig.ApiURL = c.provider.apiURL
	}

	if c.provider.token != "" {
		credentialsName := c.getCredentialsName()
		scmConfig.Credentials = &credentialsName
		c.addCredentials(credentialsName)
	}

	c.cfg.Environments[DefaultEnv].SCM = scmConfig
}

// buildStorageConfig configures the storage backend
func (c *quickstart) buildStorageConfig() {
	storageConfig := &config.StorageConfig{
		Remote: c.repo.url,
		Branch: c.repo.branch,
		Backend: config.StorageBackendConfig{
			Type: config.MemoryStorageBackendType,
		},
		PollInterval: 30 * time.Second,
	}

	// Add credentials to storage if set
	if c.provider.token != "" && c.provider.typ != ProviderGit {
		storageConfig.Credentials = c.getCredentialsName()
	}

	c.cfg.Storage = config.StoragesConfig{
		DefaultStorage: storageConfig,
	}
}

// getCredentialsName returns the credentials name for the current provider
func (c *quickstart) getCredentialsName() string {
	return fmt.Sprintf("%s-api", strings.ToLower(c.provider.name))
}

// addCredentials adds the access token to pending credentials
func (c *quickstart) addCredentials(credentialsName string) {
	c.pendingCredentials[credentialsName] = map[string]any{
		"type":         "access_token",
		"access_token": c.provider.token,
	}
}

// buildConfiguration assembles the complete configuration
func (c *quickstart) buildConfiguration() {
	c.buildEnvironmentConfig()
	c.buildSCMConfig()
	c.buildStorageConfig()
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

// successScreenSection represents a section of the success screen
type successScreenSection struct {
	badge       lipgloss.Style
	badgeText   string
	heading     string
	content     []string
	bulletItems []string
}

// render creates the formatted success screen section
func (s *successScreenSection) render() string {
	if len(s.bulletItems) > 0 {
		section := &contentSection{
			badge:       s.badge,
			badgeText:   s.badgeText,
			heading:     s.heading,
			bulletItems: s.bulletItems,
		}
		return section.render()
	}

	section := &contentSection{
		badge:       s.badge,
		badgeText:   s.badgeText,
		heading:     s.heading,
		configItems: s.content,
	}
	return section.render()
}

// createSuccessScreenSections creates all sections for the success screen
func (c *quickstart) createSuccessScreenSections() []successScreenSection {
	// Configuration summary
	configSummary := []string{
		renderKeyValue("Config file", AccentStyle.Render(c.configFile)),
		renderKeyValue("Repository", AccentStyle.Render(c.repo.url)),
		renderKeyValue("Branch", ValueStyle.Render(c.repo.branch)),
	}

	// Command to start server
	command := fmt.Sprintf("flipt server --config %q", c.configFile)
	commandLine := lipgloss.NewStyle().
		Background(CodeBlockBg).
		Foreground(Green).
		Padding(0, 1).
		Render(command)

	commandContent := []string{
		commandLine,
		"",
		HelperTextStyle.Render("Start the Flipt server with your new configuration"),
		HintStyle.Render("The server will run on http://localhost:8080 by default"),
	}

	// Next steps
	nextSteps := []string{
		"Access the Flipt UI at " + AccentStyle.Render("http://localhost:8080"),
		"Create your first feature flag",
		"Integrate with your application using our SDKs",
		"Set up your team's workflow with pull requests",
	}

	// Resources
	resources := []string{
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

	return []successScreenSection{
		{BadgeSuccessStyle, "CONFIG", "Configuration created successfully", configSummary, nil},
		{BadgeInfoStyle, "RUN", "Start the server", commandContent, nil},
		{BadgeInfoStyle, "NEXT", "What to do next", nil, nextSteps},
		{BadgeInfoStyle, "RESOURCES", "Keep exploring", resources, nil},
	}
}

// renderSuccessScreen creates a celebratory success screen
func (c *quickstart) renderSuccessScreen() {
	fmt.Println(c.heroHeader("Setup complete!", "Your Flipt configuration is ready to sync."))

	sections := c.createSuccessScreenSections()
	for _, section := range sections {
		fmt.Println(applySectionSpacing(section.render()))
	}

	fmt.Println(applySectionSpacing(lipgloss.NewStyle().
		Foreground(Purple).
		Bold(true).
		Align(lipgloss.Center).
		Width(c.availableWidth()).
		Render("Thank you for choosing Flipt!")))
	fmt.Println()
}

// prepareConfigOutput prepares the configuration for YAML output
func (c *quickstart) prepareConfigOutput() ([]byte, error) {
	yamlConfig := c.convertConfigToYAML()
	if c.pendingCredentials != nil && len(c.pendingCredentials) > 0 {
		yamlConfig["credentials"] = c.pendingCredentials
	}

	out, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return nil, fmt.Errorf("marshaling configuration to YAML: %w", err)
	}

	// Add schema comment
	content := yamlSchemaComment + string(out)
	return []byte(content), nil
}

// writeConfig creates the configuration file and shows the success screen
func (c *quickstart) writeConfig() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.configFile), ConfigDirPerm); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	content, err := c.prepareConfigOutput()
	if err != nil {
		return err
	}

	if err := os.WriteFile(c.configFile, content, ConfigFilePerm); err != nil {
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
