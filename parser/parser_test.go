package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xiam/symbolic/ast"
)

func TestParserBuildTree(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{
			In:  ``,
			Out: ``,
		},
		{
			In:  `[]`,
			Out: `[]`,
		},
		{
			In:  `1`,
			Out: `1`,
		},
		{
			In:  `1 3 3.4 5.6789`,
			Out: `1 3 3.4 5.6789`,
		},
		{
			In:  `[1 2 3]`,
			Out: `[1 2 3]`,
		},
		{
			In:  "[1\n\t 2\n\n3\n]",
			Out: "[1 2 3]",
		},
		{
			In:  `[1.2 2.4 3.44 5.678]`,
			Out: `[1.2 2.4 3.44 5.678]`,
		},
		{
			In:  `[1.2 2.4 3.44 5.678 [1 1.2 1.3 [2] 1.4] 4 5 [6 [7] 8] [9 10] [] 11 12 13]`,
			Out: `[1.2 2.4 3.44 5.678 [1 1.2 1.3 [2] 1.4] 4 5 [6 [7] 8] [9 10] [] 11 12 13]`,
		},
		{
			In:  `[1 [ 1 2 3 ] 3] 4 [5 6] 7 8`,
			Out: `[1 [1 2 3] 3] 4 [5 6] 7 8`,
		},
		{
			In:  `[1 [ 1 2 3 ] 3]`,
			Out: `[1 [1 2 3] 3]`,
		},
		{
			In:  `(1 2 3)`,
			Out: `(1 2 3)`,
		},
		{
			In:  `() (1) ()`,
			Out: `() (1) ()`,
		},
		{
			In:  `(1 2 [] [3[4[5]]] 6 (7))`,
			Out: `(1 2 [] [3 [4 [5]]] 6 (7))`,
		},
		{
			In:  `[(1[2])]`,
			Out: `[(1 [2])]`,
		},
		{
			In:  `([(1[2])]3)`,
			Out: `([(1 [2])] 3)`,
		},
		{
			In: "(a		b c def GHIJ 1 1.23)",
			Out: "(a b c def GHIJ 1 1.23)",
		},
		{
			In: "(a :b :cdef :GHI [jkl		:hijk])",
			Out: "(a :b :cdef :GHI [jkl :hijk])",
		},
		{
			In:  ":true\n\n\n\n:false\n:nil\na\nCBD",
			Out: ":true :false :nil a CBD",
		},
		{
			In:  ":true\n\n\n\n:false\n:nil\na\nCBD",
			Out: ":true :false :nil a CBD",
		},
		{
			In: `"ABC		DEF	[] GHI :jkl mno" :aBC def ghij "foo BAR" # extra`,
			Out: `"ABC\t\tDEF\t[] GHI :jkl mno" :aBC def ghij "foo BAR"`,
		},
		{
			In: "\"ABC	\\n	DEF	[] GHI :jkl mno\" # AABBCBCC\n:aBC #def ghij\n \"foo\" # BAR",
			Out: `"ABC\t\\n\tDEF\t[] GHI :jkl mno" :aBC "foo"`,
		},
		{
			In:  `{}`,
			Out: `{}`,
		},
		{
			In:  `{:foo 1}`,
			Out: `{:foo 1}`,
		},
		{
			In:  `[{:foo 1}]`,
			Out: `[{:foo 1}]`,
		},
		{
			In:  "set {:foo 1\n:bar 2} {:baz {{{[(1)]}}}}",
			Out: `set {:foo 1 :bar 2} {:baz {{{[(1)]}}}}`,
		},
		{
			In:  "(fn [a b c]\n\n\t [\n(print a\n\n b c)])",
			Out: `(fn [a b c] [(print a b c)])`,
		},
		{
			In:  `{:a 1.11 :b "STRING VALUE"} ()`,
			Out: `{:a 1.11 :b "STRING VALUE"} ()`,
		},
		{
			In:  `(fn word [] [(print "foo")]) (word "a" "b" "c")`,
			Out: `(fn word [] [(print "foo")]) (word "a" "b" "c")`,
		},
		{
			In:  `(print "hello world" "beautiful world!") (echo :brave :new :world)`,
			Out: `(print "hello world" "beautiful world!") (echo :brave :new :world)`,
		},
		{
			In:  `(+ 1 2 3 4)`,
			Out: `(+ 1 2 3 4)`,
		},
		{
			In:  `{+ 1 2 3 4}`,
			Out: `{+ 1 2 3 4}`,
		},
		{
			In:  `[+ 1 2 3 4]`,
			Out: `[+ 1 2 3 4]`,
		},
		{
			In:  `[+ -1 55 +6.3 +2 -3.23 4.01]`,
			Out: `[+ -1 55 6.3 2 -3.23 4.01]`,
		},
	}

	for i := range testCases {
		root, err := Parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		var dump strings.Builder
		assert.NoError(t, ast.Dump(&dump, root))
		t.Log("\n" + dump.String())

		s := ast.Encode(root)

		assert.Equal(t, testCases[i].Out, string(s))
	}
}

