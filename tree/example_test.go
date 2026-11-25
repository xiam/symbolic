package tree_test

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/xiam/symbolic/tree"
)

// Dir is a directory with files and subdirectories. Its method Children
// makes it a tree.Node[*Dir].
type Dir struct {
	Name string
	Subs []*Dir
}

func (d *Dir) Children() iter.Seq[*Dir] { return slices.Values(d.Subs) }

// Example walks a tree of a node type that is a struct pointer, not an
// interface, and prints each node indented by its depth.
func Example() {
	root := &Dir{Name: "/", Subs: []*Dir{
		{Name: "etc"},
		{Name: "usr", Subs: []*Dir{{Name: "bin"}, {Name: "lib"}}},
	}}

	for path, d := range tree.WalkPath(root) {
		fmt.Println(strings.Repeat("  ", len(path)) + d.Name)
	}
	// Output:
	// /
	//   etc
	//   usr
	//     bin
	//     lib
}
