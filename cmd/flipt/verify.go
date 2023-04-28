package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/cue"
)

// Verify command should take in an argument representing an array/slice of
// filenames of which we will run the cue evaluate against.
//
// We can collect the errors if any in a multierror and return back to the caller.

type verifyCommand struct{}

func newVerifyCommand() *cobra.Command {
	v := &verifyCommand{}

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify cue file against",
		RunE:  v.run,
	}

	return cmd
}

func (v *verifyCommand) run(cmd *cobra.Command, args []string) error {
	fmt.Println(args)

	// Within this function should we get we marshal all the files to get the byte slice
	// and then pass the results to Validate? or should we do something else?

	cue.Validate()
	return nil
}
