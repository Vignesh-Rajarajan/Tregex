package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── palette ──────────────────────────────────────────────────────────────────

var (
	textMain  = lipgloss.Color("#f3f4f6")
	textMuted = lipgloss.Color("#8b8d98")
	accent    = lipgloss.Color("#3b82f6")
	border    = lipgloss.Color("#3a3a3a")
	borderAct = lipgloss.Color("#6b6b6b")
	match1Bg  = lipgloss.Color("#3b2d59")
	match1Fg  = lipgloss.Color("#d8b4fe")
	match2Bg  = lipgloss.Color("#501e3b")
	match2Fg  = lipgloss.Color("#f9a8d4")
	successFg = lipgloss.Color("#10b981")
	errorFg   = lipgloss.Color("#ef4444")
	flagOnFg  = lipgloss.Color("#3b82f6")
	flagOffFg = lipgloss.Color("#8b8d98")
)

// ── styles ────────────────────────────────────────────────────────────────────
// RULE: never use Margin* on any style used inside a panel.
// Margins produce unstyled blank lines that expose the terminal background as
// a dark stripe. Use Padding or explicit " " spacers instead.

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	logoStyle     = lipgloss.NewStyle().Foreground(textMain).Bold(true)
	logoSpanStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	taglineStyle  = lipgloss.NewStyle().Foreground(textMuted)
	shortcutStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1)

	panelStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border).Padding(0, 1)
	panelActiveStyle = panelStyle.Copy().BorderForeground(borderAct)

	// PaddingBottom only — no MarginBottom.
	panelHeaderStyle = lipgloss.NewStyle().
				Foreground(textMuted).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(border).
				PaddingBottom(1)

	badgeSuccessStyle = lipgloss.NewStyle().Foreground(successFg)
	badgeErrorStyle   = lipgloss.NewStyle().Foreground(errorFg)
	badgeDimStyle     = lipgloss.NewStyle().Foreground(textMuted)

	// No MarginRight — spacing is added with explicit " " between buttons.
	flagOnStyle = lipgloss.NewStyle().
			Foreground(flagOnFg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(flagOnFg).
			Padding(0, 1)
	flagOffStyle = lipgloss.NewStyle().
			Foreground(flagOffFg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1)
	// Cursor highlight when the flags row itself is focused.
	flagFocusOnStyle  = flagOnStyle.Copy().BorderForeground(accent).Bold(true)
	flagFocusOffStyle = flagOffStyle.Copy().BorderForeground(borderAct)

	// No MarginRight on statStyle.
	statStyle    = lipgloss.NewStyle().Foreground(textMuted)
	statValStyle = lipgloss.NewStyle().Foreground(textMain).Bold(true)

	match1Style = lipgloss.NewStyle().Background(match1Bg).Foreground(match1Fg)
	match2Style = lipgloss.NewStyle().Background(match2Bg).Foreground(match2Fg)

	listItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(border)
)

// ── focus areas ───────────────────────────────────────────────────────────────

type focus int

const (
	focusPattern focus = iota
	focusText
	focusFlags // Tab-stops here; g/i/m/s toggle the flags directly
)

// ── flag descriptors (order matches keyboard shortcuts) ───────────────────────

type flagDef struct {
	key   string // key to press when flags row is focused
	short string // single-letter label
	label string // full label shown in button
}

var flagDefs = []flagDef{
	{"g", "g", "g  global"},
	{"i", "i", "i  ignore case"},
	{"m", "m", "m  multiline"},
	{"s", "s", "s  dotall"},
}

// ── model ─────────────────────────────────────────────────────────────────────

type Model struct {
	patternInput textinput.Model
	textInput    textinput.Model
	activeFocus  focus

	// regex flags
	flagG bool // global — return all matches (default on)
	flagI bool // case insensitive  (?i)
	flagM bool // multiline anchors (?m)
	flagS bool // dot matches \n   (?s)

	matcher  Matcher
	ranges   []Range
	matchErr error

	matchCount int
	groupCount int

	width  int
	height int
}

