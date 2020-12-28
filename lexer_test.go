package cptest_test

import (
	"testing"

	"github.com/kuredoro/cptest"
)

func TestLexerScan(t *testing.T) {

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
	t.Run("two different lexems", func(t *testing.T) {
		target := []string{"x", "xar"}
		source := []string{"one", "x"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []cptest.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
			{target[1], lexer.GenMaskForString(target[1], source[1])},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("got less than want", func(t *testing.T) {
		target := []string{"x"}
		source := []string{"x", "y"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []cptest.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
		}

		cptest.AssertDiffSuccess(t, ok)
		cptest.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("got more than want", func(t *testing.T) {
		target := []string{"x", "yz"}
		source := []string{"x"}

		lexer := cptest.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []cptest.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
			{target[1], []bool{true, true}},
		}

		cptest.AssertDiffFailure(t, ok)
		cptest.AssertEnrichedLexSequence(t, got, want)
	})
}

func TestGenMaskForString(t *testing.T) {
	lexer := &cptest.Lexer{}

	t.Run("equal strings", func(t *testing.T) {
		lexeme := "test"
		other := "test"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, false, false, false}

		cptest.AssertRichTextMask(t, got, want)
	})

	t.Run("target is shorter", func(t *testing.T) {
		lexeme := "123"
		other := "12345"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, false, false}

		cptest.AssertRichTextMask(t, got, want)
	})

	t.Run("target is longer", func(t *testing.T) {
		lexeme := "12345"
		other := "123"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, false, false, true, true}

		cptest.AssertRichTextMask(t, got, want)
	})

	t.Run("checkerboard, lengths equal", func(t *testing.T) {
		lexeme := "a.c.e"
		other := "abcde"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, true, false, true, false}

		cptest.AssertRichTextMask(t, got, want)
	})
}
