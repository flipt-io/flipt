package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

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
	validator, err := cue.NewFeaturesValidator(v.format)
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

		res, err := validator.Validate(arg, f)
		if err != nil && !errors.Is(err, cue.ErrValidationFailed) {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(res.Errors) > 0 {
			if v.format == jsonFormat {
				_ = json.NewEncoder(os.Stdout).Encode(res)
				os.Exit(v.issueExitCode)
				return
			}

			var sb strings.Builder
			sb.WriteString("❌ Validation failure!\n")

			for _, e := range res.Errors {
				msg := fmt.Sprintf(`
- Message  : %s
	File     : %s
	Line     : %d
	Column   : %d
`, e.Message, e.Location.File, e.Location.Line, e.Location.Column)

				sb.WriteString(msg)
			}

			fmt.Println(sb.String())
			os.Exit(v.issueExitCode)
		}
	}
}
