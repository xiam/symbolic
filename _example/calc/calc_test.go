package calc

import (
	"errors"
	"strings"
	"testing"

	"github.com/xiam/symbolic/tree"
)

// format writes the tree under n with every operation in parentheses, so a
// test can check the structure: "1 - 2 - 3" is "((1 - 2) - 3)".
func format(n Node) string {
	switch n := n.(type) {
	case *Paren:
		return "[" + format(n.X) + "]"
	case *Unary:
		return "(" + n.Op.String() + format(n.X) + ")"
	case *Binary:
		return "(" + format(n.X) + " " + n.Op.String() + " " + format(n.Y) + ")"
	}
	kind, value := describe(n)
	if kind != "Number" {
		return kind
	}
	return value
}

func TestEval(t *testing.T) {
	tests := []struct {
		in   string
		tree string
		want float64
	}{
		{"42", "42", 42},
		{"3.25", "3.25", 3.25},
		{"007", "7", 7},
		{"1 + 2", "(1 + 2)", 3},
		{"1 - 2 - 3", "((1 - 2) - 3)", -4},
		{"2 + 3 * 4", "(2 + (3 * 4))", 14},
		{"2 * 3 + 4", "((2 * 3) + 4)", 10},
		{"-(2 + 3) * 4", "((-[(2 + 3)]) * 4)", -20},
		{"(2 + 3) * 4", "([(2 + 3)] * 4)", 20},
		{"8 / 4 / 2", "((8 / 4) / 2)", 1},
		{"8 / (4 / 2)", "(8 / [(4 / 2)])", 4},
		{"7 / 2", "(7 / 2)", 3.5},
		{"-3", "(-3)", -3},
		{"+3", "(+3)", 3},
		{"--3", "(-(-3))", 3},
		{"- - 3", "(-(-3))", 3},
		{"2 * -3", "(2 * (-3))", -6},
		{"1 - -1", "(1 - (-1))", 2},
		{"-2 * 3", "((-2) * 3)", -6},
		{"((1))", "[[1]]", 1},
		{"1 + 2 * 3 - 4 / 2", "((1 + (2 * 3)) - (4 / 2))", 5},
		{"0.5 * (3.25 - -0.75)", "(0.5 * [(3.25 - (-0.75))])", 2},
		{"1+2*3", "(1 + (2 * 3))", 7},
		{" \t1\r\n+\n2 ", "(1 + 2)", 3},
		{"0 / 5", "(0 / 5)", 0},
		{"10 - 2 * 3 * 2 + 1", "((10 - ((2 * 3) * 2)) + 1)", -1},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			root, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.in, err)
			}
			if got := format(root); got != tt.tree {
				t.Errorf("Parse(%q) = %s, want %s", tt.in, got, tt.tree)
			}
			got, err := Eval(root)
			if err != nil {
				t.Fatalf("Eval(%q): %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Eval(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		in    string
		cause error
		pos   string // line:column
		token string
		text  string // the text of Error
	}{
		// Unexpected token.
		{"1 2", ErrUnexpectedToken, "1:2", "2", `1:2: unexpected token "2"`},
		{"2 * * 3", ErrUnexpectedToken, "1:4", "*", `1:4: unexpected token "*"`},
		{")", ErrUnexpectedToken, "1:0", ")", `1:0: unexpected token ")"`},
		{"(1 + 2]", ErrUnexpectedToken, "1:6", "]", `1:6: unexpected token "]"`},
		{"1 + x", ErrUnexpectedToken, "1:4", "x", `1:4: unexpected token "x"`},
		{"1.", ErrUnexpectedToken, "1:1", ".", `1:1: unexpected token "."`},
		{"()", ErrUnexpectedToken, "1:1", ")", `1:1: unexpected token ")"`},

		// Unclosed parenthesis: the position of the opening parenthesis.
		{"(1 + 2", ErrUnclosedParen, "1:0", "", `1:0: unclosed parenthesis`},
		{"1 * (2 - (3\n)", ErrUnclosedParen, "1:4", "", `1:4: unclosed parenthesis`},

		// Unexpected end of input: the position of the end of input.
		{"", ErrUnexpectedEOF, "1:0", "", `1:0: unexpected end of input`},
		{"4 +", ErrUnexpectedEOF, "1:3", "", `1:3: unexpected end of input`},
		{"4 *\n  ", ErrUnexpectedEOF, "2:2", "", `2:2: unexpected end of input`},
		{"(", ErrUnexpectedEOF, "1:1", "", `1:1: unexpected end of input`},
		{"-", ErrUnexpectedEOF, "1:1", "", `1:1: unexpected end of input`},

		// Division by zero: the position of the operator.
		{"1 / 0", ErrDivisionByZero, "1:2", "", `1:2: division by zero`},
		{"1 + 2 / (3 - 3)", ErrDivisionByZero, "1:6", "", `1:6: division by zero`},
		{"1 +\n  4 / 0.0 / 2", ErrDivisionByZero, "2:4", "", `2:4: division by zero`},

		// A number out of range.
		{"1" + strings.Repeat("0", 400), ErrRange, "1:0", "1" + strings.Repeat("0", 400), ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			root, err := Parse(tt.in)
			if err == nil {
				if root == nil {
					t.Fatalf("Parse(%q) returned no tree and no error", tt.in)
				}
				_, err = Eval(root)
			} else if root != nil {
				t.Fatalf("Parse(%q) returned a tree and an error", tt.in)
			}

			var e *Error
			if !errors.As(err, &e) {
				t.Fatalf("%q: got error %v, want an *Error", tt.in, err)
			}
			if !errors.Is(err, tt.cause) {
				t.Errorf("%q: got cause %v, want %v", tt.in, e.Err, tt.cause)
			}
			if got := e.Pos.String(); got != tt.pos {
				t.Errorf("%q: got position %s, want %s", tt.in, got, tt.pos)
			}
			if e.Token != tt.token {
				t.Errorf("%q: got token %q, want %q", tt.in, e.Token, tt.token)
			}
			if tt.text != "" && err.Error() != tt.text {
				t.Errorf("%q: got text %q, want %q", tt.in, err.Error(), tt.text)
			}
		})
	}
}

