package calc

import (
	"errors"
	"fmt"

	"github.com/xiam/bytescan/cursor"
)

// The causes of an Error. Use errors.Is to test them.
var (
	// ErrUnexpectedToken is a token where the grammar does not allow it,
	// such as the second number of "1 2", or a rune that no rule of the lexer
	// matches.
	ErrUnexpectedToken = errors.New("unexpected token")

	// ErrUnexpectedEOF is the end of input where an operand is expected, such
	// as after "1 +".
	ErrUnexpectedEOF = errors.New("unexpected end of input")

	// ErrUnclosedParen is the end of input where a closing parenthesis is
	// expected. The position is that of the opening parenthesis.
	ErrUnclosedParen = errors.New("unclosed parenthesis")

	// ErrRange is a number too large for a float64.
	ErrRange = errors.New("number out of range")

	// ErrDivisionByZero is a division by zero in Eval. The position is that
	// of the operator.
	ErrDivisionByZero = errors.New("division by zero")
)

// Error is an error of Parse or Eval at a position of the source text.
type Error struct {
	// Err is the cause, such as ErrUnexpectedToken, or an error of the
	// scanner.
	Err error

	// Pos is the position of the error.
	Pos cursor.Pos

	// Token is the text of the token at Pos, with ErrUnexpectedToken and
	// ErrRange.
	Token string
}

// Error returns the position, the cause and the token, such as
// `1:2: unexpected token "2"`.
func (e *Error) Error() string {
	if e.Token != "" {
		return fmt.Sprintf("%v: %v %q", e.Pos, e.Err, e.Token)
	}
	return fmt.Sprintf("%v: %v", e.Pos, e.Err)
}

// Unwrap returns the cause of the error.
func (e *Error) Unwrap() error {
	return e.Err
}
