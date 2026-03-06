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
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
