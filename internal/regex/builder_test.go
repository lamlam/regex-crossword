package regex

import (
	"math/rand"
	"regexp"
	"testing"

	"github.com/lamlam/regex-crossword/internal/puzzle"
)

func TestBuildPatternMatchesTarget(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	alphabet := "ABCDE12345"

	for _, diff := range []puzzle.Difficulty{puzzle.Easy, puzzle.Medium, puzzle.Hard} {
		strategy := StrategyFor(diff)
		builder := NewBuilder(rng, strategy, alphabet)

		for i := 0; i < 50; i++ {
			// Generate a random target string.
			length := 3 + rng.Intn(5)
			target := make([]byte, length)
			for j := range target {
				target[j] = alphabet[rng.Intn(len(alphabet))]
			}
			targetStr := string(target)

			pat := builder.BuildPattern(targetStr)
			re, err := regexp.Compile(pat)
			if err != nil {
				t.Errorf("difficulty=%v: BuildPattern(%q) produced invalid regex %q: %v", diff, targetStr, pat, err)
				continue
			}
			if !re.MatchString(targetStr) {
				t.Errorf("difficulty=%v: BuildPattern(%q) = %q doesn't match target", diff, targetStr, pat)
			}
		}
	}
}

func TestEscapeLiteral(t *testing.T) {
	tests := []struct {
		ch   byte
		want string
	}{
		{'A', "A"},
		{'0', "0"},
	}
	for _, tt := range tests {
		if got := EscapeLiteral(tt.ch); got != tt.want {
			t.Errorf("EscapeLiteral(%q) = %q, want %q", tt.ch, got, tt.want)
		}
	}
}

func TestCompactClass(t *testing.T) {
	tests := []struct {
		chars []byte
		want  string
	}{
		{[]byte{'A', 'B', 'C'}, "A-C"},
		{[]byte{'A', 'C'}, "AC"},
		{[]byte{'A', 'B', 'C', 'E', 'F', 'G'}, "A-CE-G"},
	}
	for _, tt := range tests {
		if got := CompactClass(tt.chars); got != tt.want {
			t.Errorf("CompactClass(%v) = %q, want %q", tt.chars, got, tt.want)
		}
	}
}