func NewModel(matcher Matcher) Model {
	pi := textinput.New()
	pi.Placeholder = "(foo|bar)+"
	pi.SetValue("(foo|bar)+")
	pi.Prompt = "/ "
	pi.PromptStyle = lipgloss.NewStyle().Foreground(textMuted)
	pi.TextStyle = lipgloss.NewStyle().Foreground(textMain)
	pi.Focus()

	ti := textinput.New()
	ti.Placeholder = "this bar be better than foo"
	ti.SetValue("this bar be better than foo")
	ti.Prompt = ""
	ti.TextStyle = lipgloss.NewStyle().Foreground(textMain)

	m := Model{
		patternInput: pi,
		textInput:    ti,
		activeFocus:  focusPattern,
		flagG:        true,
		matcher:      matcher,
		width:        100,
	}
	m.syncWidths()
	m.refresh()
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncWidths()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			m.cycleFocus(false)
			return m, textinput.Blink

		case "shift+tab":
			m.cycleFocus(true)
			return m, textinput.Blink
		}

		// When the flags row is focused, single letters toggle the flag directly.
		if m.activeFocus == focusFlags {
			changed := true
			switch msg.String() {
			case "g":
				m.flagG = !m.flagG
			case "i":
				m.flagI = !m.flagI
			case "m":
				m.flagM = !m.flagM
			case "s":
				m.flagS = !m.flagS
			default:
				changed = false
			}
			if changed {
				m.refresh()
				return m, nil
			}
		}
	}

	// Route remaining keys to whichever text input is active.
	if m.activeFocus == focusPattern || m.activeFocus == focusText {
		oldPattern := m.patternInput.Value()
		oldText := m.textInput.Value()

		if m.activeFocus == focusPattern {
			m.patternInput, cmd = m.patternInput.Update(msg)
		} else {
			m.textInput, cmd = m.textInput.Update(msg)
		}

		if oldPattern != m.patternInput.Value() || oldText != m.textInput.Value() {
			m.refresh()
		}
	}

	return m, cmd
}

func (m *Model) cycleFocus(backwards bool) {
	order := []focus{focusPattern, focusText, focusFlags}
	cur := 0
	for i, f := range order {
		if f == m.activeFocus {
			cur = i
		}
	}
	if backwards {
		cur = (cur + len(order) - 1) % len(order)
	} else {
		cur = (cur + 1) % len(order)
	}
	m.activeFocus = order[cur]
	m.patternInput.Blur()
	m.textInput.Blur()
	switch m.activeFocus {
	case focusPattern:
		m.patternInput.Focus()
	case focusText:
		m.textInput.Focus()
	}
}

// ── layout helpers ────────────────────────────────────────────────────────────

// layoutWidths returns (leftOuterWidth, rightOuterWidth, showSidebar).
// Outer width = border (2) + padding (2) + inner content.
func (m Model) layoutWidths() (int, int, bool) {
	usable := m.width - 4 // appStyle Padding(1,2) removes 2 chars each side
	if usable < 44 {
		usable = 44
	}
	if usable >= 90 {
		rightOuter := 32
		leftOuter := usable - rightOuter - 2 // 2-char gap
		return leftOuter, rightOuter, true
	}
	return usable, 0, false
}

