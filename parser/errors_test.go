package parser

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The three error tests of output 5 of #1. Each error names a position in
// bytescan's coordinates: a 1-based line and a 0-based rune column.

func TestErrorUnclosedBracket(t *testing.T) {
	root, err := Parse([]byte("(a\n  [b"))
	assert.Nil(t, root)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedEOF)
	// The end of input, and the bracket that is not closed.
	assert.Contains(t, err.Error(), "(around (line: 2) (column 4))")
	assert.Contains(t, err.Error(), `"[" at (line: 2) (column 2) is not closed`)
}

func TestErrorUnclosedBracketAfterComment(t *testing.T) {
	// The comment reads up to the end of input; the error still names the
	// end-of-input position, not a zero position.
	_, err := Parse([]byte("(1 # c"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "(around (line: 1) (column 6))")
	assert.Contains(t, err.Error(), `"(" at (line: 1) (column 0) is not closed`)
}

func TestErrorUnexpectedClosingBracket(t *testing.T) {
	root, err := Parse([]byte("(a)\n  }"))
	assert.Nil(t, root)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedToken)
	assert.Contains(t, err.Error(), `"}" (around (line 2) (column 2))`)
}

func TestErrorUnterminatedString(t *testing.T) {
	root, err := Parse([]byte("(a\n \"bc d"))
	assert.Nil(t, root)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedEOF)
	assert.Contains(t, err.Error(), "(around (line: 2) (column 6))")
	assert.Contains(t, err.Error(), `"\"" at (line: 2) (column 1) is not closed`)
}

// TestAutoCloseOnEOFKeepsStringError checks that AutoCloseOnEOF closes
// brackets only: an unterminated string is still an error.
func TestAutoCloseOnEOFKeepsStringError(t *testing.T) {
	p := NewParser(strings.NewReader(`(a "bc`))
	p.SetOptions(ParserOptions{AutoCloseOnEOF: true})
	err := p.Parse()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedEOF)
}

// TestErrorType checks the fields of *Error (decision 3 of #2).
func TestErrorType(t *testing.T) {
	testCases := []struct {
		In   string
		Want *Error
		// Positions as {line, column}.
		Pos, OpenerPos [2]int
	}{
		{
			In:        "(a\n  [b",
			Want:      &Error{Err: ErrUnexpectedEOF, Opener: "["},
			Pos:       [2]int{2, 4},
			OpenerPos: [2]int{2, 2},
		},
		{
			In:   "(a)\n  }",
			Want: &Error{Err: ErrUnexpectedToken, Token: "}"},
			Pos:  [2]int{2, 2},
		},
		{
			In:        "(a\n \"bc d",
			Want:      &Error{Err: ErrUnexpectedEOF, Opener: `"`},
			Pos:       [2]int{2, 6},
			OpenerPos: [2]int{2, 1},
		},
	}

	for _, tc := range testCases {
		_, err := Parse([]byte(tc.In))
		var perr *Error
		require.ErrorAs(t, err, &perr, "input %q", tc.In)
		assert.Equal(t, tc.Want.Err, perr.Err, "input %q", tc.In)
		assert.ErrorIs(t, err, tc.Want.Err, "input %q", tc.In)
		assert.Equal(t, tc.Want.Token, perr.Token, "input %q", tc.In)
		assert.Equal(t, tc.Want.Opener, perr.Opener, "input %q", tc.In)
		assert.Equal(t, tc.Pos, [2]int{perr.Pos.Line(), perr.Pos.Column()}, "input %q", tc.In)
		if tc.Want.Opener != "" {
			assert.Equal(t, tc.OpenerPos, [2]int{perr.OpenerPos.Line(), perr.OpenerPos.Column()}, "input %q", tc.In)
		}
	}

	// A number that does not fit in 64 bits unwraps to its strconv error.
	_, err := Parse([]byte("99999999999999999999999"))
	var numErr *strconv.NumError
	assert.ErrorAs(t, err, &numErr)
	assert.Contains(t, err.Error(), "(around (line: 1) (column 0))")
}
