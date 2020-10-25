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
