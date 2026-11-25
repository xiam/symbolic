# Building an AST

> How to build a typed abstract syntax tree for a grammar on
> [bytescan](https://github.com/xiam/bytescan) and package `tree`, in five
> steps.

This repository has two ASTs that follow the same five steps:

- **The S-expression AST**: the rules in [`lexer`](../lexer), the nodes in
  [`ast`](../ast) and the parser in [`parser`](../parser).
- **The calculator**: an arithmetic grammar with numbers, `+ - * /` and
  parentheses, all in one package, [`_example/calc`](../_example/calc).

Each code block below is copied from a file of this repository, and the text
above the block names that file. The tests build every one of these files.

The five steps:

1. [The rules](#step-1-the-rules): split the input into tokens.
2. [The node types](#step-2-the-node-types): one struct type per kind of
   node.
3. [Children](#step-3-children): each node lists its direct children.
4. [The parser](#step-4-the-parser): build the tree from the tokens.
5. [Walking the tree](#step-5-walking-the-tree): package `tree` does it for
   you.

## Step 1: the rules

A bytescan scanner splits text into tokens. Give it one rule per kind of
token. Each rule has a token type, which is a name that you choose.

The calculator has a table of rules, in
[`_example/calc/lexer.go`](../_example/calc/lexer.go):

```go
var rules = []struct {
	typ  bytescan.TokenType
	rule func() bytescan.Rule
}{
	// A number is an integer, such as 42, or a decimal, such as 3.25. Choice
	// takes the longer match, so "3.25" is one decimal and not the integer 3.
	// A sign is not part of a number: "-3" is a unary minus and a number.
	{tokenNumber, func() bytescan.Rule { return core.Choice(base.Decimal(), base.Integer()) }},
	{tokenPlus, func() bytescan.Rule { return core.Literal("+") }},
	{tokenMinus, func() bytescan.Rule { return core.Literal("-") }},
	{tokenStar, func() bytescan.Rule { return core.Literal("*") }},
	{tokenSlash, func() bytescan.Rule { return core.Literal("/") }},
	{tokenLParen, func() bytescan.Rule { return core.Literal("(") }},
	{tokenRParen, func() bytescan.Rule { return core.Literal(")") }},
	{tokenSpace, base.Whitespace},
	{tokenNewline, base.Newline},
}
```

The scanner takes the longest match. Whitespace and line breaks are tokens
too: the parser skips them.

The S-expression rules are in [`lexer/rules.go`](../lexer/rules.go). They
build their own rules from the `core` package of bytescan, such as
`core.RuneRun` for a word.

## Step 2: the node types

Write one struct type for each kind of node, and one interface that all of
them satisfy. The interface has two methods that every node needs, and one
unexported method that keeps other packages from adding node types. From
[`ast/node.go`](../ast/node.go):

```go
type Node interface {
	// Span returns the half-open range of the source text of the node.
	Span() cursor.Span

	// Children returns the direct children of the node, in source order. A
	// leaf has no children.
	Children() iter.Seq[Node]

	// node seals the interface: no type outside this package implements
	// Node.
	node()
}
```

A node type holds what the grammar needs, with Go types. An S-expression
container holds a slice of nodes, from [`ast/node.go`](../ast/node.go):

```go
type Expr struct {
	Items []Node
	Loc   cursor.Span
}
```

A calculator operation holds its two operands in named fields, from
[`_example/calc/node.go`](../_example/calc/node.go):

```go
type Binary struct {
	Op    Op
	OpPos cursor.Pos
	X, Y  Node
	Loc   cursor.Span
}
```

Notes:

- The span field is named `Loc`, because the method `Span` takes the name
  `Span`.
- A span is often not enough. The calculator reports a division by zero at
  the operator, so `Binary` keeps the position of the operator in `OpPos`.
- A consumer reads a tree with a type switch on the node types. The
  calculator's `Eval` in [`_example/calc/eval.go`](../_example/calc/eval.go)
  is one type switch.

Each node type has three one-line methods. From
[`_example/calc/node.go`](../_example/calc/node.go):

```go
func (n *Number) Span() cursor.Span { return n.Loc }
func (n *Paren) Span() cursor.Span  { return n.Loc }
func (n *Unary) Span() cursor.Span  { return n.Loc }
func (n *Binary) Span() cursor.Span { return n.Loc }
```

## Step 3: Children

`Children` returns the direct children of a node as an `iter.Seq`, in source
order. How you write it depends on the shape of the node.

A leaf has no children, and a node with a slice returns the slice. From
[`ast/node.go`](../ast/node.go):

```go
// noChildren is the empty sequence of the children of a leaf.
func noChildren(func(Node) bool) {}

func (n *Int) Children() iter.Seq[Node]    { return noChildren }
```

```go
func (n *Expr) Children() iter.Seq[Node]   { return slices.Values(n.Items) }
```

A node with fixed fields yields each field. Stop when `yield` returns
false: then a `break` in the caller's loop stops at once. From
[`_example/calc/node.go`](../_example/calc/node.go):

```go
func (n *Binary) Children() iter.Seq[Node] {
	return func(yield func(Node) bool) {
		_ = yield(n.X) && yield(n.Y)
	}
}
```

`Children` is the only method that package `tree` uses.

## Step 4: the parser

The parser reads the tokens and builds the nodes. It also sets the span of
each node.

The calculator is a recursive descent parser: one function per level of
precedence. The loop of one level builds a `Binary` from left to right, so
`1 - 2 - 3` is `(1 - 2) - 3`. From
[`_example/calc/parser.go`](../_example/calc/parser.go):

```go
	for {
		op, ok := p.op(ops...)
		if !ok {
			return x, nil
		}
		opPos := p.tok.Span().Start()
		if err := p.advance(); err != nil {
			return nil, err
		}
		y, err := operand()
		if err != nil {
			return nil, err
		}
		span := cursor.NewSpan(x.Span().Start(), y.Span().End())
		x = &Binary{Op: op, OpPos: opPos, X: x, Y: y, Loc: span}
	}
```

A bytescan token has a span, so a leaf takes the span of its token, and an
inner node spans from its first token to its last.

The S-expression parser in [`parser/parser.go`](../parser/parser.go) is a
different kind of parser: a set of state functions. Each container state
reads items into the `Items` of its node until the closing bracket. Its spans
follow the same rule: its lexer keeps the span of each bytescan token, and a
node spans from the start of its first token to the end of its last.

Each grammar has its own error type with the position of the error: see
`*calc.Error` in [`_example/calc/errors.go`](../_example/calc/errors.go) and
`*parser.Error` in [`parser/errors.go`](../parser/errors.go).

## Step 5: walking the tree

Do not write the walk. Package [`tree`](../tree) walks every node type that
has a method `Children`. Its constraint, from
[`tree/tree.go`](../tree/tree.go):

```go
type Node[N any] interface {
	Children() iter.Seq[N]
}
```

Both `ast.Node` and `calc.Node` satisfy it. Package `tree` has four
functions. The first three return an iterator, and a `break` in the loop
stops the walk:

- `tree.Walk(root)`: every node, depth first, parents before children;
- `tree.WalkPath(root)`: the same nodes, each with the path of its
  ancestors;
- `tree.All[T](root)`: the nodes of one type `T`;
- `tree.Find(root, match)`: the first node that matches.

The calculator prints a tree with `tree.WalkPath`: the length of the path is
the depth of the node. From [`_example/calc/dump.go`](../_example/calc/dump.go):

```go
	for path, n := range tree.WalkPath(root) {
		kind, value := describe(n)
		if value != "" {
			value = " " + value
		}
		indent := strings.Repeat("  ", len(path))
```

`tree.All` takes the node type that you want, and Go infers the rest from
the root. From
[`_example/calc/example_test.go`](../_example/calc/example_test.go):

```go
	for n := range tree.All[*calc.Binary](root) {
		fmt.Println(n.Op, n.Span())
	}
```

A grammar package can also keep its own names for the walk, so that its
users do not need to know package `tree`. Package `ast` does this with
one-line functions, from [`ast/walk.go`](../ast/walk.go):

```go
func All[T Node](root Node) iter.Seq[T] { return tree.All[T](root) }
```

## What `tree` does not do

Package `tree` only reads a tree. It does not:

- change a tree, or rebuild it with new children;
- call a visitor interface: you write a `for` loop and a type switch;
- treat the tree as a graph: it keeps no set of visited nodes. A node that
  two parents share is visited twice, and a cycle never ends;
- know spans, errors, kind names or printing: those belong to each grammar.

The package documentation is at
[pkg.go.dev/github.com/xiam/symbolic/tree](https://pkg.go.dev/github.com/xiam/symbolic/tree).
