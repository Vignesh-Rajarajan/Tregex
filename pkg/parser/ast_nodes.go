package parser

import (
	"fmt"
	"sort"
	"strings"
)

type ASTNode interface {
	isASTNode()
}

type CharNode struct {
	Char string
}

func (*CharNode) isASTNode() {}

func (n *CharNode) String() string {
	return fmt.Sprintf("Char(%q)", n.Char)
}

type DotNode struct{}

func (*DotNode) isASTNode() {}

func (*DotNode) String() string {
	return "Dot(.)"
}

type CharClassNode struct {
	Chars   map[string]struct{} // [a-z] is a char class and map stores entire range
	Negated bool                // ^[a-z] is a negated char class
}

func (*CharClassNode) isASTNode() {}

func (n *CharClassNode) String() string {
	prefix := ""
	if n.Negated {
		prefix = "^"
	}
	if len(n.Chars) == 0 {
		return "CharClass([" + prefix + "])"
	}

	chars := make([]string, 0, len(n.Chars))
	for c := range n.Chars {
		chars = append(chars, c)
	}
	sort.Strings(chars)

	limit := 10
	if len(chars) < limit {
		limit = len(chars)
	}
	charsStr := strings.Join(chars[:limit], "")
	if len(chars) > 10 {
		charsStr += "..."
	}
	return "CharClass([" + prefix + charsStr + "])"
}

type PredefinedClassNode struct {
	ClassType string // \d, \D, \w, \W, \s, \S, \b, \B
}

func (*PredefinedClassNode) isASTNode() {}

func (n *PredefinedClassNode) String() string {
	return "PredefinedClass(\\" + n.ClassType + ")"
}

type QuantifierNode struct {
	Child    ASTNode
	MinCount int  // for range quantifiers {n,m} for example {1,3} is a quantifier and 1 is the min count and 3 is the max count
	MaxCount *int // for range quantifiers {n,m} for example {1,3} is a quantifier and 3 is the max count
	Greedy   bool // greedy quantifier is * and tries to match as much as possible
}

func (*QuantifierNode) isASTNode() {}

func (n *QuantifierNode) String() string {
	var q string
	switch {
	case n.MinCount == 0 && n.MaxCount != nil && *n.MaxCount == 1:
		if n.Greedy {
			q = "?"
		} else {
			q = "??"
		}
	case n.MinCount == 0 && n.MaxCount == nil:
		if n.Greedy {
			q = "*"
		} else {
			q = "*?"
		}
	case n.MinCount == 1 && n.MaxCount == nil:
		if n.Greedy {
			q = "+"
		} else {
			q = "+?"
		}
	default:
		maxText := "None"
		if n.MaxCount != nil {
			maxText = fmt.Sprintf("%d", *n.MaxCount)
		}
		q = fmt.Sprintf("{%d,%s}", n.MinCount, maxText)
		if !n.Greedy {
			q += "?"
		}
	}
	return fmt.Sprintf("Quantifier(%v %s)", n.Child, q)
}

type ConcatNode struct {
	Children []ASTNode
}

func (*ConcatNode) isASTNode() {}

func (n *ConcatNode) String() string {
	return fmt.Sprintf("Concat(%v)", n.Children)
}

type AlternationNode struct {
	Alternatives []ASTNode
}

func (*AlternationNode) isASTNode() {}

func (n *AlternationNode) String() string {
	return fmt.Sprintf("Alternation(%v)", n.Alternatives)
}

type BackreferenceNode struct {
	GroupNumber int
}

func (*BackreferenceNode) isASTNode() {}

func (n *BackreferenceNode) String() string {
	return fmt.Sprintf("Backref(\\%d)", n.GroupNumber)
}

type AnchorNode struct {
	AnchorType string
}

func (*AnchorNode) isASTNode() {}

func (n *AnchorNode) String() string {
	symbols := map[string]string{
		"^": "^",
		"$": "$",
		"b": `\b`,
		"B": `\B`,
	}
	if symbol, ok := symbols[n.AnchorType]; ok {
		return "Anchor(" + symbol + ")"
	}
	return "Anchor(" + n.AnchorType + ")"
}

type NonCapturingGroupNode struct {
	Child ASTNode
}

func (*NonCapturingGroupNode) isASTNode() {}

func (n *NonCapturingGroupNode) String() string {
	return fmt.Sprintf("NonCapturingGroup(%v)", n.Child)
}

type GroupNode struct {
	Child       ASTNode
	GroupNumber int
}

func (*GroupNode) isASTNode() {}

func (n *GroupNode) String() string {
	return fmt.Sprintf("Group#%d(%v)", n.GroupNumber, n.Child)
}

type LookaheadNode struct {
	Child    ASTNode
	Positive bool
}

func (*LookaheadNode) isASTNode() {}

func (n *LookaheadNode) String() string {
	prefix := "?!"
	if n.Positive {
		prefix = "?="
	}
	return fmt.Sprintf("Lookahead(%s%v)", prefix, n.Child)
}

type LookbehindNode struct {
	Child    ASTNode
	Positive bool
}

func (*LookbehindNode) isASTNode() {}

func (n *LookbehindNode) String() string {
	prefix := "?<!"
	if n.Positive {
		prefix = "?<="
	}
	return fmt.Sprintf("Lookbehind(%s%v)", prefix, n.Child)
}
