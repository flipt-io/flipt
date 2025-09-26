package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/keygen-sh/keygen-go/v3"
	"github.com/keygen-sh/machineid"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/license"
	"go.uber.org/zap"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

// License types
const (
	LicenseTypeProMonthly = "Pro Monthly"
	LicenseTypeProAnnual  = "Pro Annual"
)

// License activation steps
type LicenseStep int

const (
	LicenseStepWelcome LicenseStep = iota
	LicenseStepType
	LicenseStepKey
	LicenseStepOffline
	LicenseStepValidation
	LicenseStepComplete
)

var licenseStepNames = map[LicenseStep]string{
	LicenseStepWelcome:    "Welcome",
	LicenseStepType:       "License Type",
	LicenseStepKey:        "License Key",
	LicenseStepOffline:    "Offline Setup",
	LicenseStepValidation: "Validation",
	LicenseStepComplete:   "Complete",
}

func (s LicenseStep) String() string {
	if name, ok := licenseStepNames[s]; ok {
		return name
	}
	return "Unknown"
}

// Constants are now imported from quickstart.go

// Note: Common styles are imported from styles.go to maintain consistency across commands

// checkCommand handles license checking
type checkCommand struct{}

func (c *checkCommand) availableWidth() int {
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

func (c *checkCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	// Header with responsive width
	width := c.availableWidth()
	fmt.Println(c.heroHeader("Flipt License Center", "Review your subscription configuration", width))

	// Load configuration
	path, _ := determineConfig(providedConfigFile)
	res, err := config.Load(ctx, path)
	if err != nil {
		fmt.Println(ErrorStyle.Render("✗ Failed to load configuration"))
		fmt.Println(LabelStyle.Render("Error: ") + ValueStyle.Render(err.Error()))
		fmt.Println()
		fmt.Println(HelperTextStyle.Render("Please check your configuration file path."))
		return nil
	}

	cfg := res.Config

	// Check if license is configured
	if cfg.License.Key == "" {
		section := &contentSection{
			badge:      BadgeWarnStyle,
			badgeText:  "ACTION REQUIRED",
			heading:    "No License Configured",
			helperText: "You are currently using the OSS edition of Flipt.",
			bulletItems: []string{
				"Enterprise GitOps integration requires Flipt Pro",
				"Access to secrets management and GPG signing",
				"Dedicated Slack support channel",
			},
		}

		callToAction := lipgloss.JoinVertical(lipgloss.Left,
			"",
			LabelStyle.Render("To activate Pro features, run:"),
			lipgloss.NewStyle().MarginLeft(2).Render(AccentStyle.Render("flipt license activate")),
		)

		noLicenseCard := applySectionSpacing(WarningCardStyle.Render(
			section.render() + callToAction,
		))
		fmt.Println(noLicenseCard)
		return nil
	}

	// Display current license configuration using content section
	configDetails := []string{
		renderKeyValue("License Key", ValueStyle.Render("*** (configured)")),
	}

	if cfg.License.File != "" {
		configDetails = append(configDetails,
			renderKeyValue("License Source", ValueStyle.Render("Offline file")),
			renderKeyValue("File Path", AccentStyle.Render(cfg.License.File)),
		)
	} else {
		configDetails = append(configDetails, renderKeyValue("License Source", ValueStyle.Render("Online validation")))
	}

	configSection := &contentSection{
		badge:       BadgeInfoStyle,
		badgeText:   "CONFIG",
		heading:     "License Configuration",
		configItems: configDetails,
	}

	configCard := applySectionSpacing(InfoCardStyle.Render(configSection.render()))
	fmt.Println(configCard)

	// Initialize license manager to check validity
	logger := zap.NewNop() // Use a no-op logger for cleaner output
	licenseManagerOpts := []license.LicenseManagerOption{}

	if keygenVerifyKey != "" {
		licenseManagerOpts = append(licenseManagerOpts, license.WithVerificationKey(keygenVerifyKey))
	}

	// Show checking status with animation feel
	fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "CHECKING", "Validating license...")))

	// Create license manager
	licenseManager, licenseManagerShutdown := license.NewManager(ctx, logger, keygenAccountID, keygenProductID, &cfg.License, licenseManagerOpts...)

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = licenseManagerShutdown(shutdownCtx)
		cancel()
	}()

	// Check the license status
	product := licenseManager.Product()

	if product == "pro" {
		// Valid license status section
		statusSection := &contentSection{
			badge:     BadgeSuccessStyle,
			badgeText: "ACTIVE",
			heading:   "License Status",
			configItems: []string{
				renderKeyValue("Product", ValueStyle.Render("Flipt Pro")),
				renderKeyValue("Status", SuccessStyle.Render("Active")),
				renderKeyValue("Features", ValueStyle.Render("All Pro capabilities unlocked")),
				renderKeyValue("Last Checked", ValueStyle.Render(time.Now().Format("Jan 2, 2006 15:04 MST"))),
			},
		}

		validCard := applySectionSpacing(SuccessCardStyle.Render(statusSection.render()))
		fmt.Println(validCard)

		// Pro features highlights section
		featuresSection := &contentSection{
			badge:     BadgeInfoStyle,
			badgeText: "UNLOCKED",
			heading:   "Pro Feature Highlights",
			bulletItems: []string{
				"Enterprise GitOps integration (GitHub, GitLab, Bitbucket, Azure DevOps)",
				"Merge proposals directly from the Flipt UI",
				"Integrated secrets management",
				"GPG commit signing for security and auditability",
				"Air-gapped environment support (annual license)",
				"Dedicated Slack support channel",
			},
		}

		featuresCard := applySectionSpacing(InfoCardStyle.Render(featuresSection.render()))
		fmt.Println(featuresCard)
	} else {
		// Invalid license section
		invalidSection := &contentSection{
			badge:      BadgeErrorStyle,
			badgeText:  "ATTENTION",
			heading:    "License Invalid or Expired",
			helperText: "Your license could not be validated. This may be due to:",
			bulletItems: []string{
				"Invalid or malformed license key",
				"Expired subscription",
				"Network connectivity issues",
				"License not activated for this machine",
			},
		}

		guidance := lipgloss.JoinVertical(lipgloss.Left,
			"",
			LabelStyle.Render("To activate a new license, run:"),
			lipgloss.NewStyle().MarginLeft(2).Render(AccentStyle.Render("flipt license activate")),
		)

		invalidCard := applySectionSpacing(CardStyle.Copy().
			BorderForeground(Red).
			Render(invalidSection.render() + guidance))
		fmt.Println(invalidCard)
	}

	return nil
}

