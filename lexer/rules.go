package lexer

import (
	"github.com/xiam/bytescan"
	"github.com/xiam/bytescan/rules/core"
)

// Rule pairs a bytescan rule with the token type it produces.
type Rule struct {
	Type TokenType
	Rule bytescan.Rule
}

// Rules returns the rule set of the lexer, one rule for each token type that
// consumes input, in the order AddRules registers them. TokenEOF has no rule:
// it consumes no input, and the Lexer produces it when the scanner reports
// the end of input.
//
// Each call builds new rules. They follow these boundaries:
//
//   - one rune for each of ( ) [ ] { } " # : . \ and the newline;
//   - a run of [ \f\t\r] is a separator (TokenWhitespace);
//   - a run of [A-Za-z_] is a word;
//   - an optional sign followed by a run of [0-9] is an integer, so "-123"
//     is one integer;
//   - any other rune starts a sequence, which runs up to a separator, a
//     newline, a double quote or a bracket. A sign followed by a digit is an
//     integer, not the start of a sequence.
func Rules() []Rule {
	rules := []Rule{}
	for _, tt := range singleRuneTypes {
		rules = append(rules, Rule{tt, core.Literal(string(tokenValues[tt]))})
	}
	return append(rules,
		Rule{TokenWhitespace, core.RuneRun(isWhitespace)},
		Rule{TokenWord, core.RuneRun(isWord)},
		Rule{TokenInteger, core.Sequence(
			core.Optional(core.Rune(isArithmeticSign)),
			core.RuneRun(isInteger),
		)},
		Rule{TokenSequence, sequenceRule()},
	)
}

// AddRules registers the rule set of Rules on sc. Each bytescan token type
// is the String of the lexer token type, for example "open_list".
func AddRules(sc *bytescan.Scanner) error {
	for _, r := range Rules() {
		if err := sc.AddRule(bytescan.TokenType(r.Type.String()), r.Rule); err != nil {
			return err
		}
	}
	return nil
}

// singleRuneTypes are the token types made of exactly one fixed rune.
var singleRuneTypes = []TokenType{
	TokenOpenExpression,
	TokenCloseExpression,
	TokenOpenList,
	TokenCloseList,
	TokenOpenMap,
	TokenCloseMap,
	TokenNewLine,
	TokenDoubleQuote,
	TokenHash,
	TokenColon,
	TokenDot,
	TokenBackslash,
}

var (
	isWhitespace     = core.Class(isTokenType(TokenWhitespace))
	isWord           = core.Class(isTokenType(TokenWord))
	isInteger        = core.Class(isTokenType(TokenInteger))
	isArithmeticSign = core.RuneSet("+-")

	// isSequenceStop holds the runes that end a sequence.
	isSequenceStop = isWhitespace.Or(core.RuneSet("\n\"()[]{}"))
)

// isSequenceStart reports whether r starts a sequence: r starts no other
// token. A sign is checked apart, by sequenceRule.
func isSequenceStart(r rune) bool {
	if isWhitespace(r) || isWord(r) || isInteger(r) {
		return false
	}
	for _, tt := range singleRuneTypes {
		if tokenValues[tt][0] == r {
			return false
		}
	}
	return true
}

// sequenceRule matches a sequence. A sign followed by a digit is left to the
// integer rule: the rule rejects the digit and keeps no match longer than the
// sign, so the longer integer match wins.
func sequenceRule() bytescan.Rule {
	var rest bytescan.Rule
	rest = func(s bytescan.Symbol) (bytescan.Rule, bytescan.State) {
		if s.IsEOF() || isSequenceStop(s.Rune()) {
			return nil, bytescan.StateReject
		}
		return rest, bytescan.StateAccept
	}

	afterSign := func(s bytescan.Symbol) (bytescan.Rule, bytescan.State) {
		if s.IsEOF() || isInteger(s.Rune()) || isSequenceStop(s.Rune()) {
			return nil, bytescan.StateReject
		}
		return rest, bytescan.StateAccept
	}

	return func(s bytescan.Symbol) (bytescan.Rule, bytescan.State) {
		if s.IsEOF() {
			return nil, bytescan.StateReject
		}
		switch r := s.Rune(); {
		case isArithmeticSign(r):
			return afterSign, bytescan.StateAccept
		case isSequenceStart(r):
			return rest, bytescan.StateAccept
		}
		return nil, bytescan.StateReject
	}
}
