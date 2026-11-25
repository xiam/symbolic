# Example reference outputs at 9d7e25c

These files hold the output of the three programs under `_example/` at commit
`9d7e25c`, the last commit before the port onto bytescan.

- `parser.txt`: `go run ./_example/parser`, unchanged source.
- `xml_format.txt`: `go run ./_example/xml_format`, unchanged source.
- `lexer.txt`: `go run ./_example/lexer`, with one correction that makes it
  compile. At `9d7e25c`, line 24 reads `line, col := tok.Pos()`, but `Pos`
  returns one `text/scanner.Position`. The recorded run used
  `line, col := tok.Pos().Line, tok.Pos().Column`.

Compare a program with its reference:

```sh
go run ./_example/parser | diff _example/reference/9d7e25c/parser.txt -
```
