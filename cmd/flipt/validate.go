package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/cue"
	"go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
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
		Use:   "validate",
		Short: "Validate flipt flag state (.yaml, .yml) files",
		Run:   v.run,
	}

	cmd.Flags().IntVar(&v.issueExitCode, "issue-exit-code", 1, "Exit code to use when issues are found")

	cmd.Flags().StringVarP(
		&v.format,
		"format", "F",
		"text",
		"output format: json, text",
	)

	return cmd
}

func (v *validateCommand) run(cmd *cobra.Command, args []string) {
	var err error
	if len(args) == 0 {
		_, err = fs.SnapshotFromFS(zap.NewNop(), os.DirFS("."))
	} else {
		_, err = fs.SnapshotFromPaths(os.DirFS("."), args...)
	}

	errs, ok := cue.Unwrap(err)
	if !ok {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(errs) > 0 {
		if v.format == jsonFormat {
			if err := json.NewEncoder(os.Stdout).Encode(errs); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(v.issueExitCode)
			return
		}

		fmt.Println("Validation failed!")

		for _, err := range errs {
			fmt.Printf("%v\n", err)
		}

		os.Exit(v.issueExitCode)
	}
}