// activateCommand handles license activation
type activateCommand struct {
	// License information
	licenseType        string
	licenseKey         string
	offlineLicensePath string
	useOffline         bool

	// Configuration
	configFile string
	cfg        *config.Config

	// Internal state
	currentStep LicenseStep
	totalSteps  int
}

func (c *activateCommand) availableWidth() int {
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

type licenseLayout struct {
	activateCommand *activateCommand
	base            huh.Layout
}

func (l *licenseLayout) View(f *huh.Form) string {
	body := l.base.View(f)
	header := l.activateCommand.renderHeader()
	return stack(header, body)
}

func (l *licenseLayout) GroupWidth(f *huh.Form, g *huh.Group, w int) int {
	return l.base.GroupWidth(f, g, w)
}

func (c *activateCommand) newForm(groups ...*huh.Group) *huh.Form {
	form := huh.NewForm(groups...).WithTheme(huh.ThemeCharm())
	form = form.WithLayout(&licenseLayout{
		activateCommand: c,
		base:            huh.LayoutDefault,
	})
	form = form.WithProgramOptions(
		tea.WithOutput(os.Stdout),
		tea.WithAltScreen(),
		tea.WithReportFocus(),
	)
	return form
}

func (c *activateCommand) noteFor(content string) *huh.Note {
	return huh.NewNote().
		Description(content).
		Height(lipgloss.Height(content))
}

func (c *activateCommand) renderHeader() string {
	// Determine title and subtitle based on current step
	var title, subtitle string
	
	switch c.currentStep {
	case LicenseStepWelcome:
		title = "Flipt License Activation"
		subtitle = "Secure your Pro subscription in a few guided steps"
	case LicenseStepType:
		title = "License Type Selection"
		subtitle = "Choose your Flipt Pro subscription plan"
	case LicenseStepKey:
		title = "License Key"
		subtitle = "Enter your Flipt Pro license key"
	case LicenseStepOffline:
		title = "Offline Configuration"
		subtitle = "Configure offline license validation"
	case LicenseStepValidation:
		title = "License Validation"
		subtitle = "Verifying your license with Flipt services"
	case LicenseStepComplete:
		title = "Activation Complete!"
		subtitle = "Your Flipt Pro license is ready to use."
	default:
		title = "Flipt License Activation"
		subtitle = ""
	}
	
	return c.heroHeader(title, subtitle)
}

func (c *activateCommand) runWelcomeStep() error {
	// Welcome content section
	welcomeSection := &contentSection{
		badge:     BadgeInfoStyle,
		badgeText: "OVERVIEW",
		heading:   "What We'll Do",
		bulletItems: []string{
			"Select your license type (Monthly or Annual)",
			"Enter your license key securely",
			"Configure offline validation (Annual licenses only)",
			"Validate and activate your license",
			"Update your Flipt configuration",
		},
	}

	welcomeContent := welcomeSection.render()
	note := c.noteFor(welcomeContent)

	var proceed bool
	confirmGroup := huh.NewGroup(
		note,
		huh.NewConfirm().
			Title("Ready to begin?").
			Description("This process will take about 2 minutes").
			Value(&proceed).
			Affirmative("Yes, let's activate").
			Negative("No, maybe later"),
	)

	form := c.newForm(confirmGroup)
	if err := form.Run(); err != nil {
		return err
	}

	if !proceed {
		return tea.ErrInterrupted
	}

	return nil
}

func (c *activateCommand) runLicenseTypeStep() error {
	// License type comparison section
	comparisonSection := &contentSection{
		badge:     BadgeInfoStyle,
		badgeText: "OPTIONS",
		heading:   "Available License Types",
		configItems: []string{
			SectionHeaderStyle.Render("Pro Monthly"),
			ConfigItemStyle.Render("• Requires continuous internet connectivity"),
			ConfigItemStyle.Render("• Automatic updates and validation"),
			ConfigItemStyle.Render("• Ideal for cloud-native deployments"),
			"",
			SectionHeaderStyle.Render("Pro Annual"),
			ConfigItemStyle.Render("• Supports offline validation"),
			ConfigItemStyle.Render("• Air-gapped environment compatible"),
			ConfigItemStyle.Render("• Cryptographically signed license files"),
		},
	}

	comparisonContent := comparisonSection.render()
	note := c.noteFor(comparisonContent)

	licenseTypeGroup := huh.NewGroup(
		note,
		huh.NewSelect[string]().
			Title("Select License Type").
			Description("Choose the type of license you want to activate").
			Options(
				huh.NewOption(LicenseTypeProMonthly, LicenseTypeProMonthly),
				huh.NewOption(LicenseTypeProAnnual, LicenseTypeProAnnual),
			).
			Value(&c.licenseType),
	)

	form := c.newForm(licenseTypeGroup)
	return form.Run()
}

func (c *activateCommand) runLicenseKeyStep() error {
	// Key format guidance section
	guidanceSection := &contentSection{
		badge:      BadgeInfoStyle,
		badgeText:  "FORMAT",
		heading:    "License Key Information",
		helperText: "Your license key should be in the format XXXXX-XXXXX-XXXXX-XXXXX",
		bulletItems: []string{
			"The key will be hidden as you type for security",
			"Check your purchase confirmation email for the key",
			"Contact support@flipt.io if you can't find your key",
		},
	}

	guidanceContent := guidanceSection.render()
	note := c.noteFor(guidanceContent)

	keyGroup := huh.NewGroup(
		note,
		huh.NewInput().
			Title(InputLabelStyle.Render("License Key")).
			Description("Enter your Flipt Pro license key").
			Placeholder("XXXXX-XXXXX-XXXXX-XXXXX").
			Value(&c.licenseKey).
			EchoMode(huh.EchoModePassword).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("license key is required")
				}
				if len(strings.ReplaceAll(s, "-", "")) < 20 {
					return fmt.Errorf("license key appears too short")
				}
				return nil
			}),
	)

	form := c.newForm(keyGroup)
	return form.Run()
}

