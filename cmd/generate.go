package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lamlam/regex-crossword/internal/display"
	"github.com/lamlam/regex-crossword/internal/generator"
	"github.com/lamlam/regex-crossword/internal/puzzle"
)

var (
	rows         int
	cols         int
	difficulty   string
	seed         int64
	showSolution bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new regex crossword puzzle",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rows < 1 || rows > 8 {
			return fmt.Errorf("rows must be between 1 and 8, got %d", rows)
		}
		if cols < 1 || cols > 8 {
			return fmt.Errorf("cols must be between 1 and 8, got %d", cols)
		}

		diff, err := puzzle.ParseDifficulty(difficulty)
		if err != nil {
			return err
		}

		opts := generator.Options{
			Rows:       rows,
			Cols:       cols,
			Difficulty: diff,
			Seed:       seed,
		}

		p, err := generator.Generate(opts)
		if err != nil {
			return err
		}

		display.Print(os.Stdout, p, showSolution)
		return nil
	},
}

func init() {
	generateCmd.Flags().IntVarP(&rows, "rows", "r", 3, "Number of rows (1-8)")
	generateCmd.Flags().IntVarP(&cols, "cols", "c", 3, "Number of columns (1-8)")
	generateCmd.Flags().StringVarP(&difficulty, "difficulty", "d", "medium", "Difficulty level (easy, medium, hard)")
	generateCmd.Flags().Int64VarP(&seed, "seed", "s", 0, "Random seed (0 for random)")
	generateCmd.Flags().BoolVar(&showSolution, "show-solution", false, "Show the solution")

	rootCmd.AddCommand(generateCmd)
}
