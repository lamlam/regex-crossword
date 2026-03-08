package display

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/lamlam/regex-crossword/internal/puzzle"
)

// PuzzleJSON is the JSON-serializable representation of a puzzle.
type PuzzleJSON struct {
	Rows       int        `json:"rows"`
	Cols       int        `json:"cols"`
	Difficulty string     `json:"difficulty"`
	Seed       int64      `json:"seed"`
	Alphabet   string     `json:"alphabet"`
	RowHints   []string   `json:"rowHints"`
	ColHints   []string   `json:"colHints"`
	Solution   [][]string `json:"solution,omitempty"`
}

// toPuzzleJSON converts a Puzzle to its JSON representation.
func toPuzzleJSON(p *puzzle.Puzzle) PuzzleJSON {
	solution := make([][]string, p.Rows)
	for r := 0; r < p.Rows; r++ {
		solution[r] = make([]string, p.Cols)
		for c := 0; c < p.Cols; c++ {
			solution[r][c] = string(p.Solution[r][c])
		}
	}

	return PuzzleJSON{
		Rows:       p.Rows,
		Cols:       p.Cols,
		Difficulty: strings.ToLower(p.Difficulty.String()),
		Seed:       p.Seed,
		Alphabet:   p.Alphabet,
		RowHints:   p.RowHints,
		ColHints:   p.ColHints,
		Solution:   solution,
	}
}

// fromPuzzleJSON converts a PuzzleJSON back to a Puzzle.
func fromPuzzleJSON(pj PuzzleJSON) (*puzzle.Puzzle, error) {
	diff, err := puzzle.ParseDifficulty(pj.Difficulty)
	if err != nil {
		return nil, err
	}

	p := &puzzle.Puzzle{
		Rows:       pj.Rows,
		Cols:       pj.Cols,
		Difficulty: diff,
		Seed:       pj.Seed,
		Alphabet:   pj.Alphabet,
		RowHints:   pj.RowHints,
		ColHints:   pj.ColHints,
	}

	if pj.Solution != nil {
		solution := make([][]byte, pj.Rows)
		for r := 0; r < pj.Rows; r++ {
			solution[r] = make([]byte, pj.Cols)
			for c := 0; c < pj.Cols; c++ {
				if len(pj.Solution[r][c]) != 1 {
					return nil, fmt.Errorf("invalid solution cell [%d][%d]: expected single character, got %q", r, c, pj.Solution[r][c])
				}
				solution[r][c] = pj.Solution[r][c][0]
			}
		}
		p.Solution = solution
	}

	return p, nil
}

// ReadJSON reads a JSON-encoded puzzle from the given reader.
func ReadJSON(r io.Reader) (*puzzle.Puzzle, error) {
	var pj PuzzleJSON
	if err := json.NewDecoder(r).Decode(&pj); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return fromPuzzleJSON(pj)
}

// WriteJSON writes the puzzle as indented JSON to the given writer.
// When showSolution is false, the solution field is omitted.
func WriteJSON(w io.Writer, p *puzzle.Puzzle, showSolution bool) error {
	pj := toPuzzleJSON(p)
	if !showSolution {
		pj.Solution = nil
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(pj)
}
