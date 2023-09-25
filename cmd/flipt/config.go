package main

import (
	"bytes"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

type initCommand struct {
	force bool
}

func (c *initCommand) run(cmd *cobra.Command, args []string) error {
	var (
		file        string
		defaultFile = providedConfigFile
	)

	if defaultFile == "" {
		defaultFile = userConfigFile
	}

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

	overwrite := c.force
	if !overwrite {
		// check if file exists
		if _, err := os.Stat(file); err == nil {
			// file exists
			prompt := &survey.Confirm{
				Message: "File exists. Overwrite?",
			}
			if err := survey.AskOne(prompt, &overwrite); err != nil {
				return err
			}
		}
	}

	cfg := config.Default()
	cfg.Version = config.Version // write version for backward compatibility
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	b.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/main/config/flipt.schema.json\n\n")
	b.Write(out)

	return os.WriteFile(file, b.Bytes(), 0600)
}

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Flipt configuration",
	}

	initCmd := &initCommand{}

	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize Flipt configuration",
		RunE:  initCmd.run,
	}

	init.Flags().BoolVarP(&initCmd.force, "force", "y", false, "Overwrite existing configuration file")

	cmd.PersistentFlags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.AddCommand(init)

	return cmd
}
