package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cmd/util"
	"go.flipt.io/flipt/internal/config"
	"gopkg.in/yaml.v2"
)

type initCommand struct {
	force bool
}

func (c *initCommand) run(cmd *cobra.Command, args []string) error {
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
			overwrite, err := util.PromptConfirm("File exists. Overwrite?", false)
			if err != nil {
				return err
			}
			if !overwrite {
				return nil
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	// get path to file, create directory if not exists
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return err
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

	if err := os.WriteFile(file, b.Bytes(), 0600); err != nil {
		return err
	}

	fmt.Printf("Configuration file written to %s\n", file)
	return nil
}

type editCommand struct{}

func (c *editCommand) run(cmd *cobra.Command, args []string) error {
	// TODO: check if no TTY
	file := providedConfigFile

	if file == "" {
		file = userConfigFile
	}

	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
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

	return nil
}

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage Flipt configuration",
		Aliases: []string{"cfg"},
	}

	var (
		initCmd = &initCommand{}
		editCmd = &editCommand{}
	)

	var init = &cobra.Command{
		Use:   "init",
		Short: "Initialize Flipt configuration",
		RunE:  initCmd.run,
	}

	init.Flags().BoolVarP(&initCmd.force, "force", "y", false, "Overwrite existing configuration file")

	var edit = &cobra.Command{
		Use:   "edit",
		Short: "Edit Flipt configuration",
		RunE:  editCmd.run,
	}

	cmd.PersistentFlags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.AddCommand(init)
	cmd.AddCommand(edit)

	return cmd
}
