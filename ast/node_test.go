package ast

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every node type implements Node.
var (
	_ Node = (*Int)(nil)
	_ Node = (*Float)(nil)
	_ Node = (*String)(nil)
	_ Node = (*Symbol)(nil)
	_ Node = (*Atom)(nil)
	_ Node = (*Expr)(nil)
	_ Node = (*List)(nil)
	_ Node = (*Map)(nil)
	_ Node = (*File)(nil)
)

func TestLeavesHaveNoChildren(t *testing.T) {
	for _, n := range []Node{&Int{}, &Float{}, &String{}, &Symbol{}, &Atom{}} {
		assert.Empty(t, slices.Collect(n.Children()), "%T", n)
	}
}

func TestContainerChildren(t *testing.T) {
	items := []Node{sym("a"), sym("b")}
	for _, n := range []Node{&Expr{Items: items}, &List{Items: items}, &Map{Items: items}, &File{Items: items}} {
		assert.Equal(t, items, slices.Collect(n.Children()), "%T", n)
	}
}

func TestMapPairs(t *testing.T) {
	type pair struct{ k, v string }
	pairs := func(m *Map) []pair {
		var out []pair
		for k, v := range m.Pairs() {
			out = append(out, pair{label(k), label(v)})
		}
		return out
	}

	even := &Map{Items: []Node{sym("a"), sym("1"), sym("b"), sym("2")}}
	assert.False(t, even.Odd())
	assert.Equal(t, []pair{{"a", "1"}, {"b", "2"}}, pairs(even))

	// {+ 1 2 3 4}: the last item has no value, and Pairs skips it.
	odd := &Map{Items: []Node{sym("+"), sym("1"), sym("2"), sym("3"), sym("4")}}
	assert.True(t, odd.Odd())
	assert.Equal(t, []pair{{"+", "1"}, {"2", "3"}}, pairs(odd))

	empty := &Map{}
	assert.False(t, empty.Odd())
	assert.Empty(t, pairs(empty))

	// A break stops Pairs.
	count := 0
	for range even.Pairs() {
		count++
		break
	}
	assert.Equal(t, 1, count)
}

func TestEncode(t *testing.T) {
	tree := &File{Items: []Node{
		&Expr{Items: []Node{
			sym("fn"),
			&Int{Value: -12},
			&Float{Value: 3.27},
			&String{Value: "say \"hi\""},
			&List{Items: []Node{&Atom{Name: ":a"}}},
			&Map{},
		}},
		&List{},
	}}
	assert.Equal(t, `(fn -12 3.27 "say \"hi\"" [:a] {}) []`, string(Encode(tree)))

	// Only the File is written without brackets.
	assert.Equal(t, `[1 2]`, string(Encode(&List{Items: []Node{&Int{Value: 1}, &Int{Value: 2}}})))
	assert.Equal(t, ``, string(Encode(&File{})))
}

func TestDump(t *testing.T) {
	var b strings.Builder
	require.NoError(t, Dump(&b, sampleTree()))
	assert.Equal(t, `File 0:0>0:0
  Expr 0:0>0:0
    Symbol a 0:0>0:0
    List 0:0>0:0
      Symbol b 0:0>0:0
      Symbol c 0:0>0:0
    Map 0:0>0:0
      Atom :k 0:0>0:0
      Int 1 0:0>0:0
  Symbol d 0:0>0:0
`, b.String())
}

// failWriter fails after n writes.
type failWriter struct{ n int }

var errWrite = errors.New("write failed")

func (w *failWriter) Write(p []byte) (int, error) {
	if w.n == 0 {
		return 0, errWrite
	}
	w.n--
	return len(p), nil
}

func TestWriteErrors(t *testing.T) {
	assert.ErrorIs(t, Fprint(&failWriter{n: 2}, sampleTree()), errWrite)
	assert.ErrorIs(t, Dump(&failWriter{n: 2}, sampleTree()), errWrite)
}
