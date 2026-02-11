package solver

import (
	"testing"

	"github.com/lamlam/regex-crossword/internal/puzzle"
)

func TestSolverUniqueSolution(t *testing.T) {
	// A simple 2x2 puzzle with a known unique solution: AB / CD
	p := &puzzle.Puzzle{
		Rows:     2,
		Cols:     2,
		Alphabet: "ABCD",
		RowHints: []string{"^AB$", "^CD$"},
		ColHints: []string{"^AC$", "^BD$"},
		Solution: [][]byte{{'A', 'B'}, {'C', 'D'}},
	}

	if !HasUniqueSolution(p) {
		t.Error("expected unique solution for trivial puzzle")
	}
}

func TestSolverMultipleSolutions(t *testing.T) {
	// A puzzle with multiple solutions.
	p := &puzzle.Puzzle{
		Rows:     2,
		Cols:     2,
		Alphabet: "AB",
		RowHints: []string{"^..$", "^..$"},
		ColHints: []string{"^..$", "^..$"},
		Solution: [][]byte{{'A', 'B'}, {'A', 'B'}},
	}

	s := New(p)
	result := s.Solve()
	if result.Unique {
		t.Error("expected multiple solutions for unconstrained puzzle")
	}
	if len(result.Solutions) < 2 {
		t.Errorf("expected at least 2 solutions, got %d", len(result.Solutions))
	}
}

func TestSolverNoSolution(t *testing.T) {
	// A puzzle with contradictory constraints.
	p := &puzzle.Puzzle{
		Rows:     2,
		Cols:     2,
		Alphabet: "AB",
		RowHints: []string{"^AA$", "^BB$"},
		ColHints: []string{"^AB$", "^BA$"}, // Row says AA/BB, col says AB/BA - contradictory
		Solution: [][]byte{{'A', 'A'}, {'B', 'B'}},
	}

	s := New(p)
	result := s.Solve()
	// This should have 0 solutions since rows demand AA,BB but cols demand AB,BA
	// which means col0=AB (A then B - ok) and col1=AB (A then B - but row0 is AA so col1[0]=A, row1 is BB so col1[1]=B)
	// Actually: row0=AA, row1=BB => grid is [[A,A],[B,B]]
	// col0=AB matches ^AB$, col1=AB matches ^BA$? No, AB doesn't match ^BA$.
	// So no solution exists.
	if len(result.Solutions) != 0 {
		t.Errorf("expected no solutions, got %d", len(result.Solutions))
	}
}

func TestSolverFindsCorrectSolution(t *testing.T) {
	p := &puzzle.Puzzle{
		Rows:     3,
		Cols:     3,
		Alphabet: "ABCDE",
		RowHints: []string{"^A[BC]D$", "^E[AB]C$", "^B[DE]A$"},
		ColHints: []string{"^AEB$", "^[BC][AB][DE]$", "^DCA$"},
		Solution: [][]byte{{'A', 'B', 'D'}, {'E', 'A', 'C'}, {'B', 'D', 'A'}},
	}

	s := New(p)
	result := s.Solve()

	if len(result.Solutions) == 0 {
		t.Fatal("expected at least one solution")
	}

	sol := result.Solutions[0]
	expected := []string{"ABD", "EAC", "BDA"}
	for r, exp := range expected {
		if string(sol[r]) != exp {
			t.Errorf("row %d: got %q, want %q", r, string(sol[r]), exp)
		}
	}
}

func BenchmarkSolver3x3(b *testing.B) {
	p := &puzzle.Puzzle{
		Rows:     3,
		Cols:     3,
		Alphabet: "ABCDE",
		RowHints: []string{"^A[BC]D$", "^E[AB]C$", "^B[DE]A$"},
		ColHints: []string{"^AEB$", "^[BC][AB][DE]$", "^DCA$"},
		Solution: [][]byte{{'A', 'B', 'D'}, {'E', 'A', 'C'}, {'B', 'D', 'A'}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := New(p)
		s.Solve()
	}
}
