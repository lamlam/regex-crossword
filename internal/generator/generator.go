package generator

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/lamlam/regex-crossword/internal/puzzle"
	"github.com/lamlam/regex-crossword/internal/regex"
	"github.com/lamlam/regex-crossword/internal/solver"
)

const maxRetries = 50

// Options holds the configuration for puzzle generation.
type Options struct {
	Rows       int
	Cols       int
	Difficulty puzzle.Difficulty
	Seed       int64
}

// Generate creates a new regex crossword puzzle.
func Generate(opts Options) (*puzzle.Puzzle, error) {
	if opts.Seed == 0 {
		opts.Seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(opts.Seed))

	for attempt := 0; attempt < maxRetries; attempt++ {
		p := generateAndRefine(rng, opts)
		if p != nil {
			return p, nil
		}
	}

	return nil, fmt.Errorf("failed to generate a puzzle with a unique solution after %d attempts", maxRetries)
}

func generateAndRefine(rng *rand.Rand, opts Options) *puzzle.Puzzle {
	// Select a per-puzzle alphabet.
	alphaSize := puzzle.AlphabetSize(opts.Rows, opts.Cols, opts.Difficulty)
	alpha := selectAlphabet(rng, alphaSize)

	// Generate random solution grid.
	solution := make([][]byte, opts.Rows)
	for r := 0; r < opts.Rows; r++ {
		solution[r] = make([]byte, opts.Cols)
		for c := 0; c < opts.Cols; c++ {
			solution[r][c] = alpha[rng.Intn(len(alpha))]
		}
	}

	// Generate initial regex hints.
	strategy := regex.StrategyFor(opts.Difficulty)
	builder := regex.NewBuilder(rng, strategy, alpha)

	rowHints := make([]string, opts.Rows)
	for r := 0; r < opts.Rows; r++ {
		rowHints[r] = builder.BuildPattern(string(solution[r]))
	}

	colHints := make([]string, opts.Cols)
	for c := 0; c < opts.Cols; c++ {
		colBuf := make([]byte, opts.Rows)
		for r := 0; r < opts.Rows; r++ {
			colBuf[r] = solution[r][c]
		}
		colHints[c] = builder.BuildPattern(string(colBuf))
	}

	p := &puzzle.Puzzle{
		Rows:       opts.Rows,
		Cols:       opts.Cols,
		Difficulty: opts.Difficulty,
		Seed:       opts.Seed,
		Alphabet:   alpha,
		RowHints:   rowHints,
		ColHints:   colHints,
		Solution:   solution,
	}

	// Iteratively refine patterns to achieve uniqueness.
	for refine := 0; refine < 100; refine++ {
		s := solver.New(p)
		result := s.Solve()

		if result.Unique {
			return p
		}

		if len(result.Solutions) < 2 {
			return nil // No solution found
		}

		// Find a cell where solutions differ and tighten that row or column.
		sol1 := result.Solutions[0]
		sol2 := result.Solutions[1]

		tightened := false
		for r := 0; r < opts.Rows && !tightened; r++ {
			for c := 0; c < opts.Cols && !tightened; c++ {
				if sol1[r][c] != sol2[r][c] {
					// Tighten either the row or column hint for this cell.
					if rng.Float64() < 0.5 {
						// Tighten row hint to exclude sol2's row.
						newHint := tightenPattern(p.RowHints[r], string(solution[r]), string(sol2[r]), rng)
						if newHint != p.RowHints[r] {
							p.RowHints[r] = newHint
							tightened = true
						}
					}
					if !tightened {
						// Tighten column hint to exclude sol2's column.
						solCol := make([]byte, opts.Rows)
						altCol := make([]byte, opts.Rows)
						for ri := 0; ri < opts.Rows; ri++ {
							solCol[ri] = solution[ri][c]
							altCol[ri] = sol2[ri][c]
						}
						newHint := tightenPattern(p.ColHints[c], string(solCol), string(altCol), rng)
						if newHint != p.ColHints[c] {
							p.ColHints[c] = newHint
							tightened = true
						}
					}
				}
			}
		}

		if !tightened {
			return nil // Can't tighten further
		}
	}

	return nil
}

