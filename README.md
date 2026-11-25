# symbolic [![GoDoc](https://godoc.org/github.com/xiam/symbolic?status.svg)](https://godoc.org/github.com/xiam/symbolic)

<a href="https://github.com/xiam/symbolic/actions?query=workflow%3Aunit-tests"><img alt="xiam/symbolic unit tests status" src="https://github.com/xiam/symbolic/workflows/unit-tests/badge.svg"></a>

`symbolic` is a Go library that provides a general purpose
[S-expression][1] parser and a lexer. The lexer scans a stream of bytes for a
finite set of patterns (tokens) and identifies lexemes, the parser takes those
lexemes as input and builds abstract syntax trees ([AST][2]) with them.

## Install

```
go get github.com/xiam/symbolic@latest
```

`symbolic` needs Go 1.23 or later. The module path was
`github.com/xiam/s-expr` before v0.1.0; see the [changelog](CHANGELOG.md) for
the upgrade notes.

## What's this for?

ASTs can be used for several purposes:

* Configuration files (like Emacs).
* The front end of a small Lisp-like language: parse the code, then walk the
  tree to run it.
* Representation of lambda terms.
* Build [transformational grammar](https://en.wikipedia.org/wiki/Transformational_grammar) graphs.
* ...or any expression that can be represented as a tree.

`symbolic` is also a worked example of how to build a typed syntax tree for
any grammar. To build one for your own grammar, see
[Building an AST](docs/building-an-ast.md).

### Example

The following *bracketed expression*:

```
[S [N John] [VP [V hit] [NP [D the] [N ball]]]]
```

that corresponds to this [constituency-based parse tree][5]:

![Constituency-based parse trees](https://upload.wikimedia.org/wikipedia/commons/5/54/Parse_tree_1.jpg)

could be processed by `symbolic`, converted into an AST for easy manipulation and
then transformed into something else, like XML:

```xml
<list>
  <list>
    <symbol>S</symbol>
    <list>
      <symbol>N</symbol>
      <symbol>John</symbol>
    </list>
    <list>
      <symbol>VP</symbol>
      <list>
        <symbol>V</symbol>
        <symbol>hit</symbol>
      </list>
      <list>
        <symbol>NP</symbol>
        <list>
          <symbol>D</symbol>
          <symbol>the</symbol>
        </list>
        <list>
          <symbol>N</symbol>
          <symbol>ball</symbol>
        </list>
      </list>
    </list>
  </list>
</list>
```

## Built on bytescan

The lexer is built on [bytescan](https://github.com/xiam/bytescan), a library
that splits a stream of text into labeled pieces. Each kind of token is a
bytescan rule. `lexer.Rules` returns the rules, and `lexer.AddRules`
registers them on a `bytescan.Scanner`, so you can also use the rules
directly:

```go
sc := bytescan.New(strings.NewReader("(+ -12 3)"))
if err := lexer.AddRules(sc); err != nil {
  // ...
}

for {
  tok, err := sc.Next()
  if err == io.EOF {
    break
  }
  if err != nil {
    // ...
  }
  fmt.Printf("%d:%d %s %q\n", tok.Line(), tok.Column(), tok.Type(), tok.Text())
}
```

Output:

```
1:0 open_expression "("
1:1 sequence "+"
1:2 separator " "
1:3 integer "-12"
1:6 separator " "
1:7 integer "3"
1:8 close_expression ")"
```

Positions follow bytescan: lines start at 1 and columns start at 0. A column
counts runes, not bytes.

## Subpackages

### Lexer

When a stream of bytes matches any of the following patterns:

`(`, `)`, `[`, `]`, `{`, `}`, `"`, `#`, `[a-zA-Z_]`, `:`, `.`, `\`

a new instance of that token is created and recorded along with the line and
column where the match was found.

#### Example

The following example converts a byte slice into an slice of tokens and prints
details about each one of these tokens:

```go
package main

import (
  ...

  "github.com/xiam/symbolic/lexer"
)

func main() {
  input := `
    (fn_a # comment
      (fn_b [89 :A :B [67 3.27]])
      (fn_c 66 3 53 "Hello world!")
    )
  `

  tokens, err := lexer.Tokenize([]byte(input))
  if err != nil {
    // ...
  }

  for i, tok := range tokens {
    line, col := tok.Pos().Line(), tok.Pos().Column()
    lexeme := tok.Text()
    tt := tok.Type().String()

    // Example output:
    // token[1] (type: separator, line: 2, col: 0)
    //   -> "    "
    fmt.Printf("token[%d] (type: %v, line: %d, col: %d)\n\t-> %q\n\n", i, tt, line, col, lexeme)
  }
}
```

For more information see the [list of token types][3].

### Parser

The parser reads the tokens of the lexer and builds an [AST][2]. The root of
the tree is an `*ast.File`, which holds the top-level items of the input. Each
kind of node has its own type in package `ast`:

| Pattern                    | Type          | Fields          | Example   |
|----------------------------|---------------|-----------------|-----------|
| `(`...`)`                  | `*ast.Expr`   | `Items []Node`  | `(fib 1)` |
| `[`...`]`                  | `*ast.List`   | `Items []Node`  | `[1 2 3]` |
| `{`...`}`                  | `*ast.Map`    | `Items []Node`  | `{:A 65}` |
| `[+-]?[0-9]+`              | `*ast.Int`    | `Value int64`   | `-123`    |
| `[+-]?[0-9]+\.[0-9]+`      | `*ast.Float`  | `Value float64` | `1.23`    |
| `[a-zA-Z][a-zA-Z0-9_]+`    | `*ast.Symbol` | `Name string`   | `hello`   |
| `:[a-zA-Z][a-zA-Z0-9_]+`   | `*ast.Atom`   | `Name string`   | `:hello`  |
| `"`...`"`                  | `*ast.String` | `Value string`  | `"hello"` |

A map is a flat sequence of keys and values; `Map.Pairs` reads it as pairs.
Every node satisfies the interface `ast.Node`, which only package `ast`
implements:

```go
type Node interface {
  Span() cursor.Span        // the half-open range of the source text
  Children() iter.Seq[Node] // the direct children; a leaf has none
  // and one unexported method
}
```

`Span` returns a [bytescan][6] `cursor.Span`: lines start at 1 and columns
start at 0. A string spans its double quotes, and a container spans its
brackets.

#### Example

```go
package main

import (
  ...

  "github.com/xiam/symbolic/ast"
  "github.com/xiam/symbolic/parser"
)

func main() {
  input := `(fn_a
    (fn_b [89 :A :B [67 3.27]])
    (fn_c 66 3 53 "Hello world!" 😊))
  `

  root, err := parser.Parse([]byte(input))
  if err != nil {
    log.Fatal("parser.Parse:", err)
  }

  // One line per node: its kind, its value and its span.
  if err := ast.Dump(os.Stdout, root); err != nil {
    log.Fatal("ast.Dump:", err)
  }
}
```

A syntax error is a `*parser.Error` with the position of the error and, for
an unclosed bracket or string, the position of the opener. `errors.Is` tells
`parser.ErrUnexpectedEOF` from `parser.ErrUnexpectedToken`. An error of the
lexer, such as an error of the reader, is a `*parser.Error` too: its cause
wraps `parser.ErrLexer` and the error of the lexer.

See the [documentation of package ast][4].

### Tree

Package [`tree`](tree) walks the syntax tree of any grammar, not only
S-expressions. A grammar gives each node one method, `Children`, and gets
`Walk`, `WalkPath`, `All` and `Find`. The [calculator](_example/calc) uses it
for its own tree. See [AST traversal](#ast-traversal) below for an example.

## AST

The following byte stream:

```lisp
(fn_a (fn_b [89 :A :B [67 3.27]]) (fn_c 66 3 53 "Hello world!"))
```

gets transformed into an AST that looks like this:

![AST](_res/ast.png)

`ast.Encode` writes a tree back as text, and `ast.Dump` writes one line per
node, for debugging.

### AST traversal

Package `ast` walks a tree with range-over-func iterators:

* `ast.Walk(root)`: every node, depth first, parents before children;
* `ast.WalkPath(root)`: the same nodes, each with the path of its ancestors;
* `ast.All[T](root)`: the nodes of one type;
* `ast.Find(root, match)`: the first node that matches.

These four functions call package [`tree`](tree), which walks the tree of
any grammar whose nodes have a method `Children`. A `break` stops the walk.
The following example prints every symbol of a tree with its position:

```go
root, err := parser.Parse([]byte(`(fn_a (fn_b [89 :A]) (fn_c "hi" 3.27))`))
if err != nil {
  log.Fatal(err)
}

for sym := range ast.All[*ast.Symbol](root) {
  fmt.Println(sym.Name, "at", sym.Span().Start())
}
```

Output:

```
fn_a at 1:1
fn_b at 1:7
fn_c at 1:22
```

A type switch reads the fields of each kind of node. The
[AST to XML](_example/xml_format/xml_format_example.go) example prints the
tree above as XML-like text:

```xml
<list>
  <expression>
    <symbol>fn_a</symbol>
    <expression>
      <symbol>fn_b</symbol>
      <list>
        <int>89</int>
        <atom>:A</atom>
        <atom>:B</atom>
        <list>
          <int>67</int>
          <float>3.27</float>
        </list>
      </list>
    </expression>
    <expression>
      <symbol>fn_c</symbol>
      <int>66</int>
      <int>3</int>
      <int>53</int>
      <string>Hello world!</string>
    </expression>
  </expression>
</list>
```

## Examples

* [Lexer](_example/lexer/lexer_example.go)
* [Parser](_example/parser/parser_example.go)
* [AST to XML](_example/xml_format/xml_format_example.go)
* [Calculator](_example/calc): a second grammar with its own typed AST

## Build an AST for your own grammar

The guide [Building an AST](docs/building-an-ast.md) shows, in five steps, how
to build a typed syntax tree on bytescan and package `tree`, with the
S-expression AST and the calculator as examples.

[1]: https://en.wikipedia.org/wiki/S-expression
[2]: https://en.wikipedia.org/wiki/Abstract_syntax_tree
[3]: https://godoc.org/github.com/xiam/symbolic/lexer#TokenType
[4]: https://pkg.go.dev/github.com/xiam/symbolic/ast
[5]: https://en.wikipedia.org/wiki/Parse_tree#Constituency-based_parse_trees
[6]: https://github.com/xiam/bytescan
