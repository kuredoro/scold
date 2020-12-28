package cptest

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/sanity-io/litter"
)

// AssertTest compare the inputs and outputs with respective expected ones
// for equivalence.
func AssertTest(t *testing.T, got Test, want Test) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %v\nwant %v", litter.Sdump(got), litter.Sdump(want))
	}
}

// AssertTests will compare received array of tests with the expected one.
func AssertTests(t *testing.T, got []Test, want []Test) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %v\nwant %v", litter.Sdump(got), litter.Sdump(want))
	}
}

// AssertNoErrors will check if the array of errors is empty. If it's not
// empty, the test will be failed and the errors will be reported.
func AssertNoErrors(t *testing.T, errs []error) {
	t.Helper()

	if errs != nil && len(errs) != 0 {
		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("expected no errors, but got %d:\n", len(errs)))

		for _, err := range errs {
			msg.WriteString(fmt.Sprintf("\t%v\n", err))
		}

		t.Error(msg.String())
	}
}

// AssertErrors compared received array of errors with the expected one.
func AssertErrors(t *testing.T, got, want []error) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("got %d errors, want %d", len(got), len(want))
	}

	for i, err := range got {
		if i == len(want) {
			break
		}

		if !errors.Is(err, want[i]) {
			t.Errorf("got error #%d '%v', want '%v'", i+1, errors.Unwrap(err), want[i])
		}
	}
}

// AssertVerdicts checks that received and expected verdict maps contain the
// same keys, and then checks that the values for these keys equal.
func AssertVerdicts(t *testing.T, got, want map[int]Verdict) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d verdicts, want %d", len(got), len(want))
	}

	for testID, got := range got {
		if got != want[testID] {
			t.Errorf("for test %d got verdict %v, want %v", testID, got, want[testID])
		}
	}
}

// AssertCallCount checks that the received and expected number of calls are
// equal.
func AssertCallCount(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("test was called %d times, want %d", got, want)
	}
}

// AssertConfig checks whether received and expected config key-value sets
// are equal.
func AssertConfig(t *testing.T, got, want map[string]string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot config %v\nwant %v", litter.Sdump(got), litter.Sdump(want))
	}
}

// AssertErrorLines checks that each error in the received array of errors
// is wrapping a LinedError error. At the same time, it checks that the line
// numbers are equal to the expected ones.
func AssertErrorLines(t *testing.T, errs []error, lines []int) {
	t.Helper()

	if len(errs) < len(lines) {
		t.Fatalf("got %d errors, want %d", len(errs), len(lines))
	}

	for i := range lines {
        err := errs[i]
		var e *LinedError
		if !errors.As(err, &e) {
			t.Errorf("got error without line info, want one with line %d. Error: %v", lines[i], err)
			continue
		}

		if e.Line != lines[i] {
			t.Errorf("got error #%d at line %d, want at line %d", i+1, e.Line, lines[i])
		}
	}
}

// AssertNoConfig checks that the received key-value set is empty. If it's not,
// the test is failed and the its contents are printed.
func AssertNoConfig(t *testing.T, got map[string]string) {
	t.Helper()

	if got == nil {
		t.Error("expected empty config, but config isn't even initialized")
	}

	if len(got) != 0 {
		t.Errorf("expected emtpy config, but got %v", litter.Sdump(got))
	}
}

// AssertTimes check whether the received and expected timestampts for the
// test cases both exist and are equal.
func AssertTimes(t *testing.T, got, want map[int]time.Duration) {
	if len(got) != len(want) {
		t.Errorf("got %d timestamps, want %d\ngot %v\nwant %v\n",
			len(got), len(want), litter.Sdump(got), litter.Sdump(want))
		return
	}

	for id, wantTime := range want {
		gotTime, exists := got[id]

		if !exists {
			t.Errorf("expected time #%d to exist, but doesn't", id)
			continue
		}

		if gotTime != wantTime {
			t.Errorf("id=%d: got time %v, want %v", id, gotTime, wantTime)
		}
	}
}

// AssertLexSequence compares if the two LexSequences are equal.
func AssertLexemes(t *testing.T, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

// AssertDiffSuccess chacks if lexeme comparison returned ok = true.
func AssertDiffSuccess(t *testing.T, ok bool) {
	t.Helper()

	if !ok {
		t.Errorf("lex diff failed, but wanted to pass")
	}
}

// AssertDiffSuccess chacks if lexeme comparison returned ok = true.
func AssertDiffFailure(t *testing.T, ok bool) {
	t.Helper()

	if ok {
		t.Errorf("lexer compare succeeded, but wanted to fail")
	}
}

func AssertRichText(t *testing.T, got, want RichText) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got '%s', want '%s'", got.Colorize(aurora.ReverseFm),
			want.Colorize(aurora.ReverseFm))
	}
}

func AssertRichTextMask(t *testing.T, got, want []bool) {
	t.Helper()

	gotRt := RichText{
		Str:  strings.Repeat("#", len(got)),
		Mask: got,
	}

	wantRt := RichText{
		Str:  strings.Repeat("#", len(want)),
		Mask: want,
	}

	AssertRichText(t, gotRt, wantRt)
}

func AssertEnrichedLexSequence(t *testing.T, got, want []RichText) {
	t.Helper()

	commonLen := len(got)
	if len(want) < commonLen {
		commonLen = len(want)
	}

	for i := 0; i < commonLen; i++ {
		AssertRichText(t, got[i], want[i])
	}

	for i := commonLen; i < len(got); i++ {
		AssertRichText(t, got[i], RichText{"", []bool{}})
	}

	for i := commonLen; i < len(want); i++ {
		AssertRichText(t, RichText{"", []bool{}}, want[i])
	}
}
