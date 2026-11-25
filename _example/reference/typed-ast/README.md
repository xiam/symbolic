# Example reference outputs of the typed AST

These files hold the output of the programs under `_example/` whose output
changed with the typed AST of issue #2.

- `parser.txt`: `go run ./_example/parser`. The program prints the tree with
  `ast.Dump`: one line per node, with its kind, its value and its span.

The other programs print what they printed before:
`go run ./_example/xml_format` prints `../9d7e25c/xml_format.txt`, and
`go run ./_example/lexer` prints what it printed at `a99f386`.

Compare a program with its reference:

```sh
go run ./_example/parser | diff _example/reference/typed-ast/parser.txt -
```
