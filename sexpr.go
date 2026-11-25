// Package symbolic is a Go library that provides a general purpose
// S-expression parser and a lexer. The lexer scans a stream of bytes for a
// finite set of patterns (tokens) and identifies lexemes, the parser takes
// those lexemes as input and builds abstract syntax trees (ASTs) with them.
//
// The code lives in the subpackages: lexer, parser, ast and tree.
package symbolic
