package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/cue"
)

type verifyCommand struct{}

func newVerifyCommand() *cobra.Command {
	v := &verifyCommand{}

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify cue file against list of yaml configurations",
		Run:   v.run,
	}

	return cmd
}

func (v *verifyCommand) run(cmd *cobra.Command, args []string) {
	err := cue.ValidateFiles(args)
	if err != nil {
		os.Exit(1)
	}
}
