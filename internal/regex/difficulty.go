package regex

import "github.com/lamlam/regex-crossword/internal/puzzle"

// Strategy defines the pattern generation parameters for a difficulty level.
type Strategy struct {
	// MinLiteralRatio is the minimum fraction of positions that should be
	// literal (exact single character match).
	MinLiteralRatio float64
	// MaxDotRatio is the maximum fraction of positions that can use dot (.).
	MaxDotRatio float64
	// AllowQuantifiers enables {n} quantifiers.
	AllowQuantifiers bool
	// AllowNegatedClasses enables [^...] negated character classes.
	AllowNegatedClasses bool
	// AllowAlternation enables (X|Y) alternation patterns.
	AllowAlternation bool
	// MaxClassSize is the maximum size of a character class [ABC...].
	MaxClassSize int
}

// StrategyFor returns the pattern generation strategy for the given difficulty.
func StrategyFor(d puzzle.Difficulty) Strategy {
	switch d {
	case puzzle.Easy:
		return Strategy{
			MinLiteralRatio:     0.7,
			MaxDotRatio:         0.0,
			AllowQuantifiers:    false,
			AllowNegatedClasses: false,
			AllowAlternation:    true,
			MaxClassSize:        2,
		}
	case puzzle.Medium:
		return Strategy{
			MinLiteralRatio:     0.5,
			MaxDotRatio:         0.1,
			AllowQuantifiers:    true,
			AllowNegatedClasses: false,
			AllowAlternation:    true,
			MaxClassSize:        3,
		}
	case puzzle.Hard:
		return Strategy{
			MinLiteralRatio:     0.45,
			MaxDotRatio:         0.1,
			AllowQuantifiers:    true,
			AllowNegatedClasses: true,
			AllowAlternation:    true,
			MaxClassSize:        4,
		}
	default:
		return StrategyFor(puzzle.Medium)
	}
}
