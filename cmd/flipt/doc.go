package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// docCommand represents the documentation command
func newDocCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "doc [path]",
		Short:  "Generate command documentation",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "./tmp/docs"
			if len(args) > 0 {
				path = args[0]
			}
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create docs directory: %w", err)
			}

			if err := doc.GenMarkdownTree(cmd.Root(), path); err != nil {
				return fmt.Errorf("failed to generate docs: %w", err)
			}

			fmt.Printf("Documentation generated in: %s\n", path)

			return nil
		},
	}
}
