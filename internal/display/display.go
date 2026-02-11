package display

import (
	"fmt"
	"io"
	"strings"

	"github.com/lamlam/regex-crossword/internal/puzzle"
)

// Print renders the puzzle to the given writer.
func Print(w io.Writer, p *puzzle.Puzzle, showSolution bool) {
	// Header.
	fmt.Fprintf(w, "Regex Crossword  [%dx%d, %s]\n", p.Rows, p.Cols, p.Difficulty)
	fmt.Fprintf(w, "Seed: %d\n", p.Seed)
	fmt.Fprintf(w, "Alphabet: %s\n\n", p.Alphabet)

	// Calculate column widths based on hint lengths.
	cellWidth := 8
	for _, h := range p.ColHints {
		if len(h)+2 > cellWidth {
			cellWidth = len(h) + 2
		}
	}
	if cellWidth > 16 {
		cellWidth = 16
	}

	// Row label width.
	rowLabelWidth := 4 // "R1  " etc.

	// Print column labels.
	fmt.Fprintf(w, "%s", strings.Repeat(" ", rowLabelWidth+2))
	for c := 0; c < p.Cols; c++ {
		label := fmt.Sprintf("C%d", c+1)
		fmt.Fprintf(w, "%-*s", cellWidth, centerString(label, cellWidth))
	}
	fmt.Fprintln(w)

	// Print column hints.
	fmt.Fprintf(w, "%s", strings.Repeat(" ", rowLabelWidth+2))
	for c := 0; c < p.Cols; c++ {
		hint := p.ColHints[c]
		if len(hint) > cellWidth-1 {
			hint = hint[:cellWidth-1]
		}
		fmt.Fprintf(w, "%-*s", cellWidth, centerString(hint, cellWidth))
	}
	fmt.Fprintln(w)

	// Print separator.
	fmt.Fprintf(w, "%s+", strings.Repeat(" ", rowLabelWidth+1))
	for c := 0; c < p.Cols; c++ {
		fmt.Fprintf(w, "%s+", strings.Repeat("-", cellWidth))
	}
	fmt.Fprintln(w)

	// Print rows.
	for r := 0; r < p.Rows; r++ {
		label := fmt.Sprintf("R%d", r+1)
		fmt.Fprintf(w, "%-*s |", rowLabelWidth, label)
		for c := 0; c < p.Cols; c++ {
			if showSolution {
				cell := string(p.Solution[r][c])
				fmt.Fprintf(w, "%s|", centerString(cell, cellWidth))
			} else {
				fmt.Fprintf(w, "%s|", strings.Repeat(" ", cellWidth))
			}
		}
		fmt.Fprintf(w, "  %s\n", p.RowHints[r])

		// Row separator.
		fmt.Fprintf(w, "%s+", strings.Repeat(" ", rowLabelWidth+1))
		for c := 0; c < p.Cols; c++ {
			fmt.Fprintf(w, "%s+", strings.Repeat("-", cellWidth))
		}
		fmt.Fprintln(w)
	}
}

// centerString centers s within a field of the given width.
func centerString(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
