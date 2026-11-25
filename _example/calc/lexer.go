package calc

import (
	"io"

	"github.com/xiam/bytescan"
	"github.com/xiam/bytescan/cursor"
	"github.com/xiam/bytescan/rules/base"
	"github.com/xiam/bytescan/rules/core"
)

// The token types of the calculator's rule set.
const (
	tokenNumber  bytescan.TokenType = "number"
	tokenPlus    bytescan.TokenType = "plus"
	tokenMinus   bytescan.TokenType = "minus"
	tokenStar    bytescan.TokenType = "star"
	tokenSlash   bytescan.TokenType = "slash"
	tokenLParen  bytescan.TokenType = "lparen"
	tokenRParen  bytescan.TokenType = "rparen"
	tokenSpace   bytescan.TokenType = "space"
	tokenNewline bytescan.TokenType = "newline"
)

// rules is the rule set of the lexer, in the order that the scanner registers
// it. The scanner takes the longest match, and on a tie the rule registered
// first. A rune that no rule matches becomes a one-rune token of type
// bytescan.TokenTypeUnknown, which the parser reports as unexpected.
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

// newScanner returns a scanner of src with the calculator's rule set, and the
// cursor that it reads, which gives the position of the end of input.
func newScanner(src io.Reader) (*bytescan.Scanner, *cursor.Reader) {
	cur := cursor.NewReader(src)
	sc := bytescan.NewWithCursor(cur)
	for _, r := range rules {
		// The rule set has distinct, non-empty types and no nil rule, the only
		// reasons for AddRule to fail.
		sc.MustAddRule(r.typ, r.rule())
	}
	return sc, cur
}