// TestSpans pins the span of every node of a multi-line expression. A span
// is half-open: it ends at the position after the last rune of the node.
func TestSpans(t *testing.T) {
	src := "(1.5 +\n  2) *\n\t-(3 - 4)  \n"
	want := `Binary * 1:0>3:9
  Paren 1:0>2:4
    Binary + 1:1>2:3
      Number 1.5 1:1>1:4
      Number 2 2:2>2:3
  Unary - 3:1>3:9
    Paren 3:2>3:9
      Binary - 3:3>3:8
        Number 3 3:3>3:4
        Number 4 3:7>3:8
`
	root, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	var got strings.Builder
	if err := Dump(&got, root); err != nil {
		t.Fatal(err)
	}
	if got.String() != want {
		t.Errorf("Dump:\n%s\nwant:\n%s", got.String(), want)
	}

	// The byte offsets of each span cut the source text of its node.
	texts := map[string]bool{}
	for n := range tree.Walk(root) {
		texts[src[n.Span().Start().ByteOffset():n.Span().End().ByteOffset()]] = true
	}
	for _, text := range []string{"(1.5 +\n  2) *\n\t-(3 - 4)", "(1.5 +\n  2)", "1.5 +\n  2", "-(3 - 4)", "3 - 4", "1.5"} {
		if !texts[text] {
			t.Errorf("no node has the text %q", text)
		}
	}
}

// TestEvalInvalid checks the trees that Parse does not build: a program can
// build them by hand.
func TestEvalInvalid(t *testing.T) {
	one := &Number{Value: 1}
	for _, n := range []Node{
		nil,
		&Unary{Op: Mul, X: one},
		&Binary{Op: Op('%'), X: one, Y: one},
	} {
		if _, err := Eval(n); err == nil {
			t.Errorf("Eval(%#v) returned no error", n)
		}
	}
}
