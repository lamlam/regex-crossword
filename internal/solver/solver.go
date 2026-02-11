package solver

import (
	"regexp"

	"github.com/lamlam/regex-crossword/internal/puzzle"
)

const maxEnumLimit = 500000

// Solver solves regex crossword puzzles.
type Solver struct {
	rows, cols   int
	alphabet     string
	rowRegexps   []*regexp.Regexp
	colRegexps   []*regexp.Regexp
	maxSolutions int
}

// New creates a new Solver from the given puzzle hints.
func New(p *puzzle.Puzzle) *Solver {
	rowRes := make([]*regexp.Regexp, p.Rows)
	for i, h := range p.RowHints {
		rowRes[i] = regexp.MustCompile(h)
	}
	colRes := make([]*regexp.Regexp, p.Cols)
	for i, h := range p.ColHints {
		colRes[i] = regexp.MustCompile(h)
	}
	alpha := p.Alphabet
	if alpha == "" {
		alpha = puzzle.FullAlphabet
	}
	return &Solver{
		rows:         p.Rows,
		cols:         p.Cols,
		alphabet:     alpha,
		rowRegexps:   rowRes,
		colRegexps:   colRes,
		maxSolutions: 2,
	}
}

// Result holds the solver output.
type Result struct {
	Solutions [][][]byte
	Unique    bool
}

// Solve finds solutions to the puzzle.
func (s *Solver) Solve() Result {
	// Step 1: Compute per-cell candidate sets.
	cellCands := s.computeCellCandidates()

	// Check for empty cells.
	for r := 0; r < s.rows; r++ {
		for c := 0; c < s.cols; c++ {
			if len(cellCands[r][c]) == 0 {
				return Result{}
			}
		}
	}

	// Step 2: Iterative arc consistency propagation.
	s.propagate(cellCands)
	for r := 0; r < s.rows; r++ {
		for c := 0; c < s.cols; c++ {
			if len(cellCands[r][c]) == 0 {
				return Result{}
			}
		}
	}

	// Step 3: Enumerate valid rows using filtered candidates.
	rowMatches := make([][]string, s.rows)
	for r := 0; r < s.rows; r++ {
		matches := s.enumRowMatches(r, cellCands[r])
		if len(matches) == 0 {
			return Result{}
		}
		rowMatches[r] = matches
	}

	// Step 4: Backtrack row by row.
	grid := make([][]byte, s.rows)
	for r := 0; r < s.rows; r++ {
		grid[r] = make([]byte, s.cols)
	}

	var solutions [][][]byte
	s.solveRows(grid, rowMatches, 0, &solutions)

	return Result{
		Solutions: solutions,
		Unique:    len(solutions) == 1,
	}
}

// computeCellCandidates determines valid characters for each cell.
// For each cell (r,c), a character ch is valid if:
// - ch is in the puzzle alphabet
// - There exists a string of length cols containing ch at position c that matches rowRegexps[r]
// - There exists a string of length rows containing ch at position r that matches colRegexps[c]
func (s *Solver) computeCellCandidates() [][]string {
	// Row-based candidates.
	rowCands := make([][]map[byte]bool, s.rows)
	for r := 0; r < s.rows; r++ {
		rowCands[r] = make([]map[byte]bool, s.cols)
		for c := 0; c < s.cols; c++ {
			rowCands[r][c] = make(map[byte]bool)
		}
		s.findPositionCandidates(s.rowRegexps[r], s.cols, rowCands[r])
	}

	// Column-based candidates.
	colCands := make([][]map[byte]bool, s.cols)
	for c := 0; c < s.cols; c++ {
		colCands[c] = make([]map[byte]bool, s.rows)
		for r := 0; r < s.rows; r++ {
			colCands[c][r] = make(map[byte]bool)
		}
		s.findPositionCandidates(s.colRegexps[c], s.rows, colCands[c])
	}

	// Intersect.
	result := make([][]string, s.rows)
	for r := 0; r < s.rows; r++ {
		result[r] = make([]string, s.cols)
		for c := 0; c < s.cols; c++ {
			var chars []byte
			for _, ch := range []byte(s.alphabet) {
				if rowCands[r][c][ch] && colCands[c][r][ch] {
					chars = append(chars, ch)
				}
			}
			result[r][c] = string(chars)
		}
	}
	return result
}

