package ast_test

import (
	"fmt"
	"log"
	"os"

	"github.com/xiam/symbolic/ast"
	"github.com/xiam/symbolic/parser"
)

// ExampleAll prints every symbol of a tree with its position. The README
// shows this example.
func ExampleAll() {
	root, err := parser.Parse([]byte(`(fn_a (fn_b [89 :A]) (fn_c "hi" 3.27))`))
	if err != nil {
		log.Fatal(err)
	}

	for sym := range ast.All[*ast.Symbol](root) {
		fmt.Println(sym.Name, "at", sym.Span().Start())
	}
	// Output:
	// fn_a at 1:1
	// fn_b at 1:7
	// fn_c at 1:22
}

// ExampleWalk sums the integers of a tree with a type switch.
func ExampleWalk() {
	root, err := parser.Parse([]byte(`(+ 1 [2 3] {:a 4})`))
	if err != nil {
		log.Fatal(err)
	}

	var sum int64
	for n := range ast.Walk(root) {
		switch n := n.(type) {
		case *ast.Int:
			sum += n.Value
		}
	}
	fmt.Println(sum)
	// Output:
	// 10
}

// ExampleMap_Pairs reads the keys and values of a map.
func ExampleMap_Pairs() {
	root, err := parser.Parse([]byte(`{:a 1 :b "two"}`))
	if err != nil {
		log.Fatal(err)
	}

	m := root.Items[0].(*ast.Map)
	for k, v := range m.Pairs() {
		fmt.Printf("%s = %s\n", ast.Encode(k), ast.Encode(v))
	}
	// Output:
	// :a = 1
	// :b = "two"
}

// ExampleDump writes one line per node, with its span.
func ExampleDump() {
	root, err := parser.Parse([]byte(`(a "b" 1)`))
	if err != nil {
		log.Fatal(err)
	}

	if err := ast.Dump(os.Stdout, root); err != nil {
		log.Fatal(err)
	}
	// Output:
	// File 1:0>1:9
	//   Expr 1:0>1:9
	//     Symbol a 1:1>1:2
	//     String "b" 1:3>1:6
	//     Int 1 1:7>1:8
}
