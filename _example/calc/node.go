// Package calc is an arithmetic calculator with its own typed abstract
// syntax tree. It is the second concrete AST of symbolic, next to the
// S-expression AST of package ast, and it has the same shape: one struct type
// per kind of node behind a sealed interface and a span on every node. Package
// tree of symbolic walks its trees, through the method Children:
//
//	for n := range tree.All[*calc.Number](root) { ... }
//
// The grammar has numbers, the operators + - * /, unary + and -, and
// parentheses, with the usual precedence and left associativity:
//
//	expr    = term { ("+" | "-") term } .
//	term    = unary { ("*" | "/") unary } .
//	unary   = ("+" | "-") unary | primary .
//	primary = number | "(" expr ")" .
//
// Parse builds a tree, and Eval computes its value. A consumer reads a tree
// with a type switch and the exported fields of each type:
//
//	switch n := n.(type) {
//	case *calc.Number:
//		fmt.Println("number", n.Value)
//	case *calc.Binary:
//		fmt.Println("operator", n.Op)
//	}
//
// Every node records the half-open span of its source text. A node that a
// program builds by hand has the zero span. The operands of a node must not
// be nil.
package calc

import (
	"iter"

	"github.com/xiam/bytescan/cursor"
)

// Node is a node of a calculator tree. Only the types of this package
// implement it: Number, Paren, Unary and Binary.
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

// Op is an arithmetic operator.
type Op byte

// The operators. A Unary node uses Add and Sub, a Binary node uses all four.
const (
	Add Op = '+'
	Sub Op = '-'
	Mul Op = '*'
	Div Op = '/'
)

// String returns the operator as it is written, such as "+".
func (op Op) String() string { return string(rune(op)) }

// Number is a number, such as 42 or 3.25.
type Number struct {
	Value float64
	Loc   cursor.Span
}

// Paren is an expression in parentheses: ( X ). Loc includes both
// parentheses.
type Paren struct {
	X   Node
	Loc cursor.Span
}

// Unary is a unary operation: Op X, such as -3. Op is Add or Sub.
type Unary struct {
	Op  Op
	X   Node
	Loc cursor.Span
}

// Binary is a binary operation: X Op Y, such as 2 * 3. OpPos is the position
// of the operator, where Eval reports a division by zero.
type Binary struct {
	Op    Op
	OpPos cursor.Pos
	X, Y  Node
	Loc   cursor.Span
}

func (n *Number) Span() cursor.Span { return n.Loc }
func (n *Paren) Span() cursor.Span  { return n.Loc }
func (n *Unary) Span() cursor.Span  { return n.Loc }
func (n *Binary) Span() cursor.Span { return n.Loc }

// noChildren is the empty sequence of the children of a leaf.
func noChildren(func(Node) bool) {}

func (n *Number) Children() iter.Seq[Node] { return noChildren }
func (n *Paren) Children() iter.Seq[Node]  { return one(n.X) }
func (n *Unary) Children() iter.Seq[Node]  { return one(n.X) }

func (n *Binary) Children() iter.Seq[Node] {
	return func(yield func(Node) bool) {
		_ = yield(n.X) && yield(n.Y)
	}
}

// one returns the sequence of the single child x.
func one(x Node) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		yield(x)
	}
}

func (*Number) node() {}
func (*Paren) node()  {}
func (*Unary) node()  {}
func (*Binary) node() {}
