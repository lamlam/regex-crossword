package puzzle

import "fmt"

// FullAlphabet is the complete set of characters available for puzzles.
const FullAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Difficulty represents the puzzle difficulty level.
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	default:
		return "Unknown"
	}
}

// ParseDifficulty converts a string to a Difficulty.
func ParseDifficulty(s string) (Difficulty, error) {
	switch s {
	case "easy":
		return Easy, nil
	case "medium":
		return Medium, nil
	case "hard":
		return Hard, nil
	default:
		return 0, fmt.Errorf("unknown difficulty: %q (must be easy, medium, or hard)", s)
	}
}

// Puzzle represents a regex crossword puzzle.
type Puzzle struct {
	Rows       int
	Cols       int
	Difficulty Difficulty
	Seed       int64
	// Alphabet is the character set used in this puzzle.
	Alphabet string
	// RowHints contains the regex pattern for each row.
	RowHints []string
	// ColHints contains the regex pattern for each column.
	ColHints []string
	// Solution is the answer grid (row-major order).
	Solution [][]byte
}

// SolutionString returns the solution row at index r as a string.
func (p *Puzzle) SolutionString(r int) string {
	return string(p.Solution[r])
}

// ColString returns the solution column at index c as a string.
func (p *Puzzle) ColString(c int) string {
	buf := make([]byte, p.Rows)
	for r := 0; r < p.Rows; r++ {
		buf[r] = p.Solution[r][c]
	}
	return string(buf)
}

// AlphabetSize returns the recommended alphabet size for the given parameters.
// The alphabet must be small enough that the solver can enumerate matches
// (alphabet^maxDim should be under ~500K for the solver to work efficiently).
func AlphabetSize(rows, cols int, d Difficulty) int {
	maxDim := rows
	if cols > maxDim {
		maxDim = cols
	}

	// Target: alphabet^maxDim <= 500000
	// maxDim=3: alpha<=79, maxDim=4: alpha<=26, maxDim=5: alpha<=13
	// maxDim=6: alpha<=8, maxDim=7: alpha<=7, maxDim=8: alpha<=6
	var base int
	switch {
	case maxDim <= 3:
		base = 8
	case maxDim <= 4:
		base = 7
	case maxDim <= 5:
		base = 7
	case maxDim <= 6:
		base = 6
	case maxDim <= 7:
		base = 6
	default: // 8
		base = 5
	}

	// Adjust for difficulty: easier puzzles use fewer characters.
	// Hard puzzles get more complex patterns but not necessarily more characters,
	// since the solver needs to enumerate all matching strings.
	switch d {
	case Easy:
		if base > 5 {
			base--
		}
	case Hard:
		// Only add a character if it won't make enumeration infeasible.
		if maxDim <= 5 {
			base++
		}
	}

	if base < 4 {
		base = 4
	}
	if base > 10 {
		base = 10
	}
	return base
}
