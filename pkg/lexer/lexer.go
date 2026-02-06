package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes a regex pattern string.
type Lexer struct {
	pattern string
	pos     int
	bytePos int
	length  int
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{
		pattern: input,
		pos:     0,
		bytePos: 0,
		length:  utf8.RuneCountInString(input),
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() (Token, error) {
	if l.pos >= l.length {
		return Token{Type: TokenEOF, Pos: l.pos}, nil
	}

	pos := l.pos
	ch := l.currentChar()

	switch ch {
	case '\x00':
		return Token{Type: TokenEOF, Pos: pos}, nil
	case '\\':
		return l.handleEscape(pos)
	case '*':
		l.advance()
		return Token{Type: TokenStar, Value: "*", Pos: pos}, nil
	case '+':
		l.advance()
		return Token{Type: TokenPlus, Value: "+", Pos: pos}, nil
	case '?':
		l.advance()
		return Token{Type: TokenQuestion, Value: "?", Pos: pos}, nil
	case '{':
		l.advance()
		return Token{Type: TokenLBrace, Value: "{", Pos: pos}, nil
	case '}':
		l.advance()
		return Token{Type: TokenRBrace, Value: "}", Pos: pos}, nil
	case ',':
		l.advance()
		return Token{Type: TokenComma, Value: ",", Pos: pos}, nil
	case '|':
		l.advance()
		return Token{Type: TokenPipe, Value: "|", Pos: pos}, nil
	case '(':
		return l.handleGroupStart(pos)
	case ')':
		l.advance()
		return Token{Type: TokenRParen, Value: ")", Pos: pos}, nil
	case '[':
		l.advance()
		return Token{Type: TokenLBracket, Value: "[", Pos: pos}, nil
	case ']':
		l.advance()
		return Token{Type: TokenRBracket, Value: "]", Pos: pos}, nil
	case '^':
		l.advance()
		return Token{Type: TokenCaret, Value: "^", Pos: pos}, nil
	case '$':
		l.advance()
		return Token{Type: TokenDollar, Value: "$", Pos: pos}, nil
	case '.':
		l.advance()
		return Token{Type: TokenDot, Value: ".", Pos: pos}, nil
	case '-':
		l.advance()
		return Token{Type: TokenMinus, Value: "-", Pos: pos}, nil
	default:
		l.advance()
		return Token{Type: TokenChar, Value: string(ch), Pos: pos}, nil
	}
}

// handleEscape processes escape sequences after a backslash.
func (l *Lexer) handleEscape(pos int) (Token, error) {
	l.advance()
	next := l.currentChar()
	if next == '\x00' {
		return Token{}, fmt.Errorf("pattern cannot end with backslash at position %d", pos)
	}

	switch next {
	case 'd':
		l.advance()
		return Token{Type: TokenDigit, Value: `\d`, Pos: pos}, nil
	case 'D':
		l.advance()
		return Token{Type: TokenNonDigit, Value: `\D`, Pos: pos}, nil
	case 'w':
		l.advance()
		return Token{Type: TokenWord, Value: `\w`, Pos: pos}, nil
	case 'W':
		l.advance()
		return Token{Type: TokenNonWord, Value: `\W`, Pos: pos}, nil
	case 's':
		l.advance()
		return Token{Type: TokenWhitespace, Value: `\s`, Pos: pos}, nil
	case 'S':
		l.advance()
		return Token{Type: TokenNonWhitespace, Value: `\S`, Pos: pos}, nil
	case 'b':
		l.advance()
		return Token{Type: TokenWordBoundary, Value: `\b`, Pos: pos}, nil
	case 'B':
		l.advance()
		return Token{Type: TokenNonWordBoundary, Value: `\B`, Pos: pos}, nil
	case 'n':
		l.advance()
		return Token{Type: TokenChar, Value: "\n", Pos: pos}, nil
	case 't':
		l.advance()
		return Token{Type: TokenChar, Value: "\t", Pos: pos}, nil
	case 'r':
		l.advance()
		return Token{Type: TokenChar, Value: "\r", Pos: pos}, nil
	default:
		if unicode.IsDigit(next) {
			var digits []rune
			for {
				cur := l.currentChar()
				if cur == '\x00' || !unicode.IsDigit(cur) {
					break
				}
				digits = append(digits, cur)
				l.advance()
			}
			return Token{Type: TokenBackref, Value: string(digits), Pos: pos}, nil
		}
		l.advance()
		return Token{Type: TokenChar, Value: string(next), Pos: pos}, nil
	}
}

func (l *Lexer) handleGroupStart(pos int) (Token, error) {
	l.advance()
	if l.currentChar() != '?' {
		return Token{Type: TokenLParen, Value: "(", Pos: pos}, nil
	}

	l.advance()
	next := l.currentChar()
	switch next {
	case ':':
		l.advance()
		return Token{Type: TokenNonCapturing, Value: "(?:", Pos: pos}, nil
	case '=':
		l.advance()
		return Token{Type: TokenLookaheadPos, Value: "(?=", Pos: pos}, nil
	case '!':
		l.advance()
		return Token{Type: TokenLookaheadNeg, Value: "(?!", Pos: pos}, nil
	case '<':
		l.advance()
		look := l.currentChar()
		switch look {
		case '=':
			l.advance()
			return Token{Type: TokenLookbehindPos, Value: "(?<=", Pos: pos}, nil
		case '!':
			l.advance()
			return Token{Type: TokenLookbehindNeg, Value: "(?<!", Pos: pos}, nil
		default:
			return Token{}, fmt.Errorf("invalid group syntax at position %d", pos)
		}
	case '\x00':
		return Token{}, fmt.Errorf("unknown group modifier '?' at position %d", pos)
	default:
		return Token{}, fmt.Errorf("unknown group modifier '?%c' at position %d", next, pos)
	}
}

func (l *Lexer) currentChar() rune {
	if l.bytePos >= len(l.pattern) {
		return '\x00'
	}
	ch, _ := utf8.DecodeRuneInString(l.pattern[l.bytePos:])
	return ch
}

func (l *Lexer) advance() rune {
	if l.bytePos >= len(l.pattern) {
		return '\x00'
	}
	ch, size := utf8.DecodeRuneInString(l.pattern[l.bytePos:])
	if ch == utf8.RuneError && size == 1 {
		return '\x00'
	}
	l.bytePos += size
	l.pos++
	return ch
}

// Peek returns the current rune without advancing.
func (l *Lexer) Peek() rune {
	return l.currentChar()
}

// Position returns the current position in the input.
func (l *Lexer) Position() int {
	return l.pos
}

func (l *Lexer) Tokenize() ([]Token, error) {
	tokens := []Token{}
	for {
		token, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
		if token.Type == TokenEOF {
			break
		}
	}
	return tokens, nil
}
