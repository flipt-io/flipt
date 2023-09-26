package main

import (
	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "completion [SHELL]",
		Short:  "Generate completion scripts",
		Args:   cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				_ = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				_ = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			}

			return nil
		},
	}
}
