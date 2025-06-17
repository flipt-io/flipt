package main

import (
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
)

type licenseCmd struct {
	configFile string
	cfg        *config.Config
}

func newLicenseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "license",
		Short:   "Manage Flipt license",
		Aliases: []string{"lic"},
	}

	var (
		initCmd = &initCommand{}
		editCmd = &editCommand{}
	)

	var init = &cobra.Command{
		Use:   "set",
		Short: "Set Flipt license",
		RunE:  licenseCmd.run,
	}

	var check = &cobra.Command{
		Use:   "check",
		Short: "Check Flipt license",
		RunE:  checkCmd.run,
	}

	cmd.PersistentFlags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.AddCommand(init)
	cmd.AddCommand(check)

	return cmd
}
