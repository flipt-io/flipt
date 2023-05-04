package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/cue"
)

type validateCommand struct{}

func newValidateCommand() *cobra.Command {
	v := &validateCommand{}

	cmd := &cobra.Command{
		Use:    "validate",
		Short:  "Validate a list yaml configurations against cue file",
		Run:    v.run,
		Hidden: true,
	}

	return cmd
}

func (v *validateCommand) run(cmd *cobra.Command, args []string) {
	err := cue.ValidateFiles(args)
	if err != nil {
		os.Exit(1)
	}
}
