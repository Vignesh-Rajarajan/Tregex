package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	patternInput textinput.Model
	textInput    textinput.Model
	focusIndex   int

	matcher   Matcher
	ranges    []Range
	explain   string
	matchErr  error
	width     int
	height    int
}

func NewModel(matcher Matcher) Model {
	pattern := textinput.New()
	pattern.Prompt = "Pattern: "
	pattern.Placeholder = "e.g. (foo|bar)+"
	pattern.Focus()

	text := textinput.New()
	text.Prompt = "Text:    "
	text.Placeholder = "Type text to test"

	m := Model{
		patternInput: pattern,
		textInput:    text,
		focusIndex:   0,
		matcher:      matcher,
	}
	m.refresh()
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab", "up", "down":
			m.cycleFocus(msg.String())
			return m, nil
		}
	}

	oldPattern := m.patternInput.Value()
	oldText := m.textInput.Value()

	var cmdPattern, cmdText tea.Cmd
	m.patternInput, cmdPattern = m.patternInput.Update(msg)
	m.textInput, cmdText = m.textInput.Update(msg)

	if oldPattern != m.patternInput.Value() || oldText != m.textInput.Value() {
		m.refresh()
	}

	return m, tea.Batch(cmdPattern, cmdText)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(m.width - 4)

	explainStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(m.width - 4)

	highlightStyles := []lipgloss.Style{
		lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")),
		lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("230")),
	}

	inputs := lipgloss.JoinVertical(
		lipgloss.Left,
		m.patternInput.View(),
		m.textInput.View(),
	)

	previewContent := RenderHighlighted(m.textInput.Value(), m.ranges, highlightStyles)
	if previewContent == "" {
		previewContent = "Enter text to see highlights."
	}

	explainContent := m.explain
	if m.matchErr != nil {
		explainContent = fmt.Sprintf("Error: %v", m.matchErr)
	}
	if explainContent == "" {
		explainContent = "No matcher wired. Implement tui.Matcher and pass it to NewModel."
	}

	header := headerStyle.Render(inputs)
	preview := previewStyle.Render(previewContent)
	explain := explainStyle.Render(explainContent)

	return lipgloss.JoinVertical(lipgloss.Left, header, preview, explain)
}

func (m *Model) cycleFocus(key string) {
	if key == "up" || key == "shift+tab" {
		m.focusIndex--
	} else {
		m.focusIndex++
	}
	if m.focusIndex > 1 {
		m.focusIndex = 0
	}
	if m.focusIndex < 0 {
		m.focusIndex = 1
	}

	if m.focusIndex == 0 {
		m.patternInput.Focus()
		m.textInput.Blur()
	} else {
		m.patternInput.Blur()
		m.textInput.Focus()
	}
}

func (m *Model) refresh() {
	m.matchErr = nil
	if m.matcher == nil {
		m.ranges = nil
		m.explain = ""
		return
	}

	ranges, err := m.matcher.MatchRanges(m.patternInput.Value(), m.textInput.Value())
	if err != nil {
		m.matchErr = err
		m.ranges = nil
		m.explain = ""
		return
	}
	m.ranges = ranges

	explain, err := m.matcher.Explain(m.patternInput.Value())
	if err != nil {
		m.matchErr = err
		m.explain = ""
		return
	}
	m.explain = explain
}
