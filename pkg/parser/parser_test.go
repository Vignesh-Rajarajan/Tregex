package parser

import (
	"reflect"
	"testing"

	"tregex/pkg/lexer"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    ASTNode
		wantErr string
	}{
		{
			name:    "empty pattern",
			pattern: "",
			want:    &ConcatNode{Children: []ASTNode{}},
		},
		{
			name:    "single char",
			pattern: "a",
			want:    &CharNode{Char: "a"},
		},
		{
			name:    "concatenation",
			pattern: "ab",
			want: &ConcatNode{Children: []ASTNode{
				&CharNode{Char: "a"},
				&CharNode{Char: "b"},
			}},
		},
		{
			name:    "alternation",
			pattern: "a|b|c",
			want: &AlternationNode{Alternatives: []ASTNode{
				&CharNode{Char: "a"},
				&CharNode{Char: "b"},
				&CharNode{Char: "c"},
			}},
		},
		{
			name:    "alternation trailing empty",
			pattern: "a|",
			want: &AlternationNode{Alternatives: []ASTNode{
				&CharNode{Char: "a"},
				&ConcatNode{Children: []ASTNode{}},
			}},
		},
		{
			name:    "alternation leading empty",
			pattern: "|a",
			want: &AlternationNode{Alternatives: []ASTNode{
				&ConcatNode{Children: []ASTNode{}},
				&CharNode{Char: "a"},
			}},
		},
		{
			name:    "literal dash and comma",
			pattern: "-,",
			want: &ConcatNode{Children: []ASTNode{
				&CharNode{Char: "-"},
				&CharNode{Char: ","},
			}},
		},
		{
			name:    "star lazy",
			pattern: "a*?",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 0, MaxCount: nil, Greedy: false},
		},
		{
			name:    "plus lazy",
			pattern: "a+?",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 1, MaxCount: nil, Greedy: false},
		},
		{
			name:    "question lazy",
			pattern: "a??",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 0, MaxCount: intPtr(1), Greedy: false},
		},
		{
			name:    "range quantifier exact",
			pattern: "a{2}",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 2, MaxCount: intPtr(2), Greedy: true},
		},
		{
			name:    "range quantifier open",
			pattern: "a{2,}",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 2, MaxCount: nil, Greedy: true},
		},
		{
			name:    "range quantifier lazy",
			pattern: "a{2,4}?",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 2, MaxCount: intPtr(4), Greedy: false},
		},
		{
			name:    "range quantifier multi-digit",
			pattern: "a{12,34}",
			want:    &QuantifierNode{Child: &CharNode{Char: "a"}, MinCount: 12, MaxCount: intPtr(34), Greedy: true},
		},
		{
			name:    "anchors and predefined",
			pattern: "^\\d+$",
			want: &ConcatNode{Children: []ASTNode{
				&AnchorNode{AnchorType: "^"},
				&QuantifierNode{Child: &PredefinedClassNode{ClassType: "d"}, MinCount: 1, MaxCount: nil, Greedy: true},
				&AnchorNode{AnchorType: "$"},
			}},
		},
		{
			name:    "backreference multi-digit",
			pattern: "\\12",
			want:    &BackreferenceNode{GroupNumber: 12},
		},
		{
			name:    "capturing group",
			pattern: "(ab)",
			want: &GroupNode{
				GroupNumber: 1,
				Child: &ConcatNode{Children: []ASTNode{
					&CharNode{Char: "a"},
					&CharNode{Char: "b"},
				}},
			},
		},
		{
			name:    "nested groups",
			pattern: "(a(b))",
			want: &GroupNode{
				GroupNumber: 1,
				Child: &ConcatNode{Children: []ASTNode{
					&CharNode{Char: "a"},
					&GroupNode{GroupNumber: 2, Child: &CharNode{Char: "b"}},
				}},
			},
		},
		{
			name:    "non-capturing group",
			pattern: "(?:ab)",
			want: &NonCapturingGroupNode{Child: &ConcatNode{Children: []ASTNode{
				&CharNode{Char: "a"},
				&CharNode{Char: "b"},
			}}},
		},
		{
			name:    "lookahead positive",
			pattern: "(?=a)",
			want:    &LookaheadNode{Child: &CharNode{Char: "a"}, Positive: true},
		},
		{
			name:    "lookahead negative",
			pattern: "(?!a)",
			want:    &LookaheadNode{Child: &CharNode{Char: "a"}, Positive: false},
		},
		{
			name:    "lookbehind positive",
			pattern: "(?<=a)",
			want:    &LookbehindNode{Child: &CharNode{Char: "a"}, Positive: true},
		},
		{
			name:    "lookbehind negative",
			pattern: "(?<!a)",
			want:    &LookbehindNode{Child: &CharNode{Char: "a"}, Positive: false},
		},
		{
			name:    "char class basic",
			pattern: "[abc]",
			want:    &CharClassNode{Chars: setFromString("abc"), Negated: false},
		},
		{
			name:    "char class range",
			pattern: "[a-c]",
			want:    &CharClassNode{Chars: setFromString("abc"), Negated: false},
		},
		{
			name:    "char class negated",
			pattern: "[^a]",
			want:    &CharClassNode{Chars: setFromString("a"), Negated: true},
		},
		{
			name:    "char class dash literal tail",
			pattern: "[a-]",
			want:    &CharClassNode{Chars: setFromString("a-"), Negated: false},
		},
		{
			name:    "char class dash literal head",
			pattern: "[-a]",
			want:    &CharClassNode{Chars: setFromString("-a"), Negated: false},
		},
		{
			name:    "char class digit",
			pattern: "[\\d]",
			want:    &CharClassNode{Chars: digitSet(), Negated: false},
		},
		{
			name:    "char class whitespace",
			pattern: "[\\s]",
			want:    &CharClassNode{Chars: setFromString(" \t\n\r\f\v"), Negated: false},
		},
		{
			name:    "char class literals",
			pattern: "[+*?]",
			want:    &CharClassNode{Chars: setFromString("+*?"), Negated: false},
		},
		{
			name:    "unclosed class",
			pattern: "[abc",
			wantErr: "Unclosed character class",
		},
		{
			name:    "empty class",
			pattern: "[]",
			wantErr: "Empty character class",
		},
		{
			name:    "empty negated class",
			pattern: "[^]",
			wantErr: "Empty character class",
		},
		{
			name:    "invalid range",
			pattern: "[z-a]",
			wantErr: "Invalid range z-a: start > end",
		},
		{
			name:    "range missing min",
			pattern: "a{,}",
			wantErr: "Expected number in quantifier at position 2",
		},
		{
			name:    "range invalid max",
			pattern: "a{1,x}",
			wantErr: "Expected number or '}' at position 4",
		},
		{
			name:    "unexpected closing paren",
			pattern: "a)",
			wantErr: "Unexpected token RPAREN at position 1",
		},
		{
			name:    "missing closing paren",
			pattern: "(a",
			wantErr: "Expected RPAREN, got EOF at position 2",
		},
		{
			name:    "unexpected leading quantifier",
			pattern: "*a",
			wantErr: "Unexpected token STAR at position 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := lexer.New(tt.pattern).Tokenize()
			if err != nil {
				t.Fatalf("tokenize error: %v", err)
			}
			parser := NewParser(tokens)
			got, err := parser.Parse()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ast mismatch:\nwant: %v\ngot:  %v", tt.want, got)
			}
		})
	}
}

