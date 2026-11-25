package calc

import (
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/xiam/bytescan"
	"github.com/xiam/bytescan/cursor"
)

// Parse parses src as one expression and returns its tree. On a syntax error
// it returns no tree and an *Error with the position.
func Parse(src string) (Node, error) {
	sc, cur := newScanner(strings.NewReader(src))
	p := &parser{sc: sc, cur: cur}
	if err := p.advance(); err != nil {
		return nil, err
	}

	x, err := p.expr()
	if err != nil {
		return nil, err
	}
	if p.tok != nil {
		return nil, p.unexpected()
	}
	return x, nil
}

// parser is a recursive descent parser. It pulls the tokens from the
// scanner one at a time and looks one token ahead.
type parser struct {
	sc  *bytescan.Scanner
	cur *cursor.Reader

	// tok is the next token that is not whitespace, or nil at the end of
	// input. eof is the position of the end of input, once tok is nil.
	tok *bytescan.Token
	eof cursor.Pos
}

// advance reads the next token that is not whitespace or a line break into
// p.tok.
func (p *parser) advance() error {
	for {
		tok, err := p.sc.Next()
		switch {
		case err == io.EOF:
			p.tok, p.eof = nil, p.cur.Cursor()
			return nil
		case err != nil:
			return &Error{Err: err, Pos: p.cur.Cursor()}
		}
		if t := tok.Type(); t != tokenSpace && t != tokenNewline {
			p.tok = tok
			return nil
		}
	}
}

// operators maps the token types of the operators to their Op.
var operators = map[bytescan.TokenType]Op{
	tokenPlus:  Add,
	tokenMinus: Sub,
	tokenStar:  Mul,
	tokenSlash: Div,
}

// op reports whether the next token is one of the operators ops, and which.
func (p *parser) op(ops ...Op) (Op, bool) {
	if p.tok == nil {
		return 0, false
	}
	op, ok := operators[p.tok.Type()]
	return op, ok && slices.Contains(ops, op)
}

// expr parses: term { ("+" | "-") term }.
func (p *parser) expr() (Node, error) {
	return p.binary(p.term, Add, Sub)
}

// term parses: unary { ("*" | "/") unary }.
func (p *parser) term() (Node, error) {
	return p.binary(p.unary, Mul, Div)
}

// binary parses operand { op operand } for the operators ops of one level of
// precedence. The loop makes the operators left associative: "1 - 2 - 3" is
// (1 - 2) - 3.
func (p *parser) binary(operand func() (Node, error), ops ...Op) (Node, error) {
	x, err := operand()
	if err != nil {
		return nil, err
	}
	for {
		op, ok := p.op(ops...)
		if !ok {
			return x, nil
		}
		opPos := p.tok.Span().Start()
		if err := p.advance(); err != nil {
			return nil, err
		}
		y, err := operand()
		if err != nil {
			return nil, err
		}
		span := cursor.NewSpan(x.Span().Start(), y.Span().End())
		x = &Binary{Op: op, OpPos: opPos, X: x, Y: y, Loc: span}
	}
}

// unary parses: ("+" | "-") unary | primary.
func (p *parser) unary() (Node, error) {
	op, ok := p.op(Add, Sub)
	if !ok {
		return p.primary()
	}
	start := p.tok.Span().Start()
	if err := p.advance(); err != nil {
		return nil, err
	}
	x, err := p.unary()
	if err != nil {
		return nil, err
	}
	return &Unary{Op: op, X: x, Loc: cursor.NewSpan(start, x.Span().End())}, nil
}

// primary parses: number | "(" expr ")".
func (p *parser) primary() (Node, error) {
	tok := p.tok
	if tok == nil {
		return nil, &Error{Err: ErrUnexpectedEOF, Pos: p.eof}
	}

	switch tok.Type() {
	case tokenNumber:
		text := tok.Text()
		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			// The rule set matches only digits and a point, so the only
			// error is a number out of range.
			return nil, &Error{Err: ErrRange, Pos: tok.Span().Start(), Token: text}
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Number{Value: v, Loc: tok.Span()}, nil

	case tokenLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		x, err := p.expr()
		if err != nil {
			return nil, err
		}
		if p.tok == nil {
			return nil, &Error{Err: ErrUnclosedParen, Pos: tok.Span().Start()}
		}
		if p.tok.Type() != tokenRParen {
			return nil, p.unexpected()
		}
		span := cursor.NewSpan(tok.Span().Start(), p.tok.Span().End())
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Paren{X: x, Loc: span}, nil
	}

	return nil, p.unexpected()
}

// unexpected returns the error for the next token, which the grammar does not
// allow where it is.
func (p *parser) unexpected() error {
	text := p.tok.Text()
	return &Error{Err: ErrUnexpectedToken, Pos: p.tok.Span().Start(), Token: text}
}
