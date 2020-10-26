package cptest

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/sanity-io/litter"
)

func AssertTest(t *testing.T, got Test, want Test) {
    t.Helper()
    if !reflect.DeepEqual(got, want) {
        t.Errorf("\ngot %v\nwant %v", litter.Sdump(got), litter.Sdump(want))
    }
}

func AssertTests(t *testing.T, got []Test, want []Test) {
    t.Helper()
    if !reflect.DeepEqual(got, want) {
        t.Errorf("\ngot %v\nwant %v", litter.Sdump(got), litter.Sdump(want))
    }
}

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
            t.Errorf("got error #%d '%v', want '%v'", i + 1, errors.Unwrap(err), want[i])
        }
    }
}

func AssertVerdicts(t *testing.T, got, want map[int]Verdict) {
    t.Helper()

    if len(got) != len(want) {
        t.Fatalf("got %d verdicts, want %d", len(got), len(want))
    }

    for testId, got := range got {
        if got != want[testId] {
            t.Errorf("for test %d got verdict %v, want %v", testId, got, want[testId])
        }
    }
}

func AssertCallCount(t *testing.T, got, want int) {
    t.Helper()

    if got != want {
        t.Errorf("test was called %d times, want %d", got, want)
    }
}

func AssertConfig(t *testing.T, got, want map[string]string) {
    t.Helper()

    if !reflect.DeepEqual(got, want) {
        t.Errorf("\ngot config %v\nwant %v", litter.Sdump(got), litter.Sdump(want))
    }
}

func AssertErrorLines(t *testing.T, errs []error, lines []int) {
    t.Helper()

    if len(errs) != len(lines) {
        t.Fatalf("got %d errors, want %d", len(errs), len(lines))
    }

    for i, err := range errs {
        var e *LinedError
        if !errors.As(err, &e) {
            t.Errorf("got error without line info, want one with line %d. Error: %v", lines[i], err)
            continue
        }

        if e.Line != lines[i] {
            t.Errorf("got error #%d at line %d, want at line %d", i + 1, e.Line, lines[i])
        }
    }
}
