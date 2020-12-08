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
				{"foo", []int{3}}, {"bar", []int{3}},
			},
			Want: []cptest.RichText{
				{"foo", []int{3}}, {"bar", []int{3}},
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
				{"x", []int{0, 1}}, {"bar", []int{0, 3}},
			},
			Want: []cptest.RichText{
				{"one", []int{0, 3}}, {"x", []int{0, 1}},
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
				{"one", []int{3}}, {"two", []int{0, 3}},
			},
			Want: []cptest.RichText{
				{"one", []int{3}},
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
				{"one", []int{3}},
			},
			Want: []cptest.RichText{
				{"one", []int{3}}, {"two", []int{0, 3}},
			},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})

	t.Run("only unequal characters are highlighted", func(t *testing.T) {
		a := []string{"abcd", ".b.d."}
		b := []string{"a.c.e", "abcd"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(a, b)

		want := cptest.LexComparison{
			Got: []cptest.RichText{
				{"abcd", []int{1, 2, 3, 4}}, {".b.de", []int{0, 1, 2, 3, 4, 5}},
			},
			Want: []cptest.RichText{
				{"a.c.e", []int{1, 2, 3, 5}}, {"abcd", []int{0, 1, 2, 3, 4}},
			},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertLexDiff(t, got, want)
	})
}
