package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xiam/symbolic/ast"
	"github.com/xiam/symbolic/lexer"
)

// positionInput has two lines, a string, a comment, nested brackets and the
// non-ASCII rune "é" before the token "b" on the second line.
const positionInput = "(a \"hi there\" # note\n  [é b {:c 1}])"

func linecol(tok *lexer.Token) [2]int {
	pos := tok.Pos()
	return [2]int{pos.Line(), pos.Column()}
}

// dump returns the text of ast.Dump for root: one line per node, with its
// kind, its value and its span.
func dump(t *testing.T, root ast.Node) string {
	t.Helper()
	var b strings.Builder
	require.NoError(t, ast.Dump(&b, root))
	return b.String()
}

// TestPositions pins the start and the end of every node of positionInput
// (output 3 of #2). Coordinates follow bytescan: 1-based lines and 0-based
// rune columns; a span is half-open.
func TestPositions(t *testing.T) {
	root, err := Parse([]byte(positionInput))
	require.NoError(t, err)

	// A string runs from its opening to after its closing double quote, and
	// a container from its opening to after its closing bracket. Columns
	// count runes: "é" is two bytes and one column.
	assert.Equal(t, `File 1:0>2:15
  Expr 1:0>2:15
    Symbol a 1:1>1:2
    String "hi there" 1:3>1:13
    List 2:2>2:14
      Symbol é 2:3>2:4
      Symbol b 2:5>2:6
      Map 2:7>2:13
        Atom :c 2:8>2:10
        Int 1 2:11>2:12
`, dump(t, root))

	b, ok := ast.Find(root, func(n ast.Node) bool {
		sym, ok := n.(*ast.Symbol)
		return ok && sym.Name == "b"
	})
	require.True(t, ok)
	// Line 1 has 21 bytes, and "  [é " has 6 bytes.
	assert.Equal(t, 27, b.Span().Start().ByteOffset(), "b")
	assert.Equal(t, 28, b.Span().End().ByteOffset(), "b")
	assert.Equal(t, len(positionInput), root.Span().End().ByteOffset(), "File")

	tokens, err := lexer.Tokenize([]byte(positionInput))
	require.NoError(t, err)

	var newline, eof *lexer.Token
	for i := range tokens {
		switch tokens[i].Type() {
		case lexer.TokenNewLine:
			newline = &tokens[i]
		case lexer.TokenEOF:
			eof = &tokens[i]
		}
	}
	// A newline token is on the line it ends, at the column of the "\n".
	require.NotNil(t, newline)
	assert.Equal(t, [2]int{1, 20}, linecol(newline))
	// TokenEOF is the last token, at the end of input.
	require.NotNil(t, eof)
	assert.Equal(t, eof, &tokens[len(tokens)-1])
	assert.Equal(t, [2]int{2, 15}, linecol(eof))

	// After a final newline, TokenEOF is at column 0 of the next line.
	tokens, err = lexer.Tokenize([]byte("a\n"))
	require.NoError(t, err)
	assert.Equal(t, [2]int{2, 0}, linecol(&tokens[len(tokens)-1]))
}

// TestSpans pins the spans of output 3 of #2 for an empty string, an empty
// input, signed numbers and a byte order mark.
func TestSpans(t *testing.T) {
	testCases := []struct {
		In   string
		Dump string
	}{
		{
			// An empty string runs from its opening to after its closing
			// double quote, like any string.
			In: `""`,
			Dump: `File 1:0>1:2
  String "" 1:0>1:2
`,
		},
		{
			In: `(a "")`,
			Dump: `File 1:0>1:6
  Expr 1:0>1:6
    Symbol a 1:1>1:2
    String "" 1:3>1:5
`,
		},
		{
			In:   ``,
			Dump: "File 1:0>1:0\n",
		},
		{
			// A signed number starts at its sign.
			In: `-123 +45.5`,
			Dump: `File 1:0>1:10
  Int -123 1:0>1:4
  Float 45.5 1:5>1:10
`,
		},
		{
			// The File starts after a skipped byte order mark.
			In: "\ufeff(a 1)\n",
			Dump: `File 1:0>2:0
  Expr 1:0>1:5
    Symbol a 1:1>1:2
    Int 1 1:3>1:4
`,
		},
	}

	for _, tc := range testCases {
		root, err := Parse([]byte(tc.In))
		require.NoError(t, err, "input %q", tc.In)
		assert.Equal(t, tc.Dump, dump(t, root), "input %q", tc.In)
	}
}

// TestSpansAutoCloseOnEOF checks that, with AutoCloseOnEOF, an unclosed
// container ends at the end of input.
func TestSpansAutoCloseOnEOF(t *testing.T) {
	p := NewParser(strings.NewReader("(a [b {"))
	p.SetOptions(ParserOptions{AutoCloseOnEOF: true})
	require.NoError(t, p.Parse())
	assert.Equal(t, `File 1:0>1:7
  Expr 1:0>1:7
    Symbol a 1:1>1:2
    List 1:3>1:7
      Symbol b 1:4>1:5
      Map 1:6>1:7
`, dump(t, p.RootNode()))
}

// TestEmptyString checks that an empty string parses; it panicked before the
// port of #1.
func TestEmptyString(t *testing.T) {
	root, err := Parse([]byte(`(a "")`))
	require.NoError(t, err)

	node, ok := root.Items[0].(*ast.Expr).Items[1].(*ast.String)
	require.True(t, ok)
	assert.Equal(t, "", node.Value)
}

// TestSignedNumberNodes pins owner decision 5 of #1 in the tree: a sign and
// all its digits are one number node.
func TestSignedNumberNodes(t *testing.T) {
	testCases := []struct {
		In   string
		Want ast.Node
	}{
		{"-123", &ast.Int{Value: -123}},
		{"+45.5", &ast.Float{Value: 45.5}},
		{"-1", &ast.Int{Value: -1}},
	}

	for _, tc := range testCases {
		root, err := Parse([]byte(tc.In))
		require.NoError(t, err)
		require.Len(t, root.Items, 1, "input %q", tc.In)
		switch got := root.Items[0].(type) {
		case *ast.Int:
			assert.Equal(t, tc.Want.(*ast.Int).Value, got.Value, "input %q", tc.In)
		case *ast.Float:
			assert.Equal(t, tc.Want.(*ast.Float).Value, got.Value, "input %q", tc.In)
		default:
			t.Errorf("input %q: got %T, want %T", tc.In, got, tc.Want)
		}
	}
}

// TestByteOrderMark checks that one byte order mark at the start of the input
// creates no node, as before the port, and that a later one stays in its
// symbol.
func TestByteOrderMark(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{"\ufeff", ``},
		{"\ufeffabc", `abc`},
		{"\ufeff(a 1)", `(a 1)`},
		{"(a \ufeffb)", "(a \ufeffb)"},
	}

	for _, tc := range testCases {
		root, err := Parse([]byte(tc.In))
		require.NoError(t, err, "input %q", tc.In)
		assert.Equal(t, tc.Out, string(ast.Encode(root)), "input %q", tc.In)
	}
}