func (c *activateCommand) runOfflineStep() error {
	// Offline benefits section
	offlineSection := &contentSection{
		badge:     BadgeInfoStyle,
		badgeText: "OFFLINE",
		heading:   "Offline License Benefits",
		bulletItems: []string{
			"Works in air-gapped environments",
			"No internet connectivity required for validation",
			"Cryptographically signed for security",
			"Ideal for high-security deployments",
		},
	}

	offlineContent := offlineSection.render()
	note := c.noteFor(offlineContent)

	offlineGroup := huh.NewGroup(
		note,
		huh.NewConfirm().
			Title("Use Offline License File?").
			Description("Annual licenses support offline validation using a license file").
			Value(&c.useOffline).
			Affirmative("Yes, configure offline license").
			Negative("No, use online validation"),
	)

	form := c.newForm(offlineGroup)
	if err := form.Run(); err != nil {
		return err
	}

	if c.useOffline {
		pathGroup := huh.NewGroup(
			huh.NewInput().
				Title(InputLabelStyle.Render("License File Path")).
				Description("Enter the path where you saved your offline license file").
				Placeholder("/path/to/license.cert").
				Value(&c.offlineLicensePath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("license file path is required")
					}
					// Check if file exists
					if _, err := os.Stat(s); err != nil {
						return fmt.Errorf("license file not found: %w", err)
					}
					return nil
				}),
		)

		pathForm := c.newForm(pathGroup)
		return pathForm.Run()
	}

	return nil
}

