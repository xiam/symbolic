package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/xiam/symbolic/ast"
	"github.com/xiam/symbolic/parser"
)

// printTree prints node and its children as XML-like text, one node per line.
func printTree(node ast.Node, level int) {
	indent := strings.Repeat("  ", level)

	// A leaf prints its value between its tags.
	leaf := func(tag string, value any) {
		fmt.Printf("%s<%s>%v</%s>\n", indent, tag, value, tag)
	}
	// A container prints its items between its tags.
	container := func(tag string, items []ast.Node) {
		fmt.Printf("%s<%s>\n", indent, tag)
		for _, item := range items {
			printTree(item, level+1)
		}
		fmt.Printf("%s</%s>\n", indent, tag)
	}

	switch n := node.(type) {
	case *ast.Int:
		leaf("int", n.Value)
	case *ast.Float:
		leaf("float", n.Value)
	case *ast.String:
		leaf("string", n.Value)
	case *ast.Symbol:
		leaf("symbol", n.Name)
	case *ast.Atom:
		leaf("atom", n.Name)
	case *ast.Expr:
		container("expression", n.Items)
	case *ast.List:
		container("list", n.Items)
	case *ast.Map:
		container("map", n.Items)
	case *ast.File:
		// The root prints as a list, as the synthetic root list did before.
		container("list", n.Items)
	}
}

func main() {
	input := `(fn_a (fn_b [89 :A :B [67 3.27]]) (fn_c 66 3 53 "Hello world!" 😊))`

	root, err := parser.Parse([]byte(input))
	if err != nil {
		log.Fatal("parser.Parse:", err)
	}

	printTree(root, 0)
}
