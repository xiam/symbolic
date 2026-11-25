package lexer

import (
	"fmt"

	"github.com/xiam/bytescan/cursor"
)

// Token represents a known sequence of characters (lexical unit)
type Token struct {
	tt     TokenType
	lexeme string

	span cursor.Span
}

// NewToken creates a lexical unit that starts at pos and has an empty span.
// A nil pos gives the zero position. Use NewTokenWithSpan for a token that
// covers source text.
func NewToken(tt TokenType, lexeme string, pos *cursor.Pos) *Token {
	if pos == nil {
		pos = &cursor.Pos{}
	}
	return NewTokenWithSpan(tt, lexeme, cursor.NewSpan(*pos, *pos))
}

// NewTokenWithSpan creates a lexical unit that covers span.
func NewTokenWithSpan(tt TokenType, lexeme string, span cursor.Span) *Token {
	return &Token{
		tt:     tt,
		lexeme: lexeme,
		span:   span,
	}
}

// Type returns the type of the lexical unit
func (t Token) Type() TokenType {
	return t.tt
}

// Pos returns the position where the lexical unit starts, in bytescan's
// coordinates: a 1-based line and a 0-based rune column. A newline token is
// on the line it ends. A TokenEOF token is at the end of input.
func (t Token) Pos() cursor.Pos {
	return t.span.Start()
}

// Span returns the source text that the lexical unit covers: from its first
// rune to after its last rune. A TokenEOF token has the empty span at the end
// of input.
func (t Token) Span() cursor.Span {
	return t.span
}

// Text returns the raw text of the lexical unit
func (t Token) Text() string {
	return t.lexeme
}

func (t Token) String() string {
	return fmt.Sprintf("(:%v %q [%d %d])", t.tt, t.lexeme, t.Pos().Line(), t.Pos().Column())
}
