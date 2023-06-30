package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cue"
)

type validateCommand struct {
	issueExitCode int
	format        string
}

const (
	jsonFormat = "json"
	textFormat = "text"
)

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
		"output format",
	)

	return cmd
}

func (v *validateCommand) run(cmd *cobra.Command, args []string) {
	validator, err := cue.NewValidator(v.format)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, arg := range args {
		f, err := os.ReadFile(arg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		verrs, err := validator.Validate(arg, f)
		if err != nil && !errors.Is(err, cue.ErrValidationFailed) {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(verrs) > 0 {
			for _, e := range verrs {
				fmt.Printf("%s\n", e.Message)
				fmt.Printf("%s: %d:%d\n", e.Location.File, e.Location.Line, e.Location.Column)
			}
			os.Exit(v.issueExitCode)
		}
	}

}