// tightenPattern modifies a regex pattern to still match `target` but reject `reject`.
// It does this by finding a position where target and reject differ and making
// that position more specific.
func tightenPattern(currentPattern, target, reject string, rng *rand.Rand) string {
	// Find positions where target and reject differ.
	var diffPositions []int
	for i := 0; i < len(target) && i < len(reject); i++ {
		if target[i] != reject[i] {
			diffPositions = append(diffPositions, i)
		}
	}

	if len(diffPositions) == 0 {
		return currentPattern
	}

	// Try replacing one differing position with a literal in a new pattern.
	// Build a new pattern with the target character as a literal at one diff position.
	pos := diffPositions[rng.Intn(len(diffPositions))]

	// Build a simple pattern: use literals at the chosen position and
	// the original pattern's constraints elsewhere.
	// Simple approach: create a character class that includes target[pos] but not reject[pos].
	parts := make([]string, len(target))
	for i := 0; i < len(target); i++ {
		if i == pos {
			parts[i] = regex.EscapeLiteral(target[i])
		} else {
			parts[i] = "."
		}
	}

	// Create a new pattern as (original | new_specific) intersection.
	// Since regex doesn't support intersection, we use a different approach:
	// Build a pattern that matches target but not reject by using alternation.
	newPat := "^" + strings.Join(parts, "") + "$"
	newRe, err := regexp.Compile(newPat)
	if err != nil || !newRe.MatchString(target) {
		return currentPattern
	}

	// If the new pattern already rejects the alternative, use an intersection approach.
	// Since we can't intersect regexes, use alternation within a group that
	// constrains specific positions.
	// Simplest: modify the current pattern by adding a constraint at the diff position.
	// Use a lookahead-like approach... but RE2 doesn't support lookaheads.

	// Practical approach: generate a new pattern that's more specific.
	// Combine: must match both the original pattern and have the right char at pos.
	// This is equivalent to intersecting patterns. Since RE2 can't do intersection,
	// we build a new concrete pattern.
	return buildIntersectedPattern(currentPattern, target, pos, rng)
}

// buildIntersectedPattern creates a new pattern that's at least as tight as
// currentPattern, with position pos fixed to target[pos].
func buildIntersectedPattern(currentPattern, target string, fixedPos int, rng *rand.Rand) string {
	re := regexp.MustCompile(currentPattern)

	// Find all matching strings for the current pattern (up to a limit)
	// that are the same length as target.
	n := len(target)

	// Instead of full enumeration, build a new pattern position by position.
	// For the fixed position, use the literal. For others, keep the original
	// constraint (approximated by testing which chars match).
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		if i == fixedPos {
			parts[i] = regex.EscapeLiteral(target[i])
		} else {
			// Test which chars from the alphabet match at this position.
			var validChars []byte
			testBuf := []byte(target)
			for _, ch := range []byte(puzzle.FullAlphabet) {
				testBuf[i] = ch
				// Also set the fixed position.
				testBuf[fixedPos] = target[fixedPos]
				if re.Match(testBuf) {
					validChars = append(validChars, ch)
				}
				// Restore.
				testBuf[i] = target[i]
			}

			if len(validChars) == 0 {
				validChars = []byte{target[i]}
			}
			if len(validChars) == 1 {
				parts[i] = regex.EscapeLiteral(validChars[0])
			} else if len(validChars) == len(puzzle.FullAlphabet) {
				parts[i] = "."
			} else {
				sort.Slice(validChars, func(a, b int) bool { return validChars[a] < validChars[b] })
				parts[i] = "[" + regex.CompactClass(validChars) + "]"
			}
		}
	}

	newPat := "^" + strings.Join(parts, "") + "$"
	newRe, err := regexp.Compile(newPat)
	if err != nil || !newRe.MatchString(target) {
		return currentPattern
	}

	return newPat
}

func selectAlphabet(rng *rand.Rand, size int) string {
	full := []byte(puzzle.FullAlphabet)
	rng.Shuffle(len(full), func(i, j int) { full[i], full[j] = full[j], full[i] })
	selected := full[:size]
	sort.Slice(selected, func(i, j int) bool { return selected[i] < selected[j] })
	return string(selected)
}
