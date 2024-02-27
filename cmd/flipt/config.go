package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"gopkg.in/yaml.v2"
)

type initConfigCommand struct {
	force bool
}

func (c *initConfigCommand) run(cmd *cobra.Command, args []string) error {
	defaultFile := providedConfigFile

	if defaultFile == "" {
		defaultFile = userConfigFile
	}

	file := defaultFile

	ack := c.force
	if !ack {
		q := []*survey.Question{
			{
				Name: "file",
				Prompt: &survey.Input{
					Message: "Configuration file path:",
					Default: defaultFile,
				},
				Validate: survey.Required,
			},
		}

		if err := survey.Ask(q, &file); err != nil {
			return err
		}

		// check if file exists
		if _, err := os.Stat(file); err == nil {
			// file exists
			if err := survey.AskOne(&survey.Confirm{
				Message: "Overwrite existing configuration?",
			}, &ack); err != nil {
				return err
			}
			if !ack {
				return nil
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	cfg := config.Default()
	cfg.Version = config.Version // write version for backward compatibility
	return writeConfig(cfg, file)
}

type editConfigCommand struct{}

func (c *editConfigCommand) run(cmd *cobra.Command, args []string) error {
	// TODO: check if no TTY
	file := providedConfigFile

	if file == "" {
		file = userConfigFile
	}

	b, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading config file: %w", err)
	}

	// create default config if not exists
	if os.IsNotExist(err) {
		cfg := config.Default()
		cfg.Version = config.Version // write version for backward compatibility
		out, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		buf.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/main/config/flipt.schema.json\n\n")
		buf.Write(out)

		b = buf.Bytes()
	}

	content := string(b)

	if err := survey.AskOne(&survey.Editor{
		Message:       "Edit configuration",
		Default:       string(b),
		HideDefault:   true,
		AppendDefault: true,
		FileName:      "flipt*.yml",
	}, &content); err != nil {
		return err
	}

	// create if not exists, truncate if exists
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}

	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Printf("Configuration file written to %s\n", file)
	return nil
}

type getConfigCommand struct{}

func (c *getConfigCommand) run(cmd *cobra.Command, args []string) error {
	// TODO: check if no TTY
	file := providedConfigFile

	if file == "" {
		file = userConfigFile
	}

	b, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		fmt.Println("Flipt configuration file not found\nUse `flipt config init` to create a new configuration file")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	content := string(b)

	fmt.Println(content)
	return nil
}

type cloudConfigCommand struct{}

func (c *cloudConfigCommand) run(cmd *cobra.Command, args []string) error {
	// TODO: check if no TTY
	file := providedConfigFile

	if file == "" {
		file = userConfigFile
	}

	var (
		hasExistingConfig = hasConfig()
		setupJWT          = false
		setupWebhooks     = false
		organizationID    string
		apiKey            string
	)

	b, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading config file: %w", err)
	}

	cfg := &config.Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("unmarshalling config: %w", err)
	}

	required := survey.WithValidator(survey.Required)

	// prompt for organization identifier
	if err := survey.AskOne(&survey.Input{
		Message: "Flipt Cloud organization identifier:",
	}, &organizationID, required); err != nil {
		return err
	}

	if organizationID == "" {
		return fmt.Errorf("organization identifier is required")
	}

	if err := survey.AskOne(&survey.Confirm{
		Message: "Enable authentication to Flipt Cloud for this instance?",
	}, &setupJWT); err != nil {
		return err
	}

	if setupJWT {
		// check if JWT is already configured
		for _, method := range cfg.Authentication.Methods.EnabledMethods() {
			if method.Method == auth.Method_METHOD_JWT {
				// JWT is already configured
				setupJWT = false
				fmt.Println("JWT authentication is already configured for this instance, skipping setup..")
				break
			}
		}

		// TODO: set audiences
		var jwtConfig config.AuthenticationMethod[config.AuthenticationMethodJWTConfig]

		jwtConfig.Method = config.AuthenticationMethodJWTConfig{
			JWKSURL: "https://flipt.cloud/api/auth/jwks",
			ValidateClaims: config.AuthenticationMethodJWTValidateClaims{
				Issuer: "https://flipt.cloud",
			},
		}

		cfg.Authentication.Required = true
		cfg.Authentication.Methods.JWT = jwtConfig
	}

	if err := survey.AskOne(&survey.Confirm{
		Message: "Enable audit webhooks for Flipt Cloud for this instance?",
	}, &setupWebhooks); err != nil {
		return err
	}

	if setupWebhooks {
		// prompt for API Key
		if err := survey.AskOne(&survey.Password{
			Message: "Flipt Cloud API key:",
		}, &apiKey, required); err != nil {
			return err
		}

		if apiKey == "" {
			return fmt.Errorf("API key is required")
		}

		if cfg.Audit.Sinks.Webhook.Enabled {
			fmt.Println("Webhooks are already configured for this instance, skipping setup..")
		} else {
			cfg.Audit.Sinks.Webhook.Enabled = true
			// TODO: set default template
			cfg.Audit.Sinks.Webhook.Templates = append(cfg.Audit.Sinks.Webhook.Templates, config.WebhookTemplate{
				URL: "https://flipt.cloud/api/audit/event",
				Headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer " + apiKey,
				},
			})
		}
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	var ack bool
	if hasExistingConfig {
		// prompt if they want to overwrite existing config
		if err := survey.AskOne(&survey.Confirm{
			Message: "Overwrite existing configuration?",
		}, &ack); err != nil {
			return err
		}

		if !ack {
			fmt.Printf("Configuration not updated. Update the configuration file with the following:\n\n")
			fmt.Println(string(out))
			return nil
		}
	}

	return writeConfig(cfg, file)
}

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage Flipt configuration",
		Aliases: []string{"cfg"},
	}

	var (
		initCmd  = &initConfigCommand{}
		editCmd  = &editConfigCommand{}
		getCmd   = &getConfigCommand{}
		cloudCmd = &cloudConfigCommand{}
	)

	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize Flipt configuration",
		RunE:  initCmd.run,
	}

	init.Flags().BoolVarP(&initCmd.force, "force", "y", false, "Overwrite existing configuration file")

	cmd.PersistentFlags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.AddCommand(init)
	cmd.AddCommand(&cobra.Command{
		Use:   "edit",
		Short: "Edit Flipt configuration",
		RunE:  editCmd.run,
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get Flipt configuration",
		RunE:  getCmd.run,
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "cloud",
		Short: "Configure Flipt cloud",
		RunE:  cloudCmd.run,
	})

	return cmd
}

func hasConfig() bool {
	// TODO: check if no TTY
	file := providedConfigFile

	if file == "" {
		file = userConfigFile
	}

	_, err := os.Stat(file)
	if os.IsNotExist(err) || err != nil {
		return false
	}

	return true
}

func writeConfig(cfg *config.Config, file string) error {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	b.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/main/config/flipt.schema.json\n\n")
	b.Write(out)

	// get path to file, create directory if not exists
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return err
	}

	if err := os.WriteFile(file, b.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}
