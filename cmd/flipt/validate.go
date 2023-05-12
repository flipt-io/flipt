package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cue"
)

type validateCommand struct {
	issueExitCode int
	format        string
}

func newValidateCommand() *cobra.Command {
	v := &validateCommand{}

	cmd := &cobra.Command{
		Use:          "validate",
		Short:        "Validate a list of flipt features.yaml files",
		Run:          v.run,
		Hidden:       true,
		SilenceUsage: true,
	}

	cmd.Flags().IntVar(&v.issueExitCode, "issue-exit-code", 1, "Exit code to use when issues are found")

	cmd.Flags().StringVarP(
		&v.format,
		"format", "F",
		"text",
		"output format.",
	)

	return cmd
}

func (v *validateCommand) run(cmd *cobra.Command, args []string) {
	if err := cue.ValidateFiles(os.Stdout, args, v.format); err != nil {
		if errors.Is(err, cue.ErrValidationFailed) && v.issueExitCode != 1 {
			os.Exit(v.issueExitCode)
		}
		os.Exit(1)
	}
}
