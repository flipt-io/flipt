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

// Define license-specific styles to avoid conflicts
var (
	licPurple     = lipgloss.Color("#7571F9")
	licGreen      = lipgloss.Color("#10B981")
	licAmber      = lipgloss.Color("#F59E0B")
	licRed        = lipgloss.Color("#EF4444")
	licMutedGray  = lipgloss.Color("#9CA3AF")
	licWhite      = lipgloss.Color("#FFFFFF")
	licBorderGray = lipgloss.Color("#4B5563")

	licTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(licPurple).
			Padding(1, 2).
			Align(lipgloss.Center)

	licSubtitleStyle = lipgloss.NewStyle().
				Foreground(licMutedGray).
				Italic(true).
				Align(lipgloss.Center)

	licSuccessStyle = lipgloss.NewStyle().
			Foreground(licGreen).
			Bold(true)

	licWarningStyle = lipgloss.NewStyle().
			Foreground(licAmber).
			Bold(true)

	licErrorStyle = lipgloss.NewStyle().
			Foreground(licRed).
			Bold(true)

	licCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(licBorderGray).
			Padding(1, 2).
			MarginBottom(1)

	licSectionHeaderStyle = lipgloss.NewStyle().
				Foreground(licWhite).
				Bold(true).
				MarginBottom(1)

	licInputLabelStyle = lipgloss.NewStyle().
				Foreground(licWhite).
				Bold(true).
				MarginBottom(1)

	licHelperTextStyle = lipgloss.NewStyle().
				Foreground(licMutedGray).
				Italic(true)

	licAccentStyle = lipgloss.NewStyle().
			Foreground(licPurple).
			Bold(true)

	licLabelStyle = lipgloss.NewStyle().
			Foreground(licMutedGray)

	licValueStyle = lipgloss.NewStyle().
			Foreground(licWhite).
			Bold(true)

	licConfigItemStyle = lipgloss.NewStyle().
				PaddingLeft(2)
)

// checkCommand handles license checking
type checkCommand struct{}

