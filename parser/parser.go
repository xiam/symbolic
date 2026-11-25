package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xiam/bytescan/cursor"
	"github.com/xiam/symbolic/ast"
	"github.com/xiam/symbolic/lexer"
)

// EOF represents the end of the file the parser is reading
var EOF = lexer.NewToken(lexer.TokenEOF, "", nil)

type parserState func(p *Parser) parserState

// Parser represents a parser
type Parser struct {
	lx   *lexer.Lexer
	root *ast.File

	lastTok *lexer.Token
	nextTok *lexer.Token

	options ParserOptions

	lastErr error
}

// New creates a new parser that reads from the given input
func NewParser(r io.Reader) *Parser {
	return &Parser{
		root:    &ast.File{},
		lx:      lexer.New(r),
		options: parserDefaultOptions,
	}
}

func (p *Parser) SetOptions(options ParserOptions) {
	p.options = options
}

func (p *Parser) Options() ParserOptions {
	return p.options
}

// RootNode returns the root of the tree that Parse builds.
func (p *Parser) RootNode() *ast.File {
	return p.root
}

// Parse tokenizes the input and transforms it into a AST. It reads the
// tokens one by one from the lexer, in the goroutine of the caller.
func (p *Parser) Parse() error {
	start := p.peek().Pos()

	for state := parserDefaultState(p); state != nil; {
		state = state(p)
	}

	if err := p.lx.Err(); err != nil {
		// The lexer stopped after its last token, if it read one.
		var pos cursor.Pos
		if tok := p.lx.Token(); tok != nil {
			pos = tok.Span().End()
		}
		return &Error{Err: fmt.Errorf("%w: %w", ErrLexer, err), Pos: pos}
	}

	// The last token is TokenEOF, at the end of input.
	p.root.Loc = cursor.NewSpan(start, p.curr().Pos())

	return p.lastErr
}

func (p *Parser) curr() *lexer.Token {
	return p.lastTok
}

func (p *Parser) read() *lexer.Token {
	if ok := p.lx.Next(); !ok {
		// After the end of input the lexer keeps its TokenEOF, which holds
		// the end-of-input position.
		if tok := p.lx.Token(); tok != nil && tok.Type() == lexer.TokenEOF {
			return tok
		}
		return EOF
	}

	return p.lx.Token()
}

func (p *Parser) peek() *lexer.Token {
	if p.nextTok != nil {
		return p.nextTok
	}

	p.nextTok = p.read()
	return p.nextTok
}

func (p *Parser) next() *lexer.Token {
	if p.nextTok != nil {
		tok := p.nextTok
		p.lastTok, p.nextTok = tok, nil
		return tok
	}

	tok := p.read()
	p.lastTok, p.nextTok = tok, nil
	return tok
}

// spanFrom returns the span of a node whose first token is start and whose
// last token is the current token: from the start of start to the end of the
// current token.
func (p *Parser) spanFrom(start *lexer.Token) cursor.Span {
	return cursor.NewSpan(start.Span().Start(), p.curr().Span().End())
}

func parserDefaultState(p *Parser) parserState {
	tok := p.next()

	switch tok.Type() {
	case lexer.TokenEOF:
		return nil

	default:
		if state := parserStateData(&p.root.Items)(p); state != nil {
			return state
		}
	}

	return parserDefaultState
}

// parserErrorState reports err at the current token.
func parserErrorState(err error) parserState {
	return func(p *Parser) parserState {
		tok := p.curr()
		text := tok.Text()
		p.lastErr = &Error{Err: err, Pos: tok.Pos(), Token: text}
		return nil
	}
}

// parserUnclosedState reports input that ends before the bracket or the
// double quote opener is closed. The error names the end of input and the
// position of the opener.
func parserUnclosedState(opener *lexer.Token) parserState {
	return func(p *Parser) parserState {
		p.lastErr = &Error{
			Err:       ErrUnexpectedEOF,
			Pos:       p.curr().Pos(),
			Opener:    opener.Text(),
			OpenerPos: opener.Pos(),
		}
		return nil
	}
}

