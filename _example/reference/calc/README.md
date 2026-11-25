# Example reference output of the calculator

`calc.txt` holds the output of the calculator program of issue #3:
`go run ./_example/calc/cmd/calc`. For each expression of the program it
holds the expression, its value or its error with the position, and the tree
that `calc.Dump` writes: one line per node, with its kind, its number or
operator, and its span.

Compare the program with its reference:

```sh
go run ./_example/calc/cmd/calc | diff _example/reference/calc/calc.txt -
```
