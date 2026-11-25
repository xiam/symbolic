package parser

import (
	"bytes"
	"testing"

	"github.com/xiam/symbolic/ast"
)

// FuzzParse feeds arbitrary bytes to Parse. Parse must return a tree or an
// error, and never panic or hang. The spans of a tree must nest, and the
// source text of each node must parse to the same node. Run it with:
//
//	go test ./parser -run '^$' -fuzz FuzzParse -fuzztime 1m
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		``,
		`(fn_a (fn_b [89 :A :B [67 3.27]]) (fn_c 66 3 53 "Hello world!" 😊))`,
		"(a \"hi there\" # note\n  [é b {:c 1}])",
		`[+ -1 55 +6.3 +2 -3.23 4.01]`,
		`-123 +45.5 -1 +1abc a+b -.5 - 1`,
		`(a "")`,
		`"\"\\"`,
		`(((1 1 1`,
		`1 ](}`,
		"a\xffb\x00c",
		`:`,
		`1.`,
		`99999999999999999999999`,
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		root, err := Parse(in)
		if err != nil {
			if root != nil {
				t.Fatalf("Parse returned a tree and an error: %v", err)
			}
			return
		}
		if root == nil {
			t.Fatal("Parse returned no tree and no error")
		}
		_ = ast.Encode(root)
		checkSpans(t, in, root)
	})
}

// checkSpans checks the spans of a tree parsed from in: each span is in its
// parent's span, siblings do not overlap, and the source text of a node
// parses to a tree with that one node.
func checkSpans(t *testing.T, in []byte, root *ast.File) {
	t.Helper()

	// Positions are counted after a skipped byte order mark.
	src := bytes.TrimPrefix(in, []byte("\ufeff"))
	if got := root.Span().End().ByteOffset(); got != len(src) {
		t.Fatalf("File ends at byte %d, want %d", got, len(src))
	}

	for path, n := range ast.WalkPath(root) {
		span := n.Span()
		if span.Start().Compare(span.End()) > 0 {
			t.Fatalf("%T: span %v ends before it starts", n, span)
		}
		if len(path) > 0 {
			parent := path[len(path)-1].Span()
			if span.Start().Compare(parent.Start()) < 0 || span.End().Compare(parent.End()) > 0 {
				t.Fatalf("%T: span %v is not in the span %v of its parent", n, span, parent)
			}
		}

		var prev ast.Node
		for c := range n.Children() {
			if prev != nil && prev.Span().End().Compare(c.Span().Start()) > 0 {
				t.Fatalf("%T: span %v overlaps its sibling %v", c, c.Span(), prev.Span())
			}
			prev = c
		}

		if _, ok := n.(*ast.File); ok {
			continue
		}
		text := src[span.Start().ByteOffset():span.End().ByteOffset()]
		sub, err := Parse(text)
		if err != nil {
			t.Fatalf("%T: the text %q of span %v does not parse: %v", n, text, span, err)
		}
		if len(sub.Items) != 1 || !bytes.Equal(ast.Encode(sub.Items[0]), ast.Encode(n)) {
			t.Fatalf("%T: the text %q of span %v parses to %q, want %q", n, text, span, ast.Encode(sub), ast.Encode(n))
		}
	}
}
