package scold_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kuredoro/scold"
)

func TestLexerScan(t *testing.T) {

	t.Run("empty string", func(t *testing.T) {
		text := ""
		var want []string

		lexer := &scold.Lexer{}
		got := lexer.Scan(text)

		scold.AssertLexemes(t, got, want)
	})

	t.Run("one word", func(t *testing.T) {
		text := "foo"
		want := []string{"foo"}

		lexer := &scold.Lexer{}
		got := lexer.Scan(text)

		scold.AssertLexemes(t, got, want)
	})

	t.Run("several words", func(t *testing.T) {
		text := " foo bar   --> "
		want := []string{"foo", "bar", "-->"}

		lexer := &scold.Lexer{}
		got := lexer.Scan(text)

		scold.AssertLexemes(t, got, want)
	})

	t.Run("newline is treated like a word", func(t *testing.T) {
		text := "one\ntwo\n\n  three \n"
		want := []string{"one", "\n", "two", "\n", "\n", "three", "\n"}

		lexer := &scold.Lexer{}
		got := lexer.Scan(text)

		scold.AssertLexemes(t, got, want)
	})
}

func TestLexerCompare(t *testing.T) {
	t.Run("two different lexems", func(t *testing.T) {
		target := []string{"x", "xar"}
		source := []string{"one", "x"}

		lexer := &scold.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []scold.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
			{target[1], lexer.GenMaskForString(target[1], source[1])},
		}

		scold.AssertDiffFailure(t, ok)
		scold.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("got less than want", func(t *testing.T) {
		target := []string{"x"}
		source := []string{"x", "y"}

		lexer := &scold.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []scold.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
		}

		scold.AssertDiffSuccess(t, ok)
		scold.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("got more than want", func(t *testing.T) {
		target := []string{"x", "yz"}
		source := []string{"x"}

		lexer := &scold.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []scold.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
			{target[1], []bool{true, true}},
		}

		scold.AssertDiffFailure(t, ok)
		scold.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("integers treated differently", func(t *testing.T) {
		target := []string{"10", "-10", "x", "10"}
		source := []string{"+10", "10", "10", "y"}

		lexer := &scold.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []scold.RichText{
			{target[0], lexer.GenMaskForInt(target[0], source[0])},
			{target[1], lexer.GenMaskForInt(target[1], source[1])},
			{target[2], lexer.GenMaskForString(target[2], source[2])},
			{target[3], lexer.GenMaskForString(target[3], source[3])},
		}

		scold.AssertDiffFailure(t, ok)
		scold.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("spurious LFs are skipped in target", func(t *testing.T) {
		target := []string{"foo", "\n", "\n", "bar"}
		source := []string{"foo", "\n", "bar"}

		lexer := &scold.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []scold.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
			{target[1], lexer.GenMaskForString(target[1], source[1])},
			{target[2], []bool{true}},
			{target[3], lexer.GenMaskForString(target[3], source[2])},
		}

		scold.AssertDiffFailure(t, ok)
		scold.AssertEnrichedLexSequence(t, got, want)
	})

	t.Run("spurious LFs are skipped in source", func(t *testing.T) {
		target := []string{"foo", "\n", "bar"}
		source := []string{"foo", "\n", "\n", "bar"}

		lexer := &scold.Lexer{}

		got, ok := lexer.Compare(target, source)

		want := []scold.RichText{
			{target[0], lexer.GenMaskForString(target[0], source[0])},
			{target[1], lexer.GenMaskForString(target[1], source[1])},
			{target[2], lexer.GenMaskForString(target[2], source[3])},
		}

		scold.AssertDiffSuccess(t, ok)
		scold.AssertEnrichedLexSequence(t, got, want)
	})
}

func TestGenMaskForString(t *testing.T) {
	lexer := &scold.Lexer{}

	t.Run("equal strings", func(t *testing.T) {
		lexeme := "test"
		other := "test"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, false, false, false}

		scold.AssertRichTextMask(t, got, want)
	})

	t.Run("target is shorter", func(t *testing.T) {
		lexeme := "123"
		other := "12345"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, false, false}

		scold.AssertRichTextMask(t, got, want)
	})

	t.Run("target is longer", func(t *testing.T) {
		lexeme := "12345"
		other := "123"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, false, false, true, true}

		scold.AssertRichTextMask(t, got, want)
	})

	t.Run("checkerboard, lengths equal", func(t *testing.T) {
		lexeme := "a.c.e"
		other := "abcde"

		got := lexer.GenMaskForString(lexeme, other)
		want := []bool{false, true, false, true, false}

		scold.AssertRichTextMask(t, got, want)
	})
}

func TestIsIntLexeme(t *testing.T) {
	cases := []struct {
		Str  string
		Want bool
	}{
		{"10", true},
		{"+10", true},
		{"-10", true},
		{"++10", false},
		{"--10", false},
		{"-10-", false},
		{"10+-", false},
		{"0", true},
		{"0xa", false},
		{strings.Repeat("1", scold.ValidIntMaxLen), true},
		{strings.Repeat("1", scold.ValidIntMaxLen+1), false},
	}

	for _, test := range cases {
		t.Run(test.Str, func(t *testing.T) {
			got := scold.IsIntLexeme(test.Str)

			if got != test.Want {
				if test.Want {
					t.Errorf("got '%s' is not INT, but it is", test.Str)
				} else {
					t.Errorf("got '%s' is INT, but it isn't", test.Str)
				}
			}
		})
	}
}

func TestGenMaskForInt(t *testing.T) {
	lexer := &scold.Lexer{}

	cases := []struct {
		Target, Source string
		Want           []bool
	}{
		{"999", "1000", []bool{true, true, true}},
		{"10", "10", []bool{false, false}},
		{"+10", "10", []bool{false, false, false}},
		{"10", "+10", []bool{false, false}},
		{"-10", "10", []bool{true, false, false}},
		{"10", "-10", []bool{false, false}},
		{"+10", "+10", []bool{false, false, false}},
		{"-10", "-10", []bool{false, false, false}},
		{"+10", "-10", []bool{true, false, false}},
		{"-10", "+10", []bool{true, false, false}},
		{"", "10", []bool{}},
		{"10", "", []bool{false, false}},
	}

	for _, test := range cases {
		title := fmt.Sprintf("%s against %s", test.Target, test.Source)
		t.Run(title, func(t *testing.T) {
			got := lexer.GenMaskForInt(test.Target, test.Source)

			scold.AssertRichTextMask(t, got, test.Want)
		})
	}
}

func TestIsFloatLexeme(t *testing.T) {
	cases := []struct {
		Str  string
		Want bool
	}{
		{"10", true},
		{"10.0", true},
		{"10.", true},
		{".10", true},
		{"-.0", true},
		{".", false},
		{".10.", false},
		{"+10.0", true},
		{"-10.0", true},
		{"-10.123456789", true},
		{"1e0", false},
		{"-10e-10", false},
	}

	for _, test := range cases {
		t.Run(test.Str, func(t *testing.T) {
			got := scold.IsFloatLexeme(test.Str)

			if got != test.Want {
				if test.Want {
					t.Errorf("got '%s' is not FLOAT, but it is", test.Str)
				} else {
					t.Errorf("got '%s' is FLOAT, but it isn't", test.Str)
				}
			}
		})
	}
}

func TestGenMaskForFloat(t *testing.T) {
	lexer := &scold.Lexer{
		Precision: 2,
	}

	cases := []struct {
		Target, Source string
		Want           []bool
	}{
		{"1.0", "1.0", []bool{false, false, false}},
		{"1.2", "1.34", []bool{false, false, true}},
		{"1.24", "1.34", []bool{false, false, true, true}},
		{"1.2455", "1.3456", []bool{false, false, true, true, false, false}},
		{"1.24", "1.3", []bool{false, false, true, true}},
		{"1.24", "2", []bool{true, false, true, true}},
		{"1.24", "2", []bool{true, false, true, true}},
		{"2.", "2", []bool{false, false}},
		{"2", "2.", []bool{false}},
		{"2.", "2.2", []bool{false, false}},
		{"2.2", "2.", []bool{false, false, true}},
		{".5", "2.5", []bool{false, false}},
		{"2.5", ".5", []bool{true, false, false}},
		{"0.5", ".5", []bool{false, false, false}},
		{"-10.5", "10.0", []bool{true, false, false, false, true}},
		{"-11.5", "10.0", []bool{true, true, true, false, true}},
	}

	for _, test := range cases {
		title := fmt.Sprintf("%s against %s", test.Target, test.Source)
		t.Run(title, func(t *testing.T) {
			got := lexer.GenMaskForFloat(test.Target, test.Source)

			scold.AssertRichTextMask(t, got, test.Want)
		})
	}
}