func (m *Model) syncWidths() {
	leftOuter, _, _ := m.layoutWidths()
	// Pattern: subtract border(2)+padding(2)+"/ " prompt(2)+" /" suffix(2) = 8.
	// Text:    subtract border(2)+padding(2) = 4.
	m.patternInput.Width = max(10, leftOuter-8)
	m.textInput.Width = max(10, leftOuter-4)
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	leftOuter, rightOuter, showSidebar := m.layoutWidths()
	innerW := leftOuter - 4 // border(2) + padding(2)

	// ── header ────────────────────────────────────────────────────────────────
	logo := logoStyle.Render("Regex") + logoSpanStyle.Render("Studio")
	tagline := taglineStyle.Render("Interactive Expression Evaluator")
	headerLeft := lipgloss.JoinVertical(lipgloss.Left, logo, tagline)

	shortcuts := lipgloss.JoinHorizontal(lipgloss.Left,
		shortcutStyle.Render("Ctrl+C quit"),
		" ",
		shortcutStyle.Render("Tab focus"),
	)

	totalW := leftOuter
	if showSidebar {
		totalW += rightOuter + 2
	}
	leftBlockW := max(20, totalW-lipgloss.Width(shortcuts))
	header := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().Width(leftBlockW).Render(headerLeft),
		shortcuts,
	)

	// ── pattern panel ─────────────────────────────────────────────────────────
	pStyle := panelStyle.Copy().Width(leftOuter - 2)
	if m.activeFocus == focusPattern {
		pStyle = panelActiveStyle.Copy().Width(leftOuter - 2)
	}

	badge := m.patternBadge()
	patHeader := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().
			Width(innerW-lipgloss.Width(badge)-1).
			Foreground(textMuted).
			Render("EXPRESSION"),
		badge,
	)

	inputLine := strings.TrimRight(m.patternInput.View(), "\n") +
		lipgloss.NewStyle().Foreground(textMuted).Render(" /")

	flagsRow := m.renderFlags()

	// Hint shown only when the flags row is focused.
	flagHint := ""
	if m.activeFocus == focusFlags {
		flagHint = lipgloss.NewStyle().Foreground(textMuted).
			Render("  press g · i · m · s to toggle")
	}

	patternBody := []string{
		panelHeaderStyle.Copy().Width(innerW).Render(patHeader),
		inputLine,
		"",
		flagsRow,
	}
	if flagHint != "" {
		patternBody = append(patternBody, flagHint)
	}

	patternPanel := pStyle.Render(lipgloss.JoinVertical(lipgloss.Left, patternBody...))

	// ── test string panel ─────────────────────────────────────────────────────
	tStyle := panelStyle.Copy().Width(leftOuter - 2)
	if m.activeFocus == focusText {
		tStyle = panelActiveStyle.Copy().Width(leftOuter - 2)
	}

	textLine := strings.TrimRight(m.textInput.View(), "\n")
	textPanel := tStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		panelHeaderStyle.Copy().Width(innerW).Render("TEST STRING"),
		textLine,
	))

	// ── evaluation panel ──────────────────────────────────────────────────────
	matchBadge := m.renderMatchBadge()
	evalHeader := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().
			Width(innerW-lipgloss.Width(matchBadge)-1).
			Foreground(textMuted).
			Render("EVALUATION"),
		matchBadge,
	)

	matchContent := RenderHighlighted(m.textInput.Value(), m.ranges,
		[]lipgloss.Style{match1Style, match2Style})
	if matchContent == "" && m.matchErr == nil {
		matchContent = lipgloss.NewStyle().Foreground(textMuted).Render("(enter text to test)")
	}
	if m.matchErr != nil {
		matchContent = lipgloss.NewStyle().Foreground(errorFg).Render(m.matchErr.Error())
	}

	statsBar := m.renderStatsBar(innerW)

	evalPanel := panelStyle.Copy().Width(leftOuter-2).Render(lipgloss.JoinVertical(lipgloss.Left,
		panelHeaderStyle.Copy().Width(innerW).Render(evalHeader),
		matchContent,
		"",
		statsBar,
	))

	leftCol := lipgloss.JoinVertical(lipgloss.Left, patternPanel, textPanel, evalPanel)

	if !showSidebar {
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", leftCol))
	}

	// ── examples sidebar ──────────────────────────────────────────────────────
	sideInnerW := rightOuter - 4
	sidePanel := panelStyle.Copy().Width(rightOuter - 2)

	examples := []struct{ name, pattern string }{
		{"Alternation", "(foo|bar)+"},
		{"Lazy tags", `<.+?>`},
		{"Digits", `\d+`},
		{"Anchors", `^foo$`},
	}

	sideHeader := panelHeaderStyle.Copy().Width(sideInnerW).Render(
		lipgloss.NewStyle().Foreground(textMuted).Render("EXAMPLES"),
	)

	var sideItems []string
	sideItems = append(sideItems, sideHeader)
	for _, ex := range examples {
		nameRow := lipgloss.NewStyle().Foreground(textMuted).Render(ex.name)
		patRow := lipgloss.NewStyle().Foreground(accent).Render("/" + ex.pattern + "/")
		item := lipgloss.JoinVertical(lipgloss.Left, nameRow, patRow)
		sideItems = append(sideItems, listItemStyle.Copy().Width(sideInnerW).Render(item))
	}

	rightCol := sidePanel.Render(lipgloss.JoinVertical(lipgloss.Left, sideItems...))

	mainGrid := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "  ", rightCol)
	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", mainGrid))
}

