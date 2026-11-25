package calc

import (
	"errors"
	"strings"
	"testing"

	"github.com/xiam/symbolic/tree"
)

// FuzzParse feeds arbitrary text to Parse. Parse must return a tree or an
// *Error, and never panic or hang. In a tree, each span lies in its parent's
// span, siblings do not overlap, the root covers the input without its
// leading and trailing whitespace, and the source text of each node parses
// to the same tree. Eval must return a value or a division by zero. Run it
// with:
//
//	go test ./_example/calc -run '^$' -fuzz FuzzParse -fuzztime 1m
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"",
		"1 + 2",
		"1 - 2 - 3",
		"-(2 + 3) * 4",
		"(1.5 +\n  2) *\n\t-(3 - 4)",
		"1 / 0",
		"(1 + 2",
		"2 * * 3",
		"((((1))))",
		"1." + strings.Repeat("0", 50),
		"1" + strings.Repeat("0", 400),
		"\xff+1",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, src string) {
		root, err := Parse(src)
		if err != nil {
			if root != nil {
				t.Fatalf("Parse returned a tree and an error: %v", err)
			}
			var e *Error
			if !errors.As(err, &e) {
				t.Fatalf("Parse returned %T, want *Error: %v", err, err)
			}
			return
		}
		if root == nil {
			t.Fatal("Parse returned no tree and no error")
		}

		if _, err := Eval(root); err != nil && !errors.Is(err, ErrDivisionByZero) {
			t.Fatalf("Eval: %v", err)
		}
		checkSpans(t, src, root)
	})
}

// checkSpans checks the spans of a tree parsed from src.
func checkSpans(t *testing.T, src string, root Node) {
	t.Helper()

	if text := source(src, root); text != strings.Trim(src, " \t\v\f\r\n") {
		t.Fatalf("the root has the text %q, want the trimmed input", text)
	}

	for path, n := range tree.WalkPath(root) {
		span := n.Span()
		if span.Start().Compare(span.End()) >= 0 {
			t.Fatalf("%T: span %v is empty or reversed", n, span)
		}
		if len(path) > 0 {
			parent := path[len(path)-1].Span()
			if span.Start().Compare(parent.Start()) < 0 || span.End().Compare(parent.End()) > 0 {
				t.Fatalf("%T: span %v is not in the span %v of its parent", n, span, parent)
			}
		}

		var prev Node
		for c := range n.Children() {
			if prev != nil && prev.Span().End().Compare(c.Span().Start()) > 0 {
				t.Fatalf("%T: span %v overlaps its sibling %v", c, c.Span(), prev.Span())
			}
			prev = c
		}

		text := source(src, n)
		sub, err := Parse(text)
		if err != nil {
			t.Fatalf("%T: the text %q of span %v does not parse: %v", n, text, span, err)
		}
		if got, want := format(sub), format(n); got != want {
			t.Fatalf("%T: the text %q of span %v parses to %s, want %s", n, text, span, got, want)
		}
	}
}

// source returns the source text of n.
func source(src string, n Node) string {
	return src[n.Span().Start().ByteOffset():n.Span().End().ByteOffset()]
}
