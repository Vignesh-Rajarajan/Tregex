package tui

// Range represents a highlight span in rune offsets.
// Start is inclusive and End is exclusive.
// Style selects an index in the highlight palette (use 0 if unsure).
type Range struct {
	Start int
	End   int
	Style int
}

// Matcher provides hooks for match highlighting and explanation text.
// Implement this in your parsing layer and pass it to NewModel.
type Matcher interface {
	MatchRanges(pattern, text string) ([]Range, error)
	Explain(pattern string) (string, error)
}
