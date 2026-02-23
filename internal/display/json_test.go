package display

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lamlam/regex-crossword/internal/puzzle"
)

func testPuzzle() *puzzle.Puzzle {
	return &puzzle.Puzzle{
		Rows:       2,
		Cols:       2,
		Difficulty: puzzle.Medium,
		Seed:       42,
		Alphabet:   "ABCD",
		RowHints:   []string{"^A.$", "^.B$"},
		ColHints:   []string{"^A.$", "^.B$"},
		Solution:   [][]byte{{'A', 'B'}, {'A', 'B'}},
	}
}

func TestWriteJSON(t *testing.T) {
	p := testPuzzle()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, p); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	var got PuzzleJSON
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if got.Rows != 2 || got.Cols != 2 {
		t.Errorf("expected 2x2, got %dx%d", got.Rows, got.Cols)
	}
	if got.Difficulty != "medium" {
		t.Errorf("expected difficulty \"medium\", got %q", got.Difficulty)
	}
	if got.Seed != 42 {
		t.Errorf("expected seed 42, got %d", got.Seed)
	}
	if got.Alphabet != "ABCD" {
		t.Errorf("expected alphabet \"ABCD\", got %q", got.Alphabet)
	}
	if len(got.RowHints) != 2 || got.RowHints[0] != "^A.$" {
		t.Errorf("unexpected rowHints: %v", got.RowHints)
	}
	if len(got.ColHints) != 2 || got.ColHints[0] != "^A.$" {
		t.Errorf("unexpected colHints: %v", got.ColHints)
	}
	if len(got.Solution) != 2 || len(got.Solution[0]) != 2 {
		t.Fatalf("unexpected solution dimensions: %v", got.Solution)
	}
	if got.Solution[0][0] != "A" || got.Solution[0][1] != "B" {
		t.Errorf("unexpected solution row 0: %v", got.Solution[0])
	}
	if got.Solution[1][0] != "A" || got.Solution[1][1] != "B" {
		t.Errorf("unexpected solution row 1: %v", got.Solution[1])
	}
}

func TestDifficultyStringConversion(t *testing.T) {
	tests := []struct {
		diff puzzle.Difficulty
		want string
	}{
		{puzzle.Easy, "easy"},
		{puzzle.Medium, "medium"},
		{puzzle.Hard, "hard"},
	}
	for _, tt := range tests {
		p := &puzzle.Puzzle{
			Rows:       1,
			Cols:       1,
			Difficulty: tt.diff,
			Alphabet:   "A",
			RowHints:   []string{"^A$"},
			ColHints:   []string{"^A$"},
			Solution:   [][]byte{{'A'}},
		}
		pj := toPuzzleJSON(p)
		if pj.Difficulty != tt.want {
			t.Errorf("difficulty %v: got %q, want %q", tt.diff, pj.Difficulty, tt.want)
		}
	}
}

func TestSolutionConversion(t *testing.T) {
	p := &puzzle.Puzzle{
		Rows:       2,
		Cols:       3,
		Difficulty: puzzle.Easy,
		Alphabet:   "ABCDEF",
		RowHints:   []string{"^...$", "^...$"},
		ColHints:   []string{"^..$", "^..$", "^..$"},
		Solution:   [][]byte{{'A', 'B', 'C'}, {'D', 'E', 'F'}},
	}
	pj := toPuzzleJSON(p)
	expected := [][]string{{"A", "B", "C"}, {"D", "E", "F"}}
	for r := range expected {
		for c := range expected[r] {
			if pj.Solution[r][c] != expected[r][c] {
				t.Errorf("solution[%d][%d]: got %q, want %q", r, c, pj.Solution[r][c], expected[r][c])
			}
		}
	}
}