func TestAutoCloseOnEOF(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{
			In:  `(`,
			Out: `()`,
		},
		{
			In:  `(1`,
			Out: `(1)`,
		},
		{
			In:  `(((`,
			Out: `((()))`,
		},
		{
			In:  `(((1 1 1`,
			Out: `(((1 1 1)))`,
		},
		{
			In:  `({`,
			Out: `({})`,
		},
		{
			In:  `(()`,
			Out: `(())`,
		},
		{
			In:  `((({{`,
			Out: `((({{}})))`,
		},
		{
			In:  `({}{`,
			Out: `({} {})`,
		},
		{
			In:  `({/{`,
			Out: `({/ {}})`,
		},
		{
			In: `(1 2 3 4
			(5 6 7 8
			(4 6
		`,
			Out: `(1 2 3 4 (5 6 7 8 (4 6)))`,
		},
		{
			In:  `(1 2 3 4 (5 6 7 8 (4 6) 7`,
			Out: `(1 2 3 4 (5 6 7 8 (4 6) 7))`,
		},
		{
			In:  `(*`,
			Out: `(*)`,
		},
		{
			In: `(1 2 3 4
			# hello world
			(5 6 7 8 # a
			# b
			(4 6
			# # c
		`,
			Out: `(1 2 3 4 (5 6 7 8 (4 6)))`,
		},
		{
			In:  `(1 # a comment`,
			Out: `(1)`,
		},
	}

	for i := range testCases {
		{
			p := NewParser(strings.NewReader(testCases[i].In))
			err := p.Parse()
			assert.Error(t, err)
		}

		{
			p := NewParser(strings.NewReader(testCases[i].In))
			p.SetOptions(ParserOptions{
				AutoCloseOnEOF: true,
			})

			err := p.Parse()
			assert.NotNil(t, p.root)
			assert.NoError(t, err)

			s := ast.Encode(p.root)
			assert.Equal(t, testCases[i].Out, string(s))
		}
	}
}

func TestParserErrors(t *testing.T) {
	testCases := []struct {
		In  string
		Err string
	}{
		{
			In: `(}`,
		},
		{
			In: `[}`,
		},
		{
			In: `[)`,
		},
		{
			In: `1 )}`,
		},
		{
			In: `1 ](}`,
		},
		{
			In: `+}`,
		},
		{
			In: `{)}`,
		},
		{
			In: `(1 2 3 4
			(5 6 7 8
			(4 6})
			)`,
		},
	}

	for i := range testCases {
		root, err := Parse([]byte(testCases[i].In))
		assert.Nil(t, root)
		assert.Error(t, err)
		t.Log(err)
	}
}

// TestOddMap pins decision 1 of #2: maps are flat, and an odd number of items
// is legal.
func TestOddMap(t *testing.T) {
	root, err := Parse([]byte(`{+ 1 2 3 4}`))
	require.NoError(t, err)
	require.Len(t, root.Items, 1)

	m, ok := root.Items[0].(*ast.Map)
	require.True(t, ok)
	assert.Len(t, m.Items, 5)
	assert.True(t, m.Odd())

	var keys []string
	for k := range m.Pairs() {
		keys = append(keys, string(ast.Encode(k)))
	}
	assert.Equal(t, []string{"+", "2"}, keys)
}
