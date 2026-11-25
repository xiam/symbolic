package calc_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/xiam/symbolic/_example/calc"
	"github.com/xiam/symbolic/tree"
)

func Example() {
	root, err := calc.Parse("-(1 + 2) * 4")
	if err != nil {
		panic(err)
	}

	v, err := calc.Eval(root)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

	if err := calc.Dump(os.Stdout, root); err != nil {
		panic(err)
	}
	// Output:
	// -12
	// Binary * 1:0>1:12
	//   Unary - 1:0>1:8
	//     Paren 1:1>1:8
	//       Binary + 1:2>1:7
	//         Number 1 1:2>1:3
	//         Number 2 1:6>1:7
	//   Number 4 1:11>1:12
}

func ExampleError() {
	_, err := calc.Parse("2 * (3 + 4")

	var e *calc.Error
	if errors.As(err, &e) && errors.Is(err, calc.ErrUnclosedParen) {
		fmt.Println("line", e.Pos.Line(), "column", e.Pos.Column())
	}
	fmt.Println(err)
	// Output:
	// line 1 column 4
	// 1:4: unclosed parenthesis
}

func Example_all() {
	root, err := calc.Parse("1 + 2 * 3")
	if err != nil {
		panic(err)
	}
	for n := range tree.All[*calc.Binary](root) {
		fmt.Println(n.Op, n.Span())
	}
	// Output:
	// + 1:0>1:9
	// * 1:4>1:9
}
