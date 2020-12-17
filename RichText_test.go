package cptest_test

import (
	"fmt"
	"testing"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

func TestRichTextColorize(t *testing.T) {

	t.Run("identity mask", func(t *testing.T) {
		rt := cptest.RichText{
			"test", make([]bool, 4),
		}

		got := rt.Colorize(0)
		want := "test"

		if got != want {
			t.Errorf("got rich text %q, want %q", got, want)
		}
	})

	t.Run("the mask may crop the string", func(t *testing.T) {
		rt := cptest.RichText{
			"test", make([]bool, 2),
		}

		got := rt.Colorize(0)
		want := "te"

		if got != want {
			t.Errorf("got rich text %q, want %q", got, want)
		}
	})

	t.Run("nil mask produces emtpy string", func(t *testing.T) {
		rt := cptest.RichText{
			"test", nil,
		}

		got := rt.Colorize(0)
		want := ""

		if got != want {
			t.Errorf("got rich text %q, want %q", got, want)
		}
	})

	t.Run("second half is highlighted", func(t *testing.T) {
		rt := cptest.RichText{
			"abcdef", []bool{false, false, false, true, true, true},
		}

		got := rt.Colorize(aurora.BoldFm)
		want := fmt.Sprint("abc", aurora.Bold("def"))

		if got != want {
			t.Errorf("got rich text '%s', want '%s'", got, want)
		}
	})

	t.Run("checkerboard", func(t *testing.T) {
		rt := cptest.RichText{
			"gray", []bool{true, false, true, false},
		}

		got := rt.Colorize(aurora.BoldFm)
		want := fmt.Sprint(aurora.Bold("g"), "r", aurora.Bold("a"), "y")

		if got != want {
			t.Errorf("got rich text '%s', want '%s'", got, want)
		}
	})
}

func TestRichTextColorful(t *testing.T) {
    t.Run("no colors", func(t *testing.T) {
        rt := cptest.RichText{
            "abc", make([]bool, 3),
        }

        if rt.Colorful() {
            t.Errorf("got rich text '%s' to be colorful, but it's not", rt.Colorize(aurora.BoldFm))
        }
    })

    t.Run("with colors", func(t *testing.T) {
        rt := cptest.RichText{
            "yay", []bool{false, true, true},
        }

        if !rt.Colorful() {
            t.Errorf("got rich text '%s' to be not colorful, but it is", rt.Colorize(aurora.BoldFm))
        }
    })
}
