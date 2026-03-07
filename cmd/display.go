package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lamlam/regex-crossword/internal/display"
)

var displayShowSolution bool

var displayCmd = &cobra.Command{
	Use:   "display [file]",
	Short: "Display a puzzle from JSON input",
	Long:  "Read a JSON-encoded puzzle from a file or stdin and display it in table format.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := os.Stdin
		if len(args) == 1 {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer f.Close()
			r = f
		}

		p, err := display.ReadJSON(r)
		if err != nil {
			return err
		}

		display.Print(os.Stdout, p, displayShowSolution)
		return nil
	},
}

func init() {
	displayCmd.Flags().BoolVar(&displayShowSolution, "show-solution", false, "Show the solution")

	rootCmd.AddCommand(displayCmd)
}
