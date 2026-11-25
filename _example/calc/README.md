# Calculator

A small arithmetic calculator with its own typed syntax tree. It is the second
example grammar of `symbolic`, next to the S-expression parser, and it is built
the same way: rules on [bytescan](https://github.com/xiam/bytescan), one Go
type per kind of node, and package [`tree`](../../tree) to walk the tree.

It reads numbers, `+ - * /`, unary `+` and `-`, and parentheses, with the
usual precedence: `2 + 3 * 4` is 14, and `1 - 2 - 3` is -4.

## Try it

Run the program. It evaluates a few expressions and prints each tree:

```sh
go run ./_example/calc/cmd/calc
```

```
"2 + 3 * 4"
= 14
Binary + 1:0>1:9
  Number 2 1:0>1:1
  Binary * 1:4>1:9
    Number 3 1:4>1:5
    Number 4 1:8>1:9
```

Each line of the tree is one node: its kind, its value or operator, and the
part of the input it covers (line:column, start and end).

Run the tests:

```sh
go test ./_example/calc/...
```

## Use it from Go

```go
root, err := calc.Parse("-(2 + 3) * 4")
if err != nil {
	log.Fatal(err) // the error names the line and column
}

v, err := calc.Eval(root)
if err != nil {
	log.Fatal(err) // for example, a division by zero
}
fmt.Println(v) // -20
```

## Where to look

- [`node.go`](node.go): the node types.
- [`lexer.go`](lexer.go) and [`parser.go`](parser.go): from text to a tree.
- [`eval.go`](eval.go): computes a value with one type switch.
- To build a grammar like this one, follow the guide
  [Building an AST](../../docs/building-an-ast.md).
