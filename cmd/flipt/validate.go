package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/core/validation"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage/fs"
)

type validateCommand struct {
	issueExitCode int
	format        string
	extraPath     string
	workDirectory string
}

const (
	jsonFormat = "json"
	textFormat = "text"
)

func newValidateCommand() *cobra.Command {
	v := &validateCommand{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate Flipt flag state (.yaml, .yml) files",
		RunE:  v.run,
	}

	cmd.Flags().IntVar(&v.issueExitCode, "issue-exit-code", 1, "exit code to use when issues are found")

	cmd.Flags().StringVarP(
		&v.format,
		"format", "F",
		"text",
		"output format: json, text",
	)

	cmd.Flags().StringVarP(
		&v.extraPath,
		"extra-schema", "e",
		"",
		"path to extra schema constraints",
	)

	cmd.Flags().StringVarP(
		&v.workDirectory,
		"work-dir", "d",
		".",
		"set the working directory",
	)

	return cmd
}

func (v *validateCommand) run(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	logger, _, err := buildConfig(ctx)
	if err != nil {
		return err
	}

	var opts []containers.Option[fs.SnapshotOption]
	if v.extraPath != "" {
		schema, err := os.ReadFile(v.extraPath)
		if err != nil {
			return err
		}

		opts = append(opts, fs.WithValidatorOption(
			validation.WithSchemaExtension(schema),
		))
	}

	if v.workDirectory == "" {
		return errors.New("non-empty working directory expected")
	}

	if len(args) == 0 {
		_, err = fs.SnapshotFromFS(logger, os.DirFS(v.workDirectory), opts...)
	} else {
		_, err = fs.SnapshotFromPaths(logger, os.DirFS(v.workDirectory), args, opts...)
	}

	errs, ok := validation.Unwrap(err)
	if !ok {
		return err
	}

	if len(errs) > 0 {
		if v.format == jsonFormat {
			if err := json.NewEncoder(os.Stdout).Encode(errs); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(v.issueExitCode)
			return nil
		}

		fmt.Println("Validation failed!")

		for _, err := range errs {
			fmt.Printf("%v\n", err)
		}

		os.Exit(v.issueExitCode)
	}

	return nil
}
