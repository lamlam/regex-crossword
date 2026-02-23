package display

import (
	"encoding/json"
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
	Solution   [][]string `json:"solution"`
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

// WriteJSON writes the puzzle as indented JSON to the given writer.
func WriteJSON(w io.Writer, p *puzzle.Puzzle) error {
	pj := toPuzzleJSON(p)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(pj)
}
