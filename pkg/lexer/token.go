package lexer

// TokenType represents the type of a regex token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenChar
	TokenDot
	TokenStar
	TokenPlus
	TokenQuestion
	TokenPipe
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenCaret
	TokenMinus
	TokenEscape
	// Range quantifiers
	TokenLBrace // { for {n}, {n,}, {n,m}
	TokenRBrace // }
	TokenComma  // , for {n,m}
	// Anchors
	TokenDollar // $ end of string anchor
	// Escape character
	TokenBackslash // \ (for escape sequences)
	// Special escape sequences
	TokenDigit         // \d - Matches any digit
	TokenNonDigit      // \D - Matches any non-digit
	TokenWord          // \w - Matches any word character
	TokenNonWord       // \W - Matches any non-word character
	TokenWhitespace    // \s - Matches any whitespace character
	TokenNonWhitespace // \S - Matches any non-whitespace character
	// Word boundaries
	TokenWordBoundary    // \b
	TokenNonWordBoundary // \B
	// Backreferences
	TokenBackref // \1, \2, etc.
	// Lookahead/Lookbehind markers
	TokenLookaheadPos  // (?=
	TokenLookaheadNeg  // (?!
	TokenLookbehindPos // (?<=
	TokenLookbehindNeg // (?<!)
	TokenNonCapturing  // (?:
)

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenChar:
		return "CHAR"
	case TokenDot:
		return "DOT"
	case TokenStar:
		return "STAR"
	case TokenPlus:
		return "PLUS"
	case TokenQuestion:
		return "QUESTION"
	case TokenPipe:
		return "PIPE"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenLBracket:
		return "LBRACKET"
	case TokenRBracket:
		return "RBRACKET"
	case TokenCaret:
		return "CARET"
	case TokenMinus:
		return "DASH"
	case TokenEscape:
		return "ESCAPE"
	case TokenLBrace:
		return "LBRACE"
	case TokenRBrace:
		return "RBRACE"
	case TokenComma:
		return "COMMA"
	case TokenDollar:
		return "DOLLAR"
	case TokenBackslash:
		return "BACKSLASH"
	case TokenDigit:
		return "DIGIT"
	case TokenNonDigit:
		return "NON_DIGIT"
	case TokenWord:
		return "WORD"
	case TokenNonWord:
		return "NON_WORD"
	case TokenWhitespace:
		return "WHITESPACE"
	case TokenNonWhitespace:
		return "NON_WHITESPACE"
	case TokenWordBoundary:
		return "WORD_BOUNDARY"
	case TokenNonWordBoundary:
		return "NON_WORD_BOUNDARY"
	case TokenBackref:
		return "BACKREF"
	case TokenLookaheadPos:
		return "LOOKAHEAD_POS"
	case TokenLookaheadNeg:
		return "LOOKAHEAD_NEG"
	case TokenLookbehindPos:
		return "LOOKBEHIND_POS"
	case TokenLookbehindNeg:
		return "LOOKBEHIND_NEG"
	case TokenNonCapturing:
		return "NON_CAPTURING"
	default:
		return "UNKNOWN"
	}
}

// Token represents a single token from the regex input.
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenChar:
		return "CHAR('" + t.Value + "')"
	case TokenDot:
		return "DOT"
	case TokenStar:
		return "STAR"
	case TokenPlus:
		return "PLUS"
	case TokenQuestion:
		return "QUESTION"
	case TokenPipe:
		return "PIPE"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenLBracket:
		return "LBRACKET"
	case TokenRBracket:
		return "RBRACKET"
	case TokenCaret:
		return "CARET"
	case TokenMinus:
		return "DASH"
	case TokenEscape:
		return "ESCAPE('" + t.Value + "')"
	case TokenLBrace:
		return "LBRACE"
	case TokenRBrace:
		return "RBRACE"
	case TokenComma:
		return "COMMA"
	case TokenDollar:
		return "DOLLAR"
	case TokenBackslash:
		return "BACKSLASH"
	case TokenDigit:
		return "DIGIT"
	case TokenNonDigit:
		return "NON_DIGIT"
	case TokenWord:
		return "WORD"
	case TokenNonWord:
		return "NON_WORD"
	case TokenWhitespace:
		return "WHITESPACE"
	case TokenNonWhitespace:
		return "NON_WHITESPACE"
	case TokenWordBoundary:
		return "WORD_BOUNDARY"
	case TokenNonWordBoundary:
		return "NON_WORD_BOUNDARY"
	case TokenBackref:
		return "BACKREF('" + t.Value + "')"
	case TokenLookaheadPos:
		return "LOOKAHEAD_POS"
	case TokenLookaheadNeg:
		return "LOOKAHEAD_NEG"
	case TokenLookbehindPos:
		return "LOOKBEHIND_POS"
	case TokenLookbehindNeg:
		return "LOOKBEHIND_NEG"
	case TokenNonCapturing:
		return "NON_CAPTURING"
	default:
		return "UNKNOWN"
	}
}
