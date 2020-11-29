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
