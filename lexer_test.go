package cptest_test

import (
	"testing"

	"github.com/kuredoro/cptest"
)

func TestLexer(t *testing.T) {

	t.Run("empty string", func(t *testing.T) {
		text := ""
		var want []string

		lexer := cptest.Lexer{}
		got := lexer.Scan(text)

		cptest.AssertLexemes(t, got, want)
	})

	t.Run("one word", func(t *testing.T) {
		text := "foo"
		want := []string{"foo"}

		lexer := cptest.Lexer{}
		got := lexer.Scan(text)

		cptest.AssertLexemes(t, got, want)
	})

	t.Run("several words", func(t *testing.T) {
		text := " foo bar   --> "
		want := []string{"foo", "bar", "-->"}

		lexer := cptest.Lexer{}
		got := lexer.Scan(text)

		cptest.AssertLexemes(t, got, want)
	})

	t.Run("newline is treated like a word", func(t *testing.T) {
		text := "one\ntwo\n\n  three \n"
		want := []string{"one", "\n", "two", "\n", "\n", "three", "\n"}

		lexer := cptest.Lexer{}
		got := lexer.Scan(text)

		cptest.AssertLexemes(t, got, want)
	})
}

func TestLexerCompare(t *testing.T) {

	t.Run("several equal strings", func(t *testing.T) {
		a := []string{"foo", "bar"}
		b := []string{"foo", "bar"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(a, b)

		want := cptest.LexComparison{
			Got: []cptest.RichText{
				{"foo", make([]bool, 3)}, {"bar", make([]bool, 3)},
			},
			Want: []cptest.RichText{
				{"foo", make([]bool, 3)}, {"bar", make([]bool, 3)},
			},
		}

		cptest.AssertDiffSuccess(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})

	t.Run("totaly different strings", func(t *testing.T) {
		a := []string{"x", "bar"}
		b := []string{"one", "x"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(a, b)

		want := cptest.LexComparison{
			Got: []cptest.RichText{
				{"x", []bool{true}}, {"bar", []bool{true, true, true}},
			},
			Want: []cptest.RichText{
				{"one", []bool{true, true, true}}, {"x", []bool{true}},
			},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})

	t.Run("got more than want", func(t *testing.T) {
		a := []string{"one", "two"}
		b := []string{"one"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(a, b)

		want := cptest.LexComparison{
			Got: []cptest.RichText{
				{"one", make([]bool, 3)}, {"two", []bool{true, true, true}},
			},
			Want: []cptest.RichText{
				{"one", make([]bool, 3)},
			},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})

	t.Run("want more than got", func(t *testing.T) {
		a := []string{"one"}
		b := []string{"one", "two"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(a, b)

		want := cptest.LexComparison{
			Got: []cptest.RichText{
				{"one", make([]bool, 3)},
			},
			Want: []cptest.RichText{
				{"one", make([]bool, 3)}, {"two", []bool{true, true, true}},
			},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})

	t.Run("only unequal characters are highlighted", func(t *testing.T) {
		a := []string{"abcd",  ".b.de"}
		b := []string{"a.c.e", "abcd"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(a, b)

		want := cptest.LexComparison{
			Got: []cptest.RichText{
				{"abcd", []bool{false, true, false, true}}, {".b.de", []bool{true, false, true, false, true}},
			},
			Want: []cptest.RichText{
				{"a.c.e", []bool{false, true, false, true, true}}, {"abcd", []bool{true, false, true, false}},
			},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})
}