// Note: successScreenSection type is imported from quickstart.go

func (c *activateCommand) createSuccessScreenSections() []successScreenSection {
	// License summary
	licenseSummary := []string{
		renderKeyValue("Type", ValueStyle.Render(c.licenseType)),
		renderKeyValue("Status", SuccessStyle.Render("Active")),
		renderKeyValue("Config", AccentStyle.Render(c.configFile)),
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
		HelperTextStyle.Render("Restart the Flipt server to apply your new license"),
	}

	// Next steps
	nextSteps := []string{
		"Run " + AccentStyle.Render("flipt license check") + " to verify status",
		"Access the Flipt UI at " + AccentStyle.Render("http://localhost:8080"),
		"Explore Pro features in the documentation",
		"Join the Pro support Slack channel",
	}

	// Resources
	resources := []string{
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("Documentation:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("https://docs.flipt.io/v2/configuration/licensing")),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("Support:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("support@flipt.io")),
		),
	}

	return []successScreenSection{
		{BadgeSuccessStyle, "COMPLETE", "License Activation", licenseSummary, nil},
		{BadgeInfoStyle, "RESTART", "Apply Your License", commandContent, nil},
		{BadgeInfoStyle, "NEXT", "What to do next", nil, nextSteps},
		{BadgeInfoStyle, "RESOURCES", "Get Help", resources, nil},
	}
}

