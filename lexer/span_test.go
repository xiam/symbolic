package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenSpan pins the span of each token: from its first rune to after its
// last rune, with "line:column" positions. TokenEOF has the empty span at the
// end of input.
func TestTokenSpan(t *testing.T) {
	tokens, err := Tokenize([]byte("(foo 12\n  :é)"))
	require.NoError(t, err)

	type span struct {
		tt         TokenType
		text       string
		start, end string
	}
	got := make([]span, 0, len(tokens))
	for _, tok := range tokens {
		got = append(got, span{tok.Type(), tok.Text(), tok.Span().Start().String(), tok.Span().End().String()})
	}

	assert.Equal(t, []span{
		{TokenOpenExpression, "(", "1:0", "1:1"},
		{TokenWord, "foo", "1:1", "1:4"},
		{TokenWhitespace, " ", "1:4", "1:5"},
		{TokenInteger, "12", "1:5", "1:7"},
		{TokenNewLine, "\n", "1:7", "2:0"},
		{TokenWhitespace, "  ", "2:0", "2:2"},
		{TokenColon, ":", "2:2", "2:3"},
		{TokenSequence, "é", "2:3", "2:4"},
		{TokenCloseExpression, ")", "2:4", "2:5"},
		{TokenEOF, "", "2:5", "2:5"},
	}, got)

	for _, tok := range tokens {
		assert.Equal(t, tok.Span().Start(), tok.Pos(), "Pos is the start of the span of %v", tok)
	}
}

// TestNewTokenSpan checks that NewToken gives the empty span at its position.
func TestNewTokenSpan(t *testing.T) {
	tok := NewToken(TokenWord, "a", nil)
	assert.True(t, tok.Span().IsEmpty())
	assert.Equal(t, tok.Pos(), tok.Span().Start())
}
