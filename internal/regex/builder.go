package regex

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
)

// Builder generates regex patterns for a given target string.
type Builder struct {
	rng      *rand.Rand
	strategy Strategy
	alphabet string
}

// NewBuilder creates a new regex pattern builder.
func NewBuilder(rng *rand.Rand, strategy Strategy, alphabet string) *Builder {
	return &Builder{
		rng:      rng,
		strategy: strategy,
		alphabet: alphabet,
	}
}

// BuildPattern generates a regex pattern that matches the target string.
// The pattern is anchored (wrapped in ^...$) and validated.
func (b *Builder) BuildPattern(target string) string {
	for attempt := 0; attempt < 50; attempt++ {
		pat := b.buildPatternOnce(target)
		fullPat := "^" + pat + "$"
		re, err := regexp.Compile(fullPat)
		if err != nil {
			continue
		}
		if re.MatchString(target) {
			return fullPat
		}
	}
	return "^" + regexp.QuoteMeta(target) + "$"
}

func (b *Builder) buildPatternOnce(target string) string {
	n := len(target)
	if n == 0 {
		return ""
	}

	// Determine how many positions get literal treatment.
	literalCount := max(1, int(b.strategy.MinLiteralRatio*float64(n)+0.5))
	if literalCount > n {
		literalCount = n
	}
	maxDots := int(b.strategy.MaxDotRatio * float64(n))

	// Assign roles to positions.
	positions := make([]int, n)
	for i := range positions {
		positions[i] = i
	}
	b.rng.Shuffle(n, func(i, j int) { positions[i], positions[j] = positions[j], positions[i] })

	roles := make([]string, n)
	for i := 0; i < n; i++ {
		if i < literalCount {
			roles[positions[i]] = "literal"
		} else {
			roles[positions[i]] = "class"
		}
	}

	dotCount := 0
	for i := 0; i < n; i++ {
		if roles[i] == "class" && dotCount < maxDots && b.rng.Float64() < 0.3 {
			roles[i] = "dot"
			dotCount++
		}
	}

	// Generate fragments.
	fragments := make([]string, n)
	for i := 0; i < n; i++ {
		ch := target[i]
		switch roles[i] {
		case "literal":
			fragments[i] = escapeLiteral(ch)
		case "dot":
			fragments[i] = "."
		default:
			fragments[i] = b.charClassFragment(ch)
		}
	}

	// Optionally combine with quantifiers.
	if b.strategy.AllowQuantifiers && n >= 2 && b.rng.Float64() < 0.25 {
		fragments = b.tryCombineQuantifiers(target, fragments)
	}

	// Optionally wrap in alternation.
	if b.strategy.AllowAlternation && b.rng.Float64() < 0.15 {
		return b.tryAlternation(target, fragments)
	}

	return strings.Join(fragments, "")
}

func (b *Builder) charClassFragment(ch byte) string {
	maxSize := b.strategy.MaxClassSize
	if maxSize < 2 {
		maxSize = 2
	}
	// Size relative to alphabet size.
	size := 2 + b.rng.Intn(maxSize-1)
	if size > len(b.alphabet) {
		size = len(b.alphabet)
	}

	if b.strategy.AllowNegatedClasses && b.rng.Float64() < 0.25 {
		return b.negatedClassFragment(ch, size)
	}

	chars := []byte{ch}
	used := map[byte]bool{ch: true}
	for len(chars) < size {
		c := b.alphabet[b.rng.Intn(len(b.alphabet))]
		if !used[c] {
			used[c] = true
			chars = append(chars, c)
		}
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })
	return "[" + compactClass(chars) + "]"
}

func (b *Builder) negatedClassFragment(ch byte, desiredPositiveSize int) string {
	// Keep desiredPositiveSize chars (including ch), exclude the rest.
	kept := map[byte]bool{ch: true}
	for len(kept) < desiredPositiveSize && len(kept) < len(b.alphabet) {
		c := b.alphabet[b.rng.Intn(len(b.alphabet))]
		if !kept[c] {
			kept[c] = true
		}
	}

	var excluded []byte
	for i := 0; i < len(b.alphabet); i++ {
		c := b.alphabet[i]
		if !kept[c] {
			excluded = append(excluded, c)
		}
	}

	if len(excluded) == 0 {
		// All chars kept, use dot instead.
		return "."
	}

	sort.Slice(excluded, func(i, j int) bool { return excluded[i] < excluded[j] })
	return "[^" + compactClass(excluded) + "]"
}

func (b *Builder) tryCombineQuantifiers(target string, fragments []string) []string {
	var result []string
	i := 0
	for i < len(fragments) {
		j := i + 1
		for j < len(fragments) && target[j] == target[i] {
			j++
		}
		runLen := j - i
		if runLen >= 2 && b.rng.Float64() < 0.5 {
			lit := escapeLiteral(target[i])
			result = append(result, lit+fmt.Sprintf("{%d}", runLen))
			i = j
		} else {
			result = append(result, fragments[i])
			i++
		}
	}
	return result
}

func (b *Builder) tryAlternation(target string, fragments []string) string {
	base := strings.Join(fragments, "")
	alt := b.generateFakeAlternative(target)
	if alt == target || alt == base {
		return base
	}
	if b.rng.Float64() < 0.5 {
		return "(" + base + "|" + regexp.QuoteMeta(alt) + ")"
	}
	return "(" + regexp.QuoteMeta(alt) + "|" + base + ")"
}

func (b *Builder) generateFakeAlternative(target string) string {
	buf := []byte(target)
	changes := 1 + b.rng.Intn(2)
	for c := 0; c < changes; c++ {
		pos := b.rng.Intn(len(buf))
		buf[pos] = b.alphabet[b.rng.Intn(len(b.alphabet))]
	}
	return string(buf)
}

// EscapeLiteral escapes a character for use in a regex pattern.
func EscapeLiteral(ch byte) string {
	return escapeLiteral(ch)
}

func escapeLiteral(ch byte) string {
	s := string(ch)
	if strings.ContainsAny(s, `\.+*?^${}()|[]`) {
		return `\` + s
	}
	return s
}

// CompactClass converts a sorted list of bytes to a character class string.
func CompactClass(chars []byte) string {
	return compactClass(chars)
}

func compactClass(chars []byte) string {
	if len(chars) == 0 {
		return ""
	}
	var parts []string
	i := 0
	for i < len(chars) {
		j := i
		for j+1 < len(chars) && chars[j+1] == chars[j]+1 {
			j++
		}
		rangeLen := j - i + 1
		if rangeLen >= 3 {
			parts = append(parts, escapeLiteral(chars[i])+"-"+escapeLiteral(chars[j]))
		} else {
			for k := i; k <= j; k++ {
				parts = append(parts, escapeLiteral(chars[k]))
			}
		}
		i = j + 1
	}
	return strings.Join(parts, "")
}

