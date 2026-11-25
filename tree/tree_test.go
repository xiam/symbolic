package tree_test

import (
	"go/build"
	"go/parser"
	"go/token"
	"iter"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/xiam/symbolic/tree"
)

// node is the node family of the tests: a leaf or a branch.
type node interface {
	Children() iter.Seq[node]
}

// leaf is a node without children.
type leaf string

// branch is a named node with children.
type branch struct {
	name string
	kids []node
}

func (leaf) Children() iter.Seq[node]      { return func(func(node) bool) {} }
func (b *branch) Children() iter.Seq[node] { return slices.Values(b.kids) }

// br returns a branch named name with the children kids.
func br(name string, kids ...node) *branch { return &branch{name: name, kids: kids} }

// sample is the tree
//
//	a
//	├── b
//	│   ├── 1
//	│   └── 2
//	├── 3
//	└── c
//	    └── d
//	        └── 4
func sample() node {
	return br("a",
		br("b", leaf("1"), leaf("2")),
		leaf("3"),
		br("c", br("d", leaf("4"))),
	)
}

// name returns the name of a branch or the text of a leaf.
func name(n node) string {
	switch n := n.(type) {
	case leaf:
		return string(n)
	case *branch:
		return n.name
	}
	return "?"
}

// names returns the names of nodes, separated by spaces.
func names(nodes []node) string {
	s := make([]string, len(nodes))
	for i, n := range nodes {
		s[i] = name(n)
	}
	return strings.Join(s, " ")
}

func TestWalk(t *testing.T) {
	got := names(slices.Collect(tree.Walk(sample())))
	// Depth first, parents before children, children in order.
	if want := "a b 1 2 3 c d 4"; got != want {
		t.Errorf("Walk: %q, want %q", got, want)
	}

	if got := names(slices.Collect(tree.Walk[node](leaf("x")))); got != "x" {
		t.Errorf("Walk of a leaf: %q, want %q", got, "x")
	}
	if n := len(slices.Collect(tree.Walk[node](nil))); n != 0 {
		t.Errorf("Walk(nil): %d nodes, want 0", n)
	}
}

// samplePaths is the result of WalkPath on sample: the names of the path, a
// slash and the name of the node.
var samplePaths = []string{
	" / a",
	"a / b",
	"a b / 1",
	"a b / 2",
	"a / 3",
	"a / c",
	"a c / d",
	"a c d / 4",
}

func TestWalkPath(t *testing.T) {
	var got []string
	for path, n := range tree.WalkPath(sample()) {
		got = append(got, names(path)+" / "+name(n))
	}
	if !slices.Equal(got, samplePaths) {
		t.Errorf("WalkPath:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(samplePaths, "\n"))
	}

	for range tree.WalkPath[node](nil) {
		t.Error("WalkPath(nil) yielded a node")
	}
}

// TestWalkPathAppend checks that an append to a yielded path does not change
// the path of a later node.
func TestWalkPathAppend(t *testing.T) {
	var got []string
	for path, n := range tree.WalkPath(sample()) {
		_ = append(path, leaf("x")) // Must not write into the reused path.
		got = append(got, names(path)+" / "+name(n))
	}
	if !slices.Equal(got, samplePaths) {
		t.Errorf("WalkPath after append:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(samplePaths, "\n"))
	}
}

func TestAll(t *testing.T) {
	var leaves []string
	for l := range tree.All[leaf](sample()) {
		leaves = append(leaves, string(l))
	}
	if want := []string{"1", "2", "3", "4"}; !slices.Equal(leaves, want) {
		t.Errorf("All[leaf]: %q, want %q", leaves, want)
	}

	var branches []string
	for b := range tree.All[*branch](sample()) {
		branches = append(branches, b.name)
	}
	if want := []string{"a", "b", "c", "d"}; !slices.Equal(branches, want) {
		t.Errorf("All[*branch]: %q, want %q", branches, want)
	}

	// A type that no node has yields nothing.
	if n := len(slices.Collect(tree.All[int](sample()))); n != 0 {
		t.Errorf("All[int]: %d nodes, want 0", n)
	}
}

func TestFind(t *testing.T) {
	n, ok := tree.Find(sample(), func(n node) bool { return name(n) == "c" })
	if !ok || name(n) != "c" {
		t.Errorf("Find c: %v, %v", n, ok)
	}

	n, ok = tree.Find(sample(), func(node) bool { return false })
	if ok || n != nil {
		t.Errorf("Find with no match: %v, %v, want nil, false", n, ok)
	}
}

// TestEarlyStop checks that a break in a range loop stops each traversal: the
// iterator does not call yield again, and Find stops at the first match.
func TestEarlyStop(t *testing.T) {
	count := 0
	for range tree.Walk(sample()) {
		count++
		if count == 3 {
			break
		}
	}
	if count != 3 {
		t.Errorf("Walk: %d nodes after break, want 3", count)
	}

	count = 0
	for range tree.WalkPath(sample()) {
		count++
		if count == 5 {
			break
		}
	}
	if count != 5 {
		t.Errorf("WalkPath: %d nodes after break, want 5", count)
	}

	count = 0
	for range tree.All[leaf](sample()) {
		count++
		break
	}
	if count != 1 {
		t.Errorf("All: %d nodes after break, want 1", count)
	}

	visited := 0
	tree.Find(sample(), func(n node) bool {
		visited++
		return name(n) == "2"
	})
	if visited != 4 {
		t.Errorf("Find visited %d nodes, want 4", visited)
	}
}

// TestDeepTree walks a tree 100,000 levels deep.
func TestDeepTree(t *testing.T) {
	const depth = 100_000

	var root node = leaf("bottom")
	for i := range depth {
		root = br(strconv.Itoa(i), root)
	}

	if n := len(slices.Collect(tree.Walk(root))); n != depth+1 {
		t.Errorf("Walk: %d nodes, want %d", n, depth+1)
	}

	deepest := 0
	for path := range tree.WalkPath(root) {
		deepest = max(deepest, len(path))
	}
	if deepest != depth {
		t.Errorf("WalkPath: longest path %d, want %d", deepest, depth)
	}

	if n, ok := tree.Find(root, func(n node) bool { _, ok := n.(leaf); return ok }); !ok || name(n) != "bottom" {
		t.Errorf("Find leaf: %v, %v", n, ok)
	}
	if n := len(slices.Collect(tree.All[*branch](root))); n != depth {
		t.Errorf("All[*branch]: %d nodes, want %d", n, depth)
	}
}

// TestStandardLibraryOnly fails when a non-test file of package tree imports
// a package outside the standard library.
func TestStandardLibraryOnly(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			pkg, err := build.Default.Import(path, ".", build.FindOnly)
			if err != nil || !pkg.Goroot {
				t.Errorf("%s imports %q, which is not in the standard library", file, path)
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no non-test file found")
	}
}
