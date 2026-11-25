// Command calc evaluates a fixed list of arithmetic expressions. For each one
// it prints the expression, its value or its error, and the tree with spans.
// Its output is the reference file _example/reference/calc/calc.txt:
//
//	go run ./_example/calc/cmd/calc | diff _example/reference/calc/calc.txt -
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xiam/symbolic/_example/calc"
)

var expressions = []string{
	"2 + 3 * 4",
	"1 - 2 - 3",
	"-(2 + 3) * 4",
	"7 / 2 / 7",
	"0.5 * (3.25 - -0.75)",
	"(1 +\n  2) *\n  -3",
	"1 / (2 - 2)",
	"(1 + 2",
	"2 * * 3",
	"4 +",
}

func main() {
	for i, src := range expressions {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%q\n", src)

		root, err := calc.Parse(src)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}

		if v, err := calc.Eval(root); err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Println("=", v)
		}

		if err := calc.Dump(os.Stdout, root); err != nil {
			log.Fatal("calc.Dump: ", err)
		}
	}
}
