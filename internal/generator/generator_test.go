package generator

import (
	"regexp"
	"testing"

	"github.com/lamlam/regex-crossword/internal/puzzle"
	"github.com/lamlam/regex-crossword/internal/solver"
)

func TestGenerateSmallPuzzle(t *testing.T) {
	opts := Options{
		Rows:       3,
		Cols:       3,
		Difficulty: puzzle.Easy,
		Seed:       42,
	}

	p, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify dimensions.
	if p.Rows != 3 || p.Cols != 3 {
		t.Errorf("expected 3x3, got %dx%d", p.Rows, p.Cols)
	}

	// Verify solution matches all hints.
	for r := 0; r < p.Rows; r++ {
		re := regexp.MustCompile(p.RowHints[r])
		if !re.Match(p.Solution[r]) {
			t.Errorf("row %d: solution %q doesn't match hint %s", r, string(p.Solution[r]), p.RowHints[r])
		}
	}
	for c := 0; c < p.Cols; c++ {
		colBuf := make([]byte, p.Rows)
		for r := 0; r < p.Rows; r++ {
			colBuf[r] = p.Solution[r][c]
		}
		re := regexp.MustCompile(p.ColHints[c])
		if !re.Match(colBuf) {
			t.Errorf("col %d: solution %q doesn't match hint %s", c, string(colBuf), p.ColHints[c])
		}
	}

	// Verify unique solution.
	if !solver.HasUniqueSolution(p) {
		t.Error("expected unique solution")
	}
}

func TestGenerateSeedReproducibility(t *testing.T) {
	opts := Options{
		Rows:       3,
		Cols:       3,
		Difficulty: puzzle.Easy,
		Seed:       12345,
	}

	p1, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate #1 failed: %v", err)
	}

	p2, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate #2 failed: %v", err)
	}

	// Same seed should produce same puzzle.
	for r := 0; r < p1.Rows; r++ {
		if p1.RowHints[r] != p2.RowHints[r] {
			t.Errorf("row hint %d differs: %q vs %q", r, p1.RowHints[r], p2.RowHints[r])
		}
	}
	for c := 0; c < p1.Cols; c++ {
		if p1.ColHints[c] != p2.ColHints[c] {
			t.Errorf("col hint %d differs: %q vs %q", c, p1.ColHints[c], p2.ColHints[c])
		}
	}
}

func TestGenerateDifficulties(t *testing.T) {
	for _, diff := range []puzzle.Difficulty{puzzle.Easy, puzzle.Medium, puzzle.Hard} {
		t.Run(diff.String(), func(t *testing.T) {
			opts := Options{
				Rows:       3,
				Cols:       3,
				Difficulty: diff,
				Seed:       100,
			}

			p, err := Generate(opts)
			if err != nil {
				t.Fatalf("Generate(%v) failed: %v", diff, err)
			}

			if p.Difficulty != diff {
				t.Errorf("expected difficulty %v, got %v", diff, p.Difficulty)
			}
		})
	}
}
