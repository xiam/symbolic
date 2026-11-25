package calc

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xiam/symbolic/tree"
)

// Dump writes one line per node of the tree under root, for debugging: the
// kind, the value of a number or the operator, and the span, indented by
// depth. For the input `-(1 + 2)` it writes:
//
//	Unary - 1:0>1:8
//	  Paren 1:1>1:8
//	    Binary + 1:2>1:7
//	      Number 1 1:2>1:3
//	      Number 2 1:6>1:7
func Dump(w io.Writer, root Node) error {
	for path, n := range tree.WalkPath(root) {
		kind, value := describe(n)
		if value != "" {
			value = " " + value
		}
		indent := strings.Repeat("  ", len(path))
		if _, err := fmt.Fprintf(w, "%s%s%s %v\n", indent, kind, value, n.Span()); err != nil {
			return err
		}
	}
	return nil
}

// describe returns the kind of n and its value or operator as text.
func describe(n Node) (kind, value string) {
	switch n := n.(type) {
	case *Number:
		return "Number", strconv.FormatFloat(n.Value, 'g', -1, 64)
	case *Paren:
		return "Paren", ""
	case *Unary:
		return "Unary", n.Op.String()
	case *Binary:
		return "Binary", n.Op.String()
	}
	return fmt.Sprintf("%T", n), ""
}
