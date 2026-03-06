package parser

import (
	"regexp"
	"strings"

	"regex-parser/internal/tui"
	"regex-parser/pkg/lexer"
)

type RegexMatcher struct{}

func NewRegexMatcher() *RegexMatcher {
	return &RegexMatcher{}
}

func (m *RegexMatcher) MatchRanges(pattern, text string) ([]tui.Range, error) {
	if pattern == "" || text == "" {
		return nil, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	matches := re.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	ranges := make([]tui.Range, 0, len(matches))
	styleIdx := 0
	for _, match := range matches {
		if len(match) >= 4 {
			ranges = append(ranges, tui.Range{
				Start: match[0],
				End:   match[1],
				Style: styleIdx % 2,
			})
			styleIdx++
		}
	}

	return ranges, nil
}

func (m *RegexMatcher) Explain(pattern string) (string, error) {
	if pattern == "" {
		return "", nil
	}

	tokens, err := lexer.New(pattern).Tokenize()
	if err != nil {
		return "", err
	}

	p := NewParser(tokens)
	ast, err := p.Parse()
	if err != nil {
		return "", err
	}

	return explainNode(ast), nil
}

func explainNode(node ASTNode) string {
	switch n := node.(type) {
	case *CharNode:
		return "matches the character '" + n.Char + "' literally"
	case *DotNode:
		return "matches any single character (except newline)"
	case *CharClassNode:
		return explainCharClass(n)
	case *PredefinedClassNode:
		return explainPredefinedClass(n)
	case *QuantifierNode:
		return explainQuantifier(n)
	case *ConcatNode:
		return explainConcat(n)
	case *AlternationNode:
		return explainAlternation(n)
	case *BackreferenceNode:
		return "matches the same text as matched by group #" + string(rune(n.GroupNumber+'0'))
	case *AnchorNode:
		return explainAnchor(n)
	case *GroupNode:
		return "capturing group #" + string(rune(n.GroupNumber+'0')) + ": " + explainNode(n.Child)
	case *NonCapturingGroupNode:
		return "non-capturing group: " + explainNode(n.Child)
	case *LookaheadNode:
		prefix := "negative lookahead: "
		if n.Positive {
			prefix = "positive lookahead: "
		}
		return prefix + explainNode(n.Child)
	case *LookbehindNode:
		prefix := "negative lookbehind: "
		if n.Positive {
			prefix = "positive lookbehind: "
		}
		return prefix + explainNode(n.Child)
	default:
		return "unknown node"
	}
}

func explainCharClass(n *CharClassNode) string {
	prefix := "matches any character in: "
	if n.Negated {
		prefix = "matches any character NOT in: "
	}

	chars := make([]string, 0, len(n.Chars))
	for c := range n.Chars {
		chars = append(chars, c)
	}

	if len(chars) > 5 {
		return prefix + strings.Join(chars[:5], ", ") + "..."
	}
	return prefix + strings.Join(chars, ", ")
}

func explainPredefinedClass(n *PredefinedClassNode) string {
	descriptions := map[string]string{
		"d": "matches any digit (0-9)",
		"D": "matches any non-digit",
		"w": "matches any word character (a-z, A-Z, 0-9, _)",
		"W": "matches any non-word character",
		"s": "matches any whitespace character",
		"S": "matches any non-whitespace character",
	}
	if desc, ok := descriptions[n.ClassType]; ok {
		return desc
	}
	return "matches predefined class \\" + n.ClassType
}

func explainQuantifier(n *QuantifierNode) string {
	childDesc := explainNode(n.Child)

	var quant string
	switch {
	case n.MinCount == 0 && n.MaxCount != nil && *n.MaxCount == 1:
		if n.Greedy {
			quant = "optional (0 or 1)"
		} else {
			quant = "optional (0 or 1, lazy)"
		}
	case n.MinCount == 0 && n.MaxCount == nil:
		if n.Greedy {
			quant = "zero or more"
		} else {
			quant = "zero or more (lazy)"
		}
	case n.MinCount == 1 && n.MaxCount == nil:
		if n.Greedy {
			quant = "one or more"
		} else {
			quant = "one or more (lazy)"
		}
	default:
		maxText := "unlimited"
		if n.MaxCount != nil {
			maxText = string(rune(*n.MaxCount + '0'))
		}
		if n.Greedy {
			quant = "between " + string(rune(n.MinCount+'0')) + " and " + maxText + " times"
		} else {
			quant = "between " + string(rune(n.MinCount+'0')) + " and " + maxText + " times (lazy)"
		}
	}

	return childDesc + " " + quant
}

func explainConcat(n *ConcatNode) string {
	if len(n.Children) == 0 {
		return "empty sequence"
	}

	parts := make([]string, len(n.Children))
	for i, child := range n.Children {
		parts[i] = explainNode(child)
	}
	return strings.Join(parts, ", then ")
}

func explainAlternation(n *AlternationNode) string {
	if len(n.Alternatives) == 0 {
		return "empty alternation"
	}

	parts := make([]string, len(n.Alternatives))
	for i, alt := range n.Alternatives {
		parts[i] = explainNode(alt)
	}
	return "matches one of: " + strings.Join(parts, " OR ")
}

func explainAnchor(n *AnchorNode) string {
	descriptions := map[string]string{
		"^": "matches the start of the string/line",
		"$": "matches the end of the string/line",
		"b": "matches a word boundary",
		"B": "matches a non-word boundary",
	}
	if desc, ok := descriptions[n.AnchorType]; ok {
		return desc
	}
	return "matches anchor " + n.AnchorType
}