func TestParseEndToEnd(t *testing.T) {
	pattern := "a\\d{1}(?:foo|bar)?"
	tokens, err := lexer.New(pattern).Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	concat, ok := ast.(*ConcatNode)
	if !ok {
		t.Fatalf("expected ConcatNode root, got %T", ast)
	}
	if len(concat.Children) != 3 {
		t.Fatalf("expected 3 concat children, got %d", len(concat.Children))
	}
}

func TestParseEndToEndAlternationAndQuantifier(t *testing.T) {
	pattern := "(ab|cd)+e"
	tokens, err := lexer.New(pattern).Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	concat, ok := ast.(*ConcatNode)
	if !ok {
		t.Fatalf("expected ConcatNode root, got %T", ast)
	}
	if len(concat.Children) != 2 {
		t.Fatalf("expected 2 concat children, got %d", len(concat.Children))
	}

	quant, ok := concat.Children[0].(*QuantifierNode)
	if !ok {
		t.Fatalf("expected QuantifierNode first child, got %T", concat.Children[0])
	}
	if quant.MinCount != 1 || quant.MaxCount != nil || !quant.Greedy {
		t.Fatalf("unexpected quantifier settings: %+v", quant)
	}

	group, ok := quant.Child.(*GroupNode)
	if !ok {
		t.Fatalf("expected GroupNode under quantifier, got %T", quant.Child)
	}
	if group.GroupNumber != 1 {
		t.Fatalf("expected quantifier child to be group #1, got #%d", group.GroupNumber)
	}

	alt, ok := group.Child.(*AlternationNode)
	if !ok {
		t.Fatalf("expected AlternationNode inside group, got %T", group.Child)
	}
	if len(alt.Alternatives) != 2 {
		t.Fatalf("expected 2 alternation options, got %d", len(alt.Alternatives))
	}
	if _, ok := concat.Children[1].(*CharNode); !ok {
		t.Fatalf("expected CharNode second child, got %T", concat.Children[1])
	}
}