func (c *activateCommand) renderSuccessScreen() {
	fmt.Println(c.heroHeader("Activation Complete!", "Your Flipt Pro license is ready to use."))

	sections := c.createSuccessScreenSections()
	for _, section := range sections {
		fmt.Println(applySectionSpacing(InfoCardStyle.Render(section.render())))
	}

	fmt.Println(applySectionSpacing(lipgloss.NewStyle().
		Foreground(Purple).
		Bold(true).
		Align(lipgloss.Center).
		Width(c.availableWidth()).
		Render("Thank you for choosing Flipt Pro!")))
	fmt.Println()
}

func (c *activateCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Check if we're in a TTY session
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("license activation requires an interactive terminal (TTY) session\n" +
			"Please run this command in an interactive terminal")
	}

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	// Initialize wizard state
	c.currentStep = LicenseStepWelcome
	c.totalSteps = 5
	c.configFile = providedConfigFile
	if c.configFile == "" {
		c.configFile = userConfigFile
	}

	// Step 1: Welcome screen
	if err := c.runWelcomeStep(); err != nil {
		if isLicenseInterruptError(err) {
			fmt.Println(HelperTextStyle.Render("Activation cancelled."))
			return nil
		}
		return err
	}

	// Clear screen for next step
	fmt.Print("\033[H\033[2J")
	c.currentStep = LicenseStepType

	// Step 2: Choose license type
	if err := c.runLicenseTypeStep(); err != nil {
		if isLicenseInterruptError(err) {
			fmt.Println(HelperTextStyle.Render("Activation cancelled."))
			return nil
		}
		return err
	}

	// Clear screen for next step
	fmt.Print("\033[H\033[2J")
	c.currentStep = LicenseStepKey

	// Step 3: Get license key
	if err := c.runLicenseKeyStep(); err != nil {
		if isLicenseInterruptError(err) {
			fmt.Println(HelperTextStyle.Render("Activation cancelled."))
			return nil
		}
		return err
	}

	// Step 4: For annual license, optionally get offline license file path
	if c.licenseType == LicenseTypeProAnnual {
		// Clear screen for offline step
		fmt.Print("\033[H\033[2J")
		c.currentStep = LicenseStepOffline

		if err := c.runOfflineStep(); err != nil {
			if isLicenseInterruptError(err) {
				fmt.Println(HelperTextStyle.Render("Activation cancelled."))
				return nil
			}
			return err
		}
	}

	// Step 5: Validate the license
	fmt.Print("\033[H\033[2J")
	c.currentStep = LicenseStepValidation
	fmt.Println(c.heroHeader("License Validation", "Verifying your license with Flipt services"))
	fmt.Println()
	fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "VALIDATING", "Checking your license details...")))

	// Configure Keygen client
	keygen.Account = keygenAccountID
	keygen.Product = keygenProductID
	keygen.LicenseKey = c.licenseKey
	if keygenVerifyKey != "" {
		keygen.PublicKey = keygenVerifyKey
	}

	// Get machine fingerprint
	fingerprint, err := machineid.ProtectedID(keygenProductID)
	if err != nil {
		return fmt.Errorf("getting machine fingerprint: %w", err)
	}

	// Validate and activate license
	lic, err := keygen.Validate(ctx, fingerprint)
	if err != nil {
		// Check if license needs activation
		if errors.Is(err, keygen.ErrLicenseNotActivated) {
			fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "ACTIVATING", "Linking the license to this machine...")))

			// Activate the license
			if _, err := lic.Activate(ctx, fingerprint); err != nil {
				fmt.Println(applySectionSpacing(renderInlineStatus(BadgeErrorStyle, "FAILED", "License activation failed")))
				fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left,
					LabelStyle.Render("Error:"),
					lipgloss.NewStyle().MarginLeft(1).Render(ErrorStyle.Render(err.Error())),
				))
				fmt.Println()
				fmt.Println(HelperTextStyle.Render("Please check your license key and try again."))
				fmt.Println(HelperTextStyle.Render("If you continue to have issues, contact support@flipt.io"))
				return nil
			}
			fmt.Println(applySectionSpacing(renderInlineStatus(BadgeSuccessStyle, "ACTIVATED", "License bound to this machine")))
		} else {
			fmt.Println(applySectionSpacing(renderInlineStatus(BadgeErrorStyle, "FAILED", "License validation failed")))
			fmt.Println(lipgloss.JoinHorizontal(lipgloss.Left,
				LabelStyle.Render("Error:"),
				lipgloss.NewStyle().MarginLeft(1).Render(ErrorStyle.Render(err.Error())),
			))
			fmt.Println()
			fmt.Println(HelperTextStyle.Render("Please verify your license key is correct."))
			fmt.Println(HelperTextStyle.Render("If you need assistance, contact support@flipt.io"))
			return nil
		}
	} else {
		fmt.Println(applySectionSpacing(renderInlineStatus(BadgeSuccessStyle, "VALID", "License validated successfully")))
	}

	// Step 6: Update configuration file
	fmt.Println()
	fmt.Println(HelperTextStyle.Render("Updating configuration..."))

	// Load existing config or create new one
	path, found := determineConfig(c.configFile)
	if found {
		res, err := config.Load(ctx, path)
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}
		c.cfg = res.Config
	} else {
		c.cfg = config.Default()
	}

	// Update license configuration
	c.cfg.License.Key = c.licenseKey
	if c.offlineLicensePath != "" {
		c.cfg.License.File = c.offlineLicensePath
	}

	// Marshal the entire updated configuration
	out, err := yaml.Marshal(c.cfg)
	if err != nil {
		return fmt.Errorf("marshaling configuration: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.configFile), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Add schema comment
	content := yamlSchemaComment + string(out)

	if err := os.WriteFile(c.configFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing configuration file: %w", err)
	}

	// Step 7: Show success screen
	fmt.Print("\033[H\033[2J")
	c.currentStep = LicenseStepComplete
	c.renderSuccessScreen()

	return nil
}

// Helper function to check if the error is a user interrupt for license commands
func isLicenseInterruptError(err error) bool {
	return errors.Is(err, tea.ErrInterrupted) || errors.Is(err, huh.ErrUserAborted)
}

// Note: contentSection and progressInfo types are imported from quickstart.go

// calculateProgress computes progress information for the current step
func (c *activateCommand) calculateProgress() progressInfo {
	currentStepNum := int(c.currentStep)
	displayStepNum := currentStepNum

	// Adjust total steps based on license type and options
	totalSteps := 4 // Default: Welcome, Type, Key, Validation
	if c.licenseType == LicenseTypeProAnnual && c.useOffline {
		totalSteps = 5 // Add offline step
	}

	progressPercent := float64(displayStepNum) / float64(totalSteps)
	if progressPercent > 1.0 {
		progressPercent = 1.0
	}

	return progressInfo{
		currentStep:   currentStepNum,
		displayStep:   displayStepNum,
		totalSteps:    totalSteps,
		progressPct:   progressPercent,
		percentageInt: int(progressPercent * 100),
	}
}

// heroHeader creates a hero header with optional progress bar for activation wizard
func (c *activateCommand) heroHeader(title, subtitle string) string {
	width := c.availableWidth()

	var titleLine string
	var progressLine string

	if c.currentStep != LicenseStepWelcome && c.currentStep != LicenseStepComplete {
		// Calculate step info
		progress := c.calculateProgress()

		// Simple title without step counter
		titleLine = TitleStyle.Copy().Width(width).Align(lipgloss.Center).Render(title)

		// Create a thin progress bar
		progressFilled := int(float64(width) * progress.progressPct)
		if progressFilled > width {
			progressFilled = width
		}

		// Build progress bar with filled and remaining sections
		filledSection := strings.Repeat("▬", progressFilled)
		remainingSection := strings.Repeat("▬", width-progressFilled)

		// Combine filled and remaining sections
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

// heroHeader for check command (no progress needed)
func (c *checkCommand) heroHeader(title, subtitle string, width int) string {
	titleLine := TitleStyle.Copy().Width(width).Align(lipgloss.Center).Render(title)

	// Decorative border
	decorativeBorder := strings.Repeat("━", width)
	progressLine := lipgloss.NewStyle().
		Foreground(PurpleAccent).
		Render(decorativeBorder)

	// Build the header sections
	lines := []string{
		"", // Top spacing
		titleLine,
		lipgloss.NewStyle().MarginTop(1).Render(progressLine),
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

func stack(lines ...string) string {
	if len(lines) == 0 {
		return ""
	}

	var stacked []string
	for _, line := range lines {
		if line == "" {
			stacked = append(stacked, "")
			continue
		}
		if len(stacked) == 0 {
			stacked = append(stacked, line)
			continue
		}
		stacked = append(stacked, lipgloss.NewStyle().Render(line))
	}

	return lipgloss.JoinVertical(lipgloss.Left, stacked...)
}

func renderSectionBadge(badge lipgloss.Style, badgeText, heading string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		badge.Render(badgeText),
		lipgloss.NewStyle().MarginLeft(1).Render(SectionHeaderStyle.Render(heading)),
	)
}

func renderKeyValue(label, value string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		LabelStyle.Render(fmt.Sprintf("%s:", label)),
		lipgloss.NewStyle().MarginLeft(1).Render(value),
	)
}

func renderBulletList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	lineStyle := lipgloss.NewStyle().Foreground(SoftGray).PaddingLeft(2)
	bullets := make([]string, len(items))
	for i, item := range items {
		bullets[i] = lineStyle.Render("› " + item)
	}

	return lipgloss.JoinVertical(lipgloss.Left, bullets...)
}

func renderInlineStatus(badge lipgloss.Style, badgeText, message string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		badge.Render(badgeText),
		HelperTextStyle.Copy().MarginLeft(1).Render(message),
	)
}

func newLicenseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license",
		Short: "Manage Flipt Pro license",
		Long: `Manage your Flipt Pro license for enterprise-grade features.

Flipt Pro provides:
  • Enterprise GitOps integration (GitHub, GitLab, Bitbucket, Azure DevOps)
  • Secrets management
  • GPG commit signing and merge proposals from UI
  • Air-gapped environment support (annual license)
  • Dedicated Slack support channel

License management commands:
  flipt license check              # Check current license status
  flipt license activate           # Activate a new license interactively`,
		Aliases: []string{"lic"},
	}

	checkCmd := &checkCommand{}
	check := &cobra.Command{
		Use:   "check",
		Short: "Check license status and validity",
		Long: `Check the current license configuration and validate it.

This command will:
  • Check if a license is configured in your Flipt configuration
  • Validate the license with the licensing service
  • Display the license status and available features`,
		RunE: checkCmd.run,
	}

	activateCmd := &activateCommand{}
	activate := &cobra.Command{
		Use:   "activate",
		Short: "Activate a new Flipt Pro license",
		Long: `Activate a new Flipt Pro license interactively.

This command will guide you through:
  • Selecting your license type (Pro Monthly or Pro Annual)
  • Entering your license key
  • Optionally configuring an offline license file (for annual licenses)
  • Validating and activating the license
  • Updating your Flipt configuration

Pro Monthly licenses require continuous internet connectivity for validation.
Pro Annual licenses support offline validation using cryptographically signed license files.`,
		RunE: activateCmd.run,
	}

	cmd.PersistentFlags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.AddCommand(check)
	cmd.AddCommand(activate)

	return cmd
}
