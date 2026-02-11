package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "regex-crossword",
	Short: "Generate regex crossword puzzles",
	Long:  "A CLI tool to generate regex crossword puzzles with guaranteed unique solutions.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
