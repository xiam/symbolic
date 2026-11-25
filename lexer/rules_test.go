package lexer

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedText struct {
	Type TokenType
	Text string
}

func tokenize(t *testing.T, in string) []typedText {
	t.Helper()
	tokens, err := Tokenize([]byte(in))
	require.NoError(t, err)
	out := make([]typedText, 0, len(tokens))
	for _, tok := range tokens {
		out = append(out, typedText{tok.Type(), tok.Text()})
	}
	return out
}

// TestBoundaries pins token boundaries that the rules keep from the lexer
// before the port onto bytescan.
func TestBoundaries(t *testing.T) {
	testCases := []struct {
		In  string
		Out []typedText
	}{
		{"+1abc", []typedText{{TokenInteger, "+1"}, {TokenWord, "abc"}, {TokenEOF, ""}}},
		{"a+b", []typedText{{TokenWord, "a"}, {TokenSequence, "+b"}, {TokenEOF, ""}}},
		{"-.5", []typedText{{TokenSequence, "-.5"}, {TokenEOF, ""}}},
		{"- 1", []typedText{{TokenSequence, "-"}, {TokenWhitespace, " "}, {TokenInteger, "1"}, {TokenEOF, ""}}},
		{"-", []typedText{{TokenSequence, "-"}, {TokenEOF, ""}}},
		{"--1", []typedText{{TokenSequence, "--1"}, {TokenEOF, ""}}},
		{"a1_b", []typedText{{TokenWord, "a"}, {TokenInteger, "1"}, {TokenWord, "_b"}, {TokenEOF, ""}}},
		{"x/y#z", []typedText{{TokenWord, "x"}, {TokenSequence, "/y#z"}, {TokenEOF, ""}}},
		{"😊(", []typedText{{TokenSequence, "😊"}, {TokenOpenExpression, "("}, {TokenEOF, ""}}},
		{":a.b\\", []typedText{{TokenColon, ":"}, {TokenWord, "a"}, {TokenDot, "."}, {TokenWord, "b"}, {TokenBackslash, "\\"}, {TokenEOF, ""}}},
		{" \t\f\r\n", []typedText{{TokenWhitespace, " \t\f\r"}, {TokenNewLine, "\n"}, {TokenEOF, ""}}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Out, tokenize(t, tc.In), "input %q", tc.In)
	}
}

// TestSignedNumbers pins owner decision 5 of #1: a sign and all its digits are
// one integer. Before the port, "-123" was the integer "-1" then "23".
func TestSignedNumbers(t *testing.T) {
	testCases := []struct {
		In  string
		Out []typedText
	}{
		{"-123", []typedText{{TokenInteger, "-123"}, {TokenEOF, ""}}},
		{"+45.5", []typedText{{TokenInteger, "+45"}, {TokenDot, "."}, {TokenInteger, "5"}, {TokenEOF, ""}}},
		{"-1", []typedText{{TokenInteger, "-1"}, {TokenEOF, ""}}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Out, tokenize(t, tc.In), "input %q", tc.In)
	}
}

// TestNoStderr checks that input text/scanner used to report on stderr (an
// invalid UTF-8 byte and a NUL) lexes without output and without error.
func TestNoStderr(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	stderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = stderr }()

	tokens, lexErr := Tokenize([]byte("a\xffb\x00c"))

	os.Stderr = stderr
	require.NoError(t, w.Close())
	written, err := io.ReadAll(r)
	require.NoError(t, err)

	assert.NoError(t, lexErr)
	assert.Empty(t, string(written))
	assert.Equal(t, TokenEOF, tokens[len(tokens)-1].Type())
}

// TestByteOrderMark checks that one byte order mark at the start of the input
// produces no token, as with text/scanner before the port, and that a byte
// order mark anywhere else is input.
func TestByteOrderMark(t *testing.T) {
	testCases := []struct {
		In  string
		Out []typedText
	}{
		{"\ufeff", []typedText{{TokenEOF, ""}}},
		{"\ufeffabc", []typedText{{TokenWord, "abc"}, {TokenEOF, ""}}},
		{"\ufeff(a)", []typedText{{TokenOpenExpression, "("}, {TokenWord, "a"}, {TokenCloseExpression, ")"}, {TokenEOF, ""}}},
		{"a \ufeffb", []typedText{{TokenWord, "a"}, {TokenWhitespace, " "}, {TokenSequence, "\ufeffb"}, {TokenEOF, ""}}},
		{"\ufeff\ufeff", []typedText{{TokenSequence, "\ufeff"}, {TokenEOF, ""}}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.Out, tokenize(t, tc.In), "input %q", tc.In)
	}

	// Positions are counted from the byte after the skipped byte order mark.
	tokens, err := Tokenize([]byte("\ufeffab\ncd"))
	require.NoError(t, err)
	assert.Equal(t, 1, tokens[0].Pos().Line())
	assert.Equal(t, 0, tokens[0].Pos().Column())
	assert.Equal(t, 0, tokens[0].Pos().ByteOffset())
	eof := tokens[len(tokens)-1]
	assert.Equal(t, [2]int{2, 2}, [2]int{eof.Pos().Line(), eof.Pos().Column()})
}
