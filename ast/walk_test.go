package ast

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sym returns a symbol node named name.
func sym(name string) *Symbol { return &Symbol{Name: name} }

// sampleTree is the tree of `(a [b c] {:k 1}) d`, built by hand.
func sampleTree() *File {
	return &File{Items: []Node{
		&Expr{Items: []Node{
			sym("a"),
			&List{Items: []Node{sym("b"), sym("c")}},
			&Map{Items: []Node{&Atom{Name: ":k"}, &Int{Value: 1}}},
		}},
		sym("d"),
	}}
}

// label returns a short name of n for the tests.
func label(n Node) string {
	kind, value := describe(n)
	if value != "" {
		return value
	}
	return kind
}

func TestWalkOrder(t *testing.T) {
	var got []string
	for n := range Walk(sampleTree()) {
		got = append(got, label(n))
	}
	// Depth first, parents before children, children in order.
	assert.Equal(t, []string{"File", "Expr", "a", "List", "b", "c", "Map", ":k", "1", "d"}, got)
}

func TestWalkLeafAndNil(t *testing.T) {
	assert.Equal(t, []Node{&Int{Value: 7}}, slices.Collect(Walk(&Int{Value: 7})))
	assert.Empty(t, slices.Collect(Walk(nil)))
}

func TestWalkPath(t *testing.T) {
	var got []string
	for path, n := range WalkPath(sampleTree()) {
		line := ""
		for _, p := range path {
			line += label(p) + "/"
		}
		got = append(got, line+label(n))
	}
	assert.Equal(t, []string{
		"File",
		"File/Expr",
		"File/Expr/a",
		"File/Expr/List",
		"File/Expr/List/b",
		"File/Expr/List/c",
		"File/Expr/Map",
		"File/Expr/Map/:k",
		"File/Expr/Map/1",
		"File/d",
	}, got)
}

// TestWalkPathAppend checks that appending to a yielded path does not change
// the path that WalkPath keeps.
func TestWalkPathAppend(t *testing.T) {
	var got []int
	for path, n := range WalkPath(sampleTree()) {
		_ = append(path, n) // must not write into the path of WalkPath
		got = append(got, len(path))
	}
	assert.Equal(t, []int{0, 1, 2, 2, 3, 3, 2, 3, 3, 1}, got)
}

func TestAll(t *testing.T) {
	var names []string
	for s := range All[*Symbol](sampleTree()) {
		names = append(names, s.Name)
	}
	assert.Equal(t, []string{"a", "b", "c", "d"}, names)

	assert.Len(t, slices.Collect(All[*Map](sampleTree())), 1)
	assert.Empty(t, slices.Collect(All[*Float](sampleTree())))
}

func TestFind(t *testing.T) {
	n, ok := Find(sampleTree(), func(n Node) bool {
		_, ok := n.(*List)
		return ok
	})
	require.True(t, ok)
	assert.Len(t, n.(*List).Items, 2)

	n, ok = Find(sampleTree(), func(Node) bool { return false })
	assert.False(t, ok)
	assert.Nil(t, n)
}

// TestEarlyStop checks that a break in a range loop stops each traversal: the
// iterator does not call yield again.
func TestEarlyStop(t *testing.T) {
	count := 0
	for range Walk(sampleTree()) {
		count++
		if count == 3 {
			break
		}
	}
	assert.Equal(t, 3, count)

	count = 0
	for range WalkPath(sampleTree()) {
		count++
		if count == 5 {
			break
		}
	}
	assert.Equal(t, 5, count)

	count = 0
	for range All[*Symbol](sampleTree()) {
		count++
		break
	}
	assert.Equal(t, 1, count)

	// Find stops at the first match.
	visited := 0
	_, ok := Find(sampleTree(), func(n Node) bool {
		visited++
		return label(n) == "a"
	})
	assert.True(t, ok)
	assert.Equal(t, 3, visited)
}

// TestDeepTree walks a tree 100,000 levels deep.
func TestDeepTree(t *testing.T) {
	const depth = 100_000

	var root Node = sym("leaf")
	for range depth {
		root = &Expr{Items: []Node{root}}
	}

	count := 0
	for range Walk(root) {
		count++
	}
	assert.Equal(t, depth+1, count)

	deepest := 0
	for path := range WalkPath(root) {
		deepest = max(deepest, len(path))
	}
	assert.Equal(t, depth, deepest)

	leaf, ok := Find(root, func(n Node) bool {
		_, ok := n.(*Symbol)
		return ok
	})
	require.True(t, ok)
	assert.Equal(t, "leaf", leaf.(*Symbol).Name)
	assert.Len(t, slices.Collect(All[*Expr](root)), depth)
}
