# Changelog

All notable changes to this project are recorded in this file. The project
follows [Semantic Versioning](https://semver.org/). While the major version is
0, a minor release can change the API.

## v0.1.0 - 2026-11-25

First tagged release. The module path is now `github.com/xiam/symbolic`, the
lexer is built on bytescan, and the parser returns a typed syntax tree.

### Upgrade notes

- The module path changed from `github.com/xiam/s-expr` to
  `github.com/xiam/symbolic`. Code that imports the old path keeps the last
  untagged version of the old module. To upgrade, change the import paths and
  run `go get github.com/xiam/symbolic@v0.1.0`.
- The minimum Go version is 1.23.
- This release changes the API of the lexer, the parser and the syntax tree.
  The sections below list the changes.

### Lexer

- The lexer is built on [bytescan](https://github.com/xiam/bytescan) v0.6.0.
  Each kind of token is a bytescan rule. `lexer.Rules` returns the rules and
  `lexer.AddRules` registers them on a `bytescan.Scanner`.
- The lexer is a pull lexer: call `Next`, then `Token`. `Scan` and `Stop` are
  removed, and the lexer starts no goroutine.
- A token keeps the whole span of its text. `Token.Span` returns it, and
  `Token.Pos` returns its start as a bytescan `cursor.Pos`.
- A signed number such as `-12` is one integer token.

### Syntax tree

- Package `ast` has one type for each kind of node: `Int`, `Float`, `String`,
  `Symbol`, `Atom`, `Expr`, `List`, `Map` and `File`. The fields are exported,
  and every node has a span.
- `ast.Node` is a sealed interface: only package `ast` defines nodes.
- `ast.Walk`, `ast.WalkPath`, `ast.All` and `ast.Find` traverse a tree with Go
  iterators.
- `ast.Fprint` and `ast.Encode` write a tree as S-expression text, and
  `ast.Dump` writes one line for each node.

### Parser

- `parser.Parse` returns an `*ast.File`.
- A parse failure is a `*parser.Error` with the position of the error. Use
  `errors.Is` with `parser.ErrUnexpectedEOF`, `parser.ErrUnexpectedToken` or
  `parser.ErrLexer` to test the cause.
- A lexer failure, such as an error of the reader, is a `*parser.Error` that
  wraps `parser.ErrLexer` and the cause.

### New package: tree

- Package `tree` has the traversal functions `Walk`, `WalkPath`, `All` and
  `Find` for the syntax tree of any grammar. A node type needs one method,
  `Children`.

### Documentation and examples

- [Building an AST](docs/building-an-ast.md) is a guide to building a typed
  syntax tree for your own grammar.
- `_example/calc` is a calculator with its own lexer, parser and typed syntax
  tree, built with the same method.