func TestParseEndToEndLookaroundAndAnchors(t *testing.T) {
	pattern := "^(?=ab)ab$"
	tokens, err := lexer.New(pattern).Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	concat, ok := ast.(*ConcatNode)
	if !ok {
		t.Fatalf("expected ConcatNode root, got %T", ast)
	}
	if len(concat.Children) != 5 {
		t.Fatalf("expected 5 concat children, got %d", len(concat.Children))
	}
	if _, ok := concat.Children[0].(*AnchorNode); !ok {
		t.Fatalf("expected AnchorNode first child, got %T", concat.Children[0])
	}
	if _, ok := concat.Children[1].(*LookaheadNode); !ok {
		t.Fatalf("expected LookaheadNode second child, got %T", concat.Children[1])
	}
	lookahead, ok := concat.Children[1].(*LookaheadNode)
	if !ok {
		t.Fatalf("expected LookaheadNode second child, got %T", concat.Children[1])
	}
	lookaheadConcat, ok := lookahead.Child.(*ConcatNode)
	if !ok {
		t.Fatalf("expected lookahead child to be ConcatNode, got %T", lookahead.Child)
	}
	if len(lookaheadConcat.Children) != 2 {
		t.Fatalf("expected lookahead concat to contain 2 chars, got %d", len(lookaheadConcat.Children))
	}
	if _, ok := concat.Children[2].(*CharNode); !ok {
		t.Fatalf("expected CharNode third child, got %T", concat.Children[2])
	}
	if _, ok := concat.Children[3].(*CharNode); !ok {
		t.Fatalf("expected CharNode fourth child, got %T", concat.Children[3])
	}
	if _, ok := concat.Children[4].(*AnchorNode); !ok {
		t.Fatalf("expected AnchorNode fifth child, got %T", concat.Children[4])
	}
}

func TestParseEndToEndCharClassConcat(t *testing.T) {
	pattern := "[a-c]\\w"
	tokens, err := lexer.New(pattern).Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	concat, ok := ast.(*ConcatNode)
	if !ok {
		t.Fatalf("expected ConcatNode root, got %T", ast)
	}
	if len(concat.Children) != 2 {
		t.Fatalf("expected 2 concat children, got %d", len(concat.Children))
	}
	if _, ok := concat.Children[0].(*CharClassNode); !ok {
		t.Fatalf("expected CharClassNode first child, got %T", concat.Children[0])
	}
	if _, ok := concat.Children[1].(*PredefinedClassNode); !ok {
		t.Fatalf("expected PredefinedClassNode second child, got %T", concat.Children[1])
	}
}

func intPtr(v int) *int {
	return &v
}

func setFromString(chars string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, ch := range chars {
		set[string(ch)] = struct{}{}
	}
	return set
}

func digitSet() map[string]struct{} {
	set := map[string]struct{}{}
	for ch := '0'; ch <= '9'; ch++ {
		set[string(ch)] = struct{}{}
	}
	return set
}
