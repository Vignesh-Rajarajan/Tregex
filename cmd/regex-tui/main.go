package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"regex-parser/internal/tui"
	"regex-parser/pkg/parser"
)

func main() {
	matcher := parser.NewRegexMatcher()
	model := tui.NewModel(matcher)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