// renderFlags builds the four flag toggle buttons for the pattern panel.
func (m Model) renderFlags() string {
	flagsFocused := m.activeFocus == focusFlags
	states := []bool{m.flagG, m.flagI, m.flagM, m.flagS}

	parts := make([]string, 0, len(flagDefs)*2)
	for idx, fd := range flagDefs {
		on := states[idx]
		var btn string
		switch {
		case flagsFocused && on:
			btn = flagFocusOnStyle.Render(fd.label)
		case flagsFocused && !on:
			btn = flagFocusOffStyle.Render(fd.label)
		case on:
			btn = flagOnStyle.Render(fd.label)
		default:
			btn = flagOffStyle.Render(fd.label)
		}
		parts = append(parts, btn)
		if idx < len(flagDefs)-1 {
			parts = append(parts, " ") // explicit space — no MarginRight
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// ── badges ────────────────────────────────────────────────────────────────────

func (m Model) patternBadge() string {
	if m.matchErr != nil {
		return badgeErrorStyle.Render("error")
	}
	if m.patternInput.Value() == "" {
		return badgeDimStyle.Render("idle")
	}
	return badgeSuccessStyle.Render("valid")
}

func (m Model) renderMatchBadge() string {
	if m.matchErr != nil {
		return badgeErrorStyle.Render("error")
	}
	if m.matchCount == 0 {
		return badgeDimStyle.Render("no match")
	}
	if m.matchCount == 1 {
		return badgeSuccessStyle.Render("1 match")
	}
	return badgeSuccessStyle.Render(fmt.Sprintf("%d matches", m.matchCount))
}

// renderStatsBar builds "Matches N   Groups N" with proper spacing.
// The gap is plain spaces — NOT a nested Render call — so the inner
// reset escape from statValStyle doesn't bleed into surrounding text.
func (m Model) renderStatsBar(width int) string {
	left := statStyle.Render("Matches ") + statValStyle.Render(fmt.Sprintf("%d", m.matchCount))
	right := statStyle.Render("Groups ") + statValStyle.Render(fmt.Sprintf("%d", m.groupCount))
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

// ── matching ──────────────────────────────────────────────────────────────────

// effectivePattern prepends a Go inline-flag group for any active regex flags.
func (m Model) effectivePattern() string {
	var fb strings.Builder
	if m.flagI {
		fb.WriteString("i")
	}
	if m.flagM {
		fb.WriteString("m")
	}
	if m.flagS {
		fb.WriteString("s")
	}
	pattern := m.patternInput.Value()
	if fb.Len() > 0 {
		return "(?" + fb.String() + ")" + pattern
	}
	return pattern
}

func (m *Model) refresh() {
	m.matchErr = nil
	m.matchCount = 0
	m.groupCount = 0
	m.ranges = nil

	if m.matcher == nil {
		return
	}

	pattern := m.effectivePattern()
	text := m.textInput.Value()

	if pattern == "" || text == "" {
		return
	}

	ranges, err := m.matcher.MatchRanges(pattern, text)
	if err != nil {
		m.matchErr = err
		return
	}

	// global=false → show only the first match.
	if !m.flagG && len(ranges) > 1 {
		ranges = ranges[:1]
	}

	m.ranges = ranges
	m.matchCount = len(ranges)
	m.groupCount = m.matchCount
}

// ── helpers ───────────────────────────────────────────────────────────────────

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
