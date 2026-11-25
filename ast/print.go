package ast

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Fprint writes the tree under n back as S-expression text. Items are
// separated by one space; whitespace and comments of the source are not
// kept. The items of a File are written without brackets.
func Fprint(w io.Writer, n Node) error {
	p := printer{w: w}
	p.node(n)
	return p.err
}

// Encode returns the text that Fprint writes for n.
func Encode(n Node) []byte {
	var buf bytes.Buffer
	_ = Fprint(&buf, n) // A bytes.Buffer does not fail.
	return buf.Bytes()
}

// printer writes nodes to w and keeps the first write error.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) print(s string) {
	if p.err == nil {
		_, p.err = io.WriteString(p.w, s)
	}
}

func (p *printer) node(n Node) {
	switch n := n.(type) {
	case *Int:
		p.print(strconv.FormatInt(n.Value, 10))
	case *Float:
		p.print(strconv.FormatFloat(n.Value, 'g', -1, 64))
	case *String:
		p.print(strconv.Quote(n.Value))
	case *Symbol:
		p.print(n.Name)
	case *Atom:
		p.print(n.Name)
	case *Expr:
		p.items("(", n.Items, ")")
	case *List:
		p.items("[", n.Items, "]")
	case *Map:
		p.items("{", n.Items, "}")
	case *File:
		p.items("", n.Items, "")
	}
}

func (p *printer) items(open string, items []Node, close string) {
	p.print(open)
	for i, item := range items {
		if i > 0 {
			p.print(" ")
		}
		p.node(item)
	}
	p.print(close)
}

// Dump writes one line per node of the tree under root, for debugging: the
// kind, the value of a leaf and the span, indented by depth. For the input
// `(a 1)` it writes:
//
//	File 1:0>1:5
//	  Expr 1:0>1:5
//	    Symbol a 1:1>1:2
//	    Int 1 1:3>1:4
func Dump(w io.Writer, root Node) error {
	for path, n := range WalkPath(root) {
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

// describe returns the kind of n and, for a leaf, its value as text.
func describe(n Node) (kind, value string) {
	switch n := n.(type) {
	case *Int:
		return "Int", strconv.FormatInt(n.Value, 10)
	case *Float:
		return "Float", strconv.FormatFloat(n.Value, 'g', -1, 64)
	case *String:
		return "String", strconv.Quote(n.Value)
	case *Symbol:
		return "Symbol", n.Name
	case *Atom:
		return "Atom", n.Name
	case *Expr:
		return "Expr", ""
	case *List:
		return "List", ""
	case *Map:
		return "Map", ""
	case *File:
		return "File", ""
	}
	return fmt.Sprintf("%T", n), ""
}
