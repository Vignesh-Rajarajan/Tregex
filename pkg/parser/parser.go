package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"tregex/pkg/lexer"
)

type Parser struct {
	tokens     []lexer.Token
	pos        int
	groupCount int
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:     tokens,
		pos:        0,
		groupCount: 0,
	}
}

func (p *Parser) CurrentToken() lexer.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	if len(p.tokens) == 0 {
		return lexer.Token{Type: lexer.TokenEOF, Value: "", Pos: p.pos}
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) PeekToken(offset int) lexer.Token {
	pos := p.pos + offset
	if pos < len(p.tokens) {
		return p.tokens[pos]
	}
	if len(p.tokens) == 0 {
		return lexer.Token{Type: lexer.TokenEOF, Value: "", Pos: p.pos}
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) expect(tokenType lexer.TokenType) (lexer.Token, error) {
	token := p.CurrentToken()
	if token.Type != tokenType {
		return lexer.Token{}, fmt.Errorf("Expected %s, got %s at position %d", tokenType, token.Type, token.Pos)
	}
	p.advance()
	return token, nil
}

func (p *Parser) advance() {
	token := p.CurrentToken()
	if token.Type == lexer.TokenEOF {
		return
	}
	p.pos++
}

func (p *Parser) Parse() (ASTNode, error) {
	ast, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}
	if p.CurrentToken().Type != lexer.TokenEOF {
		token := p.CurrentToken()
		return nil, fmt.Errorf("Unexpected token %s at position %d", token.Type, token.Pos)
	}
	return ast, nil
}

func (p *Parser) parseAlternation() (ASTNode, error) {
	alternatives := []ASTNode{}

	first, err := p.parseConcatenation()
	if err != nil {
		return nil, err
	}
	alternatives = append(alternatives, first)

	for p.CurrentToken().Type == lexer.TokenPipe {
		p.advance()
		next, err := p.parseConcatenation()
		if err != nil {
			return nil, err
		}
		alternatives = append(alternatives, next)
	}

	if len(alternatives) == 1 {
		return alternatives[0], nil
	}

	return &AlternationNode{Alternatives: alternatives}, nil
}

func (p *Parser) parseConcatenation() (ASTNode, error) {
	children := []ASTNode{}
	for {
		token := p.CurrentToken()
		if token.Type == lexer.TokenEOF || token.Type == lexer.TokenPipe || token.Type == lexer.TokenRParen {
			break
		}

		if token.Type == lexer.TokenMinus || token.Type == lexer.TokenComma {
			p.advance()
			children = append(children, &CharNode{Char: token.Value})
			continue
		}

		node, err := p.parseQuantified()
		if err != nil {
			return nil, err
		}
		children = append(children, node)
	}

	if len(children) == 0 {
		return &ConcatNode{Children: []ASTNode{}}, nil
	}
	if len(children) == 1 {
		return children[0], nil
	}
	return &ConcatNode{Children: children}, nil
}

func (p *Parser) parseQuantified() (ASTNode, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	token := p.CurrentToken()
	switch token.Type {
	case lexer.TokenStar:
		p.advance()
		greedy := !p.checkLazyModifier()
		return &QuantifierNode{Child: atom, MinCount: 0, MaxCount: nil, Greedy: greedy}, nil
	case lexer.TokenPlus:
		p.advance()
		greedy := !p.checkLazyModifier()
		return &QuantifierNode{Child: atom, MinCount: 1, MaxCount: nil, Greedy: greedy}, nil
	case lexer.TokenQuestion:
		p.advance()
		greedy := !p.checkLazyModifier()
		max := 1
		return &QuantifierNode{Child: atom, MinCount: 0, MaxCount: &max, Greedy: greedy}, nil
	case lexer.TokenLBrace:
		return p.parseRangeQuantifier(atom)
	default:
		return atom, nil
	}
}

