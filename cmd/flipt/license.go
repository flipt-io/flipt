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
	"gopkg.in/yaml.v2"
)

// License types
const (
	LicenseTypeProMonthly = "Pro Monthly"
	LicenseTypeProAnnual  = "Pro Annual"
)

// Note: Common styles are imported from styles.go to maintain consistency across commands

// checkCommand handles license checking
type checkCommand struct{}

func (c *checkCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	// Header
	fmt.Println(renderHeroHeader("Flipt License Center", "Review your subscription configuration"))

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
		callToAction := lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("To activate Pro features, run:"),
			lipgloss.NewStyle().MarginLeft(2).Render(AccentStyle.Render("flipt license activate")),
		)

		noLicenseCard := applySectionSpacing(CardStyle.Copy().
			BorderForeground(Amber).
			Background(SurfaceMuted).
			Render(
				stack(
					renderSectionBadge(BadgeWarnStyle, "ACTION REQUIRED", "No License Configured"),
					HelperTextStyle.Render("You are currently using the OSS edition of Flipt."),
					HelperTextStyle.Render("Enterprise GitOps, secrets, and automation require a Flipt Pro license."),
					callToAction,
				),
			))
		fmt.Println(noLicenseCard)
		return nil
	}

	// Display current license configuration
	configDetails := []string{
		renderKeyValue("License Key", ValueStyle.Render("*** (configured)")),
	}

	if cfg.License.File != "" {
		configDetails = append(configDetails,
			renderKeyValue("License Source", ValueStyle.Render("Offline file")),
			renderKeyValue("File Path", AccentStyle.Render(cfg.License.File)),
		)
	} else {
		configDetails = append(configDetails, renderKeyValue("License Source", ValueStyle.Render("Online (Keygen)")))
	}

	configCard := applySectionSpacing(CardStyle.Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "CONFIG", "License Configuration"),
			ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, configDetails...)),
		),
	))
	fmt.Println(configCard)

	// Initialize license manager to check validity
	logger := zap.NewNop() // Use a no-op logger for cleaner output
	licenseManagerOpts := []license.LicenseManagerOption{}

	if keygenVerifyKey != "" {
		licenseManagerOpts = append(licenseManagerOpts, license.WithVerificationKey(keygenVerifyKey))
	}

	// Show checking status
	fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "CHECKING", "Validating license with Flipt services...")))

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
		statusDetails := []string{
			renderKeyValue("Product", ValueStyle.Render("Flipt Pro")),
			renderKeyValue("Status", SuccessStyle.Render("Active")),
			renderKeyValue("Features", ValueStyle.Render("All Pro capabilities available")),
		}

		validCard := applySectionSpacing(CardStyle.Copy().
			BorderForeground(Green).
			Background(SurfaceMuted).
			Render(
				stack(
					renderSectionBadge(BadgeSuccessStyle, "ACTIVE", "License Status"),
					ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, statusDetails...)),
				),
			))
		fmt.Println(validCard)

		featuresCard := applySectionSpacing(CardStyle.Render(
			stack(
				renderSectionBadge(BadgeInfoStyle, "UNLOCKED", "Pro Feature Highlights"),
				ConfigItemStyle.Render(renderBulletList([]string{
					"Enterprise GitOps integration (GitHub, GitLab, Bitbucket, Azure DevOps)",
					"Merge proposals directly from the Flipt UI",
					"Integrated secrets management",
					"GPG commit signing for security and auditability",
					"Air-gapped environment support (annual license)",
					"Dedicated Slack support channel",
				})),
			),
		))
		fmt.Println(featuresCard)
	} else {
		guidance := lipgloss.JoinVertical(lipgloss.Left,
			LabelStyle.Render("To activate a new license, run:"),
			lipgloss.NewStyle().MarginLeft(2).Render(AccentStyle.Render("flipt license activate")),
		)

		invalidCard := applySectionSpacing(CardStyle.Copy().
			BorderForeground(Red).
			Background(SurfaceMuted).
			Render(
				stack(
					renderSectionBadge(BadgeErrorStyle, "ATTENTION", "License Invalid or Expired"),
					HelperTextStyle.Render("Your license could not be validated. This may be due to:"),
					ConfigItemStyle.Render(renderBulletList([]string{
						"Invalid or malformed license key",
						"Expired subscription",
						"Network connectivity issues",
						"License not activated for this machine",
					})),
					guidance,
				),
			))
		fmt.Println(invalidCard)
	}

	return nil
}

// activateCommand handles license activation
type activateCommand struct{}

