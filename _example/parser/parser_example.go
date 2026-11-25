package main

import (
	"log"
	"os"

	"github.com/xiam/symbolic/ast"
	"github.com/xiam/symbolic/parser"
)

func main() {
	input := `(fn_a (fn_b [89 :A :B [67 3.27]]) (fn_c 66 3 53 "Hello world!" 😊))`

	root, err := parser.Parse([]byte(input))
	if err != nil {
		log.Fatal("parser.Parse:", err)
	}

	if err := ast.Dump(os.Stdout, root); err != nil {
		log.Fatal("ast.Dump:", err)
	}
}