func (p *Parser) parseAtom() (ASTNode, error) {
	token := p.CurrentToken()
	if token.Type == lexer.TokenChar {
		p.advance()
		return &CharNode{Char: token.Value}, nil
	}
	if token.Type == lexer.TokenDot {
		p.advance()
		return &DotNode{}, nil
	}
	if token.Type == lexer.TokenCaret {
		p.advance()
		return &AnchorNode{AnchorType: "^"}, nil
	}
	if token.Type == lexer.TokenDollar {
		p.advance()
		return &AnchorNode{AnchorType: "$"}, nil
	}
	if token.Type == lexer.TokenWordBoundary {
		p.advance()
		return &AnchorNode{AnchorType: "b"}, nil
	}
	if token.Type == lexer.TokenNonWordBoundary {
		p.advance()
		return &AnchorNode{AnchorType: "B"}, nil
	}
	if token.Type == lexer.TokenDigit {
		p.advance()
		return &PredefinedClassNode{ClassType: "d"}, nil
	}
	if token.Type == lexer.TokenNonDigit {
		p.advance()
		return &PredefinedClassNode{ClassType: "D"}, nil
	}
	if token.Type == lexer.TokenWord {
		p.advance()
		return &PredefinedClassNode{ClassType: "w"}, nil
	}
	if token.Type == lexer.TokenNonWord {
		p.advance()
		return &PredefinedClassNode{ClassType: "W"}, nil
	}
	if token.Type == lexer.TokenWhitespace {
		p.advance()
		return &PredefinedClassNode{ClassType: "s"}, nil
	}
	if token.Type == lexer.TokenNonWhitespace {
		p.advance()
		return &PredefinedClassNode{ClassType: "S"}, nil
	}
	if token.Type == lexer.TokenBackref {
		p.advance()
		num, err := strconv.Atoi(token.Value)
		if err != nil {
			return nil, fmt.Errorf("Invalid backreference %q at position %d", token.Value, token.Pos)
		}
		return &BackreferenceNode{GroupNumber: num}, nil
	}
	if token.Type == lexer.TokenLBracket {
		return p.parseCharClass()
	}
	if token.Type == lexer.TokenLParen {
		return p.parseGroup()
	}
	if token.Type == lexer.TokenNonCapturing {
		return p.parseNonCapturingGroup()
	}
	if token.Type == lexer.TokenLookaheadPos {
		return p.parseLookahead(true)
	}
	if token.Type == lexer.TokenLookaheadNeg {
		return p.parseLookahead(false)
	}
	if token.Type == lexer.TokenLookbehindPos {
		return p.parseLookbehind(true)
	}
	if token.Type == lexer.TokenLookbehindNeg {
		return p.parseLookbehind(false)
	}
	return nil, fmt.Errorf("Unexpected token %s at position %d", token.Type, token.Pos)
}