// findPositionCandidates determines which alphabet characters can appear at each
// position for a given regex of a given length.
// Uses exhaustive enumeration if the search space is small enough,
// otherwise adds all alphabet characters as candidates (no filtering).
func (s *Solver) findPositionCandidates(re *regexp.Regexp, length int, posCands []map[byte]bool) {
	alpha := []byte(s.alphabet)
	alphaLen := len(alpha)

	// Calculate search space size.
	searchSpace := 1
	overflow := false
	for i := 0; i < length; i++ {
		searchSpace *= alphaLen
		if searchSpace > 500000 {
			overflow = true
			break
		}
	}

	if overflow {
		// Search space too large for enumeration. Use all alphabet chars.
		for pos := 0; pos < length; pos++ {
			for _, ch := range alpha {
				posCands[pos][ch] = true
			}
		}
		return
	}

	// Exhaustive enumeration.
	buf := make([]byte, length)
	s.enumAllForCandidates(re, alpha, buf, 0, posCands)
}

func (s *Solver) enumAllForCandidates(re *regexp.Regexp, alpha []byte, buf []byte, pos int, posCands []map[byte]bool) {
	if pos == len(buf) {
		if re.Match(buf) {
			for p := 0; p < len(buf); p++ {
				posCands[p][buf[p]] = true
			}
		}
		return
	}
	for _, ch := range alpha {
		buf[pos] = ch
		s.enumAllForCandidates(re, alpha, buf, pos+1, posCands)
	}
}

// propagate does iterative arc consistency.
func (s *Solver) propagate(cellCands [][]string) {
	for iter := 0; iter < 10; iter++ {
		changed := false

		// For each row, enumerate valid combinations and keep only valid candidates.
		for r := 0; r < s.rows; r++ {
			newCands := make([]map[byte]bool, s.cols)
			for c := 0; c < s.cols; c++ {
				newCands[c] = make(map[byte]bool)
			}

			count := 0
			s.enumValidCombs(s.rowRegexps[r], cellCands[r], newCands, &count)

			if count <= maxEnumLimit {
				for c := 0; c < s.cols; c++ {
					filtered := filterString(cellCands[r][c], newCands[c])
					if filtered != cellCands[r][c] {
						cellCands[r][c] = filtered
						changed = true
					}
				}
			}
		}

		// For each column.
		for c := 0; c < s.cols; c++ {
			colSlice := make([]string, s.rows)
			for r := 0; r < s.rows; r++ {
				colSlice[r] = cellCands[r][c]
			}

			newCands := make([]map[byte]bool, s.rows)
			for r := 0; r < s.rows; r++ {
				newCands[r] = make(map[byte]bool)
			}

			count := 0
			s.enumValidCombs(s.colRegexps[c], colSlice, newCands, &count)

			if count <= maxEnumLimit {
				for r := 0; r < s.rows; r++ {
					filtered := filterString(colSlice[r], newCands[r])
					if filtered != colSlice[r] {
						cellCands[r][c] = filtered
						changed = true
					}
				}
			}
		}

		if !changed {
			break
		}
	}
}

// enumValidCombs enumerates valid character combinations for a regex.
func (s *Solver) enumValidCombs(re *regexp.Regexp, candidates []string, validChars []map[byte]bool, count *int) {
	buf := make([]byte, len(candidates))
	s.enumCombHelper(re, candidates, buf, 0, validChars, count)
}

