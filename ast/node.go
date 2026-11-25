// Package ast defines the abstract syntax tree of an S-expression.
//
// Each kind of node is its own struct type, used through a pointer, and
// every node satisfies the interface Node. A consumer reads a tree with a type
// switch and the exported fields of each type:
//
//	switch n := n.(type) {
//	case *ast.Int:
//		fmt.Println("integer", n.Value)
//	case *ast.Expr:
//		fmt.Println("expression with", len(n.Items), "items")
//	}
//
// Walk, WalkPath, All and Find traverse a tree. They call package tree, which
// uses only the method Children of a node.
//
// Every node records the half-open span of its source text. A node that a
// program builds by hand has the zero span. The Items of a container must not
// hold a nil node.
package ast

import (
	"iter"
	"slices"

	"github.com/xiam/bytescan/cursor"
)

// Node is a node of an S-expression tree. Only the types of this package
// implement it: Int, Float, String, Symbol, Atom, Expr, List, Map and File.
type Node interface {
	// Span returns the half-open range of the source text of the node.
	Span() cursor.Span

	// Children returns the direct children of the node, in source order. A
	// leaf has no children.
	Children() iter.Seq[Node]

	// node seals the interface: no type outside this package implements
	// Node.
	node()
}

// Int is an integer, such as 42 or -7.
type Int struct {
	Value int64
	Loc   cursor.Span
}

// Float is a floating point number, such as 3.27 or -0.5.
type Float struct {
	Value float64
	Loc   cursor.Span
}

// String is a double-quoted string. Value holds the text between the
// quotes, and Loc includes both quotes.
type String struct {
	Value string
	Loc   cursor.Span
}

// Symbol is a bare word, such as fn_a or +.
type Symbol struct {
	Name string
	Loc  cursor.Span
}

// Atom is a word preceded by a colon, such as :A. Name includes the colon.
type Atom struct {
	Name string
	Loc  cursor.Span
}

// Expr is an expression in parentheses: ( ... ).
type Expr struct {
	Items []Node
	Loc   cursor.Span
}

// List is a list in square brackets: [ ... ].
type List struct {
	Items []Node
	Loc   cursor.Span
}

// Map is a map in curly brackets: { ... }. Items is a flat sequence of keys
// and values: key, value, key, value. An odd number of items is legal; the
// last item then has no value (see Odd).
type Map struct {
	Items []Node
	Loc   cursor.Span
}

// File is the root of a parsed input: the sequence of top-level nodes. Its
// span covers the whole input, after a skipped byte order mark.
type File struct {
	Items []Node
	Loc   cursor.Span
}

// Pairs returns the key and value pairs of the map, in source order. When the
// map has an odd number of items, the last item has no value and Pairs
// skips it.
func (m *Map) Pairs() iter.Seq2[Node, Node] {
	return func(yield func(Node, Node) bool) {
		for i := 0; i+1 < len(m.Items); i += 2 {
			if !yield(m.Items[i], m.Items[i+1]) {
				return
			}
		}
	}
}

// Odd reports whether the map has an odd number of items, that is, a last
// key without a value.
func (m *Map) Odd() bool {
	return len(m.Items)%2 == 1
}

func (n *Int) Span() cursor.Span    { return n.Loc }
func (n *Float) Span() cursor.Span  { return n.Loc }
func (n *String) Span() cursor.Span { return n.Loc }
func (n *Symbol) Span() cursor.Span { return n.Loc }
func (n *Atom) Span() cursor.Span   { return n.Loc }
func (n *Expr) Span() cursor.Span   { return n.Loc }
func (n *List) Span() cursor.Span   { return n.Loc }
func (n *Map) Span() cursor.Span    { return n.Loc }
func (n *File) Span() cursor.Span   { return n.Loc }

// noChildren is the empty sequence of the children of a leaf.
func noChildren(func(Node) bool) {}

func (n *Int) Children() iter.Seq[Node]    { return noChildren }
func (n *Float) Children() iter.Seq[Node]  { return noChildren }
func (n *String) Children() iter.Seq[Node] { return noChildren }
func (n *Symbol) Children() iter.Seq[Node] { return noChildren }
func (n *Atom) Children() iter.Seq[Node]   { return noChildren }
func (n *Expr) Children() iter.Seq[Node]   { return slices.Values(n.Items) }
func (n *List) Children() iter.Seq[Node]   { return slices.Values(n.Items) }
func (n *Map) Children() iter.Seq[Node]    { return slices.Values(n.Items) }
func (n *File) Children() iter.Seq[Node]   { return slices.Values(n.Items) }

func (*Int) node()    {}
func (*Float) node()  {}
func (*String) node() {}
func (*Symbol) node() {}
func (*Atom) node()   {}
func (*Expr) node()   {}
func (*List) node()   {}
func (*Map) node()    {}
func (*File) node()   {}