func (p *Parser) checkLazyModifier() bool {
	if p.CurrentToken().Type == lexer.TokenQuestion {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) parseRangeQuantifier(atom ASTNode) (ASTNode, error) {
	if _, err := p.expect(lexer.TokenLBrace); err != nil {
		return nil, err
	}

	token := p.CurrentToken()
	if token.Type != lexer.TokenChar || !isDigit(token.Value) {
		return nil, fmt.Errorf("Expected number in quantifier at position %d", token.Pos)
	}

	minCount, err := p.readNumber()
	if err != nil {
		return nil, err
	}

	var maxCount *int
	if p.CurrentToken().Type == lexer.TokenComma {
		p.advance()
		token = p.CurrentToken()
		if token.Type == lexer.TokenRBrace {
			maxCount = nil
		} else if token.Type == lexer.TokenChar && isDigit(token.Value) {
			max, err := p.readNumber()
			if err != nil {
				return nil, err
			}
			maxCount = &max
		} else {
			return nil, fmt.Errorf("Expected number or '}' at position %d", token.Pos)
		}
	} else {
		max := minCount
		maxCount = &max
	}

	if _, err := p.expect(lexer.TokenRBrace); err != nil {
		return nil, err
	}

	greedy := !p.checkLazyModifier()
	return &QuantifierNode{Child: atom, MinCount: minCount, MaxCount: maxCount, Greedy: greedy}, nil
}

func (p *Parser) parseCharClass() (ASTNode, error) {
	if _, err := p.expect(lexer.TokenLBracket); err != nil {
		return nil, err
	}

	negated := false
	if p.CurrentToken().Type == lexer.TokenCaret {
		negated = true
		p.advance()
	}

	chars := map[string]struct{}{}

	for p.CurrentToken().Type != lexer.TokenRBracket {
		token := p.CurrentToken()
		if token.Type == lexer.TokenEOF {
			return nil, fmt.Errorf("Unclosed character class")
		}

		switch token.Type {
		case lexer.TokenChar:
			char := token.Value
			p.advance()
			if p.CurrentToken().Type == lexer.TokenMinus {
				next := p.PeekToken(1)
				if next.Type == lexer.TokenChar {
					p.advance()
					endChar := p.CurrentToken().Value
					p.advance()
					startOrd := rune([]rune(char)[0])
					endOrd := rune([]rune(endChar)[0])
					if startOrd > endOrd {
						return nil, fmt.Errorf("Invalid range %s-%s: start > end", char, endChar)
					}
					for code := startOrd; code <= endOrd; code++ {
						chars[string(code)] = struct{}{}
					}
				} else {
					chars[char] = struct{}{}
					chars["-"] = struct{}{}
					p.advance()
				}
			} else {
				chars[char] = struct{}{}
			}
		case lexer.TokenDigit:
			p.advance()
			for ch := '0'; ch <= '9'; ch++ {
				chars[string(ch)] = struct{}{}
			}
		case lexer.TokenWord:
			p.advance()
			for ch := 'a'; ch <= 'z'; ch++ {
				chars[string(ch)] = struct{}{}
			}
			for ch := 'A'; ch <= 'Z'; ch++ {
				chars[string(ch)] = struct{}{}
			}
			for ch := '0'; ch <= '9'; ch++ {
				chars[string(ch)] = struct{}{}
			}
			chars["_"] = struct{}{}
		case lexer.TokenWhitespace:
			p.advance()
			for _, ch := range []rune{' ', '\t', '\n', '\r', '\f', '\v'} {
				chars[string(ch)] = struct{}{}
			}
		case lexer.TokenMinus:
			p.advance()
			chars["-"] = struct{}{}
		case lexer.TokenPlus, lexer.TokenStar, lexer.TokenQuestion, lexer.TokenDot, lexer.TokenPipe, lexer.TokenCaret,
			lexer.TokenDollar, lexer.TokenLBrace, lexer.TokenRBrace, lexer.TokenLParen, lexer.TokenRParen, lexer.TokenComma:
			p.advance()
			chars[token.Value] = struct{}{}
		default:
			return nil, fmt.Errorf("Unexpected token %s in character class at position %d", token.Type, token.Pos)
		}
	}

	if _, err := p.expect(lexer.TokenRBracket); err != nil {
		return nil, err
	}

	if len(chars) == 0 {
		return nil, fmt.Errorf("Empty character class")
	}

	return &CharClassNode{Chars: chars, Negated: negated}, nil
}

func (p *Parser) parseGroup() (ASTNode, error) {
	if _, err := p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}
	p.groupCount++
	groupNum := p.groupCount

	child, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	return &GroupNode{Child: child, GroupNumber: groupNum}, nil
}

func (p *Parser) parseNonCapturingGroup() (ASTNode, error) {
	p.advance()

	child, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	return &NonCapturingGroupNode{Child: child}, nil
}

func (p *Parser) parseLookahead(positive bool) (ASTNode, error) {
	p.advance()

	child, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	return &LookaheadNode{Child: child, Positive: positive}, nil
}

func (p *Parser) parseLookbehind(positive bool) (ASTNode, error) {
	p.advance()

	child, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	return &LookbehindNode{Child: child, Positive: positive}, nil
}

func (p *Parser) readNumber() (int, error) {
	token := p.CurrentToken()
	if token.Type != lexer.TokenChar || !isDigit(token.Value) {
		return 0, fmt.Errorf("Expected number in quantifier at position %d", token.Pos)
	}
	var b strings.Builder
	for token.Type == lexer.TokenChar && isDigit(token.Value) {
		b.WriteString(token.Value)
		p.advance()
		token = p.CurrentToken()
	}
	value, err := strconv.Atoi(b.String())
	if err != nil {
		return 0, fmt.Errorf("Invalid number %q at position %d", b.String(), token.Pos)
	}
	return value, nil
}

func isDigit(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsDigit(r)
}
