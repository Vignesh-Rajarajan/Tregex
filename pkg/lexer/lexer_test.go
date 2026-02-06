package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "empty string",
			input: "",
			expected: []Token{
				{Type: TokenEOF, Value: "", Pos: 0},
			},
		},
		{
			name:  "basic characters",
			input: "abc",
			expected: []Token{
				{Type: TokenChar, Value: "a", Pos: 0},
				{Type: TokenChar, Value: "b", Pos: 1},
				{Type: TokenChar, Value: "c", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "basic operators",
			input: ".*+?|",
			expected: []Token{
				{Type: TokenDot, Value: ".", Pos: 0},
				{Type: TokenStar, Value: "*", Pos: 1},
				{Type: TokenPlus, Value: "+", Pos: 2},
				{Type: TokenQuestion, Value: "?", Pos: 3},
				{Type: TokenPipe, Value: "|", Pos: 4},
				{Type: TokenEOF, Value: "", Pos: 5},
			},
		},
		{
			name:  "parentheses",
			input: "()",
			expected: []Token{
				{Type: TokenLParen, Value: "(", Pos: 0},
				{Type: TokenRParen, Value: ")", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "brackets",
			input: "[]",
			expected: []Token{
				{Type: TokenLBracket, Value: "[", Pos: 0},
				{Type: TokenRBracket, Value: "]", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "braces",
			input: "{}",
			expected: []Token{
				{Type: TokenLBrace, Value: "{", Pos: 0},
				{Type: TokenRBrace, Value: "}", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "anchors",
			input: "^$",
			expected: []Token{
				{Type: TokenCaret, Value: "^", Pos: 0},
				{Type: TokenDollar, Value: "$", Pos: 1},
				{Type: TokenEOF, Value: "", Pos: 2},
			},
		},
		{
			name:  "range quantifier",
			input: "{1,5}",
			expected: []Token{
				{Type: TokenLBrace, Value: "{", Pos: 0},
				{Type: TokenChar, Value: "1", Pos: 1},
				{Type: TokenComma, Value: ",", Pos: 2},
				{Type: TokenChar, Value: "5", Pos: 3},
				{Type: TokenRBrace, Value: "}", Pos: 4},
				{Type: TokenEOF, Value: "", Pos: 5},
			},
		},
		{
			name:  "character class with dash",
			input: "[a-z]",
			expected: []Token{
				{Type: TokenLBracket, Value: "[", Pos: 0},
				{Type: TokenChar, Value: "a", Pos: 1},
				{Type: TokenMinus, Value: "-", Pos: 2},
				{Type: TokenChar, Value: "z", Pos: 3},
				{Type: TokenRBracket, Value: "]", Pos: 4},
				{Type: TokenEOF, Value: "", Pos: 5},
			},
		},
		{
			name:  "escape sequences - digit classes",
			input: "\\d\\D",
			expected: []Token{
				{Type: TokenDigit, Value: "\\d", Pos: 0},
				{Type: TokenNonDigit, Value: "\\D", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "escape sequences - word classes",
			input: "\\w\\W",
			expected: []Token{
				{Type: TokenWord, Value: "\\w", Pos: 0},
				{Type: TokenNonWord, Value: "\\W", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "escape sequences - whitespace classes",
			input: "\\s\\S",
			expected: []Token{
				{Type: TokenWhitespace, Value: "\\s", Pos: 0},
				{Type: TokenNonWhitespace, Value: "\\S", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "escape sequences - word boundaries",
			input: "\\b\\B",
			expected: []Token{
				{Type: TokenWordBoundary, Value: "\\b", Pos: 0},
				{Type: TokenNonWordBoundary, Value: "\\B", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 4},
			},
		},
		{
			name:  "escape sequences - backreferences",
			input: "\\1\\2\\9",
			expected: []Token{
				{Type: TokenBackref, Value: "1", Pos: 0},
				{Type: TokenBackref, Value: "2", Pos: 2},
				{Type: TokenBackref, Value: "9", Pos: 4},
				{Type: TokenEOF, Value: "", Pos: 6},
			},
		},
		{
			name:  "escape sequences - literal escapes",
			input: "\\.\\*\\+\\?\\|\\(\\)",
			expected: []Token{
				{Type: TokenChar, Value: ".", Pos: 0},
				{Type: TokenChar, Value: "*", Pos: 2},
				{Type: TokenChar, Value: "+", Pos: 4},
				{Type: TokenChar, Value: "?", Pos: 6},
				{Type: TokenChar, Value: "|", Pos: 8},
				{Type: TokenChar, Value: "(", Pos: 10},
				{Type: TokenChar, Value: ")", Pos: 12},
				{Type: TokenEOF, Value: "", Pos: 14},
			},
		},
		{
			name:  "lookahead positive",
			input: "(?=abc)",
			expected: []Token{
				{Type: TokenLookaheadPos, Value: "(?=", Pos: 0},
				{Type: TokenChar, Value: "a", Pos: 3},
				{Type: TokenChar, Value: "b", Pos: 4},
				{Type: TokenChar, Value: "c", Pos: 5},
				{Type: TokenRParen, Value: ")", Pos: 6},
				{Type: TokenEOF, Value: "", Pos: 7},
			},
		},
		{
			name:  "lookahead negative",
			input: "(?!abc)",
			expected: []Token{
				{Type: TokenLookaheadNeg, Value: "(?!", Pos: 0},
				{Type: TokenChar, Value: "a", Pos: 3},
				{Type: TokenChar, Value: "b", Pos: 4},
				{Type: TokenChar, Value: "c", Pos: 5},
				{Type: TokenRParen, Value: ")", Pos: 6},
				{Type: TokenEOF, Value: "", Pos: 7},
			},
		},
		{
			name:  "lookbehind positive",
			input: "(?<=abc)",
			expected: []Token{
				{Type: TokenLookbehindPos, Value: "(?<=", Pos: 0},
				{Type: TokenChar, Value: "a", Pos: 4},
				{Type: TokenChar, Value: "b", Pos: 5},
				{Type: TokenChar, Value: "c", Pos: 6},
				{Type: TokenRParen, Value: ")", Pos: 7},
				{Type: TokenEOF, Value: "", Pos: 8},
			},
		},
		{
			name:  "lookbehind negative",
			input: "(?<!abc)",
			expected: []Token{
				{Type: TokenLookbehindNeg, Value: "(?<!", Pos: 0},
				{Type: TokenChar, Value: "a", Pos: 4},
				{Type: TokenChar, Value: "b", Pos: 5},
				{Type: TokenChar, Value: "c", Pos: 6},
				{Type: TokenRParen, Value: ")", Pos: 7},
				{Type: TokenEOF, Value: "", Pos: 8},
			},
		},
		{
			name:  "non-capturing group",
			input: "(?:abc)",
			expected: []Token{
				{Type: TokenNonCapturing, Value: "(?:", Pos: 0},
				{Type: TokenChar, Value: "a", Pos: 3},
				{Type: TokenChar, Value: "b", Pos: 4},
				{Type: TokenChar, Value: "c", Pos: 5},
				{Type: TokenRParen, Value: ")", Pos: 6},
				{Type: TokenEOF, Value: "", Pos: 7},
			},
		},
		{
			name:  "complex pattern",
			input: "^\\d+[a-z]*$",
			expected: []Token{
				{Type: TokenCaret, Value: "^", Pos: 0},
				{Type: TokenDigit, Value: "\\d", Pos: 1},
				{Type: TokenPlus, Value: "+", Pos: 3},
				{Type: TokenLBracket, Value: "[", Pos: 4},
				{Type: TokenChar, Value: "a", Pos: 5},
				{Type: TokenMinus, Value: "-", Pos: 6},
				{Type: TokenChar, Value: "z", Pos: 7},
				{Type: TokenRBracket, Value: "]", Pos: 8},
				{Type: TokenStar, Value: "*", Pos: 9},
				{Type: TokenDollar, Value: "$", Pos: 10},
				{Type: TokenEOF, Value: "", Pos: 11},
			},
		},
		{
			name:  "quantifiers with ranges",
			input: "a{1,3}?",
			expected: []Token{
				{Type: TokenChar, Value: "a", Pos: 0},
				{Type: TokenLBrace, Value: "{", Pos: 1},
				{Type: TokenChar, Value: "1", Pos: 2},
				{Type: TokenComma, Value: ",", Pos: 3},
				{Type: TokenChar, Value: "3", Pos: 4},
				{Type: TokenRBrace, Value: "}", Pos: 5},
				{Type: TokenQuestion, Value: "?", Pos: 6},
				{Type: TokenEOF, Value: "", Pos: 7},
			},
		},
		{
			name:  "mixed escape sequences",
			input: "\\d\\w\\s\\b",
			expected: []Token{
				{Type: TokenDigit, Value: "\\d", Pos: 0},
				{Type: TokenWord, Value: "\\w", Pos: 2},
				{Type: TokenWhitespace, Value: "\\s", Pos: 4},
				{Type: TokenWordBoundary, Value: "\\b", Pos: 6},
				{Type: TokenEOF, Value: "", Pos: 8},
			},
		},
		{
			name:  "question mark as quantifier",
			input: "a?b",
			expected: []Token{
				{Type: TokenChar, Value: "a", Pos: 0},
				{Type: TokenQuestion, Value: "?", Pos: 1},
				{Type: TokenChar, Value: "b", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "lookahead after question",
			input: "a(?=b)",
			expected: []Token{
				{Type: TokenChar, Value: "a", Pos: 0},
				{Type: TokenLookaheadPos, Value: "(?=", Pos: 1},
				{Type: TokenChar, Value: "b", Pos: 4},
				{Type: TokenRParen, Value: ")", Pos: 5},
				{Type: TokenEOF, Value: "", Pos: 6},
			},
		},
		{
			name:  "unicode characters",
			input: "αβγ",
			expected: []Token{
				{Type: TokenChar, Value: "α", Pos: 0},
				{Type: TokenChar, Value: "β", Pos: 1},
				{Type: TokenChar, Value: "γ", Pos: 2},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "escape sequences - multiline backref",
			input: "\\12",
			expected: []Token{
				{Type: TokenBackref, Value: "12", Pos: 0},
				{Type: TokenEOF, Value: "", Pos: 3},
			},
		},
		{
			name:  "escape sequences - newline tab carriage return",
			input: "\\n\\t\\r",
			expected: []Token{
				{Type: TokenChar, Value: "\n", Pos: 0},
				{Type: TokenChar, Value: "\t", Pos: 2},
				{Type: TokenChar, Value: "\r", Pos: 4},
				{Type: TokenEOF, Value: "", Pos: 6},
			},
		},
		{
			name:  "actual pattern",
			input: "a\\d{1}(?:foo|bar)?",
			expected: []Token{
				{Type: TokenChar, Value: "a", Pos: 0},
				{Type: TokenDigit, Value: "\\d", Pos: 1},
				{Type: TokenLBrace, Value: "{", Pos: 3},
				{Type: TokenChar, Value: "1", Pos: 4},
				{Type: TokenRBrace, Value: "}", Pos: 5},
				{Type: TokenNonCapturing, Value: "(?:", Pos: 6},
				{Type: TokenChar, Value: "f", Pos: 9},
				{Type: TokenChar, Value: "o", Pos: 10},
				{Type: TokenChar, Value: "o", Pos: 11},
				{Type: TokenPipe, Value: "|", Pos: 12},
				{Type: TokenChar, Value: "b", Pos: 13},
				{Type: TokenChar, Value: "a", Pos: 14},
				{Type: TokenChar, Value: "r", Pos: 15},
				{Type: TokenRParen, Value: ")", Pos: 16},
				{Type: TokenQuestion, Value: "?", Pos: 17},
				{Type: TokenEOF, Value: "", Pos: 18},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)
			for i, expected := range tt.expected {
				token, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("token[%d] unexpected error: %v", i, err)
				}
				if token.Type != expected.Type {
					t.Errorf("token[%d] type mismatch: expected %v, got %v", i, expected.Type, token.Type)
				}
				if token.Value != expected.Value {
					t.Errorf("token[%d] value mismatch: expected %q, got %q", i, expected.Value, token.Value)
				}
				if token.Pos != expected.Pos {
					t.Errorf("token[%d] position mismatch: expected %d, got %d", i, expected.Pos, token.Pos)
				}
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:     "simple pattern",
			input:    "abc",
			expected: []TokenType{TokenChar, TokenChar, TokenChar, TokenEOF},
		},
		{
			name:     "pattern with operators",
			input:    "a.*b+",
			expected: []TokenType{TokenChar, TokenDot, TokenStar, TokenChar, TokenPlus, TokenEOF},
		},
		{
			name:     "pattern with escape sequences",
			input:    "\\d\\w",
			expected: []TokenType{TokenDigit, TokenWord, TokenEOF},
		},
		{
			name:     "pattern with lookahead",
			input:    "(?=abc)",
			expected: []TokenType{TokenLookaheadPos, TokenChar, TokenChar, TokenChar, TokenRParen, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Errorf("token count mismatch: expected %d, got %d", len(tt.expected), len(tokens))
				return
			}
			for i, expectedType := range tt.expected {
				if tokens[i].Type != expectedType {
					t.Errorf("token[%d] type mismatch: expected %v, got %v", i, expectedType, tokens[i].Type)
				}
			}
		})
	}
}

func TestPeek(t *testing.T) {
	lexer := New("abc")
	if lexer.Peek() != 'a' {
		t.Errorf("Peek() = %c, want 'a'", lexer.Peek())
	}
	if lexer.Position() != 0 {
		t.Errorf("Position() = %d, want 0", lexer.Position())
	}

	_, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lexer.Peek() != 'b' {
		t.Errorf("Peek() = %c, want 'b'", lexer.Peek())
	}
	if lexer.Position() != 1 {
		t.Errorf("Position() = %d, want 1", lexer.Position())
	}
}

func TestPosition(t *testing.T) {
	lexer := New("abc")
	if lexer.Position() != 0 {
		t.Errorf("Initial Position() = %d, want 0", lexer.Position())
	}

	_, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lexer.Position() != 1 {
		t.Errorf("After first token Position() = %d, want 1", lexer.Position())
	}

	_, err = lexer.NextToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lexer.Position() != 2 {
		t.Errorf("After second token Position() = %d, want 2", lexer.Position())
	}
}

func TestTokenizeErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "trailing backslash",
			input:         "\\",
			expectedError: "pattern cannot end with backslash at position 0",
		},
		{
			name:          "unknown group modifier",
			input:         "(?x",
			expectedError: "unknown group modifier '?x' at position 0",
		},
		{
			name:          "invalid lookbehind group",
			input:         "(?<x",
			expectedError: "invalid group syntax at position 0",
		},
		{
			name:          "missing group modifier",
			input:         "(?",
			expectedError: "unknown group modifier '?' at position 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)
			_, err := lexer.Tokenize()
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.expectedError)
			}
			if err.Error() != tt.expectedError {
				t.Fatalf("expected error %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}
