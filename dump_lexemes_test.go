package cptest_test

import (
	"testing"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

func TestDumpLexemes(t *testing.T) {
    t.Run("single lexeme", func(t *testing.T) {
        xms := []cptest.RichText{
            {"foo", []bool{false, false, false}},
        }

        got := cptest.DumpLexemes(xms, aurora.RedFg)
        want := "foo"

        cptest.AssertText(t, got, want)
    })

    t.Run("multiple on one line", func(t *testing.T) {
        xms := []cptest.RichText{
            {"foo", []bool{false, false, false}},
            {"bar", []bool{true, true, true}},
        }

        got := cptest.DumpLexemes(xms, aurora.BoldFm)
        want := "foo " + aurora.Bold("bar").String()

        cptest.AssertText(t, got, want)
    })

    t.Run("with one new line", func(t *testing.T) {
        xms := []cptest.RichText{
            {"foo", []bool{false, false, false}},
            {"bar", []bool{true, true, true}},
            {"\n", []bool{false}},
        }

        got := cptest.DumpLexemes(xms, aurora.BoldFm)
        want := "foo " + aurora.Bold("bar").String() + "\n"

        cptest.AssertText(t, got, want)
    })

    t.Run("multiple lines", func(t *testing.T) {
        xms := []cptest.RichText{
            {"foo", []bool{false, false, false}},
            {"bar", []bool{true, true, true}},
            {"\n", []bool{false}},
            {"bar", []bool{false, false, false}},
            {"foo", []bool{true, true, true}},
            {"\n", []bool{false}},
        }

        got := cptest.DumpLexemes(xms, aurora.BoldFm)
        want := "foo " + aurora.Bold("bar").String() + "\n"
        want += "bar " + aurora.Bold("foo").String() + "\n"

        cptest.AssertText(t, got, want)
    })

    t.Run("colorized new line", func(t *testing.T) {
        xms := []cptest.RichText{
            {"foo", []bool{false, false, false}},
            {"\n", []bool{true}},
            {"\n", []bool{true}},
            {"bar", []bool{true, true, true}},
            {"\n", []bool{false}},
        }

        got := cptest.DumpLexemes(xms, aurora.BoldFm)

        colorizedLF := aurora.Bold(string(cptest.AltLineFeedChar)).String() + "\n"
        want := "foo" + colorizedLF + colorizedLF + aurora.Bold("bar").String() + "\n"

        cptest.AssertText(t, got, want)
    })
}
