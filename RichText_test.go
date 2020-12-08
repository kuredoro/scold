package cptest_test

import (
	"fmt"
	"testing"

	"github.com/kuredoro/cptest"
	"github.com/logrusorgru/aurora"
)

func TestRichTextColorize(t *testing.T) {

	t.Run("identity", func(t *testing.T) {
		rt := cptest.RichText{
			"test", []int{4},
		}

		got := rt.Colorize(0)
		want := "test"

		if got != want {
			t.Errorf("got rich text %q, want %q", got, want)
		}
	})

	t.Run("the partition may crop the string", func(t *testing.T) {
		rt := cptest.RichText{
			"test", []int{2},
		}

		got := rt.Colorize(0)
		want := "te"

		if got != want {
			t.Errorf("got rich text %q, want %q", got, want)
		}
	})

	t.Run("nil partition produces emtpy string", func(t *testing.T) {
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
			"abcdef", []int{3, 6},
		}

		got := rt.Colorize(aurora.BoldFm)
		want := fmt.Sprint("abc", aurora.Bold("def"))

		if got != want {
			t.Errorf("got rich text '%s', want '%s'", got, want)
		}
	})

	t.Run("checkerboard", func(t *testing.T) {
		rt := cptest.RichText{
			"gray", []int{0, 1, 2, 3, 4},
		}

		got := rt.Colorize(aurora.BoldFm)
		want := fmt.Sprint(aurora.Bold("g"), "r", aurora.Bold("a"), "y")

		if got != want {
			t.Errorf("got rich text '%s', want '%s'", got, want)
		}
	})

	t.Run("partition may be unsorted", func(t *testing.T) {
		rt := cptest.RichText{
			"abcdef", []int{6, 1, 5, 4, 2, 3},
		}

		got := rt.Colorize(aurora.BoldFm)
		want := fmt.Sprint("a", aurora.Bold("b"), "c", aurora.Bold("d"), "e", aurora.Bold("f"))

		if got != want {
			t.Errorf("got rich text '%s', want '%s'", got, want)
		}
	})
}
