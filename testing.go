package cptest

import (
	"reflect"
	"testing"
    "errors"

	"github.com/sanity-io/litter"
)

func AssertTests(t *testing.T, got Inputs, want []Test) {
    t.Helper()
    if !reflect.DeepEqual(got.Tests, want) {
        t.Errorf("\ngot %v\nwant %v", litter.Sdump(got.Tests), litter.Sdump(want))
    }
}

func AssertNoErrors(t *testing.T, errs []error) {
    t.Helper()

    if errs != nil && len(errs) != 0 {
        t.Errorf("expected no errors, but got %d:%v", len(errs), litter.Sdump(errs))
    }
}

func AssertErrors(t *testing.T, got []error, want []error) {
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