func (c *activateCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Check if we're in a TTY session
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("license activation requires an interactive terminal (TTY) session\n" +
			"Please run this command in an interactive terminal")
	}

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	// Header
	fmt.Println(renderHeroHeader("Flipt License Activation", "Secure your Pro subscription in a few guided steps"))
	fmt.Println()

	// Step 1: Choose license type
	var licenseType string
	licenseTypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select License Type").
				Description("Choose the type of license you want to activate").
				Options(
					huh.NewOption(LicenseTypeProMonthly, LicenseTypeProMonthly),
					huh.NewOption(LicenseTypeProAnnual, LicenseTypeProAnnual),
				).
				Value(&licenseType),
		),
	).WithTheme(huh.ThemeCharm())

	if err := licenseTypeForm.Run(); err != nil {
		if isLicenseInterruptError(err) {
			fmt.Println(HelperTextStyle.Render("Activation cancelled."))
			return nil
		}
		return fmt.Errorf("selecting license type: %w", err)
	}

	// Step 2: Get license key
	var licenseKey string
	keyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(InputLabelStyle.Render("License Key")).
				Description("Enter your Flipt Pro license key").
				Placeholder("XXXXX-XXXXX-XXXXX-XXXXX").
				Value(&licenseKey).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("license key is required")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	if err := keyForm.Run(); err != nil {
		if isLicenseInterruptError(err) {
			fmt.Println(HelperTextStyle.Render("Activation cancelled."))
			return nil
		}
		return fmt.Errorf("entering license key: %w", err)
	}

	// Step 3: For annual license, optionally get offline license file path
	var offlineLicensePath string
	if licenseType == LicenseTypeProAnnual {
		var useOffline bool
		offlineForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Use Offline License File?").
					Description("Annual licenses support offline validation using a license file.\nThis is useful for air-gapped environments.").
					Value(&useOffline).
					Affirmative("Yes, configure offline license").
					Negative("No, use online validation"),
			),
		).WithTheme(huh.ThemeCharm())

		if err := offlineForm.Run(); err != nil {
			if isLicenseInterruptError(err) {
				fmt.Println(HelperTextStyle.Render("Activation cancelled."))
				return nil
			}
			return fmt.Errorf("selecting offline option: %w", err)
		}

		if useOffline {
			pathForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(InputLabelStyle.Render("License File Path")).
						Description("Enter the path where you saved your offline license file").
						Placeholder("/path/to/license.cert").
						Value(&offlineLicensePath).
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
				),
			).WithTheme(huh.ThemeCharm())

			if err := pathForm.Run(); err != nil {
				if isLicenseInterruptError(err) {
					fmt.Println(HelperTextStyle.Render("Activation cancelled."))
					return nil
				}
				return fmt.Errorf("entering license file path: %w", err)
			}
		}
	}

	// Step 4: Validate the license
	fmt.Println()
	fmt.Println(applySectionSpacing(renderInlineStatus(BadgeInfoStyle, "VALIDATING", "Checking your license details...")))

	// Configure Keygen client
	keygen.Account = keygenAccountID
	keygen.Product = keygenProductID
	keygen.LicenseKey = licenseKey
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

	// Step 5: Update configuration file
	fmt.Println()
	fmt.Println(HelperTextStyle.Render("Updating configuration..."))

	// Load existing config or create new one
	configFile := providedConfigFile
	if configFile == "" {
		configFile = userConfigFile
	}

	var cfg *config.Config
	path, found := determineConfig(configFile)
	if found {
		res, err := config.Load(ctx, path)
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}
		cfg = res.Config
	} else {
		cfg = config.Default()
	}

	// Update license configuration
	cfg.License.Key = licenseKey
	if offlineLicensePath != "" {
		cfg.License.File = offlineLicensePath
	}

	// Marshal the entire updated configuration
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling configuration: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Add schema comment
	content := "# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/v2/config/flipt.schema.json\n\n" + string(out)

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing configuration file: %w", err)
	}

	// Success message
	fmt.Println()

	licenseSummary := []string{
		renderKeyValue("Type", ValueStyle.Render(licenseType)),
		renderKeyValue("Status", SuccessStyle.Render("Active")),
		renderKeyValue("Config", AccentStyle.Render(configFile)),
	}

	successCard := applySectionSpacing(CardStyle.Copy().
		BorderForeground(Green).
		Background(SurfaceMuted).
		Render(
			stack(
				renderSectionBadge(BadgeSuccessStyle, "COMPLETE", "License Activation"),
				ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, licenseSummary...)),
			),
		))
	fmt.Println(successCard)

	// Next steps
	nextStepsCard := applySectionSpacing(CardStyle.Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "NEXT", "Post-Activation Checklist"),
			ConfigItemStyle.Render(renderBulletList([]string{
				"Restart the Flipt server to apply the license",
				"Run " + AccentStyle.Render("flipt license check") + " to verify status",
				"Explore Pro features in the documentation",
			})),
		),
	))
	fmt.Println(nextStepsCard)

	// Resources
	resourcesRows := []string{
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("Documentation:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("https://docs.flipt.io/v2/configuration/licensing")),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render("Support:"),
			lipgloss.NewStyle().MarginLeft(1).Render(AccentStyle.Render("support@flipt.io")),
		),
	}

	resourcesCard := applySectionSpacing(CardStyle.Render(
		stack(
			renderSectionBadge(BadgeInfoStyle, "RESOURCES", "Helpful Links"),
			ConfigItemStyle.Render(lipgloss.JoinVertical(lipgloss.Left, resourcesRows...)),
		),
	))
	fmt.Println(resourcesCard)

	return nil
}

// Helper function to check if the error is a user interrupt for license commands
func isLicenseInterruptError(err error) bool {
	return errors.Is(err, tea.ErrInterrupted) || errors.Is(err, huh.ErrUserAborted)
}

func renderHeroHeader(title, subtitle string) string {
	lines := []string{
		TitleStyle.Render(title),
		DividerStyle.Render(strings.Repeat("─", contentWidth)),
	}

	if subtitle != "" {
		lines = append(lines, SubtitleStyle.Render(subtitle))
	}

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
