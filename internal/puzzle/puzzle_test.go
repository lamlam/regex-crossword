package puzzle

import "testing"

func TestParseDifficulty(t *testing.T) {
	tests := []struct {
		input string
		want  Difficulty
		err   bool
	}{
		{"easy", Easy, false},
		{"medium", Medium, false},
		{"hard", Hard, false},
		{"unknown", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseDifficulty(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParseDifficulty(%q) error = %v, wantErr %v", tt.input, err, tt.err)
		}
		if got != tt.want {
			t.Errorf("ParseDifficulty(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestDifficultyString(t *testing.T) {
	tests := []struct {
		d    Difficulty
		want string
	}{
		{Easy, "Easy"},
		{Medium, "Medium"},
		{Hard, "Hard"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("Difficulty(%d).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestAlphabetSize(t *testing.T) {
	// Ensure alphabet size is at least 4 and at most 10.
	for rows := 1; rows <= 8; rows++ {
		for cols := 1; cols <= 8; cols++ {
			for _, d := range []Difficulty{Easy, Medium, Hard} {
				size := AlphabetSize(rows, cols, d)
				if size < 4 || size > 10 {
					t.Errorf("AlphabetSize(%d, %d, %v) = %d, want 4-10", rows, cols, d, size)
				}
			}
		}
	}
}

func TestColString(t *testing.T) {
	p := &Puzzle{
		Rows:     2,
		Cols:     3,
		Solution: [][]byte{{'A', 'B', 'C'}, {'D', 'E', 'F'}},
	}
	if got := p.ColString(0); got != "AD" {
		t.Errorf("ColString(0) = %q, want %q", got, "AD")
	}
	if got := p.ColString(1); got != "BE" {
		t.Errorf("ColString(1) = %q, want %q", got, "BE")
	}
}
