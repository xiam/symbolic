package parser

import (
	"errors"
	"fmt"

	"github.com/xiam/bytescan/cursor"
)

var (
	ErrUnexpectedEOF   = errors.New("unexpected EOF")
	ErrUnexpectedToken = errors.New("unexpected token")

	// ErrLexer is the cause of an error that stops the lexer, such as an
	// error of the reader. The Err of such an Error wraps both ErrLexer
	// and the error of the lexer.
	ErrLexer = errors.New("lexer error")
)

// Error is an error that stops the parser: a syntax error, or an error of
// the lexer. Use errors.Is to test its cause, such as ErrUnexpectedEOF,
// ErrUnexpectedToken or ErrLexer, and errors.As to read its positions.
type Error struct {
	// Err is the cause: ErrUnexpectedEOF, ErrUnexpectedToken, the error of
	// a number that does not fit in 64 bits, or an error that wraps
	// ErrLexer and the error of the lexer.
	Err error

	// Pos is where the parser found the error: the unexpected token, or the
	// end of input. For a lexer error, Pos is the end of the last token that
	// the lexer read, or the zero position when it read no token.
	Pos cursor.Pos

	// Token is the text of the unexpected token, with ErrUnexpectedToken.
	Token string

	// Opener is the bracket or double quote that the input does not close,
	// and OpenerPos is its position. Opener is empty when the error has no
	// opener.
	Opener    string
	OpenerPos cursor.Pos
}

func (e *Error) Error() string {
	switch {
	case e.Opener != "":
		return fmt.Sprintf("syntax error: %v (around (line: %v) (column %v)): %q at (line: %v) (column %v) is not closed",
			e.Err, e.Pos.Line(), e.Pos.Column(), e.Opener, e.OpenerPos.Line(), e.OpenerPos.Column())
	case errors.Is(e.Err, ErrLexer):
		return fmt.Sprintf("%v (around (line: %v) (column %v))", e.Err, e.Pos.Line(), e.Pos.Column())
	case e.Err == ErrUnexpectedToken:
		return fmt.Sprintf("syntax error: %v %q (around (line %v) (column %v))", e.Err, e.Token, e.Pos.Line(), e.Pos.Column())
	}
	// TODO: extract a code snippet around line and col
	return fmt.Sprintf("syntax error: %v (around (line: %v) (column %v))", e.Err, e.Pos.Line(), e.Pos.Column())
}

// Unwrap returns the cause of the error.
func (e *Error) Unwrap() error {
	return e.Err
}