func (s *Solver) enumCombHelper(re *regexp.Regexp, candidates []string, buf []byte, pos int, validChars []map[byte]bool, count *int) {
	if *count > maxEnumLimit {
		// Too many - keep all candidates.
		for p := 0; p < len(candidates); p++ {
			for _, ch := range []byte(candidates[p]) {
				validChars[p][ch] = true
			}
		}
		return
	}

	if pos == len(candidates) {
		*count++
		if re.Match(buf) {
			for p := 0; p < len(candidates); p++ {
				validChars[p][buf[p]] = true
			}
		}
		return
	}

	for _, ch := range []byte(candidates[pos]) {
		buf[pos] = ch
		s.enumCombHelper(re, candidates, buf, pos+1, validChars, count)
	}
}

// enumRowMatches enumerates all strings matching the row regex using per-cell candidates.
func (s *Solver) enumRowMatches(r int, rowCands []string) []string {
	var results []string
	buf := make([]byte, s.cols)
	count := 0
	s.enumMatchHelper(s.rowRegexps[r], rowCands, buf, 0, &results, &count)
	return results
}

func (s *Solver) enumMatchHelper(re *regexp.Regexp, candidates []string, buf []byte, pos int, results *[]string, count *int) {
	if *count > maxEnumLimit {
		return
	}

	if pos == len(candidates) {
		*count++
		if re.Match(buf) {
			*results = append(*results, string(buf))
		}
		return
	}

	for _, ch := range []byte(candidates[pos]) {
		buf[pos] = ch
		s.enumMatchHelper(re, candidates, buf, pos+1, results, count)
	}
}

// solveRows tries each valid row string and recurses.
func (s *Solver) solveRows(grid [][]byte, rowMatches [][]string, row int, solutions *[][][]byte) {
	if len(*solutions) >= s.maxSolutions {
		return
	}

	if row == s.rows {
		if s.validateCols(grid) {
			sol := make([][]byte, s.rows)
			for r := 0; r < s.rows; r++ {
				sol[r] = make([]byte, s.cols)
				copy(sol[r], grid[r])
			}
			*solutions = append(*solutions, sol)
		}
		return
	}

	for _, match := range rowMatches[row] {
		copy(grid[row], match)

		if s.partialColsOK(grid, row) {
			s.solveRows(grid, rowMatches, row+1, solutions)
			if len(*solutions) >= s.maxSolutions {
				return
			}
		}
	}
}

func (s *Solver) partialColsOK(grid [][]byte, currentRow int) bool {
	if currentRow == s.rows-1 {
		return s.validateCols(grid)
	}

	for c := 0; c < s.cols; c++ {
		prefix := make([]byte, currentRow+1)
		for r := 0; r <= currentRow; r++ {
			prefix[r] = grid[r][c]
		}
		if !s.colPrefixViable(c, prefix) {
			return false
		}
	}
	return true
}

func (s *Solver) colPrefixViable(c int, prefix []byte) bool {
	buf := make([]byte, s.rows)
	copy(buf, prefix)
	return s.tryExtend(s.colRegexps[c], buf, len(prefix))
}

func (s *Solver) tryExtend(re *regexp.Regexp, buf []byte, filled int) bool {
	if filled == len(buf) {
		return re.Match(buf)
	}
	for _, ch := range []byte(s.alphabet) {
		buf[filled] = ch
		if s.tryExtend(re, buf, filled+1) {
			return true
		}
	}
	return false
}

func (s *Solver) validateCols(grid [][]byte) bool {
	for c := 0; c < s.cols; c++ {
		buf := make([]byte, s.rows)
		for r := 0; r < s.rows; r++ {
			buf[r] = grid[r][c]
		}
		if !s.colRegexps[c].Match(buf) {
			return false
		}
	}
	return true
}

func filterString(original string, keep map[byte]bool) string {
	var result []byte
	for _, ch := range []byte(original) {
		if keep[ch] {
			result = append(result, ch)
		}
	}
	return string(result)
}

// HasUniqueSolution checks if the puzzle has exactly one solution.
func HasUniqueSolution(p *puzzle.Puzzle) bool {
	s := New(p)
	result := s.Solve()
	return result.Unique
}
