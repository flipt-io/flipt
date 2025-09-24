package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	fmt.Println(TitleStyle.Render("Flipt License Check"))
	fmt.Println(SubtitleStyle.Render("Checking your license status"))
	fmt.Println()

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
		noLicenseCard := CardStyle.Copy().
			BorderForeground(Amber).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					WarningStyle.Render("⚠  No License Configured"),
					HelperTextStyle.Render("\nYou are currently using the OSS version of Flipt."),
					HelperTextStyle.Render("Enterprise GitOps integrations and advanced features require a Pro license."),
					"",
					LabelStyle.Render("To activate Pro features, run:"),
					AccentStyle.Render("  flipt license activate"),
				),
			)
		fmt.Println(noLicenseCard)
		return nil
	}

	// Display current license configuration
	configCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("License Configuration"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", LabelStyle.Render("License Key:"), ValueStyle.Render("*** (configured)")),
					func() string {
						if cfg.License.File != "" {
							return fmt.Sprintf("%s %s", LabelStyle.Render("License File:"), ValueStyle.Render(cfg.License.File))
						}
						return fmt.Sprintf("%s %s", LabelStyle.Render("License Type:"), ValueStyle.Render("Online"))
					}(),
				),
			),
		),
	)
	fmt.Println(configCard)

	// Initialize license manager to check validity
	logger := zap.NewNop() // Use a no-op logger for cleaner output
	licenseManagerOpts := []license.LicenseManagerOption{}

	if keygenVerifyKey != "" {
		licenseManagerOpts = append(licenseManagerOpts, license.WithVerificationKey(keygenVerifyKey))
	}

	// Show checking status
	fmt.Println(HelperTextStyle.Render("Checking license validity..."))
	fmt.Println()

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
		validCard := CardStyle.Copy().
			BorderForeground(Green).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					SuccessStyle.Render("✓ License Valid"),
					"",
					SectionHeaderStyle.Render("License Status"),
					ConfigItemStyle.Render(
						lipgloss.JoinVertical(lipgloss.Left,
							fmt.Sprintf("%s %s", LabelStyle.Render("Product:"), ValueStyle.Render("Flipt Pro")),
							fmt.Sprintf("%s %s", LabelStyle.Render("Status:"), SuccessStyle.Render("Active")),
							fmt.Sprintf("%s %s", LabelStyle.Render("Features:"), ValueStyle.Render("All Pro features enabled")),
						),
					),
				),
			)
		fmt.Println(validCard)

		// Next steps
		nextStepsCard := CardStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				SectionHeaderStyle.Render("Pro Features Available"),
				ConfigItemStyle.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						"• Enterprise GitOps Integration (GitHub, GitLab, Bitbucket, Azure DevOps)",
						"• Create merge proposals directly from Flipt UI",
						"• HashiCorp Vault secrets management integration",
						"• GPG commit signing for security and auditability",
						"• Air-gapped environment support (annual license)",
						"• Dedicated Slack support channel with same-day response",
					),
				),
			),
		)
		fmt.Println(nextStepsCard)
	} else {
		invalidCard := CardStyle.Copy().
			BorderForeground(Red).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					ErrorStyle.Render("✗ License Invalid or Expired"),
					"",
					HelperTextStyle.Render("Your license could not be validated. This may be due to:"),
					ConfigItemStyle.Render(
						lipgloss.JoinVertical(lipgloss.Left,
							"• Invalid license key",
							"• Expired license",
							"• Network connectivity issues",
							"• License not activated for this machine",
						),
					),
					"",
					LabelStyle.Render("To activate a new license, run:"),
					AccentStyle.Render("  flipt license activate"),
				),
			)
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
	fmt.Println(TitleStyle.Render("Flipt License Activation"))
	fmt.Println(SubtitleStyle.Render("Activate your Flipt Pro license"))
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
	fmt.Println(HelperTextStyle.Render("Validating license..."))

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
			fmt.Println(HelperTextStyle.Render("Activating license for this machine..."))

			// Activate the license
			if _, err := lic.Activate(ctx, fingerprint); err != nil {
				fmt.Println(ErrorStyle.Render("✗ License activation failed"))
				fmt.Println(LabelStyle.Render("Error: ") + ValueStyle.Render(err.Error()))
				fmt.Println()
				fmt.Println(HelperTextStyle.Render("Please check your license key and try again."))
				fmt.Println(HelperTextStyle.Render("If you continue to have issues, contact support@flipt.io"))
				return nil
			}
			fmt.Println(SuccessStyle.Render("✓ License activated successfully!"))
		} else {
			fmt.Println(ErrorStyle.Render("✗ License validation failed"))
			fmt.Println(LabelStyle.Render("Error: ") + ValueStyle.Render(err.Error()))
			fmt.Println()
			fmt.Println(HelperTextStyle.Render("Please verify your license key is correct."))
			fmt.Println(HelperTextStyle.Render("If you need assistance, contact support@flipt.io"))
			return nil
		}
	} else {
		fmt.Println(SuccessStyle.Render("✓ License validated successfully!"))
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
	successCard := CardStyle.Copy().
		BorderForeground(Green).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				SuccessStyle.Render("✓ License Activation Complete!"),
				"",
				SectionHeaderStyle.Render("License Details"),
				ConfigItemStyle.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						fmt.Sprintf("%s %s", LabelStyle.Render("Type:"), ValueStyle.Render(licenseType)),
						fmt.Sprintf("%s %s", LabelStyle.Render("Status:"), SuccessStyle.Render("Active")),
						fmt.Sprintf("%s %s", LabelStyle.Render("Config:"), ValueStyle.Render(configFile)),
					),
				),
			),
		)
	fmt.Println(successCard)

	// Next steps
	nextStepsCard := CardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			SectionHeaderStyle.Render("Next Steps"),
			ConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", SuccessStyle.Render("1."), "Restart Flipt server to apply the license"),
					fmt.Sprintf("%s %s", SuccessStyle.Render("2."), "Run "+AccentStyle.Render("flipt license check")+" to verify status"),
					fmt.Sprintf("%s %s", SuccessStyle.Render("3."), "Explore Pro features in the documentation"),
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
					fmt.Sprintf("%s %s", LabelStyle.Render("Documentation:"), AccentStyle.Render("https://docs.flipt.io/v2/configuration/licensing")),
					fmt.Sprintf("%s %s", LabelStyle.Render("Support:"), AccentStyle.Render("support@flipt.io")),
				),
			),
		),
	)
	fmt.Println(resourcesCard)

	return nil
}

// Helper function to check if the error is a user interrupt for license commands
func isLicenseInterruptError(err error) bool {
	return errors.Is(err, tea.ErrInterrupted) || errors.Is(err, huh.ErrUserAborted)
}

func newLicenseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license",
		Short: "Manage Flipt Pro license",
		Long: `Manage your Flipt Pro license for enterprise-grade features.

Flipt Pro provides:
  • Enterprise GitOps integration (GitHub, GitLab, Bitbucket, Azure DevOps)
  • HashiCorp Vault secrets management
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
