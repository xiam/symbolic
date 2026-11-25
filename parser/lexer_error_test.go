package parser

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

// errBoom is the error of the reader in the tests of lexer errors.
var errBoom = errors.New("boom")

func TestLexerError(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		line, column int
		msg          string
	}{
		{
			// The lexer stops after "12": the whitespace that follows it is
			// not complete when the reader fails.
			name:   "after a token",
			input:  "(foo\n  12 ",
			line:   2,
			column: 4,
			msg:    "lexer error: bytescan: ReadRune: boom (around (line: 2) (column 4))",
		},
		{
			name:  "before any token",
			input: "",
			msg:   "lexer error: bytescan: ReadRune: boom (around (line: 0) (column 0))",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := io.MultiReader(strings.NewReader(tc.input), iotest.ErrReader(errBoom))
			err := NewParser(r).Parse()

			var perr *Error
			if !errors.As(err, &perr) {
				t.Fatalf("got %T %v, want a *Error", err, err)
			}
			if !errors.Is(err, ErrLexer) {
				t.Errorf("errors.Is(err, ErrLexer) is false for %v", err)
			}
			if !errors.Is(err, errBoom) {
				t.Errorf("errors.Is(err, errBoom) is false for %v", err)
			}
			if got := perr.Pos; got.Line() != tc.line || got.Column() != tc.column {
				t.Errorf("Pos is %v, want %d:%d", got, tc.line, tc.column)
			}
			if got := err.Error(); got != tc.msg {
				t.Errorf("message is %q, want %q", got, tc.msg)
			}
		})
	}
}
