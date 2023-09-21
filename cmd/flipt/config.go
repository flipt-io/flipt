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
	file  string
	force bool
}

func (c *initCommand) run(cmd *cobra.Command, args []string) error {
	cfg := config.Default()
	cfg.Version = config.Version // write version for backward compatibility
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	b.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/flipt-io/flipt/main/config/flipt.schema.json\n\n")
	b.Write(out)

	answer := struct {
		File string
	}{}

	q := []*survey.Question{
		{
			Name: "file",
			Prompt: &survey.Input{
				Message: "Configuration file path:",
				Default: fliptConfigFile,
			},
			Validate: survey.Required,
		},
	}

	if err := survey.Ask(q, &answer); err != nil {
		return err
	}

	overwrite := c.force
	if !overwrite {
		// check if file exists
		if _, err := os.Stat(answer.File); err == nil {
			// file exists
			prompt := &survey.Confirm{
				Message: "File exists. Overwrite?",
			}
			if err := survey.AskOne(prompt, &overwrite); err != nil {
				return err
			}
		}
	}

	return os.WriteFile(answer.File, b.Bytes(), 0600)
}

func newConfigCommand() *cobra.Command {
	var configCmd = &cobra.Command{
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

	configCmd.AddCommand(init)
	return configCmd
}
