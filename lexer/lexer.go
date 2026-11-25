package lexer

import (
	"bufio"
	"bytes"
	"io"

	"github.com/xiam/bytescan"
	"github.com/xiam/bytescan/cursor"
)

// tokenTypes maps the bytescan token types of AddRules to lexer token types.
var tokenTypes = func() map[bytescan.TokenType]TokenType {
	m := map[bytescan.TokenType]TokenType{}
	for tt := range tokenNames {
		m[bytescan.TokenType(tt.String())] = tt
	}
	return m
}()

// New initializes a Lexer that reads from r. The Lexer runs the rule set of
// Rules on a bytescan.Scanner.
//
// One UTF-8 byte order mark at the start of the input is skipped, as
// text/scanner did before the port: it produces no token, and positions are
// counted from the byte after it. A byte order mark anywhere else is input.
func New(r io.Reader) *Lexer {
	cur := cursor.NewReader(skipBOM(r))
	sc := bytescan.NewWithCursor(cur)
	if err := AddRules(sc); err != nil {
		// The rule set has one rule for each type and no nil rule.
		panic(err)
	}
	return &Lexer{cur: cur, sc: sc}
}

// bom is the UTF-8 encoding of U+FEFF, the byte order mark.
var bom = []byte{0xEF, 0xBB, 0xBF}

// skipBOM returns a reader of r without one byte order mark at its start.
func skipBOM(r io.Reader) io.Reader {
	br := bufio.NewReader(r)
	if head, err := br.Peek(len(bom)); err == nil && bytes.Equal(head, bom) {
		_, _ = br.Discard(len(bom))
	}
	return br
}

// Lexer represents a lexical analyzer. It reads one token on each call to
// Next, in the goroutine of the caller.
type Lexer struct {
	cur *cursor.Reader
	sc  *bytescan.Scanner

	lastTok *Token
	lastErr error
	closed  bool
}

// Next reads the next token, which Token then returns. At the end of input it
// produces one TokenEOF token, which consumes no input. After TokenEOF, or
// after an error, Next returns false.
func (lx *Lexer) Next() bool {
	if lx.closed {
		return false
	}

	tok, err := lx.sc.Next()
	switch {
	case err == io.EOF:
		end := lx.cur.Cursor()
		lx.lastTok = NewToken(TokenEOF, "", &end)
		lx.closed = true
		return true
	case err != nil:
		lx.lastErr = err
		lx.closed = true
		return false
	}

	tt, ok := tokenTypes[tok.Type()]
	if !ok {
		// Every rune starts a token of the rule set, so the scanner returns
		// no bytescan.TokenTypeUnknown token.
		tt = TokenInvalid
	}
	lx.lastTok = NewTokenWithSpan(tt, tok.Text(), tok.Span())
	return true
}

// Token returns the most recent scanned token
func (lx *Lexer) Token() *Token {
	return lx.lastTok
}

// Err returns the error that stopped the Lexer, or nil. The end of input is
// not an error.
func (lx *Lexer) Err() error {
	return lx.lastErr
}

// Tokenize takes an array of bytes and returns all the tokens within it,
// or an error if a token can't be identified.
func Tokenize(in []byte) ([]Token, error) {
	tokens := []Token{}

	lx := New(bytes.NewReader(in))
	for lx.Next() {
		tokens = append(tokens, *lx.Token())
	}
	if err := lx.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}
