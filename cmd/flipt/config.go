package main

import (
	"bytes"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

func newConfigCommand() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage Flipt configuration",
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize Flipt configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
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

			// check if file exists
			if _, err := os.Stat(answer.File); err == nil {
				// file exists
				overwrite := false
				prompt := &survey.Confirm{
					Message: "File exists. Overwrite?",
				}
				if err := survey.AskOne(prompt, &overwrite); err != nil {
					return err
				}
			}

			return os.WriteFile(answer.File, b.Bytes(), 0600)
		},
	}

	configCmd.AddCommand(initCmd)
	return configCmd
}