func expectTokens(p *Parser, tt ...lexer.TokenType) ([]*lexer.Token, error) {
	tokens := []*lexer.Token{}
	for i := range tt {
		tok := p.next()
		if tok.Type() == lexer.TokenEOF {
			return nil, ErrUnexpectedEOF
		}
		if tok.Type() != tt[i] {
			return nil, ErrUnexpectedToken
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

// parserStateData reads one item that starts at the current token and
// appends its node to items, the items of the enclosing container.
func parserStateData(items *[]ast.Node) parserState {
	return func(p *Parser) parserState {
		tok := p.curr()

		switch tok.Type() {
		case lexer.TokenWhitespace, lexer.TokenNewLine:
			// continue

		case lexer.TokenDoubleQuote:
			if state := parserStateString(items)(p); state != nil {
				return state
			}

		case lexer.TokenInteger:
			if state := parserStateNumeric(items)(p); state != nil {
				return state
			}

		case lexer.TokenColon:
			if state := parserStateAtom(items)(p); state != nil {
				return state
			}

		case lexer.TokenWord, lexer.TokenSequence:
			if state := parserStateWord(items)(p); state != nil {
				return state
			}

		case lexer.TokenHash:
			if state := parserStateComment(p); state != nil {
				return state
			}

		case lexer.TokenOpenMap:
			node := &ast.Map{}
			*items = append(*items, node)
			if state := parserStateOpen(tok, lexer.TokenCloseMap, &node.Items, &node.Loc)(p); state != nil {
				return state
			}

		case lexer.TokenOpenList:
			node := &ast.List{}
			*items = append(*items, node)
			if state := parserStateOpen(tok, lexer.TokenCloseList, &node.Items, &node.Loc)(p); state != nil {
				return state
			}

		case lexer.TokenOpenExpression:
			node := &ast.Expr{}
			*items = append(*items, node)
			if state := parserStateOpen(tok, lexer.TokenCloseExpression, &node.Items, &node.Loc)(p); state != nil {
				return state
			}

		default:
			return parserErrorState(ErrUnexpectedToken)
		}

		return nil
	}
}

// expectNumberNode reads an integer, or a floating point number when a dot
// follows the integer token.
func expectNumberNode(p *Parser) (ast.Node, error) {
	curr := p.curr()

	next := p.peek()
	switch next.Type() {
	case lexer.TokenDot:
		// got a point, this means this is a floating point number

		mantissa, err := expectTokens(p, lexer.TokenDot, lexer.TokenInteger)
		if err != nil {
			return nil, err
		}

		f64, err := strconv.ParseFloat(curr.Text()+mantissa[0].Text()+mantissa[1].Text(), 64)
		if err != nil {
			return nil, err
		}

		return &ast.Float{Value: f64, Loc: p.spanFrom(curr)}, nil

	default:
		// natural end for an integer
		i64, err := strconv.ParseInt(curr.Text(), 10, 64)
		if err != nil {
			return nil, err
		}

		return &ast.Int{Value: i64, Loc: p.spanFrom(curr)}, nil
	}
}

// parserStateComment skips a comment, up to and including the end of its
// line.
func parserStateComment(p *Parser) parserState {
	for {
		tok := p.next()
		switch tok.Type() {
		case lexer.TokenEOF, lexer.TokenNewLine:
			return nil
		}
	}
}

func parserStateString(items *[]ast.Node) parserState {
	return func(p *Parser) parserState {
		opener := p.curr()
		var text strings.Builder

		for {
			tok := p.next()

			switch tok.Type() {
			case lexer.TokenDoubleQuote:
				// The span runs from the opening to the closing double quote.
				*items = append(*items, &ast.String{Value: text.String(), Loc: p.spanFrom(opener)})
				return nil

			case lexer.TokenEOF:
				return parserUnclosedState(opener)

			default:
				text.WriteString(tok.Text())
			}
		}
	}
}

func parserStateNumeric(items *[]ast.Node) parserState {
	return func(p *Parser) parserState {
		node, err := expectNumberNode(p)
		if err != nil {
			return parserErrorState(err)
		}
		*items = append(*items, node)
		return nil
	}
}

func parserStateWord(items *[]ast.Node) parserState {
	return func(p *Parser) parserState {
		curr := p.curr()
		*items = append(*items, &ast.Symbol{Name: curr.Text(), Loc: p.spanFrom(curr)})
		return nil
	}
}

func parserStateAtom(items *[]ast.Node) parserState {
	return func(p *Parser) parserState {
		colon := p.curr()

		name, err := expectTokens(p, lexer.TokenWord)
		if err != nil {
			return parserErrorState(err)
		}

		*items = append(*items, &ast.Atom{Name: colon.Text() + name[0].Text(), Loc: p.spanFrom(colon)})
		return nil
	}
}

// parserStateOpen reads the items of a container, from the token after the
// opener up to the closer, into items. It sets loc to the span from the
// opener to the closer. With AutoCloseOnEOF, the end of input closes the
// container, and loc ends at the end of input.
func parserStateOpen(opener *lexer.Token, closer lexer.TokenType, items *[]ast.Node, loc *cursor.Span) parserState {
	return func(p *Parser) parserState {
		for {
			tok := p.next()

			switch tok.Type() {
			case lexer.TokenEOF:
				if p.options.AutoCloseOnEOF {
					*loc = cursor.NewSpan(opener.Pos(), tok.Pos())
					return nil
				}
				return parserUnclosedState(opener)

			case closer:
				*loc = p.spanFrom(opener)
				return nil

			default:
				if state := parserStateData(items)(p); state != nil {
					return state
				}
			}
		}
	}
}

// Parse parses an array of bytes and returns a AST root
func Parse(in []byte) (*ast.File, error) {
	p := NewParser(bytes.NewReader(in))

	err := p.Parse()
	if err != nil {
		return nil, err
	}

	return p.root, nil
}
