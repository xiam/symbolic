package ast

import (
	"iter"

	"github.com/xiam/symbolic/tree"
)

// The functions of this file call package tree with the node type of this
// package, so that a caller with an S-expression tree does not need to know
// package tree.

// Walk returns every node of the tree under root, root included, depth
// first: a parent comes before its children, and children come in source
// order. A break in the range loop stops the traversal. See tree.Walk.
func Walk(root Node) iter.Seq[Node] { return tree.Walk(root) }

// WalkPath returns the nodes of Walk, in the same order, each with its
// ancestors: the path from root to the parent of the node. The path of root
// is empty.
//
// The path slice is reused for the next node: copy it (for example with
// slices.Clone) to keep it after the loop body returns. See tree.WalkPath.
func WalkPath(root Node) iter.Seq2[[]Node, Node] { return tree.WalkPath(root) }

// All returns the nodes of Walk whose dynamic type is T, such as *Symbol.
// See tree.All.
func All[T Node](root Node) iter.Seq[T] { return tree.All[T](root) }

// Find returns the first node of Walk for which match returns true. It
// reports false when no node matches. See tree.Find.
func Find(root Node, match func(Node) bool) (Node, bool) { return tree.Find(root, match) }
