package cptest_test

import (
	"testing"

	"github.com/kuredoro/cptest"
)

func TestLexer(t *testing.T) {

	t.Run("empty string",
		func(t *testing.T) {
			text := ""
			want := cptest.LexSequence{}

			lexer := cptest.Lexer{}
			got := lexer.Scan(text)

			cptest.AssertLexSequence(t, got, want)
		})

	t.Run("one word",
		func(t *testing.T) {
			text := "foo"
			want := cptest.LexSequence{"foo"}

			lexer := cptest.Lexer{}
			got := lexer.Scan(text)

			cptest.AssertLexSequence(t, got, want)
		})

	t.Run("several words",
		func(t *testing.T) {
			text := " foo bar   --> "
			want := cptest.LexSequence{"foo", "bar", "-->"}

			lexer := cptest.Lexer{}
			got := lexer.Scan(text)

			cptest.AssertLexSequence(t, got, want)
		})

	t.Run("newline is treated like a word",
		func(t *testing.T) {
			text := "one\ntwo\n\n  three \n"
			want := cptest.LexSequence{"one", "\n", "two", "\n", "\n", "three", "\n"}

			lexer := cptest.Lexer{}
			got := lexer.Scan(text)

			cptest.AssertLexSequence(t, got, want)
		})
}

func TestLexerCompare(t *testing.T) {

	t.Run("several strings",
		func(t *testing.T) {
			a := cptest.LexSequence{"foo", "bar"}
			b := cptest.LexSequence{"foo", "bar"}

			lexer := cptest.Lexer{}

			got, ok := lexer.Compare(a, b)

			want := []cptest.LexDiff{
				{
					Got:   "foo",
					Want:  "foo",
					Equal: true,
				},
				{
					Got:   "bar",
					Want:  "bar",
					Equal: true,
				},
			}

			cptest.AssertDiffSuccess(t, ok)
			cptest.AssertLexDiff(t, got, want)
		})

	t.Run("totaly different strings",
		func(t *testing.T) {
			a := cptest.LexSequence{"foo", "bar"}
			b := cptest.LexSequence{"one", "x"}

			lexer := cptest.Lexer{}

			got, ok := lexer.Compare(a, b)

			want := []cptest.LexDiff{
				{
					Got:   "foo",
					Want:  "one",
					Equal: false,
				},
				{
					Got:   "bar",
					Want:  "x",
					Equal: false,
				},
			}

			cptest.AssertDiffFailure(t, ok)
			cptest.AssertLexDiff(t, got, want)
		})

	t.Run("strings: ok fail",
		func(t *testing.T) {
			a := cptest.LexSequence{"x", "bar"}
			b := cptest.LexSequence{"one", "bar"}

			lexer := cptest.Lexer{}

			got, ok := lexer.Compare(a, b)

			want := []cptest.LexDiff{
				{
					Got:   "x",
					Want:  "one",
					Equal: false,
				},
				{
					Got:   "bar",
					Want:  "bar",
					Equal: true,
				},
			}

			cptest.AssertDiffFailure(t, ok)
			cptest.AssertLexDiff(t, got, want)
		})
}
