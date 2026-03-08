package display

import (
	"bytes"
	"encoding/json"
	"strings"
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
	if err := WriteJSON(&buf, p, true); err != nil {
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

func TestReadJSON(t *testing.T) {
	input := `{
		"rows": 2,
		"cols": 2,
		"difficulty": "medium",
		"seed": 42,
		"alphabet": "ABCD",
		"rowHints": ["^A.$", "^.B$"],
		"colHints": ["^A.$", "^.B$"],
		"solution": [["A", "B"], ["A", "B"]]
	}`
	p, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON returned error: %v", err)
	}

	if p.Rows != 2 || p.Cols != 2 {
		t.Errorf("expected 2x2, got %dx%d", p.Rows, p.Cols)
	}
	if p.Difficulty != puzzle.Medium {
		t.Errorf("expected Medium, got %v", p.Difficulty)
	}
	if p.Seed != 42 {
		t.Errorf("expected seed 42, got %d", p.Seed)
	}
	if p.Alphabet != "ABCD" {
		t.Errorf("expected alphabet \"ABCD\", got %q", p.Alphabet)
	}
	if len(p.RowHints) != 2 || p.RowHints[0] != "^A.$" {
		t.Errorf("unexpected rowHints: %v", p.RowHints)
	}
	if len(p.ColHints) != 2 || p.ColHints[0] != "^A.$" {
		t.Errorf("unexpected colHints: %v", p.ColHints)
	}
	if len(p.Solution) != 2 || len(p.Solution[0]) != 2 {
		t.Fatalf("unexpected solution dimensions")
	}
	if p.Solution[0][0] != 'A' || p.Solution[0][1] != 'B' {
		t.Errorf("unexpected solution row 0: %v", p.Solution[0])
	}
	if p.Solution[1][0] != 'A' || p.Solution[1][1] != 'B' {
		t.Errorf("unexpected solution row 1: %v", p.Solution[1])
	}
}

func TestReadJSON_InvalidJSON(t *testing.T) {
	_, err := ReadJSON(strings.NewReader("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadJSON_InvalidDifficulty(t *testing.T) {
	input := `{"rows":1,"cols":1,"difficulty":"impossible","seed":1,"alphabet":"A","rowHints":["^A$"],"colHints":["^A$"],"solution":[["A"]]}`
	_, err := ReadJSON(strings.NewReader(input))
	if err == nil {
		t.Error("expected error for invalid difficulty")
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	original := testPuzzle()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, original, true); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	restored, err := ReadJSON(&buf)
	if err != nil {
		t.Fatalf("ReadJSON returned error: %v", err)
	}

	if original.Rows != restored.Rows || original.Cols != restored.Cols {
		t.Errorf("dimensions mismatch: %dx%d vs %dx%d", original.Rows, original.Cols, restored.Rows, restored.Cols)
	}
	if original.Difficulty != restored.Difficulty {
		t.Errorf("difficulty mismatch: %v vs %v", original.Difficulty, restored.Difficulty)
	}
	if original.Seed != restored.Seed {
		t.Errorf("seed mismatch: %d vs %d", original.Seed, restored.Seed)
	}
	if original.Alphabet != restored.Alphabet {
		t.Errorf("alphabet mismatch: %q vs %q", original.Alphabet, restored.Alphabet)
	}
	for r := range original.Solution {
		for c := range original.Solution[r] {
			if original.Solution[r][c] != restored.Solution[r][c] {
				t.Errorf("solution[%d][%d] mismatch: %c vs %c", r, c, original.Solution[r][c], restored.Solution[r][c])
			}
		}
	}
}

func TestWriteJSON_WithoutSolution(t *testing.T) {
	p := testPuzzle()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, p, false); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	var got PuzzleJSON
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if got.Rows != 2 || got.Cols != 2 {
		t.Errorf("expected 2x2, got %dx%d", got.Rows, got.Cols)
	}
	if got.Solution != nil {
		t.Errorf("expected nil solution, got %v", got.Solution)
	}

	// Verify "solution" key is not present in raw JSON.
	if strings.Contains(buf.String(), `"solution"`) {
		t.Error("expected solution field to be omitted from JSON output")
	}
}

func TestReadJSON_WithoutSolution(t *testing.T) {
	input := `{
		"rows": 2,
		"cols": 2,
		"difficulty": "medium",
		"seed": 42,
		"alphabet": "ABCD",
		"rowHints": ["^A.$", "^.B$"],
		"colHints": ["^A.$", "^.B$"]
	}`
	p, err := ReadJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadJSON returned error: %v", err)
	}

	if p.Rows != 2 || p.Cols != 2 {
		t.Errorf("expected 2x2, got %dx%d", p.Rows, p.Cols)
	}
	if p.Solution != nil {
		t.Errorf("expected nil solution, got %v", p.Solution)
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
