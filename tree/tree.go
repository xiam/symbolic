// Package tree walks a tree of nodes that list their own children.
//
// A grammar defines its own node types, usually one struct type per kind of
// node behind an interface. If that interface has the method
//
//	Children() iter.Seq[N]
//
// where N is the interface itself, the functions of this package walk its
// trees: Walk, WalkPath, All and Find. They use no other method of a node.
//
// The package only reads a tree. It does not change or rebuild a tree, it has
// no visitor interface, and it does not treat a tree as a graph: a node that
// is reachable twice is visited twice, and a cycle makes a walk endless.
package tree

import (
	"iter"
	"slices"
)

// Node is the constraint of a node type N: a node returns its direct children,
// which are of type N too. A leaf returns an empty sequence.
type Node[N any] interface {
	Children() iter.Seq[N]
}

// Walk returns every node of the tree under root, root included, depth
// first: a parent comes before its children, and children come in the order
// of Children. A break in the range loop stops the traversal. A nil root
// interface has no nodes.
func Walk[N Node[N]](root N) iter.Seq[N] {
	return func(yield func(N) bool) {
		if any(root) != nil {
			walk(root, yield)
		}
	}
}

// walk yields n and then its subtrees. It returns false when yield asked to
// stop.
func walk[N Node[N]](n N, yield func(N) bool) bool {
	if !yield(n) {
		return false
	}
	for c := range n.Children() {
		if !walk(c, yield) {
			return false
		}
	}
	return true
}

// WalkPath returns the nodes of Walk, in the same order, each with its
// ancestors: the path from root to the parent of the node. The path of root
// is empty.
//
// The path slice is reused for the next node: copy it (for example with
// slices.Clone) to keep it after the loop body returns.
func WalkPath[N Node[N]](root N) iter.Seq2[[]N, N] {
	return func(yield func([]N, N) bool) {
		if any(root) == nil {
			return
		}
		var path []N
		var visit func(n N) bool
		visit = func(n N) bool {
			if !yield(slices.Clip(path), n) {
				return false
			}
			path = append(path, n)
			for c := range n.Children() {
				if !visit(c) {
					return false
				}
			}
			path = path[:len(path)-1]
			return true
		}
		visit(root)
	}
}

// All returns the nodes of Walk whose dynamic type is T, such as *Symbol. The
// caller names T, and Go infers N from root:
//
//	for sym := range tree.All[*ast.Symbol](root) { ... }
func All[T any, N Node[N]](root N) iter.Seq[T] {
	return func(yield func(T) bool) {
		for n := range Walk(root) {
			if t, ok := any(n).(T); ok && !yield(t) {
				return
			}
		}
	}
}

// Find returns the first node of Walk for which match returns true. It
// reports false, with the zero N, when no node matches.
func Find[N Node[N]](root N, match func(N) bool) (N, bool) {
	for n := range Walk(root) {
		if match(n) {
			return n, true
		}
	}
	var zero N
	return zero, false
}