func (c *checkCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	// Header
	fmt.Println(licTitleStyle.Render("Flipt License Check"))
	fmt.Println(licSubtitleStyle.Render("Checking your license status"))
	fmt.Println()

	// Load configuration
	path, _ := determineConfig(providedConfigFile)
	res, err := config.Load(ctx, path)
	if err != nil {
		fmt.Println(licErrorStyle.Render("✗ Failed to load configuration"))
		fmt.Println(licLabelStyle.Render("Error: ") + licValueStyle.Render(err.Error()))
		return fmt.Errorf("loading configuration: %w", err)
	}

	cfg := res.Config

	// Check if license is configured
	if cfg.License.Key == "" {
		noLicenseCard := licCardStyle.Copy().
			BorderForeground(licAmber).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					licWarningStyle.Render("⚠  No License Configured"),
					licHelperTextStyle.Render("\nYou are currently using the OSS version of Flipt."),
					licHelperTextStyle.Render("Pro features are not available."),
					"",
					licLabelStyle.Render("To activate Pro features, run:"),
					licAccentStyle.Render("  flipt license activate"),
				),
			)
		fmt.Println(noLicenseCard)
		return nil
	}

	// Display current license configuration
	configCard := licCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			licSectionHeaderStyle.Render("License Configuration"),
			licConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", licLabelStyle.Render("License Key:"), licValueStyle.Render("*** (configured)")),
					func() string {
						if cfg.License.File != "" {
							return fmt.Sprintf("%s %s", licLabelStyle.Render("License File:"), licValueStyle.Render(cfg.License.File))
						}
						return fmt.Sprintf("%s %s", licLabelStyle.Render("License Type:"), licValueStyle.Render("Online"))
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
	fmt.Println(licHelperTextStyle.Render("Checking license validity with Keygen..."))
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
		validCard := licCardStyle.Copy().
			BorderForeground(licGreen).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					licSuccessStyle.Render("✓ License Valid"),
					"",
					licSectionHeaderStyle.Render("License Status"),
					licConfigItemStyle.Render(
						lipgloss.JoinVertical(lipgloss.Left,
							fmt.Sprintf("%s %s", licLabelStyle.Render("Product:"), licValueStyle.Render("Flipt Pro")),
							fmt.Sprintf("%s %s", licLabelStyle.Render("Status:"), licSuccessStyle.Render("Active")),
							fmt.Sprintf("%s %s", licLabelStyle.Render("Features:"), licValueStyle.Render("All Pro features enabled")),
						),
					),
				),
			)
		fmt.Println(validCard)

		// Next steps
		nextStepsCard := licCardStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				licSectionHeaderStyle.Render("Pro Features Available"),
				licConfigItemStyle.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						"• Vault Secrets Provider",
						"• Advanced Storage Backends",
						"• Enterprise Authentication",
						"• Priority Support",
					),
				),
			),
		)
		fmt.Println(nextStepsCard)
	} else {
		invalidCard := licCardStyle.Copy().
			BorderForeground(licRed).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					licErrorStyle.Render("✗ License Invalid or Expired"),
					"",
					licHelperTextStyle.Render("Your license could not be validated. This may be due to:"),
					licConfigItemStyle.Render(
						lipgloss.JoinVertical(lipgloss.Left,
							"• Invalid license key",
							"• Expired license",
							"• Network connectivity issues",
							"• License not activated for this machine",
						),
					),
					"",
					licLabelStyle.Render("To activate a new license, run:"),
					licAccentStyle.Render("  flipt license activate"),
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
	fmt.Println(licTitleStyle.Render("Flipt License Activation"))
	fmt.Println(licSubtitleStyle.Render("Activate your Flipt Pro license"))
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
			fmt.Println(licHelperTextStyle.Render("Activation cancelled."))
			return nil
		}
		return fmt.Errorf("selecting license type: %w", err)
	}

	// Step 2: Get license key
	var licenseKey string
	keyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(licInputLabelStyle.Render("License Key")).
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
			fmt.Println(licHelperTextStyle.Render("Activation cancelled."))
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
				fmt.Println(licHelperTextStyle.Render("Activation cancelled."))
				return nil
			}
			return fmt.Errorf("selecting offline option: %w", err)
		}

		if useOffline {
			pathForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(licInputLabelStyle.Render("License File Path")).
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
					fmt.Println(licHelperTextStyle.Render("Activation cancelled."))
					return nil
				}
				return fmt.Errorf("entering license file path: %w", err)
			}
		}
	}

	// Step 4: Validate the license with Keygen
	fmt.Println()
	fmt.Println(licHelperTextStyle.Render("Validating license with Keygen..."))

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
			fmt.Println(licHelperTextStyle.Render("Activating license for this machine..."))

			// Activate the license
			if _, err := lic.Activate(ctx, fingerprint); err != nil {
				fmt.Println(licErrorStyle.Render("✗ License activation failed"))
				fmt.Println(licLabelStyle.Render("Error: ") + licValueStyle.Render(err.Error()))
				return fmt.Errorf("activating license: %w", err)
			}
			fmt.Println(licSuccessStyle.Render("✓ License activated successfully!"))
		} else {
			fmt.Println(licErrorStyle.Render("✗ License validation failed"))
			fmt.Println(licLabelStyle.Render("Error: ") + licValueStyle.Render(err.Error()))
			return fmt.Errorf("validating license: %w", err)
		}
	} else {
		fmt.Println(licSuccessStyle.Render("✓ License validated successfully!"))
	}

	// Step 5: Update configuration file
	fmt.Println()
	fmt.Println(licHelperTextStyle.Render("Updating configuration..."))

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

	// Prepare YAML output
	yamlConfig := make(map[string]any)

	// Add existing config sections if they exist
	if len(cfg.Storage) > 0 {
		storage := make(map[string]any)
		for name, s := range cfg.Storage {
			storageMap := map[string]any{
				"backend": map[string]any{
					"type": string(s.Backend.Type),
				},
			}
			if s.Remote != "" {
				storageMap["remote"] = s.Remote
			}
			if s.Branch != "" {
				storageMap["branch"] = s.Branch
			}
			if s.PollInterval > 0 {
				storageMap["poll_interval"] = s.PollInterval.String()
			}
			storage[name] = storageMap
		}
		yamlConfig["storage"] = storage
	}

	if len(cfg.Environments) > 0 {
		environments := make(map[string]any)
		for name, env := range cfg.Environments {
			envMap := map[string]any{
				"name":    env.Name,
				"storage": env.Storage,
				"default": env.Default,
			}
			if env.Directory != "" {
				envMap["directory"] = env.Directory
			}
			environments[name] = envMap
		}
		yamlConfig["environments"] = environments
	}

	// Add license configuration
	licenseConfig := map[string]any{
		"key": cfg.License.Key,
	}
	if cfg.License.File != "" {
		licenseConfig["file"] = cfg.License.File
	}
	yamlConfig["license"] = licenseConfig

	// Write configuration
	out, err := yaml.Marshal(yamlConfig)
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
	successCard := licCardStyle.Copy().
		BorderForeground(licGreen).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				licSuccessStyle.Render("✓ License Activation Complete!"),
				"",
				licSectionHeaderStyle.Render("License Details"),
				licConfigItemStyle.Render(
					lipgloss.JoinVertical(lipgloss.Left,
						fmt.Sprintf("%s %s", licLabelStyle.Render("Type:"), licValueStyle.Render(licenseType)),
						fmt.Sprintf("%s %s", licLabelStyle.Render("Status:"), licSuccessStyle.Render("Active")),
						fmt.Sprintf("%s %s", licLabelStyle.Render("Config:"), licValueStyle.Render(configFile)),
					),
				),
			),
		)
	fmt.Println(successCard)

	// Next steps
	nextStepsCard := licCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			licSectionHeaderStyle.Render("Next Steps"),
			licConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", licSuccessStyle.Render("1."), "Restart Flipt server to apply the license"),
					fmt.Sprintf("%s %s", licSuccessStyle.Render("2."), "Run "+licAccentStyle.Render("flipt license check")+" to verify status"),
					fmt.Sprintf("%s %s", licSuccessStyle.Render("3."), "Explore Pro features in the documentation"),
				),
			),
		),
	)
	fmt.Println(nextStepsCard)

	// Resources
	resourcesCard := licCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			licSectionHeaderStyle.Render("Resources"),
			licConfigItemStyle.Render(
				lipgloss.JoinVertical(lipgloss.Left,
					fmt.Sprintf("%s %s", licLabelStyle.Render("Documentation:"), licAccentStyle.Render("https://docs.flipt.io/v2/configuration/licensing")),
					fmt.Sprintf("%s %s", licLabelStyle.Render("Support:"), licAccentStyle.Render("support@flipt.io")),
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
		Long: `Manage your Flipt Pro license including checking status and activating new licenses.

The license command provides tools to:
  • Check your current license status and validity
  • Activate new Pro Monthly or Pro Annual licenses
  • Configure offline license files for air-gapped environments

Examples:
  flipt license check              # Check current license status
  flipt license activate           # Activate a new license interactively`,
		Aliases: []string{"lic"},
	}

	checkCmd := &checkCommand{}
	check := &cobra.Command{
		Use:   "check",
		Short: "Check license status and validity",
		Long: `Check the current license configuration and validate it with Keygen.

This command will:
  • Check if a license is configured in your Flipt configuration
  • Validate the license with Keygen's licensing service
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
  • Validating and activating the license with Keygen
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
