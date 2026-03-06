package tui

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderHighlighted applies styles to match ranges and returns a styled string.
// Overlapping ranges are ignored if they intersect an already rendered span.
func RenderHighlighted(text string, ranges []Range, styles []lipgloss.Style) string {
	if text == "" {
		return ""
	}
	if len(ranges) == 0 || len(styles) == 0 {
		return text
	}

	runes := []rune(text)
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].Start == ranges[j].Start {
			return ranges[i].End < ranges[j].End
		}
		return ranges[i].Start < ranges[j].Start
	})

	var b strings.Builder
	cursor := 0
	for _, r := range ranges {
		if r.Start < 0 || r.End > len(runes) || r.End <= r.Start {
			continue
		}
		if r.Start < cursor {
			continue
		}

		b.WriteString(string(runes[cursor:r.Start]))

		styleIndex := r.Style
		if styleIndex < 0 || styleIndex >= len(styles) {
			styleIndex = 0
		}
		b.WriteString(styles[styleIndex].Render(string(runes[r.Start:r.End])))

		cursor = r.End
	}
	if cursor < len(runes) {
		b.WriteString(string(runes[cursor:]))
	}

	return b.String()
}
