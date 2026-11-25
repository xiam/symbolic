package calc

import (
	"slices"
	"strings"
	"testing"

	"github.com/xiam/symbolic/tree"
)

// mustParse parses src or fails the test.
func mustParse(t *testing.T, src string) Node {
	t.Helper()
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	return root
}

// kinds returns the kind and value of each node, as Dump writes them.
func kinds(nodes []Node) string {
	var s []string
	for _, n := range nodes {
		kind, value := describe(n)
		s = append(s, strings.TrimSpace(kind+" "+value))
	}
	return strings.Join(s, ", ")
}

func TestWalk(t *testing.T) {
	root := mustParse(t, "-(1 + 2) * 3")
	got := kinds(slices.Collect(tree.Walk(root)))
	want := "Binary *, Unary -, Paren, Binary +, Number 1, Number 2, Number 3"
	if got != want {
		t.Errorf("Walk: %s, want %s", got, want)
	}

	if n := len(slices.Collect(tree.Walk[Node](nil))); n != 0 {
		t.Errorf("Walk(nil) returned %d nodes", n)
	}
}

func TestWalkPath(t *testing.T) {
	root := mustParse(t, "1 + 2 * 3")
	var got []string
	for path, n := range tree.WalkPath(root) {
		got = append(got, kinds(path)+" / "+kinds([]Node{n}))
	}
	want := []string{
		" / Binary +",
		"Binary + / Number 1",
		"Binary + / Binary *",
		"Binary +, Binary * / Number 2",
		"Binary +, Binary * / Number 3",
	}
	if !slices.Equal(got, want) {
		t.Errorf("WalkPath:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// TestWalkPathAliasing checks that an append to a yielded path does not
// change the path of a later node.
func TestWalkPathAliasing(t *testing.T) {
	root := mustParse(t, "(1 + 2) * (3 + 4)")
	var kept [][]Node
	for path, n := range tree.WalkPath(root) {
		_ = append(path, n) // Must not write into the reused path.
		kept = append(kept, slices.Clone(path))
	}
	for path, n := range tree.WalkPath(root) {
		if len(kept) == 0 || !slices.Equal(kept[0], path) {
			t.Fatalf("path of %s changed", kinds([]Node{n}))
		}
		kept = kept[1:]
	}
}

func TestAll(t *testing.T) {
	root := mustParse(t, "1 + -2 * (3 - 4)")

	var values []float64
	for n := range tree.All[*Number](root) {
		values = append(values, n.Value)
	}
	if want := []float64{1, 2, 3, 4}; !slices.Equal(values, want) {
		t.Errorf("All[*Number]: %v, want %v", values, want)
	}

	var ops []string
	for n := range tree.All[*Binary](root) {
		ops = append(ops, n.Op.String())
	}
	if want := []string{"+", "*", "-"}; !slices.Equal(ops, want) {
		t.Errorf("All[*Binary]: %v, want %v", ops, want)
	}
}

func TestFind(t *testing.T) {
	root := mustParse(t, "1 + 2 / 0")

	n, ok := tree.Find(root, func(n Node) bool {
		b, ok := n.(*Binary)
		return ok && b.Op == Div
	})
	if !ok {
		t.Fatal("Find: no division")
	}
	if got := n.Span().String(); got != "1:4>1:9" {
		t.Errorf("Find: the division has span %s, want 1:4>1:9", got)
	}

	if _, ok := tree.Find(root, func(n Node) bool { _, ok := n.(*Paren); return ok }); ok {
		t.Error("Find: found a Paren in a tree without one")
	}
}

// TestEarlyStop checks that a break stops each iterator at once.
func TestEarlyStop(t *testing.T) {
	root := mustParse(t, "(1 + 2) * (3 + 4)")

	count := 0
	for range tree.Walk(root) {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("Walk: %d nodes after break, want 2", count)
	}

	count = 0
	for range tree.WalkPath(root) {
		count++
		if count == 3 {
			break
		}
	}
	if count != 3 {
		t.Errorf("WalkPath: %d nodes after break, want 3", count)
	}

	count = 0
	for range tree.All[*Number](root) {
		count++
		break
	}
	if count != 1 {
		t.Errorf("All: %d nodes after break, want 1", count)
	}

	visited := 0
	tree.Find(root, func(Node) bool { visited++; return visited == 3 })
	if visited != 3 {
		t.Errorf("Find visited %d nodes, want 3", visited)
	}

	// Children of a Binary stop after the first child.
	b := root.(*Binary)
	count = 0
	for range b.Children() {
		count++
		break
	}
	if count != 1 {
		t.Errorf("Binary.Children: %d children after break, want 1", count)
	}
}

// TestDeepTree walks a tree 100,000 levels deep.
func TestDeepTree(t *testing.T) {
	const depth = 100_000
	var root Node = &Number{Value: 1}
	for range depth {
		root = &Unary{Op: Sub, X: root}
	}

	if n := len(slices.Collect(tree.Walk(root))); n != depth+1 {
		t.Errorf("Walk: %d nodes, want %d", n, depth+1)
	}
	maxPath := 0
	for path := range tree.WalkPath(root) {
		maxPath = max(maxPath, len(path))
	}
	if maxPath != depth {
		t.Errorf("WalkPath: longest path %d, want %d", maxPath, depth)
	}
	if _, ok := tree.Find(root, func(n Node) bool { _, ok := n.(*Number); return ok }); !ok {
		t.Error("Find: no number")
	}
	if n := len(slices.Collect(tree.All[*Number](root))); n != 1 {
		t.Errorf("All: %d numbers, want 1", n)
	}
	if v, err := Eval(root); err != nil || v != 1 {
		t.Errorf("Eval = %v, %v, want 1", v, err)
	}
}
