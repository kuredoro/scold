package scold

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/maxatome/go-testdeep/td"
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

	if len(errs) != 0 {
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
		t.Errorf("got %d errors (%v), want %d (%v)", len(got), got, len(want), want)
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
func AssertVerdicts(t *testing.T, got map[int]*TestResult, want map[int]Verdict) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d verdicts, want %d", len(got), len(want))
	}

	for testID, result := range got {
		if result.Verdict != want[testID] {
			t.Errorf("for test %d got verdict %v, want %v", testID, result.Verdict, want[testID])
		}
	}
}

// AssertCallCount checks that the received and expected number of calls are
// equal.
func AssertCallCount(t *testing.T, funcName string, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("%s was called %d times, want %d", funcName, got, want)
	}
}

// AssertDefaultConfig checks that the received key-value set is empty. If it's not,
// the test is failed and the its contents are printed.
func AssertDefaultConfig(t *testing.T, got InputsConfig) {
	t.Helper()

	td.Cmp(t, got, DefaultInputsConfig)
}

// AssertTimes check whether the received and expected timestampts for the
// test cases both exist and are equal.
func AssertTimes(t *testing.T, got map[int]*TestResult, want map[int]time.Duration) {
	gotTimes := map[int]time.Duration{}
	for k, v := range got {
		gotTimes[k] = v.Time
	}

	if len(gotTimes) != len(want) {
		t.Errorf("got %d timestamps, want %d\ngot %v\nwant %v\n",
			len(gotTimes), len(want), litter.Sdump(gotTimes), litter.Sdump(want))
		return
	}

	for id, wantTime := range want {
		gotTime, exists := gotTimes[id]

		if !exists {
			t.Errorf("expected time #%d to exist, but doesn't", id)
			continue
		}

		if gotTime != wantTime {
			t.Errorf("id=%d: got time %v, want %v", id, gotTime, wantTime)
		}
	}
}

// AssertLexemes compares if the two LexSequences are equal.
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

// AssertDiffFailure chacks if lexeme comparison returned ok = false.
func AssertDiffFailure(t *testing.T, ok bool) {
	t.Helper()

	if ok {
		t.Errorf("lexer compare succeeded, but wanted to fail")
	}
}

// AssertRichText checks that the contents and the bitmasks of both RichTexts
// are equal.
func AssertRichText(t *testing.T, got, want RichText) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got '%s', want '%s'", got.Colorize(aurora.ReverseFm),
			want.Colorize(aurora.ReverseFm))
	}
}

// AssertRichTextMask checks that both bitmasks to be used in RichText
// are equal.
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

// AssertEnrichedLexSequence checks that both lexeme sequences render to
// the same string.
func AssertEnrichedLexSequence(t *testing.T, got, want []RichText) {
	t.Helper()

	gotStr := DumpLexemes(got, aurora.BoldFm)
	gotStr = strings.ReplaceAll(gotStr, "\n", "\\n")

	wantStr := DumpLexemes(want, aurora.BoldFm)
	wantStr = strings.ReplaceAll(wantStr, "\n", "\\n")

	if gotStr != wantStr {
		t.Errorf("got lexemes '%s', want '%s'", gotStr, wantStr)
	}
}

// AssertText checks two strings for equality, but escapes all newlines.
func AssertText(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		got = strings.ReplaceAll(got, "\n", "\\n")
		want = strings.ReplaceAll(want, "\n", "\\n")
		t.Errorf("got text '%s', want '%s'", got, want)
	}
}

// AssertThreadCount thread-safely checks that the pool has the specified
// number of threads created during its lifetime.
func AssertThreadCount(t *testing.T, pool *SpyThreadPool, want int) {
	t.Helper()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	got := len(pool.DirtyThreads)
	if got != want {
		t.Errorf("got %d threads used, want %d", got, want)
	}
}

// AssertResultIDInvariant checks that each key in TestingBatch.Results
// equals to the its value`s ID field.
func AssertResultIDInvariant(t *testing.T, b *TestingBatch) {
	for k, v := range b.Results {
		if k != v.ID {
			t.Errorf("invariant violation: %d != Results[%d].ID (which is %d)", k, k, v.ID)
		}
	}
}

func AssertListenerNotified(t *testing.T, listener *SpyPrinter, wantTests []Test) {
	testCount := len(wantTests)
	expectedIDs := make([]int, testCount)
	for i := range expectedIDs {
		expectedIDs[i] = i + 1
	}

	wantTestPtrs := make([]*Test, testCount)
	for i := range wantTestPtrs {
		wantTestPtrs[i] = &wantTests[i]
	}

	td.Cmp(t, listener.StartedIDs, expectedIDs, fmt.Sprintf("%d tests started", testCount))
	td.Cmp(t, listener.FinishedIDs, td.Bag(td.Flatten(expectedIDs)), fmt.Sprintf("%d tests finished", testCount))

	for i, id := range listener.FinishedIDs {
		td.Cmp(t, listener.FinishedTests[i], wantTestPtrs[id-1], fmt.Sprintf("%d ID corresponds to test #%d", id, id))
	}

	if !listener.Finished {
		t.Error("event listener has not received a suite completion event, want one")
	}
}
