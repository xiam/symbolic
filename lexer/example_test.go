package lexer_test

import (
	"fmt"
	"io"
	"strings"

	"github.com/xiam/bytescan"
	"github.com/xiam/symbolic/lexer"
)

// The README shows this example.
func ExampleAddRules() {
	sc := bytescan.New(strings.NewReader("(+ -12 3)"))
	if err := lexer.AddRules(sc); err != nil {
		panic(err)
	}

	for {
		tok, err := sc.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Printf("%d:%d %s %q\n", tok.Line(), tok.Column(), tok.Type(), tok.Text())
	}
	// Output:
	// 1:0 open_expression "("
	// 1:1 sequence "+"
	// 1:2 separator " "
	// 1:3 integer "-12"
	// 1:6 separator " "
	// 1:7 integer "3"
	// 1:8 close_expression ")"
}
