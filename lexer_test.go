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

        cptest.AssertLexSequence(t, got, want)
    })

	t.Run("one word", func(t *testing.T) {
        text := "foo"
        want := []string{"foo"}

        lexer := cptest.Lexer{}
        got := lexer.Scan(text)

        cptest.AssertLexSequence(t, got, want)
    })

	t.Run("several words", func(t *testing.T) {
        text := " foo bar   --> "
        want := []string{"foo", "bar", "-->"}

        lexer := cptest.Lexer{}
        got := lexer.Scan(text)

        cptest.AssertLexSequence(t, got, want)
    })

	t.Run("newline is treated like a word", func(t *testing.T) {
        text := "one\ntwo\n\n  three \n"
        want := []string{"one", "\n", "two", "\n", "\n", "three", "\n"}

        lexer := cptest.Lexer{}
        got := lexer.Scan(text)

        cptest.AssertLexSequence(t, got, want)
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
        a := []string{"foo", "bar"}
        b := []string{"one", "x"}

        lexer := cptest.Lexer{}

        got, ok := lexer.Compare(a, b)

        want := cptest.LexComparison{
            Got: []cptest.RichText{
                {"foo", []int{0, 3}}, {"bar", []int{0, 3}},
            },
            Want: []cptest.RichText{
                {"one", []int{0, 3}}, {"x", []int{0, 1}},
            },
        }

        cptest.AssertDiffFailure(t, ok)
        cptest.AssertLexDiff(t, got, want)
    })
}
